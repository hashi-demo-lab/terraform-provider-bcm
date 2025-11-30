# Claude Code Langfuse Hook

Tool error analysis for Claude Code sessions via Langfuse.

## When to Use

| Approach      | Best For                       |
| ------------- | ------------------------------ |
| **OTEL only** | Cost, tokens, session metrics  |
| **This hook** | Tool errors, subagent tracking |
| **Both**      | Complete observability         |

Enable native OTEL with `CLAUDE_CODE_ENABLE_TELEMETRY=1`.

## Setup

### 1. Set Credentials

```bash
# Add to ~/.zshrc or ~/.bashrc
export LANGFUSE_PUBLIC_KEY=pk-lf-your-key
export LANGFUSE_SECRET_KEY=sk-lf-your-key
```

### 2. Build

```bash
cd .claude/hooks
npm install && npm run build
```

### 3. Configure Hooks

Add to `.claude/settings.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "cd .claude/hooks && node dist/langfuse-hook.js"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "cd .claude/hooks && node dist/langfuse-hook.js"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "cd .claude/hooks && node dist/langfuse-hook.js"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "cd .claude/hooks && node dist/langfuse-hook.js"
          }
        ]
      }
    ]
  }
}
```

## Error Types

| Type                | Trigger                      |
| ------------------- | ---------------------------- |
| `error`             | `response.error` field       |
| `failed`            | `response.success === false` |
| `exit_code`         | Non-zero exit code           |
| `http_client_error` | HTTP 4xx                     |
| `http_server_error` | HTTP 5xx                     |
| `timeout`           | `timedOut` or `timeout` flag |
| `cancelled`         | `cancelled` flag             |
| `not_found`         | `notFound` flag              |
| `permission_denied` | `permissionDenied` flag      |

## Features

- ✅ **Error tracking**: 9 error types with detailed classification
- ✅ **Performance metrics**: Tool execution timing (PreToolUse → PostToolUse)
- ✅ **Subagent detection**: Tracks `Task` and `runSubagent` tools
- ✅ **Token tracking**: Captures input/output/total tokens when available
- ✅ **Model tracking**: Records which model is being used
- ✅ **User identification**: Tracks user_id if provided
- ✅ **Span correlation**: Links tool start/end events with unique span IDs

## Filtering in Langfuse

```
metadata.success = false           # All errors
metadata.error_type = "exit_code"  # Bash failures
metadata.is_subagent = true        # Subagent calls
metadata.subagent_type = "Explore" # Specific subagent
metadata.duration_ms > 1000        # Slow tools (>1s)
metadata.tokens_total > 10000      # High token usage
```

## Environment Variables

| Variable               | Required | Default                         |
| ---------------------- | -------- | ------------------------------- |
| `LANGFUSE_PUBLIC_KEY`  | Yes      | -                               |
| `LANGFUSE_SECRET_KEY`  | Yes      | -                               |
| `LANGFUSE_HOST`        | No       | `https://cloud.langfuse.com`    |
| `LANGFUSE_RELEASE`     | No       | `claude-code`                   |
| `LANGFUSE_ENVIRONMENT` | No       | `development`                   |
| `LANGFUSE_LOG_LEVEL`   | No       | (INFO, set `DEBUG` for verbose) |

## Development

```bash
npm run build    # Compile
npm test         # Run tests
npm run watch    # Watch mode
```

## Test Manually

```bash
echo '{"session_id":"test","cwd":"/workspace","hook_event_name":"PostToolUse","tool_name":"Bash","tool_response":{"exit_code":0}}' | node dist/langfuse-hook.js
```
