/**
 * Tracing module for Langfuse integration.
 *
 * This module provides a type-safe API for creating Langfuse observations
 * using the v4 SDK with asType support for proper observation types.
 *
 * @example
 * ```typescript
 * import {
 *   initTracing,
 *   createConfigFromEnv,
 *   createSessionObservation,
 *   createToolObservation,
 *   shutdownTracing,
 * } from "./tracing/index.js";
 *
 * // Initialize tracing
 * initTracing(createConfigFromEnv());
 *
 * // Create observations with proper types
 * const session = createSessionObservation({ sessionId: "xxx", cwd: "/path" });
 * const agent = createToolObservation({ toolName: "Task", isSubagent: true, ... }, undefined, session);
 * const tool = createToolObservation({ toolName: "Bash", ... }, undefined, session);
 *
 * // Shutdown before exit
 * await shutdownTracing();
 * ```
 */

// Type exports
export type {
  SpanState,
  ActiveSpanInfo,
  GitContext,
  SessionContext,
  ToolContext,
  ToolResult,
  StartObservationOptions,
  TracingConfig,
  ObservationLevel,
  ObservationType,
  SessionMetrics,
  TokenUsage,
} from "./types.js";

// Provider exports
export {
  initTracing,
  shutdownTracing,
  forceFlush,
  getTracingConfig,
  isTracingInitialized,
  createConfigFromEnv,
} from "./provider.js";

// Observation factory exports
export {
  createSessionTraceId,
  createParentContext,
  createSessionObservation,
  createSessionObservationWithParent,
  createToolObservation,
  createToolObservationWithContext,
  createEventObservation,
  finalizeToolObservation,
  finalizeSessionObservation,
  recordEvent,
  recordEventWithContext,
  // Traceparent helpers for cross-process context propagation
  createTraceparent,
  parseTraceparent,
  withParentContext,
  type SessionObservation,
  type ToolObservation,
  type CreateObservationOptions,
  type FinalizeSessionOptions,
} from "./observations.js";

// Persistence exports for cross-process span linking
export {
  loadSpanState,
  saveSpanState,
  deleteSpanState,
  cleanupOldStates,
  registerActiveSpan,
  popActiveSpan,
  getSessionInfo,
  initSession,
  createEmptyMetrics,
  updateSessionMetrics,
  getSessionMetrics,
  calculateAggregateMetrics,
  // Pending parent context for subagent linking
  storePendingParentContext,
  findPendingParentContext,
  removePendingParentContext,
  cleanupPendingParentContexts,
  type PersistedSpanState,
  type TokenData,
  type PendingParentContext,
} from "./persistence.js";
