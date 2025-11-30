#!/usr/bin/env node
import { config } from "dotenv";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

// Load .env from the hooks directory (not CWD)
const __dirname = dirname(fileURLToPath(import.meta.url));
config({ path: join(__dirname, "..", ".env"), override: true });

import { createInterface } from "node:readline";
import {
  type ClaudeCodeEvent,
  isValidEvent,
  analyzeToolResult,
  getSubagentInfo,
  isSubagentTool,
  getGitContext,
} from "./utils.js";
import {
  initTracing,
  shutdownTracing,
  forceFlush,
  createConfigFromEnv,
  createSessionObservation,
  createToolObservation,
  finalizeToolObservation,
  finalizeSessionObservation,
  recordEvent,
  createParentContext,
  // Persistence functions for cross-process span linking
  registerActiveSpan,
  popActiveSpan,
  getSessionInfo,
  initSession,
  deleteSpanState,
  cleanupOldStates,
  // Metrics tracking functions
  updateSessionMetrics,
  getSessionMetrics,
  calculateAggregateMetrics,
  type LangfuseAgent,
  type LangfuseTool,
  type ToolContext,
  type ToolResult,
} from "./tracing/index.js";

/**
 * Claude Code Langfuse Hook
 * Uses @langfuse/tracing native API with proper observation types
 */

// Configuration
const DEBUG = process.env.LANGFUSE_LOG_LEVEL === "DEBUG";

const log = (level: string, msg: string) =>
  console.error(`[Langfuse] ${level === "ERROR" ? "ERROR: " : ""}${msg}`);

// Track active spans by tool_use_id
interface ActiveObservation {
  observation: LangfuseTool | LangfuseAgent;
  startTime: number;
  ctx: ToolContext;
}
const activeObservations = new Map<string, ActiveObservation>();

// Session observations by session_id
const sessionObservations = new Map<string, LangfuseAgent>();

