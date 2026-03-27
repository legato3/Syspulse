import { cleanup, render, screen, waitFor } from '@solidjs/testing-library';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DiscoveryTab } from './DiscoveryTab';

const getDiscoveryMock = vi.fn();
const getDiscoveryInfoMock = vi.fn();
const getConnectedAgentsMock = vi.fn();
const getSettingsMock = vi.fn();

vi.mock('../../api/discovery', () => ({
  getDiscovery: (...args: unknown[]) => getDiscoveryMock(...args),
  getDiscoveryInfo: (...args: unknown[]) => getDiscoveryInfoMock(...args),
  triggerDiscovery: vi.fn(),
  updateDiscoveryNotes: vi.fn(),
  formatDiscoveryAge: vi.fn(() => 'just now'),
  getCategoryDisplayName: vi.fn((value: string) => value),
  getConfidenceLevel: vi.fn(() => ({ label: 'High confidence', color: 'text-green-600' })),
  getConnectedAgents: (...args: unknown[]) => getConnectedAgentsMock(...args),
}));

vi.mock('@/api/ai', () => ({
  AIAPI: {
    getSettings: (...args: unknown[]) => getSettingsMock(...args),
  },
}));

vi.mock('../../api/guestMetadata', () => ({
  GuestMetadataAPI: {
    updateMetadata: vi.fn(),
  },
}));

describe('DiscoveryTab', () => {
  beforeEach(() => {
    getDiscoveryMock.mockReset();
    getDiscoveryInfoMock.mockReset();
    getConnectedAgentsMock.mockReset();
    getSettingsMock.mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it('does not fetch discovery data when AI discovery is disabled', async () => {
    getSettingsMock.mockResolvedValue({ discovery_enabled: false });

    render(() => (
      <DiscoveryTab
        resourceType="vm"
        hostId="node1"
        resourceId="100"
        hostname="vm100"
        guestId="pve1:node1:100"
      />
    ));

    await waitFor(() => {
      expect(screen.getByText('AI Discovery Disabled')).toBeInTheDocument();
    });

    expect(getDiscoveryMock).not.toHaveBeenCalled();
    expect(getDiscoveryInfoMock).not.toHaveBeenCalled();
    expect(getConnectedAgentsMock).not.toHaveBeenCalled();
  });
});
