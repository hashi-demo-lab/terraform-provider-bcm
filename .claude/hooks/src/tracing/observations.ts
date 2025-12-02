/**
 * Observation factory functions for creating Langfuse observations.
 * Uses the v4 SDK with asType support for proper observation types (agent, tool, generation).
 */

import {
  startObservation,
  type LangfuseSpan,
  type LangfuseAgent,
  type LangfuseTool,
  type LangfuseGeneration,
} from "@langfuse/tracing";
import {
  context,
  trace,
  SpanContext,
  TraceFlags,
  ROOT_CONTEXT,
} from "@opentelemetry/api";
import type {
  SessionContext,
  ToolContext,
  ToolResult,
  SessionMetrics,
  ObservationLevel,
  TokenUsage,
} from "./types.js";
import { createHash } from "crypto";

/**
 * Create a W3C traceparent string from trace and span IDs.
 * Format: {version}-{trace-id}-{span-id}-{flags}
 * Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
 *
 * @param traceId - 32 hex character trace ID
 * @param spanId - 16 hex character span ID
 * @returns W3C traceparent string
 */
export function createTraceparent(traceId: string, spanId: string): string {
  // Ensure traceId is 32 chars and spanId is 16 chars
  const normalizedTraceId = traceId.padStart(32, "0").substring(0, 32);
  const normalizedSpanId = spanId.padStart(16, "0").substring(0, 16);
  return `00-${normalizedTraceId}-${normalizedSpanId}-01`;
}

/**
 * Parse a W3C traceparent string into its components.
 *
 * @param traceparent - W3C traceparent string
 * @returns Parsed components or null if invalid
 */
export function parseTraceparent(
  traceparent: string
): { traceId: string; spanId: string; traceFlags: number } | null {
  const match = traceparent.match(
    /^00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})$/i
  );
  if (!match) return null;
  return {
    traceId: match[1],
    spanId: match[2],
    traceFlags: parseInt(match[3], 16),
  };
}

/**
 * Create an OTel SpanContext from a traceparent string.
 *
 * @param traceparent - W3C traceparent string
 * @returns SpanContext or null if invalid
 */
export function spanContextFromTraceparent(traceparent: string): SpanContext | null {
  const parsed = parseTraceparent(traceparent);
  if (!parsed) return null;
  return {
    traceId: parsed.traceId,
    spanId: parsed.spanId,
    traceFlags: parsed.traceFlags as TraceFlags,
    isRemote: true,
  };
}

/**
 * Execute a function within the context of a parent span.
 * This enables cross-process span linking by restoring OTel context.
 *
 * @param traceparent - W3C traceparent string
 * @param fn - Function to execute within the context
 * @returns The result of the function
 */
export function withParentContext<T>(traceparent: string, fn: () => T): T {
  const spanContext = spanContextFromTraceparent(traceparent);
  if (!spanContext) {
    // Invalid traceparent, execute without context
    return fn();
  }

  // Create a context with the parent span
  const parentContext = trace.setSpanContext(ROOT_CONTEXT, spanContext);

  // Execute the function within the parent context
  return context.with(parentContext, fn);
}

// Type for any observation that can be a parent
type AnyObservation = LangfuseSpan | LangfuseAgent | LangfuseTool | LangfuseGeneration;

/**
 * Wrapper for session (trace-level) observation.
 * Uses the v4 SDK span as the root observation for a session.
 */
export interface SessionObservation {
  /** The trace ID */
  traceId: string;
  /** The root observation ID */
  id: string;
  /** Update trace metadata */
  updateTrace(params: {
    name?: string;
    sessionId?: string;
    userId?: string;
    tags?: string[];
  }): void;
  /** Update observation with output/metadata */
  update(params: {
    output?: unknown;
    level?: ObservationLevel;
    metadata?: Record<string, unknown>;
  }): void;
  /** End the observation */
  end(): void;
  /** Internal span reference */
  _span: LangfuseSpan;
}

/**
 * Wrapper for tool observation.
 * Uses asType to set proper observation type (agent, tool, generation).
 */
