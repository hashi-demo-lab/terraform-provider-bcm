/**
 * Tests for Langfuse Hook utilities
 * Uses realistic Claude Code event structures
 */

import { describe, expect, it } from '@jest/globals';
import {
  truncate,
  stringify,
  isValidEvent,
  analyzeToolResult,
  getSubagentInfo,
  type ClaudeCodeEvent,
} from './utils.js';

// Realistic test fixtures matching Claude Code events
const fixtures = {
  bashSuccess: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Bash',
    tool_use_id: 'toolu_01ABC123',
    tool_input: {
      command: 'go test -v ./internal/provider/...',
      description: 'Run provider tests',
      timeout: 120000,
    },
    tool_response: {
      exit_code: 0,
      stdout: 'PASS\nok  \tterraform-provider-bcm/internal/provider\t1.234s',
      stderr: '',
    },
  } as ClaudeCodeEvent,

  bashError: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Bash',
    tool_use_id: 'toolu_01DEF456',
    tool_input: {
      command: 'make build',
      description: 'Build the provider',
    },
    tool_response: {
      exit_code: 2,
      stdout: '',
      stderr: 'internal/provider/resource_category.go:45:12: undefined: someVar',
    },
  } as ClaudeCodeEvent,

  readFile: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Read',
    tool_use_id: 'toolu_01GHI789',
    tool_input: {
      file_path: '/workspace/terraform-provider-bcm/internal/provider/provider.go',
    },
    tool_response: {
      content: 'package provider\n\nimport (\n\t"context"\n...',
    },
  } as ClaudeCodeEvent,

  globSearch: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Glob',
    tool_use_id: 'toolu_01JKL012',
    tool_input: {
      pattern: '**/*_test.go',
      path: '/workspace/terraform-provider-bcm',
    },
    tool_response: {
      files: [
        '/workspace/terraform-provider-bcm/internal/provider/provider_test.go',
        '/workspace/terraform-provider-bcm/internal/provider/resource_category_test.go',
      ],
    },
  } as ClaudeCodeEvent,

  subagentExplore: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Task',
    tool_use_id: 'toolu_01MNO345',
    tool_input: {
      subagent_type: 'Explore',
      description: 'Find auth handlers',
      model: 'sonnet',
      prompt: 'Search the codebase for authentication and authorization handling code. Look for login, session management, and permission checks.',
    },
    tool_response: {
      result: 'Found authentication code in internal/provider/bcm_client.go...',
    },
  } as ClaudeCodeEvent,

  subagentCodeReview: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'Task',
    tool_use_id: 'toolu_01PQR678',
    tool_input: {
      subagent_type: 'pr-review-toolkit:code-reviewer',
      description: 'Review recent changes',
      prompt: 'Review the unstaged changes for code quality issues.',
    },
    tool_response: {
      result: 'Code review complete. Found 2 minor issues...',
    },
  } as ClaudeCodeEvent,

  webFetch: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'WebFetch',
    tool_use_id: 'toolu_01STU901',
    tool_input: {
      url: 'https://developer.hashicorp.com/terraform/plugin/framework',
      prompt: 'Extract the key concepts for plugin framework',
    },
    tool_response: {
      statusCode: 200,
      content: 'The Terraform Plugin Framework is...',
    },
  } as ClaudeCodeEvent,

  webFetchError: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'PostToolUse',
    tool_name: 'WebFetch',
    tool_use_id: 'toolu_01VWX234',
    tool_input: {
      url: 'https://example.com/nonexistent',
      prompt: 'Get content',
    },
    tool_response: {
      statusCode: 404,
      error: 'Page not found',
    },
  } as ClaudeCodeEvent,

  stopEvent: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'Stop',
  } as ClaudeCodeEvent,

  subagentStopEvent: {
    session_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    cwd: '/workspace/terraform-provider-bcm',
    permission_mode: 'default',
    hook_event_name: 'SubagentStop',
    stop_hook_active: true,
  } as ClaudeCodeEvent,
};

