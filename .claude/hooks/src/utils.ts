/** Shared utilities for Langfuse hook */

// Types
export interface ToolAnalysis {
  success: boolean;
  error: string | null;
  errorType: string | null;
  exitCode: number | null;
}

export interface ClaudeCodeEvent {
  session_id: string;
  cwd: string;
  permission_mode: string;
  hook_event_name: string;
  tool_name?: string;
  tool_input?: Record<string, unknown>;
  tool_response?: unknown;
  tool_use_id?: string;
  stop_hook_active?: boolean;
}

// Constants
export const VALID_EVENTS = ['PostToolUse', 'SubagentStop', 'Stop', 'PreToolUse', 'UserPromptSubmit'];

// String Utilities
export const truncate = (s: string, max = 500): string => s.length > max ? s.slice(0, max - 3) + '...' : s;

export function stringify(v: unknown): string {
  if (typeof v === 'string') return truncate(v);
  try { return truncate(JSON.stringify(v)); } catch { return String(v); }
}

// Validation
export function isValidEvent(data: unknown): data is ClaudeCodeEvent {
  if (!data || typeof data !== 'object') return false;
  const d = data as Record<string, unknown>;
  return typeof d.session_id === 'string' && d.session_id.length > 0 &&
         typeof d.cwd === 'string' &&
         typeof d.hook_event_name === 'string' && VALID_EVENTS.includes(d.hook_event_name);
}

// Error detection rules (data-driven for maintainability)
type ErrorRule = { check: (r: Record<string, unknown>, exitCode: number | null) => boolean; getError: (r: Record<string, unknown>, exitCode: number | null) => [string, string] };
const errorRules: ErrorRule[] = [
  { check: r => !!r.error, getError: r => [stringify(r.error), 'error'] },
  { check: r => r.success === false, getError: r => [stringify(r.message ?? r.reason ?? 'Failed'), 'failed'] },
  { check: (_, e) => typeof e === 'number' && e !== 0, getError: (r, e) => [truncate(typeof r.stderr === 'string' && r.stderr ? r.stderr : `Exit code ${e}`), 'exit_code'] },
  { check: r => typeof r.statusCode === 'number' && r.statusCode >= 400, getError: r => [`HTTP ${r.statusCode}`, (r.statusCode as number) >= 500 ? 'http_server_error' : 'http_client_error'] },
  { check: r => !!(r.timedOut || r.timeout), getError: () => ['Timed out', 'timeout'] },
  { check: r => !!r.cancelled, getError: () => ['Cancelled', 'cancelled'] },
  { check: r => !!r.notFound, getError: () => ['Not found', 'not_found'] },
  { check: r => !!r.permissionDenied, getError: () => ['Permission denied', 'permission_denied'] },
];

export function analyzeToolResult(response: unknown): ToolAnalysis {
  const result: ToolAnalysis = { success: true, error: null, errorType: null, exitCode: null };
  if (!response || typeof response !== 'object') return result;

  const r = response as Record<string, unknown>;
  const exitCode = (r.exit_code ?? r.exitCode) as number | null;
  if (typeof exitCode === 'number') result.exitCode = exitCode;

  for (const rule of errorRules) {
    if (rule.check(r, exitCode)) {
      const [error, errorType] = rule.getError(r, exitCode);
      return { ...result, success: false, error, errorType };
    }
  }
  return result;
}

// Subagent Extraction
export function getSubagentInfo(input?: Record<string, unknown>) {
  if (!input || typeof input.subagent_type !== 'string') return null;
  return {
    type: input.subagent_type,
    description: typeof input.description === 'string' ? input.description : '',
    model: typeof input.model === 'string' ? input.model : undefined,
    prompt_preview: typeof input.prompt === 'string' ? truncate(input.prompt, 200) : '',
  };
}