export interface ToolObservation {
  /** The trace ID this observation belongs to */
  traceId: string;
  /** The observation ID */
  id: string;
  /** The observation type (agent, tool, span) */
  observationType: "agent" | "tool" | "span";
  /** Update observation with output/metadata */
  update(params: {
    output?: unknown;
    level?: ObservationLevel;
    statusMessage?: string;
    metadata?: Record<string, unknown>;
    /** Token usage (for agents/generations) */
    usage?: TokenUsage;
    /** Model name (for agents) */
    model?: string;
  }): void;
  /** End the observation */
  end(): void;
  /** Internal observation reference */
  _observation: AnyObservation;
}

/**
 * Options for creating observations.
 */
export interface CreateObservationOptions {
  /** Parent trace ID for linking */
  parentTraceId?: string;
  /** Parent span ID for nesting */
  parentSpanId?: string;
  /** Custom start time */
  startTime?: Date;
}

/**
 * Create a deterministic trace ID from a session ID.
 * This allows linking observations across stateless processes.
 *
 * @param sessionId - The session identifier
 * @returns A deterministic trace ID (32 hex chars)
 */
export function createSessionTraceId(sessionId: string): string {
  const hash = createHash("sha256").update(sessionId).digest("hex");
  return hash.substring(0, 32);
}

/**
 * Create a session-level observation.
 * Sessions are tracked as root spans in Langfuse v4.
 *
 * @param ctx - Session context
 * @param options - Optional creation settings
 * @returns A SessionObservation wrapper
 */
export function createSessionObservation(
  ctx: SessionContext,
  _options?: CreateObservationOptions
): SessionObservation {
  // Build input with optional git context
  const input: Record<string, unknown> = {
    cwd: ctx.cwd,
    permission_mode: ctx.permissionMode,
  };

  if (ctx.git?.isGitRepo) {
    input.git = {
      repo: ctx.git.repoName,
      branch: ctx.git.branch,
      commit: ctx.git.commitSha,
      is_dirty: ctx.git.isDirty,
    };
  }

  // Build metadata
  const metadata: Record<string, unknown> = {
    session_id: ctx.sessionId,
    user_id: ctx.userId || "unknown",
  };

  if (ctx.git?.isGitRepo) {
    metadata.git_repo = ctx.git.repoName;
    metadata.git_branch = ctx.git.branch;
    metadata.git_commit = ctx.git.commitSha;
  }

  // Create root span using v4 SDK
  const span = startObservation("claude-code-session", {
    input,
    metadata,
  });

  // Set trace-level attributes (sessionId, userId)
  span.updateTrace({
    sessionId: ctx.sessionId,
    userId: ctx.userId || "unknown",
  });

  return {
    traceId: span.traceId,
    id: span.id,
    _span: span,
    updateTrace(params) {
      span.update({
        metadata: {
          ...metadata,
          name: params.name,
          tags: params.tags,
        },
      });
    },
    update(params) {
      span.update({
        output: params.output,
        metadata: params.metadata,
        level: params.level,
      });
    },
    end() {
      span.end();
    },
  };
}

/**
 * Create a session-level observation linked to a parent trace.
 * Used for subagent sessions to link them to the parent's trace hierarchy.
 *
 * @param ctx - Session context
 * @param traceparent - W3C traceparent string for linking to parent trace
 * @param parentContext - Additional parent context for metadata
 * @returns A SessionObservation wrapper linked to the parent trace
 */
