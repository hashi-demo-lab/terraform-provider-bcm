/**
 * Tests for persistence module - cross-process span linking
 */

import { describe, expect, it, beforeEach, afterEach } from 'vitest';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { existsSync, mkdirSync, rmSync, writeFileSync, readFileSync } from 'node:fs';
import {
  loadSpanState,
  saveSpanState,
  deleteSpanState,
  cleanupOldStates,
  registerActiveSpan,
  popActiveSpan,
  getSessionInfo,
  initSession,
  type PersistedSpanState,
} from './persistence.js';

const PERSISTENCE_DIR = join(tmpdir(), "langfuse-claude-code");

// Helper to clean up test state files
function cleanupTestFiles(sessionPrefix: string) {
  try {
    if (existsSync(PERSISTENCE_DIR)) {
      const fs = require("node:fs");
      const files = fs.readdirSync(PERSISTENCE_DIR);
      for (const file of files) {
        if (file.includes(sessionPrefix)) {
          fs.unlinkSync(join(PERSISTENCE_DIR, file));
        }
      }
    }
  } catch {
    // Ignore cleanup errors
  }
}

describe('Persistence Module', () => {
  const testSessionPrefix = 'test-session';

  beforeEach(() => {
    cleanupTestFiles(testSessionPrefix);
  });

  afterEach(() => {
    cleanupTestFiles(testSessionPrefix);
  });

  describe('loadSpanState / saveSpanState', () => {
    it('returns null for non-existent session', () => {
      const state = loadSpanState('non-existent-session-123');
      expect(state).toBeNull();
    });

    it('saves and loads span state correctly', () => {
      const sessionId = `${testSessionPrefix}-001`;
      const state: PersistedSpanState = {
        traceId: 'trace-123',
        sessionSpanId: 'span-456',
        activeSpans: {
          'tool-1': { spanId: 'span-tool-1', startTime: Date.now() },
        },
        createdAt: Date.now(),
      };

      saveSpanState(sessionId, state);
      const loaded = loadSpanState(sessionId);

      expect(loaded).not.toBeNull();
      expect(loaded!.traceId).toBe('trace-123');
      expect(loaded!.sessionSpanId).toBe('span-456');
      expect(loaded!.activeSpans['tool-1'].spanId).toBe('span-tool-1');
    });

    it('handles special characters in session ID', () => {
      const sessionId = `${testSessionPrefix}-special@#$%`;
      const state: PersistedSpanState = {
        traceId: 'trace-special',
        sessionSpanId: 'span-special',
        activeSpans: {},
        createdAt: Date.now(),
      };

      saveSpanState(sessionId, state);
      const loaded = loadSpanState(sessionId);

      expect(loaded).not.toBeNull();
      expect(loaded!.traceId).toBe('trace-special');
    });
  });

  describe('deleteSpanState', () => {
    it('deletes existing state file', () => {
      const sessionId = `${testSessionPrefix}-delete`;
      const state: PersistedSpanState = {
        traceId: 'trace-delete',
        sessionSpanId: 'span-delete',
        activeSpans: {},
        createdAt: Date.now(),
      };

      saveSpanState(sessionId, state);
      expect(loadSpanState(sessionId)).not.toBeNull();

      deleteSpanState(sessionId);
      expect(loadSpanState(sessionId)).toBeNull();
    });

    it('handles non-existent session gracefully', () => {
      // Should not throw
      deleteSpanState('non-existent-session-xyz');
    });
  });

  describe('registerActiveSpan / popActiveSpan', () => {
    it('registers and pops active span correctly', () => {
      const sessionId = `${testSessionPrefix}-span`;
      const toolUseId = 'tool-use-001';

      // Register span with new object-based API
      registerActiveSpan(sessionId, toolUseId, {
        spanId: 'span-id-001',
        startTime: Date.now(),
        traceId: 'trace-id-001',
        parentSpanId: 'session-span-001',
      });

      // Pop span
      const popped = popActiveSpan(sessionId, toolUseId);

      expect(popped).not.toBeNull();
      expect(popped!.spanId).toBe('span-id-001');
      expect(popped!.traceId).toBe('trace-id-001');
      expect(popped!.sessionSpanId).toBe('session-span-001');
      expect(popped!.startTime).toBeLessThanOrEqual(Date.now());

      // Pop again should return null
      const poppedAgain = popActiveSpan(sessionId, toolUseId);
      expect(poppedAgain).toBeNull();
    });

    it('handles multiple active spans in same session', () => {
      const sessionId = `${testSessionPrefix}-multi`;

      // Register multiple spans
      registerActiveSpan(sessionId, 'tool-1', { spanId: 'span-1', startTime: Date.now(), traceId: 'trace-1', parentSpanId: 'session-1' });
      registerActiveSpan(sessionId, 'tool-2', { spanId: 'span-2', startTime: Date.now(), traceId: 'trace-1', parentSpanId: 'session-1' });
      registerActiveSpan(sessionId, 'tool-3', { spanId: 'span-3', startTime: Date.now(), traceId: 'trace-1', parentSpanId: 'session-1' });

      // Pop in different order
      const pop2 = popActiveSpan(sessionId, 'tool-2');
      expect(pop2!.spanId).toBe('span-2');

      const pop1 = popActiveSpan(sessionId, 'tool-1');
      expect(pop1!.spanId).toBe('span-1');

      const pop3 = popActiveSpan(sessionId, 'tool-3');
      expect(pop3!.spanId).toBe('span-3');
    });

    it('returns null for non-existent tool use id', () => {
      const sessionId = `${testSessionPrefix}-nonexistent`;
      registerActiveSpan(sessionId, 'tool-1', { spanId: 'span-1', startTime: Date.now(), traceId: 'trace-1', parentSpanId: 'session-1' });

      const result = popActiveSpan(sessionId, 'tool-wrong');
      expect(result).toBeNull();
    });

    it('returns null for non-existent session', () => {
      const result = popActiveSpan('non-existent-session', 'tool-1');
      expect(result).toBeNull();
    });
  });

  describe('initSession / getSessionInfo', () => {
    it('initializes session and retrieves info correctly', () => {
      const sessionId = `${testSessionPrefix}-init`;

      initSession(sessionId, 'trace-init', 'session-span-init');

      const info = getSessionInfo(sessionId);
      expect(info).not.toBeNull();
      expect(info!.traceId).toBe('trace-init');
      expect(info!.sessionSpanId).toBe('session-span-init');
    });

    it('does not overwrite existing session', () => {
      const sessionId = `${testSessionPrefix}-no-overwrite`;

      initSession(sessionId, 'trace-first', 'session-first');
      initSession(sessionId, 'trace-second', 'session-second');

      const info = getSessionInfo(sessionId);
      expect(info).not.toBeNull();
      expect(info!.traceId).toBe('trace-first');
      expect(info!.sessionSpanId).toBe('session-first');
    });

    it('returns null for non-existent session', () => {
      const info = getSessionInfo('non-existent-session-info');
      expect(info).toBeNull();
    });
  });

  describe('cleanupOldStates', () => {
    it('removes old state files', () => {
      const sessionId = `${testSessionPrefix}-cleanup`;

      // Create state with old timestamp
      const oldState: PersistedSpanState = {
        traceId: 'trace-old',
        sessionSpanId: 'span-old',
        activeSpans: {},
        createdAt: Date.now() - (25 * 60 * 60 * 1000), // 25 hours ago
      };

      saveSpanState(sessionId, oldState);
      expect(loadSpanState(sessionId)).not.toBeNull();

      cleanupOldStates();

      expect(loadSpanState(sessionId)).toBeNull();
    });

    it('keeps recent state files', () => {
      const sessionId = `${testSessionPrefix}-keep`;

      // Create state with recent timestamp
      const recentState: PersistedSpanState = {
        traceId: 'trace-recent',
        sessionSpanId: 'span-recent',
        activeSpans: {},
        createdAt: Date.now() - (1 * 60 * 60 * 1000), // 1 hour ago
      };

      saveSpanState(sessionId, recentState);
      expect(loadSpanState(sessionId)).not.toBeNull();

      cleanupOldStates();

      expect(loadSpanState(sessionId)).not.toBeNull();
    });
  });

  describe('Cross-process simulation', () => {
    it('simulates PreToolUse -> PostToolUse across processes', () => {
      const sessionId = `${testSessionPrefix}-cross-process`;
      const toolUseId = 'toolu_01ABC123';
      const traceId = 'trace-cross-process';
      const sessionSpanId = 'session-span-cross-process';

      // === Process 1: PreToolUse ===
      // Initialize session (first event)
      initSession(sessionId, traceId, sessionSpanId);

      // Create and register span
      const observationId = 'span-tool-01';
      const startTime = Date.now();
      registerActiveSpan(sessionId, toolUseId, {
        spanId: observationId,
        startTime,
        traceId,
        parentSpanId: sessionSpanId,
      });

      // Process 1 exits - in-memory state is lost
      // But file persistence remains

      // === Process 2: PostToolUse ===
      // Get session info (from file)
      const sessionInfo = getSessionInfo(sessionId);
      expect(sessionInfo).not.toBeNull();
      expect(sessionInfo!.traceId).toBe(traceId);
      expect(sessionInfo!.sessionSpanId).toBe(sessionSpanId);

      // Pop active span (from file)
      const spanInfo = popActiveSpan(sessionId, toolUseId);
      expect(spanInfo).not.toBeNull();
      expect(spanInfo!.spanId).toBe(observationId);
      expect(spanInfo!.traceId).toBe(traceId);
      expect(spanInfo!.sessionSpanId).toBe(sessionSpanId);
      expect(spanInfo!.startTime).toBeLessThanOrEqual(Date.now());
      expect(spanInfo!.startTime).toBeGreaterThanOrEqual(startTime - 100); // Allow small time diff

      // Calculate duration
      const durationMs = Date.now() - spanInfo!.startTime;
      expect(durationMs).toBeGreaterThanOrEqual(0);
      expect(durationMs).toBeLessThan(1000); // Should be quick in test
    });

    it('handles multiple tools in cross-process scenario', () => {
      const sessionId = `${testSessionPrefix}-multi-tools`;
      const traceId = 'trace-multi';
      const sessionSpanId = 'session-multi';

      // === Process 1: Multiple PreToolUse ===
      initSession(sessionId, traceId, sessionSpanId);
      registerActiveSpan(sessionId, 'tool-1', { spanId: 'span-1', startTime: Date.now(), traceId, parentSpanId: sessionSpanId });
      registerActiveSpan(sessionId, 'tool-2', { spanId: 'span-2', startTime: Date.now(), traceId, parentSpanId: sessionSpanId });
      registerActiveSpan(sessionId, 'tool-3', { spanId: 'span-3', startTime: Date.now(), traceId, parentSpanId: sessionSpanId });

      // === Process 2: PostToolUse for tool-2 ===
      const span2 = popActiveSpan(sessionId, 'tool-2');
      expect(span2!.spanId).toBe('span-2');

      // === Process 3: PostToolUse for tool-1 and tool-3 ===
      const span1 = popActiveSpan(sessionId, 'tool-1');
      const span3 = popActiveSpan(sessionId, 'tool-3');
      expect(span1!.spanId).toBe('span-1');
      expect(span3!.spanId).toBe('span-3');

      // All spans consumed
      expect(popActiveSpan(sessionId, 'tool-1')).toBeNull();
      expect(popActiveSpan(sessionId, 'tool-2')).toBeNull();
      expect(popActiveSpan(sessionId, 'tool-3')).toBeNull();
    });

    it('handles Stop event cleanup', () => {
      const sessionId = `${testSessionPrefix}-stop`;
      const traceId = 'trace-stop';
      const sessionSpanId = 'session-stop';

      // Create session with active spans
      initSession(sessionId, traceId, sessionSpanId);
      registerActiveSpan(sessionId, 'tool-1', { spanId: 'span-1', startTime: Date.now(), traceId, parentSpanId: sessionSpanId });

      // Verify state exists
      expect(getSessionInfo(sessionId)).not.toBeNull();

      // Stop event cleanup
      deleteSpanState(sessionId);

      // Verify state is cleaned up
      expect(getSessionInfo(sessionId)).toBeNull();
      expect(popActiveSpan(sessionId, 'tool-1')).toBeNull();
    });
  });
});
