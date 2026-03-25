import type { ChatSession } from '@/api/aiChat';

export function normalizeChatSessions(value: unknown): ChatSession[] {
  return Array.isArray(value) ? value : [];
}