export function createSessionObservationWithParent(
  ctx: SessionContext,
  traceparent: string,
  parentContext?: {
    parentSessionId?: string;
    parentObservationId?: string;
    subagentType?: string;
  }
): SessionObservation {
  // Build input with optional git context
  const input: Record<string, unknown> = {
    cwd: ctx.cwd,
    permission_mode: ctx.permissionMode,
  };

  if (ctx.git?.isGitRepo) {
    input.git = {
      repo: ctx.git.repoName,
      branch: ctx.git.branch,
      commit: ctx.git.commitSha,
      is_dirty: ctx.git.isDirty,
    };
  }

  // Build metadata with parent linking info
  const metadata: Record<string, unknown> = {
    session_id: ctx.sessionId,
    user_id: ctx.userId || "unknown",
    is_subagent: true, // Mark as subagent session
  };

  if (ctx.git?.isGitRepo) {
    metadata.git_repo = ctx.git.repoName;
    metadata.git_branch = ctx.git.branch;
    metadata.git_commit = ctx.git.commitSha;
  }

  // Add parent context metadata for hierarchy visualization
  if (parentContext) {
    if (parentContext.parentSessionId) {
      metadata.parent_session_id = parentContext.parentSessionId;
    }
    if (parentContext.parentObservationId) {
      metadata.parent_observation_id = parentContext.parentObservationId;
    }
    if (parentContext.subagentType) {
      metadata.subagent_type = parentContext.subagentType;
    }
  }

  // Parse traceparent to get parent span context for cross-process linking
  const parsedParent = parseTraceparent(traceparent);
  const parentSpanContext = parsedParent ? {
    traceId: parsedParent.traceId,
    spanId: parsedParent.spanId,
    traceFlags: parsedParent.traceFlags as TraceFlags,
    isRemote: true,
  } : undefined;

  // Create root span with parent span context for proper trace linking
  const span = startObservation("claude-code-subagent-session", {
    input,
    metadata,
  }, {
    parentSpanContext,
  });

  // Set trace-level attributes (sessionId, userId)
  span.updateTrace({
    sessionId: ctx.sessionId,
    userId: ctx.userId || "unknown",
  });

  return {
    traceId: span.traceId,
    id: span.id,
    _span: span,
    updateTrace(params) {
      span.update({
        metadata: {
          ...metadata,
          name: params.name,
          tags: params.tags,
        },
      });
    },
    update(params) {
      span.update({
        output: params.output,
        metadata: params.metadata,
        level: params.level,
      });
    },
    end() {
      span.end();
    },
  };
}

/**
 * Create a tool-level observation using v4 SDK with asType.
 * - Subagents (Task tool) use asType: "agent"
 * - Regular tools use asType: "tool"
 *
 * @param ctx - Tool context
 * @param options - Optional creation settings (including parent observation)
 * @param parentObservation - Optional parent observation for hierarchy
 * @returns A ToolObservation wrapper
 */
export function createToolObservation(
  ctx: ToolContext,
  _options?: CreateObservationOptions,
  parentObservation?: SessionObservation | ToolObservation
): ToolObservation {
  const metadata: Record<string, unknown> = {
    tool_name: ctx.toolName,
    tool_use_id: ctx.toolUseId,
  };

  // Subagents get "agent" type for proper hierarchy visualization
  if (ctx.isSubagent) {
    const agentName = ctx.subagentType
      ? `Agent:${ctx.subagentType}`
      : `Agent:${ctx.toolName}`;

    if (ctx.subagentType) metadata.subagent_type = ctx.subagentType;
    if (ctx.subagentDescription) metadata.description = ctx.subagentDescription;
    if (ctx.subagentModel) metadata.model = ctx.subagentModel;

    // Create agent observation - either as child or root
    let observation: LangfuseAgent;
    if (parentObservation) {
      // Create as child of parent
      const parent = "_span" in parentObservation
        ? parentObservation._span
        : parentObservation._observation;
      observation = parent.startObservation(agentName, {
        input: ctx.toolInput,
        metadata,
      }, { asType: "agent" });
    } else {
      // Create as root observation
      observation = startObservation(agentName, {
        input: ctx.toolInput,
        metadata,
      }, { asType: "agent" });
    }

    return {
      traceId: observation.traceId,
      id: observation.id,
      observationType: "agent",
      _observation: observation,
      update(params) {
        observation.update({
          output: params.output,
          level: params.level,
          statusMessage: params.statusMessage,
          metadata: params.metadata,
        });
      },
      end() {
        observation.end();
      },
    };
  }

  // Regular tools get "tool" type
  const toolName = ctx.toolName;

  if (ctx.model) {
    metadata.model = ctx.model;
  }

  // Create tool observation - either as child or root
  let observation: LangfuseTool;
  if (parentObservation) {
    // Create as child of parent
    const parent = "_span" in parentObservation
      ? parentObservation._span
      : parentObservation._observation;
    observation = parent.startObservation(toolName, {
      input: ctx.toolInput,
      metadata,
    }, { asType: "tool" });
  } else {
    // Create as root observation
    observation = startObservation(toolName, {
      input: ctx.toolInput,
      metadata,
    }, { asType: "tool" });
  }

  return {
    traceId: observation.traceId,
    id: observation.id,
    observationType: "tool",
    _observation: observation,
    update(params) {
      observation.update({
        output: params.output,
        level: params.level,
        statusMessage: params.statusMessage,
        metadata: params.metadata,
      });
    },
    end() {
      observation.end();
    },
  };
}

