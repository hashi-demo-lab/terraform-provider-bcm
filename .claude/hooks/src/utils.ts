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
  /** Parent tool use ID for nested tools/subagents */
  parent_tool_use_id?: string;
  stop_hook_active?: boolean;
  timestamp?: string;
  model?: string;
  tokens?: {
    input?: number;
    output?: number;
    total?: number;
  };
  user_id?: string;
  // Compact event fields
  trigger?: "manual" | "auto";
  custom_instructions?: string;
  // UserPromptSubmit fields
  /** User prompt content (for UserPromptSubmit events) */
  prompt?: string;
}

// Constants
export const VALID_EVENTS = [
  "PostToolUse",
  "SubagentStop",
  "Stop",
  "PreToolUse",
  "UserPromptSubmit",
  "PreCompact",
  "PostCompact",
];

// String Utilities
export const truncate = (s: string, max = 500): string =>
  s.length > max ? s.slice(0, max - 3) + "..." : s;

export function stringify(v: unknown): string {
  if (typeof v === "string") return truncate(v);
  try {
    return truncate(JSON.stringify(v));
  } catch {
    return String(v);
  }
}

// Validation
export function isValidEvent(data: unknown): data is ClaudeCodeEvent {
  if (!data || typeof data !== "object") return false;
  const d = data as Record<string, unknown>;
  return (
    typeof d.session_id === "string" &&
    d.session_id.length > 0 &&
    typeof d.cwd === "string" &&
    typeof d.hook_event_name === "string" &&
    VALID_EVENTS.includes(d.hook_event_name)
  );
}

// Error detection rules (data-driven for maintainability)
type ErrorRule = {
  check: (r: Record<string, unknown>, exitCode: number | null) => boolean;
  getError: (
    r: Record<string, unknown>,
    exitCode: number | null
  ) => [string, string];
};
const errorRules: ErrorRule[] = [
  { check: (r) => !!r.error, getError: (r) => [stringify(r.error), "error"] },
  {
    check: (r) => r.success === false,
    getError: (r) => [stringify(r.message ?? r.reason ?? "Failed"), "failed"],
  },
  {
    check: (_, e) => typeof e === "number" && e !== 0,
    getError: (r, e) => [
      truncate(
        typeof r.stderr === "string" && r.stderr ? r.stderr : `Exit code ${e}`
      ),
      "exit_code",
    ],
  },
  {
    check: (r) => typeof r.statusCode === "number" && r.statusCode >= 400,
    getError: (r) => [
      `HTTP ${r.statusCode}`,
      (r.statusCode as number) >= 500
        ? "http_server_error"
        : "http_client_error",
    ],
  },
  {
    check: (r) => !!(r.timedOut || r.timeout),
    getError: () => ["Timed out", "timeout"],
  },
  { check: (r) => !!r.cancelled, getError: () => ["Cancelled", "cancelled"] },
  { check: (r) => !!r.notFound, getError: () => ["Not found", "not_found"] },
  {
    check: (r) => !!r.permissionDenied,
    getError: () => ["Permission denied", "permission_denied"],
  },
];

export function analyzeToolResult(response: unknown): ToolAnalysis {
  const result: ToolAnalysis = {
    success: true,
    error: null,
    errorType: null,
    exitCode: null,
  };
  if (!response || typeof response !== "object") return result;

  const r = response as Record<string, unknown>;
  const exitCode = (r.exit_code ?? r.exitCode) as number | null;
  if (typeof exitCode === "number") result.exitCode = exitCode;

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
  if (!input) return null;

  // Must have either subagent_type or prompt field to be a subagent
  if (
    typeof input.subagent_type !== "string" &&
    typeof input.prompt !== "string"
  ) {
    return null;
  }

  return {
    type:
      typeof input.subagent_type === "string"
        ? input.subagent_type
        : "subagent",
    description: typeof input.description === "string" ? input.description : "",
    model: typeof input.model === "string" ? input.model : undefined,
    prompt_preview:
      typeof input.prompt === "string" ? truncate(input.prompt, 200) : "",
  };
}

// Check if tool is a subagent invocation
export function isSubagentTool(toolName?: string): boolean {
  return toolName === "Task" || toolName === "runSubagent";
}

// Git Context
export interface GitContext {
  /** Whether we're in a git repository */
  isGitRepo: boolean;
  /** Current branch name */
  branch?: string;
  /** Remote origin URL */
  remoteUrl?: string;
  /** Repository name (extracted from remote URL or directory name) */
  repoName?: string;
  /** Current commit SHA (short) */
  commitSha?: string;
  /** Whether there are uncommitted changes */
  isDirty?: boolean;
}

/**
 * Execute a git command safely using execFileSync (no shell).
 */
function execGitCommand(args: string[], cwd: string): string | null {
  try {
    const { execFileSync } = require("node:child_process");
    return execFileSync("git", args, {
      cwd,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "pipe"],
    }).trim();
  } catch {
    return null;
  }
}

/**
 * Extract repository name from a git remote URL.
 * Handles HTTPS, SSH, and other URL formats.
 */
function extractRepoName(remoteUrl: string): string {
  // Remove trailing .git
  let url = remoteUrl.replace(/\.git$/, "");

  // Handle SSH format: git@github.com:owner/repo
  const sshMatch = url.match(/[:/]([^/:]+\/[^/:]+)$/);
  if (sshMatch) {
    return sshMatch[1];
  }

  // Handle HTTPS format: https://github.com/owner/repo
  const httpsMatch = url.match(/\/([^/]+\/[^/]+)$/);
  if (httpsMatch) {
    return httpsMatch[1];
  }

  return url;
}

/**
 * Get git context for a working directory.
 * Returns information about the current git repository state.
 *
 * @param cwd - The working directory to check
 * @returns Git context information
 */
export function getGitContext(cwd: string): GitContext {
  // Check if we're in a git repo
  const gitDir = execGitCommand(["rev-parse", "--git-dir"], cwd);
  if (!gitDir) {
    return { isGitRepo: false };
  }

  const context: GitContext = { isGitRepo: true };

  // Get current branch
  const branch = execGitCommand(["rev-parse", "--abbrev-ref", "HEAD"], cwd);
  if (branch) {
    context.branch = branch;
  }

  // Get remote URL
  const remoteUrl = execGitCommand(["remote", "get-url", "origin"], cwd);
  if (remoteUrl) {
    context.remoteUrl = remoteUrl;
    context.repoName = extractRepoName(remoteUrl);
  } else {
    // Fall back to directory name if no remote
    const { basename } = require("node:path");
    context.repoName = basename(cwd);
  }

  // Get current commit SHA (short)
  const commitSha = execGitCommand(["rev-parse", "--short", "HEAD"], cwd);
  if (commitSha) {
    context.commitSha = commitSha;
  }

  // Check if working directory is dirty
  const status = execGitCommand(["status", "--porcelain"], cwd);
  if (status !== null) {
    context.isDirty = status.length > 0;
  }

  return context;
}
