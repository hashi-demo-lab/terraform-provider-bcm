/**
 * Langfuse tracing provider using the v4 SDK with asType support.
 * Initializes OpenTelemetry with LangfuseSpanProcessor for proper observation types.
 */

import { NodeTracerProvider } from "@opentelemetry/sdk-trace-node";
import { LangfuseSpanProcessor } from "@langfuse/otel";
import { setLangfuseTracerProvider } from "@langfuse/tracing";
import type { TracingConfig } from "./types.js";

// Module-level state
let isInitialized = false;
let provider: NodeTracerProvider | null = null;
let currentConfig: TracingConfig | null = null;

/**
 * Default configuration values.
 */
const DEFAULTS = {
  baseUrl: "https://cloud.langfuse.com",
  environment: "development",
  release: "claude-code",
} as const;

/**
 * Initialize the Langfuse tracing provider with OpenTelemetry.
 * Must be called before any tracing operations.
 *
 * @param config - Tracing configuration with API keys
 * @returns true if initialization succeeded, false otherwise
 */
export function initTracing(config: TracingConfig): boolean {
  if (isInitialized && provider) {
    return true;
  }

  const {
    publicKey,
    secretKey,
    baseUrl = DEFAULTS.baseUrl,
    environment = DEFAULTS.environment,
    release = DEFAULTS.release,
  } = config;

  if (!publicKey || !secretKey) {
    console.error("[Langfuse] ERROR: Missing LANGFUSE_PUBLIC_KEY or LANGFUSE_SECRET_KEY");
    return false;
  }

  try {
    // Set environment variables for Langfuse SDK
    process.env.LANGFUSE_PUBLIC_KEY = publicKey;
    process.env.LANGFUSE_SECRET_KEY = secretKey;
    process.env.LANGFUSE_BASE_URL = baseUrl;
    process.env.LANGFUSE_RELEASE = release;
    process.env.LANGFUSE_ENVIRONMENT = environment;

    // Create NodeTracerProvider with LangfuseSpanProcessor
    provider = new NodeTracerProvider({
      spanProcessors: [new LangfuseSpanProcessor()],
    });

    // Set as the Langfuse tracer provider for @langfuse/tracing
    setLangfuseTracerProvider(provider);

    currentConfig = { ...config, environment, release };
    isInitialized = true;

    console.error(`[Langfuse] Initialized (${release}/${environment})`);

    return true;
  } catch (error) {
    console.error(`[Langfuse] ERROR: Failed to initialize: ${error}`);
    return false;
  }
}

/**
 * Get the current tracing configuration.
 */
export function getTracingConfig(): TracingConfig | null {
  return currentConfig;
}

/**
 * Force flush all pending spans to Langfuse.
 * Call this before process exit to ensure data is exported.
 */
export async function forceFlush(): Promise<void> {
  if (provider) {
    try {
      await provider.forceFlush();
    } catch {
      // Ignore flush errors during shutdown
    }
  }
}

/**
 * Shutdown the tracing provider gracefully.
 * Flushes pending spans and cleans up resources.
 */
export async function shutdownTracing(): Promise<void> {
  if (!isInitialized) {
    return;
  }

  try {
    await forceFlush();
    if (provider) {
      await provider.shutdown();
    }
  } catch {
    // Ignore shutdown errors
  } finally {
    provider = null;
    currentConfig = null;
    isInitialized = false;
  }
}

/**
 * Check if tracing has been initialized.
 */
export function isTracingInitialized(): boolean {
  return isInitialized;
}

/**
 * Create tracing config from environment variables.
 */
export function createConfigFromEnv(): TracingConfig {
  return {
    publicKey: process.env.LANGFUSE_PUBLIC_KEY || "",
    secretKey: process.env.LANGFUSE_SECRET_KEY || "",
    baseUrl: process.env.LANGFUSE_HOST || process.env.LANGFUSE_BASE_URL,
    environment: process.env.LANGFUSE_ENVIRONMENT,
    release: process.env.LANGFUSE_RELEASE,
    debug: process.env.LANGFUSE_LOG_LEVEL === "DEBUG",
  };
}
