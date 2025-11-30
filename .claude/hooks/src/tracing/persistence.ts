/**
 * File-based persistence for span state across process invocations.
 *
 * Each Claude Code hook invocation is a SEPARATE Node.js process:
 * - Process 1: PreToolUse event -> creates span -> exits (span context lost!)
 * - Process 2: PostToolUse event -> needs to find/update span -> can't find it
 *
 * This module provides file-based persistence to store span state across
 * process invocations, enabling proper span linking and duration calculation.
 */

import { writeFileSync, readFileSync, unlinkSync, existsSync, mkdirSync, readdirSync, renameSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import type { SpanState, ActiveSpanInfo, SessionMetrics } from "./types.js";

/** Directory for storing span state files */
const PERSISTENCE_DIR = join(tmpdir(), "langfuse-claude-code");

/** Prefix for state files */
const STATE_PREFIX = "state-";

/** Maximum age for state files before cleanup (24 hours) */
const MAX_AGE_MS = 24 * 60 * 60 * 1000;

/**
 * Extended SpanState with creation timestamp for cleanup purposes.
 */
export interface PersistedSpanState extends SpanState {
  /** Timestamp when the state was created (ms since epoch) */
  createdAt: number;
}

/**
 * Get the file path for a session's state file.
 *
 * @param sessionId - The session identifier
 * @returns The absolute path to the state file
 */
function getStatePath(sessionId: string): string {
  // Sanitize sessionId to prevent path traversal
  const sanitizedId = sessionId.replace(/[^a-zA-Z0-9_-]/g, "_");
  return join(PERSISTENCE_DIR, `${STATE_PREFIX}${sanitizedId}.json`);
}

/**
 * Ensure the persistence directory exists.
 */
function ensureDir(): void {
  if (!existsSync(PERSISTENCE_DIR)) {
    mkdirSync(PERSISTENCE_DIR, { recursive: true, mode: 0o700 });
  }
}

/**
 * Load span state for a session from disk.
 *
 * @param sessionId - The session identifier
 * @returns The persisted span state, or null if not found
 */
export function loadSpanState(sessionId: string): PersistedSpanState | null {
  try {
    const path = getStatePath(sessionId);
    if (!existsSync(path)) return null;
    const data = readFileSync(path, "utf8");
    return JSON.parse(data) as PersistedSpanState;
  } catch {
    return null;
  }
}

/**
 * Save span state for a session to disk.
 * Uses atomic write (write to temp, then rename) to prevent corruption.
 *
 * @param sessionId - The session identifier
 * @param state - The span state to persist
 */
export function saveSpanState(sessionId: string, state: PersistedSpanState): void {
  try {
    ensureDir();
    const path = getStatePath(sessionId);
    const tempPath = `${path}.tmp.${process.pid}`;

    // Write to temp file first
    writeFileSync(tempPath, JSON.stringify(state), { encoding: "utf8", mode: 0o600 });

    // Atomic rename
    renameSync(tempPath, path);
  } catch (e) {
    console.error(`[Langfuse] Failed to save state: ${e}`);
  }
}

/**
 * Delete span state for a session.
 *
 * @param sessionId - The session identifier
 */
export function deleteSpanState(sessionId: string): void {
  try {
    const path = getStatePath(sessionId);
    if (existsSync(path)) {
      unlinkSync(path);
    }
  } catch {
    // Ignore errors during cleanup
  }
}

/**
 * Clean up old state files that exceed the maximum age.
 * Should be called periodically to prevent disk space accumulation.
 */
export function cleanupOldStates(): void {
  try {
    if (!existsSync(PERSISTENCE_DIR)) return;

    const files = readdirSync(PERSISTENCE_DIR);
    const now = Date.now();

    for (const file of files) {
      if (!file.startsWith(STATE_PREFIX)) continue;

      try {
        const path = join(PERSISTENCE_DIR, file);
        const data = readFileSync(path, "utf8");
        const state = JSON.parse(data) as PersistedSpanState;

        if (now - state.createdAt > MAX_AGE_MS) {
          unlinkSync(path);
        }
      } catch {
        // Ignore individual file errors, continue with others
      }
    }
  } catch {
    // Ignore cleanup errors
  }
}

/**
 * Register a new active span for cross-process linking.
 * Called in PreToolUse to persist span info for later retrieval in PostToolUse.
 *
 * @param sessionId - The session identifier
 * @param toolUseId - The unique tool use identifier
 * @param spanId - The OpenTelemetry span ID
 * @param traceId - The trace ID for this session
 * @param sessionSpanId - The session-level span ID (parent)
 */
export function registerActiveSpan(
  sessionId: string,
  toolUseId: string,
  spanId: string,
  traceId: string,
  sessionSpanId: string
): void {
  let state = loadSpanState(sessionId);

  if (!state) {
    state = {
      traceId,
      sessionSpanId,
      activeSpans: {},
      createdAt: Date.now(),
    };
  }

  state.activeSpans[toolUseId] = {
    spanId,
    startTime: Date.now(),
  };

  saveSpanState(sessionId, state);
}

/**
 * Retrieve and remove an active span for cross-process linking.
 * Called in PostToolUse to get span info persisted by PreToolUse.
 *
 * @param sessionId - The session identifier
 * @param toolUseId - The unique tool use identifier
 * @returns The span info including trace context, or null if not found
 */
export function popActiveSpan(
  sessionId: string,
  toolUseId: string
): (ActiveSpanInfo & { traceId: string; sessionSpanId: string }) | null {
  const state = loadSpanState(sessionId);
  if (!state || !state.activeSpans[toolUseId]) return null;

  const span = state.activeSpans[toolUseId];
  delete state.activeSpans[toolUseId];
  saveSpanState(sessionId, state);

  return {
    ...span,
    traceId: state.traceId,
    sessionSpanId: state.sessionSpanId,
  };
}

/**
 * Get session info (trace ID, session span ID) for a session.
 *
 * @param sessionId - The session identifier
 * @returns The session trace context, or null if not found
 */
export function getSessionInfo(
  sessionId: string
): { traceId: string; sessionSpanId: string } | null {
  const state = loadSpanState(sessionId);
  if (!state) return null;

  return {
    traceId: state.traceId,
    sessionSpanId: state.sessionSpanId,
  };
}

/**
 * Initialize or update a session's trace context.
 * Called when a session is first seen to persist the trace/span IDs.
 *
 * @param sessionId - The session identifier
 * @param traceId - The trace ID for this session
 * @param sessionSpanId - The session-level span ID
 */
export function initSession(
  sessionId: string,
  traceId: string,
  sessionSpanId: string
): void {
  let state = loadSpanState(sessionId);

  if (!state) {
    state = {
      traceId,
      sessionSpanId,
      activeSpans: {},
      createdAt: Date.now(),
      metrics: createEmptyMetrics(),
    };
    saveSpanState(sessionId, state);
  }
}

/**
 * Create an empty SessionMetrics object.
 */
export function createEmptyMetrics(): SessionMetrics {
  return {
    toolCount: 0,
    subagentCount: 0,
    errorCount: 0,
    totalDurationMs: 0,
    toolDurations: [],
    errorsByType: {},
    toolsByName: {},
    totalInputTokens: 0,
    totalOutputTokens: 0,
    tokensByTool: {},
    toolsByModel: {},
    modelsUsed: [],
  };
}

/**
 * Token usage data for metrics tracking.
 */
export interface TokenData {
  input?: number;
  output?: number;
  total?: number;
}

/**
 * Update session metrics after a tool execution completes.
 *
 * @param sessionId - The session identifier
 * @param toolName - Name of the tool that executed
 * @param isSubagent - Whether the tool is a subagent (Task tool)
 * @param success - Whether the tool execution succeeded
 * @param errorType - Type of error if failed (optional)
 * @param durationMs - Duration of the tool execution in milliseconds (optional)
 * @param tokens - Token usage data (optional)
 * @param model - Model that executed this tool call (optional)
 */
export function updateSessionMetrics(
  sessionId: string,
  toolName: string,
  isSubagent: boolean,
  success: boolean,
  errorType?: string,
  durationMs?: number,
  tokens?: TokenData,
  model?: string
): void {
  const state = loadSpanState(sessionId);
  if (!state) return;

  // Initialize metrics if not present
  if (!state.metrics) {
    state.metrics = createEmptyMetrics();
  }

  const metrics = state.metrics;

  // Increment tool count
  metrics.toolCount++;

  // Increment subagent count if applicable
  if (isSubagent) {
    metrics.subagentCount++;
  }

  // Track tool usage by name
  metrics.toolsByName[toolName] = (metrics.toolsByName[toolName] || 0) + 1;

  // Handle errors
  if (!success) {
    metrics.errorCount++;
    if (errorType) {
      metrics.errorsByType[errorType] = (metrics.errorsByType[errorType] || 0) + 1;
    }
  }

  // Track duration if provided
  if (durationMs !== undefined) {
    metrics.totalDurationMs += durationMs;
    metrics.toolDurations.push(durationMs);
  }

  // Track token usage if provided
  if (tokens) {
    if (tokens.input !== undefined) {
      metrics.totalInputTokens += tokens.input;
    }
    if (tokens.output !== undefined) {
      metrics.totalOutputTokens += tokens.output;
    }

    // Track per-tool token usage
    if (!metrics.tokensByTool[toolName]) {
      metrics.tokensByTool[toolName] = { input: 0, output: 0, total: 0 };
    }
    const toolTokens = metrics.tokensByTool[toolName];
    if (tokens.input !== undefined) {
      toolTokens.input = (toolTokens.input || 0) + tokens.input;
    }
    if (tokens.output !== undefined) {
      toolTokens.output = (toolTokens.output || 0) + tokens.output;
    }
    if (tokens.total !== undefined) {
      toolTokens.total = (toolTokens.total || 0) + tokens.total;
    }
  }

  // Track model usage if provided
  if (model) {
    metrics.toolsByModel[model] = (metrics.toolsByModel[model] || 0) + 1;
    if (!metrics.modelsUsed.includes(model)) {
      metrics.modelsUsed.push(model);
    }
  }

  saveSpanState(sessionId, state);
}

/**
 * Get session metrics for a session.
 *
 * @param sessionId - The session identifier
 * @returns The session metrics, or null if not found
 */
export function getSessionMetrics(sessionId: string): SessionMetrics | null {
  const state = loadSpanState(sessionId);
  if (!state || !state.metrics) return null;
  return state.metrics;
}

/**
 * Calculate aggregate performance metrics from session metrics.
 *
 * @param metrics - The session metrics
 * @returns Aggregate metrics including averages, min, max, token usage, and model breakdown
 */
export function calculateAggregateMetrics(metrics: SessionMetrics): {
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
} {
  const durations = metrics.toolDurations;
  const hasDurations = durations.length > 0;

  return {
    avgDurationMs: hasDurations
      ? Math.round(metrics.totalDurationMs / durations.length)
      : 0,
    minDurationMs: hasDurations ? Math.min(...durations) : 0,
    maxDurationMs: hasDurations ? Math.max(...durations) : 0,
    toolBreakdown: { ...metrics.toolsByName },
    errorBreakdown: { ...metrics.errorsByType },
    totalInputTokens: metrics.totalInputTokens,
    totalOutputTokens: metrics.totalOutputTokens,
    totalTokens: metrics.totalInputTokens + metrics.totalOutputTokens,
    tokensByTool: { ...metrics.tokensByTool },
    modelBreakdown: { ...metrics.toolsByModel },
    modelsUsed: [...metrics.modelsUsed],
  };
}