// Process a single event
function processEvent(event: ClaudeCodeEvent) {
  // Try to get existing session info from persistence first (cross-process scenario)
  let persistedSession = getSessionInfo(event.session_id);

  // Get or create session observation - only if no persisted session exists
  // This prevents creating duplicate traces for the same session across processes
  let sessionObs = sessionObservations.get(event.session_id);
  if (!sessionObs && !persistedSession) {
    // Capture git context for the session
    const gitContext = getGitContext(event.cwd);

    sessionObs = createSessionObservation({
      sessionId: event.session_id,
      userId: event.user_id,
      cwd: event.cwd,
      permissionMode: event.permission_mode,
      git: gitContext,
    });
    sessionObservations.set(event.session_id, sessionObs);

    // Build tags including git repo if available
    const tags = ["claude-code"];
    if (gitContext.isGitRepo && gitContext.repoName) {
      tags.push(`repo:${gitContext.repoName}`);
    }
    if (gitContext.branch) {
      tags.push(`branch:${gitContext.branch}`);
    }

    // Update trace with session info
    sessionObs.updateTrace({
      name: "claude-code-session",
      sessionId: event.session_id,
      userId: event.user_id || "unknown",
      tags,
    });

    // Persist session info for cross-process linking
    initSession(event.session_id, sessionObs.traceId, sessionObs.id);

    DEBUG && log("DEBUG", `Created session: ${event.session_id}${gitContext.isGitRepo ? ` (${gitContext.repoName}@${gitContext.branch})` : ""}`);
  }

  // Get parent context - prefer persisted session if available (cross-process scenario)
  // This ensures PostToolUse can link to the correct trace even if session was created in a different process
  const parentSpanContext = persistedSession
    ? createParentContext(persistedSession.traceId, persistedSession.sessionSpanId)
    : sessionObs
      ? createParentContext(sessionObs.traceId, sessionObs.id)
      : null;

  // If we have no parent context, something is wrong - skip processing
  if (!parentSpanContext) {
    DEBUG && log("DEBUG", `No parent context available for session: ${event.session_id}`);
    return;
  }

  switch (event.hook_event_name) {
    case "PreToolUse": {
      if (!event.tool_name || !event.tool_use_id) break;

      // Check if this is a subagent (Task tool)
      const isSubagent = isSubagentTool(event.tool_name);
      const subagentInfo = isSubagent ? getSubagentInfo(event.tool_input) : null;

      const ctx: ToolContext = {
        toolName: event.tool_name,
        toolUseId: event.tool_use_id,
        toolInput: event.tool_input,
        isSubagent,
        subagentType: subagentInfo?.type,
        subagentDescription: subagentInfo?.description,
        subagentModel: subagentInfo?.model,
        model: event.model,
      };

      const observation = createToolObservation(ctx, { parentSpanContext });

      activeObservations.set(event.tool_use_id, {
        observation,
        startTime: Date.now(),
        ctx,
      });

      // Persist span info for cross-process linking
      // This allows PostToolUse in a different process to find this span
      registerActiveSpan(
        event.session_id,
        event.tool_use_id,
        observation.id,
        parentSpanContext.traceId,
        parentSpanContext.spanId
      );

      DEBUG && log("DEBUG", `PreToolUse: ${event.tool_name} (${event.tool_use_id})`);
      break;
    }

    case "PostToolUse": {
      if (!event.tool_name) break;

      const analysis = analyzeToolResult(event.tool_response);
      const isSubagent = isSubagentTool(event.tool_name);
      const subagentInfo = isSubagent ? getSubagentInfo(event.tool_input) : null;

      // Track duration for metrics (will be set by whichever branch executes)
      let toolDurationMs: number | undefined;

      // First try in-memory cache (same process scenario)
      const active = event.tool_use_id ? activeObservations.get(event.tool_use_id) : undefined;

      if (active) {
        // Same process - use in-memory observation
        const durationMs = Date.now() - active.startTime;
        toolDurationMs = durationMs;
        const result: ToolResult = {
          success: analysis.success,
          error: analysis.error ?? undefined,
          errorType: analysis.errorType ?? undefined,
          exitCode: analysis.exitCode ?? undefined,
          output: event.tool_response,
          durationMs,
        };

        finalizeToolObservation(active.observation, result, active.ctx);
        activeObservations.delete(event.tool_use_id!);

        const durationStr = ` (${durationMs}ms)`;
        log(
          "INFO",
          `${event.tool_name}${subagentInfo ? ` (${subagentInfo.type})` : ""}${durationStr}: ${
            analysis.success ? "OK" : "ERROR"
          }`
        );
      } else if (event.tool_use_id) {
        // Different process - try to retrieve persisted span info
        const persistedSpan = popActiveSpan(event.session_id, event.tool_use_id);

        if (persistedSpan) {
          // Cross-process linking: create observation with persisted start time
          const durationMs = Date.now() - persistedSpan.startTime;
          toolDurationMs = durationMs;

          const ctx: ToolContext = {
            toolName: event.tool_name,
            toolUseId: event.tool_use_id,
            toolInput: event.tool_input,
            isSubagent,
            subagentType: subagentInfo?.type,
            subagentDescription: subagentInfo?.description,
            subagentModel: subagentInfo?.model,
            model: event.model,
          };

          // Create observation with the parent context from persistence
          const persistedParentContext = createParentContext(
            persistedSpan.traceId,
            persistedSpan.sessionSpanId
          );

          // Create new observation with the preserved start time
          const observation = createToolObservation(ctx, {
            parentSpanContext: persistedParentContext,
            startTime: new Date(persistedSpan.startTime),
          });

          const result: ToolResult = {
            success: analysis.success,
            error: analysis.error ?? undefined,
            errorType: analysis.errorType ?? undefined,
            exitCode: analysis.exitCode ?? undefined,
            output: event.tool_response,
            durationMs,
          };

          finalizeToolObservation(observation, result, ctx);

          const durationStr = ` (${durationMs}ms)`;
          log(
            "INFO",
            `${event.tool_name}${subagentInfo ? ` (${subagentInfo.type})` : ""}${durationStr}: ${
              analysis.success ? "OK" : "ERROR"
            } [cross-process]`
          );
        } else {
          // No persisted span found - create standalone observation
          const ctx: ToolContext = {
            toolName: event.tool_name,
            toolUseId: event.tool_use_id,
            toolInput: event.tool_input,
            isSubagent,
            subagentType: subagentInfo?.type,
            subagentDescription: subagentInfo?.description,
            subagentModel: subagentInfo?.model,
            model: event.model,
          };

          const observation = createToolObservation(ctx, { parentSpanContext });
          const result: ToolResult = {
            success: analysis.success,
            error: analysis.error ?? undefined,
            errorType: analysis.errorType ?? undefined,
            exitCode: analysis.exitCode ?? undefined,
            output: event.tool_response,
          };

          finalizeToolObservation(observation, result, ctx);

          log(
            "INFO",
            `${event.tool_name}${subagentInfo ? ` (${subagentInfo.type})` : ""}: ${
              analysis.success ? "OK" : "ERROR"
            }`
          );
        }
      } else {
        // No tool_use_id - create standalone observation
        const ctx: ToolContext = {
          toolName: event.tool_name,
          toolUseId: "unknown",
          toolInput: event.tool_input,
          isSubagent,
          subagentType: subagentInfo?.type,
          subagentDescription: subagentInfo?.description,
          subagentModel: subagentInfo?.model,
          model: event.model,
        };

        const observation = createToolObservation(ctx, { parentSpanContext });
        const result: ToolResult = {
          success: analysis.success,
          error: analysis.error ?? undefined,
          errorType: analysis.errorType ?? undefined,
          exitCode: analysis.exitCode ?? undefined,
          output: event.tool_response,
        };

        finalizeToolObservation(observation, result, ctx);

        log(
          "INFO",
          `${event.tool_name}${subagentInfo ? ` (${subagentInfo.type})` : ""}: ${
            analysis.success ? "OK" : "ERROR"
          }`
        );
      }

      // Update session metrics after tool completion
      updateSessionMetrics(
        event.session_id,
        event.tool_name,
        isSubagent,
        analysis.success,
        analysis.errorType ?? undefined,
        toolDurationMs,
        event.tokens,
        event.model
      );

      break;
    }

    case "UserPromptSubmit": {
      recordEvent(
        "user_prompt",
        {
          permission_mode: event.permission_mode,
          timestamp: event.timestamp || new Date().toISOString(),
        },
        { parentSpanContext }
      );
      DEBUG && log("DEBUG", "UserPromptSubmit");
      break;
    }

    case "SubagentStop": {
      recordEvent(
        "subagent_completed",
        {
          stop_hook_active: event.stop_hook_active ?? false,
          timestamp: event.timestamp || new Date().toISOString(),
        },
        { parentSpanContext }
      );
      log("INFO", "Subagent completed");
      break;
    }

    case "Stop": {
      // End any orphaned observations
      if (activeObservations.size > 0) {
        DEBUG && log("DEBUG", `Cleaning up ${activeObservations.size} incomplete observations`);
        for (const [, { observation, ctx }] of activeObservations) {
          const result: ToolResult = {
            success: false,
            error: "Session ended before completion",
            errorType: "incomplete",
          };
          finalizeToolObservation(observation, result, ctx);
        }
        activeObservations.clear();
      }

      // Retrieve session metrics BEFORE deleting state
      const sessionMetrics = getSessionMetrics(event.session_id);
      const aggregateMetrics = sessionMetrics
        ? calculateAggregateMetrics(sessionMetrics)
        : undefined;

      // End session observation with metrics
      if (sessionObs) {
        finalizeSessionObservation(sessionObs, {
          ended: true,
          timestamp: event.timestamp || new Date().toISOString(),
          metrics: sessionMetrics ?? undefined,
          aggregateMetrics,
        });
        sessionObservations.delete(event.session_id);
      }

      // Log metrics summary
      if (sessionMetrics) {
        const { toolCount, subagentCount, errorCount } = sessionMetrics;
        log(
          "INFO",
          `Session ended - tools: ${toolCount}, subagents: ${subagentCount}, errors: ${errorCount}`
        );
      } else {
        log("INFO", "Session ended");
      }

      // Clean up persisted state for this session
      deleteSpanState(event.session_id);

      // Periodically clean up old state files (stale sessions)
      cleanupOldStates();

      break;
    }
  }
}

// Main entry point
async function main() {
  const tracingConfig = createConfigFromEnv();
  const initialized = initTracing(tracingConfig);

  if (!initialized) {
    process.exit(1);
  }

  const rl = createInterface({ input: process.stdin, terminal: false });

  rl.on("line", (line) => {
    try {
      const data = JSON.parse(line);
      if (isValidEvent(data)) {
        processEvent(data);
      } else {
        DEBUG && log("DEBUG", "Invalid event structure");
      }
    } catch (e) {
      DEBUG && log("DEBUG", `Parse error: ${e}`);
    }
  });

  const shutdown = async () => {
    // End any active sessions
    for (const [sessionId, sessionObs] of sessionObservations) {
      finalizeSessionObservation(sessionObs, { ended: true });
      sessionObservations.delete(sessionId);
    }

    try {
      // Explicitly flush spans before shutdown to ensure export completes
      await forceFlush();
      await shutdownTracing();
    } catch {
      /* ignore */
    }
    process.exit(0);
  };

  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
  rl.on("close", shutdown);
}

main().catch((e) => {
  log("ERROR", `Fatal: ${e}`);
  process.exit(1);
});
