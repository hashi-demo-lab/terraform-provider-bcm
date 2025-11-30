/**
 * TracerProvider setup for Langfuse integration.
 * Uses @langfuse/tracing's setLangfuseTracerProvider for native API support.
 */

import { NodeSDK } from "@opentelemetry/sdk-node";
import { LangfuseSpanProcessor } from "@langfuse/otel";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { ATTR_SERVICE_NAME } from "@opentelemetry/semantic-conventions";
import {
  setLangfuseTracerProvider,
  getLangfuseTracerProvider,
} from "@langfuse/tracing";
import type { TracingConfig } from "./types.js";

// Module-level state
let sdk: NodeSDK | null = null;
let spanProcessor: LangfuseSpanProcessor | null = null;
let isInitialized = false;

/**
 * Default configuration values.
 */
const DEFAULTS = {
  baseUrl: "https://cloud.langfuse.com",
  environment: "development",
  release: "claude-code",
  serviceName: "claude-code-hook",
  serviceVersion: "1.0.0",
} as const;

/**
 * Initialize the tracing provider with Langfuse configuration.
 *
 * @param config - Tracing configuration with API keys
 * @returns true if initialization succeeded, false otherwise
 */
export function initTracing(config: TracingConfig): boolean {
  if (isInitialized) {
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
    // Create Langfuse span processor
    spanProcessor = new LangfuseSpanProcessor({
      publicKey,
      secretKey,
      baseUrl,
      environment,
      release,
      flushAt: 1,
      exportMode: "immediate",
    });

    // Create and start NodeSDK
    sdk = new NodeSDK({
      resource: resourceFromAttributes({
        [ATTR_SERVICE_NAME]: DEFAULTS.serviceName,
        "deployment.environment": environment,
        "service.version": release,
      }),
      spanProcessors: [spanProcessor],
    });

    sdk.start();

    // Set the tracer provider for @langfuse/tracing native API
    // This enables startObservation() and other native functions
    const provider = getLangfuseTracerProvider();
    setLangfuseTracerProvider(provider);

    isInitialized = true;

    if (config.debug) {
      console.error(`[Langfuse] Initialized (${release}/${environment})`);
    }

    return true;
  } catch (error) {
    console.error(`[Langfuse] ERROR: Failed to initialize: ${error}`);
    return false;
  }
}

/**
 * Force flush all pending spans to Langfuse.
 * Call this before process exit to ensure data is exported.
 */
export async function forceFlush(): Promise<void> {
  if (spanProcessor) {
    try {
      await spanProcessor.forceFlush();
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
    // Force flush to ensure all spans are exported
    await forceFlush();

    // Shutdown the SDK
    if (sdk) {
      await sdk.shutdown();
    }
  } catch {
    // Ignore shutdown errors
  } finally {
    // Reset state
    sdk = null;
    spanProcessor = null;
    isInitialized = false;
    setLangfuseTracerProvider(null);
  }
}

/**
 * Get the span processor for direct access if needed.
 */
export function getSpanProcessor(): LangfuseSpanProcessor | null {
  return spanProcessor;
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
    baseUrl: process.env.LANGFUSE_HOST,
    environment: process.env.LANGFUSE_ENVIRONMENT,
    release: process.env.LANGFUSE_RELEASE,
    debug: process.env.LANGFUSE_LOG_LEVEL === "DEBUG",
  };
}
