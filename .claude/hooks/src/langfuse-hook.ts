#!/usr/bin/env node
import { config } from 'dotenv';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Load .env from the hooks directory (not CWD)
const __dirname = dirname(fileURLToPath(import.meta.url));
config({ path: join(__dirname, '..', '.env'), override: true });

import { Langfuse } from 'langfuse';
import { createInterface } from 'node:readline';
import { type ClaudeCodeEvent, isValidEvent, analyzeToolResult, getSubagentInfo } from './utils.js';

/**
 * Claude Code Langfuse Hook - Tool Error Analysis
 * Complements native OTEL by providing error categorization and subagent tracking.
 */

// Configuration
const DEBUG = process.env.LANGFUSE_LOG_LEVEL === 'DEBUG';
const RELEASE = process.env.LANGFUSE_RELEASE || 'claude-code';
const ENVIRONMENT = process.env.LANGFUSE_ENVIRONMENT || 'development';

const log = (level: string, msg: string) => console.error(`[Langfuse] ${level === 'ERROR' ? 'ERROR: ' : ''}${msg}`);

// Initialize Langfuse (module-level, no class needed)
function initLangfuse(): Langfuse | null {
  const publicKey = process.env.LANGFUSE_PUBLIC_KEY;
  const secretKey = process.env.LANGFUSE_SECRET_KEY;
  if (!publicKey || !secretKey) {
    log('ERROR', 'Missing LANGFUSE_PUBLIC_KEY or LANGFUSE_SECRET_KEY');
    return null;
  }
  log('INFO', `Initialized (${RELEASE}/${ENVIRONMENT})`);
  return new Langfuse({
    publicKey,
    secretKey,
    baseUrl: process.env.LANGFUSE_HOST || 'https://cloud.langfuse.com',
    flushAt: 1,
  });
}

// Build metadata for PostToolUse (extracted for clarity)
function buildToolMetadata(event: ClaudeCodeEvent, analysis: ReturnType<typeof analyzeToolResult>, subagent: ReturnType<typeof getSubagentInfo>) {
  const meta: Record<string, unknown> = {
    tool_name: event.tool_name,
    tool_use_id: event.tool_use_id,
    success: analysis.success,
    error: analysis.error,
    error_type: analysis.errorType,
    exit_code: analysis.exitCode,
    is_subagent: !!subagent,
  };
  if (subagent) {
    meta.subagent_type = subagent.type;
    meta.subagent_description = subagent.description;
    meta.subagent_model = subagent.model;
    meta.subagent_prompt_preview = subagent.prompt_preview;
  }
  return meta;
}

// Process a single event
async function processEvent(lf: Langfuse, event: ClaudeCodeEvent) {
  const trace = lf.trace({
    id: event.session_id,
    sessionId: event.session_id,
    release: RELEASE,
    metadata: { cwd: event.cwd, permission_mode: event.permission_mode, environment: ENVIRONMENT },
  });

  switch (event.hook_event_name) {
    case 'PostToolUse': {
      if (!event.tool_name) break;
      const analysis = analyzeToolResult(event.tool_response);
      const subagent = event.tool_name === 'Task' ? getSubagentInfo(event.tool_input) : null;
      trace.span({
        name: subagent ? `Subagent:${subagent.type}` : `Tool:${event.tool_name}`,
        input: event.tool_input,
        output: event.tool_response as Record<string, unknown>,
        level: analysis.success ? 'DEFAULT' : 'ERROR',
        statusMessage: analysis.error || undefined,
        metadata: buildToolMetadata(event, analysis, subagent),
      });
      log('INFO', `${event.tool_name}${subagent ? ` (${subagent.type})` : ''}: ${analysis.success ? 'OK' : 'ERROR'}`);
      break;
    }
    case 'PreToolUse':
      if (event.tool_name) {
        trace.event({ name: `PreTool:${event.tool_name}`, metadata: { tool_name: event.tool_name } });
        DEBUG && log('DEBUG', `PreToolUse: ${event.tool_name}`);
      }
      break;
    case 'UserPromptSubmit':
      trace.event({ name: 'user_prompt', metadata: { permission_mode: event.permission_mode } });
      DEBUG && log('DEBUG', 'UserPromptSubmit');
      break;
    case 'SubagentStop':
      trace.event({ name: 'subagent_completed', metadata: { stop_hook_active: event.stop_hook_active } });
      log('INFO', 'Subagent completed');
      break;
    case 'Stop':
      trace.event({ name: 'session_end' });
      log('INFO', 'Session ended');
      break;
  }

  try { await lf.flushAsync(); } catch (e) { log('ERROR', `Flush failed: ${e}`); }
}

// Main entry point
async function main() {
  const lf = initLangfuse();
  const rl = createInterface({ input: process.stdin, terminal: false });

  rl.on('line', async (line) => {
    try {
      const data = JSON.parse(line);
      if (isValidEvent(data) && lf) await processEvent(lf, data);
      else DEBUG && log('DEBUG', 'Invalid event structure');
    } catch (e) {
      DEBUG && log('DEBUG', `Parse error: ${e}`);
    }
  });

  const shutdown = async () => {
    if (lf) try { await lf.shutdownAsync(); } catch { /* ignore */ }
    process.exit(0);
  };

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
  rl.on('close', shutdown);
}

main().catch(e => { log('ERROR', `Fatal: ${e}`); process.exit(1); });
