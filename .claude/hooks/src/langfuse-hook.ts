#!/usr/bin/env node
import { config } from "dotenv";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { appendFileSync } from "node:fs";

// Load .env from the hooks directory (not CWD)
const __dirname = dirname(fileURLToPath(import.meta.url));
config({ path: join(__dirname, "..", ".env"), override: true });

// File-based debug logging (always enabled to diagnose hook invocation)
const DEBUG_LOG_FILE = "/tmp/langfuse-hook-debug.log";
const debugLog = (msg: string): void => {
  const timestamp = new Date().toISOString();
  try {
    appendFileSync(DEBUG_LOG_FILE, `[${timestamp}] ${msg}\n`);
  } catch {
    // Ignore write errors
  }
};

// Log hook startup immediately
debugLog(`Hook started - PID: ${process.pid}, CWD: ${process.cwd()}, __dirname: ${__dirname}`);

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
  createSessionObservationWithParent,
  createToolObservation,
  createToolObservationWithContext,
  createTraceparent,
  finalizeToolObservation,
  finalizeSessionObservation,
  recordEvent,
  recordEventWithContext,
  // Persistence functions for cross-process span linking
  registerActiveSpan,
  popActiveSpan,
  getSessionInfo,
  initSession,
  deleteSpanState,
  cleanupOldStates,
  // Pending parent context for subagent linking
  storePendingParentContext,
  findPendingParentContext,
  removePendingParentContext,
  cleanupPendingParentContexts,
  // Metrics tracking functions
  updateSessionMetrics,
  getSessionMetrics,
  calculateAggregateMetrics,
  // Types
  type SessionObservation,
  type ToolObservation,
  type ToolContext,
  type ToolResult,
  type PendingParentContext,
} from "./tracing/index.js";

/**
 * Claude Code Langfuse Hook (Native SDK)
 * - Uses native Langfuse SDK for clean output without OTel metadata
 * - Restores proper parent/child relationships for tools and subagents
 * - Persists full context for cross-process recovery
 * - Centralizes parent resolution
 */

// Configuration
const DEBUG = process.env.LANGFUSE_LOG_LEVEL === "DEBUG";

const log = (level: string, msg: string): void =>
  console.error(`[Langfuse] ${level === "ERROR" ? "ERROR: " : ""}${msg}`);

// Track active spans by tool_use_id
interface ActiveObservation {
  observation: ToolObservation;
  startTime: number;
  ctx: ToolContext;
}
const activeObservations = new Map<string, ActiveObservation>();

// Session observations by session_id
const sessionObservations = new Map<string, SessionObservation>();