/**
 * Create a tool observation within a restored parent context.
 * Used for cross-process scenarios where we need to attach to an existing trace.
 *
 * @param ctx - Tool context
 * @param traceparent - W3C traceparent string for context restoration
 * @param sessionId - Optional session ID to set on the trace (for cross-process sessions)
 * @returns A ToolObservation wrapper attached to the parent trace
 */
export function createToolObservationWithContext(
  ctx: ToolContext,
  traceparent: string,
  sessionId?: string
): ToolObservation {
  // Parse traceparent to get parent span context for cross-process linking
  const parsedParent = parseTraceparent(traceparent);
  const parentSpanContext = parsedParent ? {
    traceId: parsedParent.traceId,
    spanId: parsedParent.spanId,
    traceFlags: parsedParent.traceFlags as TraceFlags,
    isRemote: true,
  } : undefined;

  const metadata: Record<string, unknown> = {
    tool_name: ctx.toolName,
    tool_use_id: ctx.toolUseId,
    cross_process: true, // Mark as cross-process for debugging
  };

  // Subagents get "agent" type for proper hierarchy visualization
  if (ctx.isSubagent) {
    const agentName = ctx.subagentType
      ? `Agent:${ctx.subagentType}`
      : `Agent:${ctx.toolName}`;

    if (ctx.subagentType) metadata.subagent_type = ctx.subagentType;
    if (ctx.subagentDescription) metadata.description = ctx.subagentDescription;
    if (ctx.subagentModel) metadata.model = ctx.subagentModel;

    // Create agent observation with parent span context for proper trace linking
    const observation: LangfuseAgent = startObservation(agentName, {
      input: ctx.toolInput,
      metadata,
    }, { asType: "agent", parentSpanContext });

    // Set sessionId and name on trace for cross-process session correlation
    if (sessionId) {
      observation.updateTrace({
        sessionId,
        name: "claude-code-session",
      });
    }

    return {
      traceId: observation.traceId,
      id: observation.id,
      observationType: "agent" as const,
      _observation: observation,
      update(params: {
        output?: unknown;
        level?: ObservationLevel;
        statusMessage?: string;
        metadata?: Record<string, unknown>;
        usage?: TokenUsage;
        model?: string;
      }) {
        observation.update({
          output: params.output,
          level: params.level,
          statusMessage: params.statusMessage,
          metadata: params.metadata,
        });
      },
      end() {
        observation.end();
      },
    };
  }

  // Regular tools get "tool" type
  if (ctx.model) {
    metadata.model = ctx.model;
  }

  // Create tool observation with parent span context for proper trace linking
  const observation: LangfuseTool = startObservation(ctx.toolName, {
    input: ctx.toolInput,
    metadata,
  }, { asType: "tool", parentSpanContext });

  // Set sessionId and name on trace for cross-process session correlation
  if (sessionId) {
    observation.updateTrace({
      sessionId,
      name: "claude-code-session",
    });
  }

  return {
    traceId: observation.traceId,
    id: observation.id,
    observationType: "tool" as const,
    _observation: observation,
    update(params: {
      output?: unknown;
      level?: ObservationLevel;
      statusMessage?: string;
      metadata?: Record<string, unknown>;
      usage?: TokenUsage;
      model?: string;
    }) {
      observation.update({
        output: params.output,
        level: params.level,
        statusMessage: params.statusMessage,
        metadata: params.metadata,
      });
    },
    end() {
      observation.end();
    },
  };
}