describe('isValidEvent', () => {
  it('accepts Bash command event', () => {
    expect(isValidEvent(fixtures.bashSuccess)).toBe(true);
  });

  it('accepts Read file event', () => {
    expect(isValidEvent(fixtures.readFile)).toBe(true);
  });

  it('accepts Glob search event', () => {
    expect(isValidEvent(fixtures.globSearch)).toBe(true);
  });

  it('accepts Task/subagent event', () => {
    expect(isValidEvent(fixtures.subagentExplore)).toBe(true);
  });

  it('accepts Stop event', () => {
    expect(isValidEvent(fixtures.stopEvent)).toBe(true);
  });

  it('accepts SubagentStop event', () => {
    expect(isValidEvent(fixtures.subagentStopEvent)).toBe(true);
  });

  it('rejects missing session_id', () => {
    expect(isValidEvent({ cwd: '/workspace', hook_event_name: 'Stop' })).toBe(false);
  });

  it('rejects empty session_id', () => {
    expect(isValidEvent({ session_id: '', cwd: '/workspace', hook_event_name: 'Stop' })).toBe(false);
  });

  it('rejects invalid hook_event_name', () => {
    expect(isValidEvent({ session_id: 'test', cwd: '/workspace', hook_event_name: 'Invalid' })).toBe(false);
  });

  it('rejects null input', () => {
    expect(isValidEvent(null)).toBe(false);
  });

  it('rejects non-object input', () => {
    expect(isValidEvent('string')).toBe(false);
  });
});

describe('analyzeToolResult', () => {
  describe('Bash tool responses', () => {
    it('success: exit code 0 with stdout', () => {
      const r = analyzeToolResult(fixtures.bashSuccess.tool_response);
      expect(r.success).toBe(true);
      expect(r.exitCode).toBe(0);
      expect(r.error).toBeNull();
    });

    it('error: non-zero exit code with stderr', () => {
      const r = analyzeToolResult(fixtures.bashError.tool_response);
      expect(r.success).toBe(false);
      expect(r.exitCode).toBe(2);
      expect(r.error).toContain('undefined: someVar');
      expect(r.errorType).toBe('exit_code');
    });

    it('error: command not found (exit 127)', () => {
      const r = analyzeToolResult({ exit_code: 127, stderr: 'bash: foo: command not found' });
      expect(r.success).toBe(false);
      expect(r.exitCode).toBe(127);
      expect(r.error).toContain('command not found');
    });

    it('error: permission denied (exit 126)', () => {
      const r = analyzeToolResult({ exit_code: 126, stderr: 'bash: ./script.sh: Permission denied' });
      expect(r.success).toBe(false);
      expect(r.exitCode).toBe(126);
    });

    it('error: timeout', () => {
      const r = analyzeToolResult({ timedOut: true, exit_code: null });
      expect(r.success).toBe(false);
      expect(r.errorType).toBe('timeout');
      expect(r.error).toBe('Timed out');
    });
  });

  describe('WebFetch tool responses', () => {
    it('success: HTTP 200', () => {
      const r = analyzeToolResult(fixtures.webFetch.tool_response);
      expect(r.success).toBe(true);
    });

    it('error: HTTP 404 with error field (error takes precedence)', () => {
      const r = analyzeToolResult(fixtures.webFetchError.tool_response);
      expect(r.success).toBe(false);
      expect(r.error).toBe('Page not found');
      expect(r.errorType).toBe('error');
    });

    it('error: HTTP 404 status only', () => {
      const r = analyzeToolResult({ statusCode: 404 });
      expect(r.success).toBe(false);
      expect(r.error).toBe('HTTP 404');
      expect(r.errorType).toBe('http_client_error');
    });

    it('error: HTTP 500 status only', () => {
      const r = analyzeToolResult({ statusCode: 500 });
      expect(r.success).toBe(false);
      expect(r.error).toBe('HTTP 500');
      expect(r.errorType).toBe('http_server_error');
    });

    it('error: HTTP 503 service unavailable', () => {
      const r = analyzeToolResult({ statusCode: 503 });
      expect(r.success).toBe(false);
      expect(r.error).toBe('HTTP 503');
      expect(r.errorType).toBe('http_server_error');
    });
  });

  describe('File operation responses', () => {
    it('success: file read', () => {
      const r = analyzeToolResult(fixtures.readFile.tool_response);
      expect(r.success).toBe(true);
    });

    it('error: file not found', () => {
      const r = analyzeToolResult({ notFound: true });
      expect(r.success).toBe(false);
      expect(r.error).toBe('Not found');
      expect(r.errorType).toBe('not_found');
    });

    it('error: permission denied', () => {
      const r = analyzeToolResult({ permissionDenied: true });
      expect(r.success).toBe(false);
      expect(r.error).toBe('Permission denied');
      expect(r.errorType).toBe('permission_denied');
    });
  });

  describe('Generic responses', () => {
    it('error: explicit error field', () => {
      const r = analyzeToolResult({ error: 'Something went wrong' });
      expect(r.success).toBe(false);
      expect(r.error).toBe('Something went wrong');
      expect(r.errorType).toBe('error');
    });

    it('error: success=false with message', () => {
      const r = analyzeToolResult({ success: false, message: 'Operation failed' });
      expect(r.success).toBe(false);
      expect(r.error).toBe('Operation failed');
      expect(r.errorType).toBe('failed');
    });

    it('error: success=false with reason', () => {
      const r = analyzeToolResult({ success: false, reason: 'Invalid input' });
      expect(r.success).toBe(false);
      expect(r.error).toBe('Invalid input');
    });

    it('error: cancelled by user', () => {
      const r = analyzeToolResult({ cancelled: true });
      expect(r.success).toBe(false);
      expect(r.errorType).toBe('cancelled');
    });

    it('success: null response', () => {
      expect(analyzeToolResult(null).success).toBe(true);
    });

    it('success: undefined response', () => {
      expect(analyzeToolResult(undefined).success).toBe(true);
    });

    it('success: string response', () => {
      expect(analyzeToolResult('Operation completed').success).toBe(true);
    });
  });
});

