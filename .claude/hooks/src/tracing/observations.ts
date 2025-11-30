/**
 * Observation factory functions for creating Langfuse observations.
 * Uses the native @langfuse/tracing API with proper observation types.
 */

import {
  startObservation,
  createTraceId,
  type LangfuseAgent,
  type LangfuseTool,
  type LangfuseEvent,
  type LangfuseSpanAttributes,
  type ObservationLevel,
} from "@langfuse/tracing";
import type { SpanContext } from "@opentelemetry/api";
import { TraceFlags } from "@opentelemetry/api";
import type {
  SessionContext,
  ToolContext,
  ToolResult,
  StartObservationOptions,
  SessionMetrics,
} from "./types.js";

/**
 * Create a deterministic trace ID from a session ID.
 * This allows linking observations across stateless processes.
 *
 * @param sessionId - The session identifier
 * @returns A deterministic trace ID
 */
export async function createSessionTraceId(sessionId: string): Promise<string> {
  return createTraceId(sessionId);
}

/**
 * Create a parent span context for linking observations.
 *
 * @param traceId - The trace ID
 * @param spanId - The parent span ID
 * @returns A SpanContext object for use with startObservation
 */
export function createParentContext(traceId: string, spanId: string): SpanContext {
  return {
    traceId,
    spanId,
    traceFlags: TraceFlags.SAMPLED,
    isRemote: false,
  };
}

/**
 * Create a session-level agent observation.
 * Sessions are tracked as "agent" type in Langfuse for proper visualization.
 *
 * @param ctx - Session context
 * @param options - Optional start observation settings
 * @returns A LangfuseAgent observation
 */
export function createSessionObservation(
  ctx: SessionContext,
  options?: StartObservationOptions
): LangfuseAgent {
  // Build input with optional git context
  const input: Record<string, unknown> = {
    cwd: ctx.cwd,
    permission_mode: ctx.permissionMode,
  };

  // Build metadata with optional git context
  const metadata: Record<string, unknown> = {
    session_id: ctx.sessionId,
    user_id: ctx.userId || "unknown",
  };

  // Add git context if available
  if (ctx.git?.isGitRepo) {
    input.git = {
      repo: ctx.git.repoName,
      branch: ctx.git.branch,
      commit: ctx.git.commitSha,
      is_dirty: ctx.git.isDirty,
    };
    metadata.git_repo = ctx.git.repoName;
    metadata.git_branch = ctx.git.branch;
    metadata.git_commit = ctx.git.commitSha;
  }

  const attributes: LangfuseSpanAttributes = { input, metadata };

  return startObservation("claude-code-session", attributes, {
    asType: "agent",
    parentSpanContext: options?.parentSpanContext,
    startTime: options?.startTime,
  });
}

/**
 * Create a tool-level observation.
 * Regular tools use "tool" type, subagents use "agent" type.
 *
 * @param ctx - Tool context
 * @param options - Optional start observation settings
 * @returns A LangfuseTool or LangfuseAgent observation
 */
export function createToolObservation(
  ctx: ToolContext,
  options?: StartObservationOptions
): LangfuseTool | LangfuseAgent {
  const spanName = ctx.isSubagent && ctx.subagentType
    ? `Subagent:${ctx.subagentType}`
    : `Tool:${ctx.toolName}`;

  const baseMetadata: Record<string, unknown> = {
    tool_name: ctx.toolName,
    tool_use_id: ctx.toolUseId,
  };

  // Add model if provided
  if (ctx.model) {
    baseMetadata.model = ctx.model;
  }

  if (ctx.isSubagent) {
    if (ctx.subagentType) baseMetadata.subagent_type = ctx.subagentType;
    if (ctx.subagentDescription) baseMetadata.subagent_description = ctx.subagentDescription;
    if (ctx.subagentModel) baseMetadata.subagent_model = ctx.subagentModel;
  }

  const attributes: LangfuseSpanAttributes = {
    input: ctx.toolInput,
    metadata: baseMetadata,
  };

  if (ctx.isSubagent) {
    // Subagents are tracked as "agent" type
    return startObservation(spanName, attributes, {
      asType: "agent",
      parentSpanContext: options?.parentSpanContext,
      startTime: options?.startTime,
    });
  } else {
    // Regular tools use "tool" type
    return startObservation(spanName, attributes, {
      asType: "tool",
      parentSpanContext: options?.parentSpanContext,
      startTime: options?.startTime,
    });
  }
}

/**
 * Create an event observation for point-in-time occurrences.
 *
 * @param name - Event name
 * @param metadata - Event metadata
 * @param options - Optional start observation settings
 * @returns A LangfuseEvent observation
 */
export function createEventObservation(
  name: string,
  metadata?: Record<string, unknown>,
  options?: StartObservationOptions
): LangfuseEvent {
  const attributes: LangfuseSpanAttributes = {
    metadata: {
      ...metadata,
      timestamp: new Date().toISOString(),
    },
  };

  return startObservation(name, attributes, {
    asType: "event",
    parentSpanContext: options?.parentSpanContext,
    startTime: options?.startTime,
  });
}