/**
 * Create an event observation for point-in-time occurrences.
 * In v4 SDK, events are created using startObservation with asType: "event".
 *
 * @param name - Event name
 * @param input - Event input (e.g., user prompt content)
 * @param metadata - Event metadata
 * @param parentObservation - Optional parent observation
 */
export function createEventObservation(
  name: string,
  input?: unknown,
  metadata?: Record<string, unknown>,
  parentObservation?: SessionObservation | ToolObservation
): void {
  const eventMetadata = {
    ...metadata,
    timestamp: new Date().toISOString(),
  };

  if (parentObservation) {
    const parent = "_span" in parentObservation
      ? parentObservation._span
      : parentObservation._observation;
    // Events auto-end in v4 SDK
    parent.startObservation(name, {
      input,
      metadata: eventMetadata,
    }, { asType: "event" });
  } else {
    // Create as root event
    startObservation(name, {
      input,
      metadata: eventMetadata,
    }, { asType: "event" });
  }
}

/**
 * Record an event (create and immediately finalize).
 *
 * @param name - Event name
 * @param input - Event input (e.g., user prompt content)
 * @param metadata - Event metadata
 * @param parentObservation - Optional parent observation
 */
export function recordEvent(
  name: string,
  input?: unknown,
  metadata?: Record<string, unknown>,
  parentObservation?: SessionObservation | ToolObservation
): void {
  createEventObservation(name, input, metadata, parentObservation);
}

/**
 * Record an event within a restored parent context (cross-process).
 * Uses W3C traceparent to link the event to the correct trace.
 *
 * @param name - Event name
 * @param input - Event input (e.g., user prompt content)
 * @param metadata - Event metadata
 * @param traceparent - W3C traceparent string for context restoration
 * @param sessionId - Optional session ID to set on the trace (for cross-process sessions)
 */
export function recordEventWithContext(
  name: string,
  input?: unknown,
  metadata?: Record<string, unknown>,
  traceparent?: string,
  sessionId?: string
): void {
  if (!traceparent) {
    // No traceparent - create as root event (fallback)
    createEventObservation(name, input, metadata);
    return;
  }

  // Parse traceparent to get parent span context for cross-process linking
  const parsedParent = parseTraceparent(traceparent);
  const parentSpanContext = parsedParent ? {
    traceId: parsedParent.traceId,
    spanId: parsedParent.spanId,
    traceFlags: parsedParent.traceFlags as TraceFlags,
    isRemote: true,
  } : undefined;

  const eventMetadata = {
    ...metadata,
    timestamp: new Date().toISOString(),
    cross_process: true,
  };

  // Create event with parent span context for proper trace linking
  const observation = startObservation(name, {
    input,
    metadata: eventMetadata,
  }, { asType: "event", parentSpanContext });

  // Ensure trace has sessionId and name for cross-process events
  if (sessionId) {
    observation.updateTrace({
      sessionId,
      name: "claude-code-session",
    });
  }
}

/**
 * Update a tool observation with its result.
 *
 * @param observation - The observation to update
 * @param result - The tool result
 * @param ctx - Optional additional context
 * @param _tokens - Optional token usage (currently not used in v4 SDK for agent/tool types)
 */
export function finalizeToolObservation(
  observation: ToolObservation,
  result: ToolResult,
  ctx?: Partial<ToolContext>,
  _tokens?: TokenUsage
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
  observation: SessionObservation,
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

  // Add end timestamp to metadata
  const metadata: Record<string, unknown> = {
    end_timestamp: new Date().toISOString(),
  };

  observation.update({
    output,
    metadata,
  });

  observation.end();
}

/**
 * Create parent context info for observation options.
 * This is a helper to build CreateObservationOptions from trace/span IDs.
 */
export function createParentContext(
  traceId: string,
  spanId?: string
): CreateObservationOptions {
  return {
    parentTraceId: traceId,
    parentSpanId: spanId,
  };
}
