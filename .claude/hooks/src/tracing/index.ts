/**
 * Tracing module for Langfuse integration.
 *
 * This module provides a type-safe API for creating Langfuse observations
 * using the native @langfuse/tracing library.
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
 * // Create observations
 * const session = createSessionObservation({ sessionId: "xxx", cwd: "/path" });
 * const tool = createToolObservation({ toolName: "Bash", ... });
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
  getSpanProcessor,
  isTracingInitialized,
  createConfigFromEnv,
} from "./provider.js";

// Observation factory exports
export {
  createSessionTraceId,
  createParentContext,
  createSessionObservation,
  createToolObservation,
  createEventObservation,
  finalizeToolObservation,
  finalizeSessionObservation,
  recordEvent,
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
  type PersistedSpanState,
  type TokenData,
} from "./persistence.js";

// Re-export useful types from @langfuse/tracing
export type {
  LangfuseAgent,
  LangfuseTool,
  LangfuseEvent,
  LangfuseSpan,
  LangfuseObservation,
} from "@langfuse/tracing";