/**
 * Update a tool observation with its result.
 *
 * @param observation - The observation to update
 * @param result - The tool result
 * @param ctx - Optional additional context
 */
export function finalizeToolObservation(
  observation: LangfuseTool | LangfuseAgent,
  result: ToolResult,
  ctx?: Partial<ToolContext>
): void {
  const level: ObservationLevel = result.success ? "DEFAULT" : "ERROR";

  const metadata: Record<string, unknown> = {
    success: result.success,
  };

  if (result.durationMs !== undefined) {
    metadata.duration_ms = result.durationMs;
  }
  if (result.error) {
    metadata.error = result.error;
  }
  if (result.errorType) {
    metadata.error_type = result.errorType;
  }
  if (result.exitCode !== undefined) {
    metadata.exit_code = result.exitCode;
  }

  // Add subagent context if provided
  if (ctx?.isSubagent) {
    if (ctx.subagentType) metadata.subagent_type = ctx.subagentType;
    if (ctx.subagentDescription) metadata.subagent_description = ctx.subagentDescription;
    if (ctx.subagentModel) metadata.subagent_model = ctx.subagentModel;
  }

  observation.update({
    output: result.output,
    level,
    statusMessage: result.error,
    metadata,
  });

  observation.end();
}

/**
 * Create and immediately end an event observation.
 * Useful for one-shot events like UserPromptSubmit.
 *
 * @param name - Event name
 * @param metadata - Event metadata
 * @param options - Optional start observation settings
 */
export function recordEvent(
  name: string,
  metadata?: Record<string, unknown>,
  options?: StartObservationOptions
): void {
  const event = createEventObservation(name, metadata, options);
  event.end();
}

/**
 * Options for finalizing a session observation.
 */
export interface FinalizeSessionOptions {
  /** Whether the session ended normally */
  ended?: boolean;
  /** Timestamp of session end */
  timestamp?: string;
  /** Session metrics to include in output */
  metrics?: SessionMetrics;
  /** Aggregate metrics (avg, min, max durations, token usage, model breakdown) */
  aggregateMetrics?: {
    avgDurationMs: number;
    minDurationMs: number;
    maxDurationMs: number;
    toolBreakdown: Record<string, number>;
    errorBreakdown: Record<string, number>;
    totalInputTokens: number;
    totalOutputTokens: number;
    totalTokens: number;
    tokensByTool: Record<string, { input?: number; output?: number; total?: number }>;
    modelBreakdown: Record<string, number>;
    modelsUsed: string[];
  };
}

/**
 * Finalize a session observation.
 *
 * @param observation - The session observation to finalize
 * @param options - Optional finalization options including metrics
 */
export function finalizeSessionObservation(
  observation: LangfuseAgent,
  options?: FinalizeSessionOptions
): void {
  const output: Record<string, unknown> = {
    ended: options?.ended ?? true,
    timestamp: options?.timestamp ?? new Date().toISOString(),
  };

  // Include session metrics if provided
  if (options?.metrics) {
    output.metrics = {
      tool_count: options.metrics.toolCount,
      subagent_count: options.metrics.subagentCount,
      error_count: options.metrics.errorCount,
      total_duration_ms: options.metrics.totalDurationMs,
    };
  }

  // Include aggregate metrics if provided
  if (options?.aggregateMetrics) {
    output.performance = {
      avg_duration_ms: options.aggregateMetrics.avgDurationMs,
      min_duration_ms: options.aggregateMetrics.minDurationMs,
      max_duration_ms: options.aggregateMetrics.maxDurationMs,
      tool_breakdown: options.aggregateMetrics.toolBreakdown,
      error_breakdown: options.aggregateMetrics.errorBreakdown,
    };

    // Include token usage if any tokens were tracked
    if (options.aggregateMetrics.totalTokens > 0) {
      output.token_usage = {
        total_input_tokens: options.aggregateMetrics.totalInputTokens,
        total_output_tokens: options.aggregateMetrics.totalOutputTokens,
        total_tokens: options.aggregateMetrics.totalTokens,
        tokens_by_tool: options.aggregateMetrics.tokensByTool,
      };
    }

    // Include model usage if any models were tracked
    if (options.aggregateMetrics.modelsUsed.length > 0) {
      output.model_usage = {
        models_used: options.aggregateMetrics.modelsUsed,
        model_breakdown: options.aggregateMetrics.modelBreakdown,
      };
    }
  }

  // Determine level based on error count
  const level: ObservationLevel =
    options?.metrics && options.metrics.errorCount > 0 ? "WARNING" : "DEFAULT";

  observation.update({
    output,
    level,
    metadata: {
      end_timestamp: new Date().toISOString(),
    },
  });
  observation.end();
}