// Process a single event
function processEvent(event: ClaudeCodeEvent): void {
  // Try to get existing session info from persistence first (cross-process scenario)
  let persistedSession = getSessionInfo(event.session_id);

  // Get or create session observation - only if no persisted session exists
  let sessionObs = sessionObservations.get(event.session_id);

  // If we have a persisted session without traceparent, add one
  if (persistedSession && !persistedSession.traceparent) {
    const newTraceparent = createTraceparent(persistedSession.traceId, persistedSession.sessionSpanId);
    initSession(event.session_id, persistedSession.traceId, persistedSession.sessionSpanId, newTraceparent);
    // Refresh persisted session
    persistedSession = getSessionInfo(event.session_id);
    DEBUG && log("DEBUG", `Added traceparent to existing session: ${newTraceparent}`);
  }

  if (!sessionObs && !persistedSession) {
    const gitContext = getGitContext(event.cwd);

    // Check if there's a pending parent context (this session might be a subagent)
    const pendingParent = findPendingParentContext();

    if (pendingParent) {
      // This is a subagent session - link it to the parent trace
      sessionObs = createSessionObservationWithParent(
        {
          sessionId: event.session_id,
          userId: event.user_id,
          cwd: event.cwd,
          permissionMode: event.permission_mode,
          git: gitContext,
        },
        pendingParent.traceparent,
        {
          parentSessionId: pendingParent.parentSessionId,
          parentObservationId: pendingParent.observationId,
          subagentType: pendingParent.subagentType,
        }
      );
      sessionObservations.set(event.session_id, sessionObs);

      const tags = ["claude-code", "subagent"];
      if (gitContext.isGitRepo && gitContext.repoName)
        tags.push(`repo:${gitContext.repoName}`);
      if (gitContext.branch) tags.push(`branch:${gitContext.branch}`);
      if (pendingParent.subagentType)
        tags.push(`subagent:${pendingParent.subagentType}`);

      sessionObs.updateTrace({
        name: "claude-code-subagent-session",
        sessionId: event.session_id,
        userId: event.user_id || "unknown",
        tags,
      });

      // Create traceparent for cross-process context propagation
      const traceparent = createTraceparent(sessionObs.traceId, sessionObs.id);

      // persist session with traceparent for cross-process linking
      initSession(event.session_id, sessionObs.traceId, sessionObs.id, traceparent);

      log(
        "INFO",
        `Created subagent session: ${event.session_id} (linked to parent: ${pendingParent.parentSessionId})`
      );
    } else {
      // This is a top-level session
      sessionObs = createSessionObservation({
        sessionId: event.session_id,
        userId: event.user_id,
        cwd: event.cwd,
        permissionMode: event.permission_mode,
        git: gitContext,
      });
      sessionObservations.set(event.session_id, sessionObs);

      const tags = ["claude-code"];
      if (gitContext.isGitRepo && gitContext.repoName)
        tags.push(`repo:${gitContext.repoName}`);
      if (gitContext.branch) tags.push(`branch:${gitContext.branch}`);

      sessionObs.updateTrace({
        name: "claude-code-session",
        sessionId: event.session_id,
        userId: event.user_id || "unknown",
        tags,
      });

      // Create traceparent for cross-process context propagation
      const traceparent = createTraceparent(sessionObs.traceId, sessionObs.id);

      // persist session with traceparent for cross-process linking
      initSession(event.session_id, sessionObs.traceId, sessionObs.id, traceparent);

      DEBUG &&
        log(
          "DEBUG",
          `Created session: ${event.session_id}${
            gitContext.isGitRepo
              ? ` (${gitContext.repoName}@${gitContext.branch})`
              : ""
          }`
        );
    }
  }


  switch (event.hook_event_name) {
    case "PreToolUse": {
      if (!event.tool_name || !event.tool_use_id) break;

      const isSubagent = isSubagentTool(event.tool_name);
      const subagentInfo = isSubagent
        ? getSubagentInfo(event.tool_input)
        : null;

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

      // Resolve parent: check if this tool has a parent_tool_use_id (nested under another tool/agent)
      let actualParent: SessionObservation | ToolObservation | undefined = sessionObs;
      if (event.parent_tool_use_id) {
        const parentActive = activeObservations.get(event.parent_tool_use_id);
        if (parentActive) {
          actualParent = parentActive.observation;
        }
      }

      // Create observation with proper hierarchy using v4 SDK asType
      // If no in-memory parent but we have a persisted session, use traceparent for cross-process linking
      let observation: ToolObservation;
      if (!actualParent && persistedSession?.traceparent) {
        observation = createToolObservationWithContext(ctx, persistedSession.traceparent, event.session_id);
        DEBUG && log("DEBUG", `PreToolUse cross-process with traceparent: ${persistedSession.traceparent}`);
      } else {
        observation = createToolObservation(ctx, undefined, actualParent);
      }

      activeObservations.set(event.tool_use_id, {
        observation,
        startTime: Date.now(),
        ctx,
      });

      // Create traceparent for this span (for nested tool linking)
      const spanTraceparent = createTraceparent(observation.traceId, observation.id);

      // Persist context for cross-process retrieval
      registerActiveSpan(event.session_id, event.tool_use_id, {
        spanId: observation.id,
        traceId: observation.traceId,
        parentSpanId: actualParent?.id,
        traceparent: spanTraceparent,
        startTime: Date.now(),
        ctx,
        parent_tool_use_id: event.parent_tool_use_id ?? undefined,
      });

      // For Task (subagent) tools, store pending parent context for the spawned subagent to find
      if (isSubagent) {
        const pendingContext: PendingParentContext = {
          traceparent: spanTraceparent,
          traceId: observation.traceId,
          observationId: observation.id,
          parentSessionId: event.session_id,
          toolUseId: event.tool_use_id,
          subagentType: subagentInfo?.type,
          createdAt: Date.now(),
        };
        storePendingParentContext(pendingContext);
        DEBUG && log("DEBUG", `Stored pending parent context for subagent: ${subagentInfo?.type || "Task"}`);
      }

      log(
        "INFO",
        `PreToolUse: ${event.tool_name} (${event.tool_use_id}) parent=${
          actualParent?.id ?? "root"
        }`
      );
      break;
    }

    case "PostToolUse": {
      if (!event.tool_name) break;

      const analysis = analyzeToolResult(event.tool_response);
      const isSubagent = isSubagentTool(event.tool_name);
      const subagentInfo = isSubagent
        ? getSubagentInfo(event.tool_input)
        : null;

      let toolDurationMs;

      const active = event.tool_use_id
        ? activeObservations.get(event.tool_use_id)
        : undefined;

      if (active) {
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

        finalizeToolObservation(active.observation, result, active.ctx, event.tokens);
        activeObservations.delete(event.tool_use_id!);

        // Remove persisted active span (cleanup)
        try {
          popActiveSpan(event.session_id, event.tool_use_id!);
        } catch {
          // ignore if persistence not found
        }

        // Clean up pending parent context if this was a Task (subagent) tool
        if (isSubagent) {
          removePendingParentContext(event.session_id, event.tool_use_id!);
        }

        log(
          "INFO",
          `${event.tool_name}${
            subagentInfo ? ` (${subagentInfo.type})` : ""
          } (${durationMs}ms): ${analysis.success ? "OK" : "ERROR"}`
        );
      } else if (event.tool_use_id) {
        // Cross-process completion
        const persistedSpan = popActiveSpan(
          event.session_id,
          event.tool_use_id
        );

        if (persistedSpan) {
          const durationMs =
            Date.now() - (persistedSpan.startTime ?? Date.now());
          toolDurationMs = durationMs;

          const restoredCtx: ToolContext = persistedSpan.ctx ?? {
            toolName: event.tool_name,
            toolUseId: event.tool_use_id,
            toolInput: event.tool_input,
            isSubagent,
            subagentType: subagentInfo?.type,
            subagentDescription: subagentInfo?.description,
            subagentModel: subagentInfo?.model,
            model: event.model,
          };

          // Use traceparent for cross-process context restoration
          let observation: ToolObservation;
          if (persistedSpan.traceparent) {
            // Create observation within restored parent context
            observation = createToolObservationWithContext(restoredCtx, persistedSpan.traceparent, event.session_id);
            debugLog(`Cross-process observation with span traceparent: ${persistedSpan.traceparent}`);
          } else if (persistedSession?.traceparent) {
            // Fall back to session traceparent
            observation = createToolObservationWithContext(restoredCtx, persistedSession.traceparent, event.session_id);
            debugLog(`Cross-process observation with session traceparent: ${persistedSession.traceparent}`);
          } else {
            // Last resort: attach to in-memory session if available
            observation = createToolObservation(restoredCtx, undefined, sessionObs);
            debugLog(`Cross-process observation with in-memory session (no traceparent)`);
          }

          const result: ToolResult = {
            success: analysis.success,
            error: analysis.error ?? undefined,
            errorType: analysis.errorType ?? undefined,
            exitCode: analysis.exitCode ?? undefined,
            output: event.tool_response,
            durationMs,
          };

          finalizeToolObservation(observation, result, restoredCtx, event.tokens);

          // Clean up pending parent context if this was a Task (subagent) tool
          if (isSubagent) {
            removePendingParentContext(event.session_id, event.tool_use_id);
          }

          log(
            "INFO",
            `${event.tool_name}${
              subagentInfo ? ` (${subagentInfo.type})` : ""
            } (${durationMs}ms): ${
              analysis.success ? "OK" : "ERROR"
            } [cross-process]`
          );
        } else {
          // Fallback: no persisted span -> create observation attached to session
          const ctx: ToolContext = {
            toolName: event.tool_name ?? "unknown",
            toolUseId: event.tool_use_id ?? "unknown",
            toolInput: event.tool_input,
            isSubagent,
            subagentType: subagentInfo?.type,
            subagentDescription: subagentInfo?.description,
            subagentModel: subagentInfo?.model,
            model: event.model,
          };

          // Use session traceparent if available for cross-process linking
          let obs: ToolObservation;
          if (persistedSession?.traceparent) {
            obs = createToolObservationWithContext(ctx, persistedSession.traceparent, event.session_id);
          } else {
            obs = createToolObservation(ctx, undefined, sessionObs);
          }

          const result: ToolResult = {
            success: analysis.success,
            error: analysis.error ?? undefined,
            errorType: analysis.errorType ?? undefined,
            exitCode: analysis.exitCode ?? undefined,
            output: event.tool_response,
          };

          finalizeToolObservation(obs, result, ctx, event.tokens);

          log(
            "INFO",
            `${event.tool_name}${
              subagentInfo ? ` (${subagentInfo.type})` : ""
            }: ${analysis.success ? "OK" : "ERROR"} [no-persist]`
          );
        }
      } else {
        // No tool_use_id - create standalone observation attached to session
        const ctx: ToolContext = {
          toolName: event.tool_name ?? "unknown",
          toolUseId: "unknown",
          toolInput: event.tool_input,
          isSubagent,
          subagentType: subagentInfo?.type,
          subagentDescription: subagentInfo?.description,
          subagentModel: subagentInfo?.model,
          model: event.model,
        };

        // Use session traceparent if available for cross-process linking
        let observation: ToolObservation;
        if (persistedSession?.traceparent) {
          observation = createToolObservationWithContext(ctx, persistedSession.traceparent, event.session_id);
        } else {
          observation = createToolObservation(ctx, undefined, sessionObs);
        }

        const result: ToolResult = {
          success: analysis.success,
          error: analysis.error ?? undefined,
          errorType: analysis.errorType ?? undefined,
          exitCode: analysis.exitCode ?? undefined,
          output: event.tool_response,
        };

        finalizeToolObservation(observation, result, ctx, event.tokens);

        log(
          "INFO",
          `${event.tool_name}${
            subagentInfo ? ` (${subagentInfo.type})` : ""
          }: ${analysis.success ? "OK" : "ERROR"} [no-id]`
        );
      }

      // Update session metrics after tool completion
      updateSessionMetrics(
        event.session_id,
        event.tool_name ?? "unknown",
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
      const promptMetadata: Record<string, unknown> = {
        permission_mode: event.permission_mode,
        timestamp: event.timestamp || new Date().toISOString(),
        prompt_received: !!event.prompt,
      };

      // Capture the user prompt as input
      const promptInput = event.prompt || null;

      if (sessionObs) {
        recordEvent("user_prompt", promptInput, promptMetadata, sessionObs);
      } else if (persistedSession?.traceparent) {
        // Cross-process: use traceparent to link to correct trace
        recordEventWithContext("user_prompt", promptInput, promptMetadata, persistedSession.traceparent, event.session_id);
      }
      DEBUG && log("DEBUG", `UserPromptSubmit: ${promptInput ? "with prompt" : "no prompt field"}`);
      break;
    }

    case "PreCompact": {
      const compactMetadata: Record<string, unknown> = {
        timestamp: event.timestamp || new Date().toISOString(),
        trigger: event.trigger || "unknown", // "manual" or "auto"
        event_type: "pre_compact",
      };

      if (event.custom_instructions) {
        compactMetadata.has_custom_instructions = true;
      }

      if (sessionObs) {
        recordEvent("compact_started", null, compactMetadata, sessionObs);
      } else if (persistedSession?.traceparent) {
        recordEventWithContext("compact_started", null, compactMetadata, persistedSession.traceparent, event.session_id);
      }
      log("INFO", `PreCompact (trigger: ${event.trigger || "unknown"})`);
      break;
    }

    case "PostCompact": {
      const compactMetadata: Record<string, unknown> = {
        timestamp: event.timestamp || new Date().toISOString(),
        trigger: event.trigger || "unknown",
        event_type: "post_compact",
        compaction_complete: true,
      };

      if (sessionObs) {
        recordEvent("compact_completed", null, compactMetadata, sessionObs);
      } else if (persistedSession?.traceparent) {
        recordEventWithContext("compact_completed", null, compactMetadata, persistedSession.traceparent, event.session_id);
      }
      log("INFO", "PostCompact completed");
      break;
    }

    case "SubagentStop": {
      const eventMetadata = {
        stop_hook_active: event.stop_hook_active ?? false,
        timestamp: event.timestamp || new Date().toISOString(),
      };

      if (sessionObs) {
        // In-memory session available (same process)
        recordEvent("subagent_completed", null, eventMetadata, sessionObs);
      } else if (persistedSession?.traceparent) {
        // Cross-process: use traceparent to link to correct trace
        recordEventWithContext("subagent_completed", null, eventMetadata, persistedSession.traceparent, event.session_id);
        DEBUG && log("DEBUG", `SubagentStop with cross-process traceparent: ${persistedSession.traceparent}`);
      } else {
        // Fallback: create orphan event (will not be linked)
        recordEvent("subagent_completed", null, eventMetadata);
        DEBUG && log("DEBUG", "SubagentStop without session context (orphan event)");
      }
      log("INFO", "Subagent completed");
      break;
    }

    case "Stop": {
      // End any orphaned observations
      if (activeObservations.size > 0) {
        DEBUG &&
          log(
            "DEBUG",
            `Cleaning up ${activeObservations.size} incomplete observations`
          );
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

      // Periodically clean up expired pending parent contexts
      cleanupPendingParentContexts();

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
    debugLog(`Received line (${line.length} chars): ${line.substring(0, 200)}...`);
    try {
      const data = JSON.parse(line);
      debugLog(`Parsed event: type=${data.type || data.hook_event_name}, session=${data.session_id}, tool=${data.tool_name || 'n/a'}`);
      // Log UserPromptSubmit prompt field
      if (data.hook_event_name === "UserPromptSubmit") {
        debugLog(`UserPromptSubmit prompt: ${data.prompt ? `"${data.prompt.substring(0, 100)}..."` : 'absent'}`);
      }
      if (isValidEvent(data)) {
        processEvent(data);
        debugLog(`Event processed successfully`);
      } else {
        debugLog(`Invalid event structure - missing required fields`);
        DEBUG && log("DEBUG", "Invalid event structure");
      }
    } catch (e) {
      debugLog(`Parse error: ${e}`);
      DEBUG && log("DEBUG", `Parse error: ${e}`);
    }
  });

  const shutdown = async () => {
    debugLog(`Shutdown initiated - ${sessionObservations.size} active sessions`);
    // NOTE: Do NOT finalize sessions here!
    // Sessions should only be ended by the explicit "Stop" event.
    // Each hook invocation is a separate process, and the session span
    // should remain "open" until the Stop event comes in a future process.
    // Just flush pending spans without ending sessions.

    try {
      // Explicitly flush spans before shutdown to ensure export completes
      debugLog(`Flushing spans to Langfuse...`);
      await forceFlush();
      await shutdownTracing();
      debugLog(`Shutdown complete`);
    } catch (e) {
      debugLog(`Shutdown error: ${e}`);
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
