import { describe, expect, it } from 'vitest';

import { normalizeChatSessions } from '@/components/Settings/aiSettingsChatSessions';

describe('normalizeChatSessions', () => {
  it('returns the original array when the API payload is valid', () => {
    const sessions = [
      { id: 'session-1', title: 'First', message_count: 1, updated_at: '2026-03-25T10:00:00Z' },
    ];

    expect(normalizeChatSessions(sessions)).toEqual(sessions);
  });

  it('returns an empty array when the API payload is null', () => {
    expect(normalizeChatSessions(null)).toEqual([]);
  });

  it('returns an empty array when the API payload is not an array', () => {
    expect(normalizeChatSessions({ id: 'session-1' })).toEqual([]);
  });
});