describe('getSubagentInfo', () => {
  it('extracts Explore subagent info', () => {
    const info = getSubagentInfo(fixtures.subagentExplore.tool_input);
    expect(info).not.toBeNull();
    expect(info!.type).toBe('Explore');
    expect(info!.description).toBe('Find auth handlers');
    expect(info!.model).toBe('sonnet');
    expect(info!.prompt_preview).toContain('Search the codebase');
  });

  it('extracts code-reviewer subagent info', () => {
    const info = getSubagentInfo(fixtures.subagentCodeReview.tool_input);
    expect(info).not.toBeNull();
    expect(info!.type).toBe('pr-review-toolkit:code-reviewer');
    expect(info!.description).toBe('Review recent changes');
  });

  it('truncates long prompts to 200 chars', () => {
    const longPrompt = 'Search for '.repeat(50);
    const info = getSubagentInfo({
      subagent_type: 'Explore',
      prompt: longPrompt,
    });
    expect(info!.prompt_preview.length).toBeLessThanOrEqual(200);
    expect(info!.prompt_preview.endsWith('...')).toBe(true);
  });

  it('returns null for non-Task tool input', () => {
    const info = getSubagentInfo(fixtures.bashSuccess.tool_input);
    expect(info).toBeNull();
  });

  it('returns null for undefined input', () => {
    const info = getSubagentInfo(undefined);
    expect(info).toBeNull();
  });

  it('returns null for empty object', () => {
    const info = getSubagentInfo({});
    expect(info).toBeNull();
  });

  it('handles missing optional fields', () => {
    const info = getSubagentInfo({ subagent_type: 'Explore' });
    expect(info).not.toBeNull();
    expect(info!.type).toBe('Explore');
    expect(info!.description).toBe('');
    expect(info!.model).toBeUndefined();
    expect(info!.prompt_preview).toBe('');
  });
});

describe('stringify', () => {
  it('returns short strings as-is', () => {
    expect(stringify('hello')).toBe('hello');
  });

  it('truncates long strings', () => {
    const long = 'a'.repeat(600);
    const r = stringify(long);
    expect(r.length).toBe(500);
    expect(r.endsWith('...')).toBe(true);
  });

  it('stringifies objects', () => {
    expect(stringify({ command: 'ls -la' })).toBe('{"command":"ls -la"}');
  });

  it('stringifies arrays', () => {
    expect(stringify(['a', 'b'])).toBe('["a","b"]');
  });

  it('handles null', () => {
    expect(stringify(null)).toBe('null');
  });

  it('handles numbers', () => {
    expect(stringify(42)).toBe('42');
  });
});

describe('truncate', () => {
  it('returns short strings as-is', () => {
    expect(truncate('hello')).toBe('hello');
  });

  it('truncates at default max (500)', () => {
    const long = 'a'.repeat(600);
    const r = truncate(long);
    expect(r.length).toBe(500);
    expect(r.endsWith('...')).toBe(true);
  });

  it('truncates at custom max', () => {
    const r = truncate('hello world', 8);
    expect(r).toBe('hello...');
    expect(r.length).toBe(8);
  });

  it('handles exact length', () => {
    const r = truncate('hello', 5);
    expect(r).toBe('hello');
  });

  it('handles length just over limit', () => {
    const r = truncate('hello!', 5);
    expect(r).toBe('he...');
  });
});
