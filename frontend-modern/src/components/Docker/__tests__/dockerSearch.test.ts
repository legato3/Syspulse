import { describe, expect, it } from 'vitest';

import type { DockerContainer, DockerHost } from '@/types/api';
import {
  containerMatchesDockerSearch,
  getDockerHostDisplayName,
  parseDockerSearchTerm,
} from '@/components/Docker/dockerSearch';

const makeHost = (overrides: Partial<DockerHost> = {}): DockerHost =>
  ({
    id: 'host-1',
    hostname: 'prod-host',
    displayName: 'Production Host',
    containers: [],
    services: [],
    tasks: [],
    ...overrides,
  }) as DockerHost;

const makeContainer = (overrides: Partial<DockerContainer> = {}): DockerContainer =>
  ({
    id: 'abc123',
    name: 'postgres-db',
    image: 'postgres:16',
    state: 'running',
    status: 'Up 1 hour',
    health: 'healthy',
    updateStatus: { updateAvailable: true },
    ...overrides,
  }) as DockerContainer;

describe('dockerSearch', () => {
  it('parses keyed and free-text search tokens', () => {
    expect(parseDockerSearchTerm('image:postgres host:prod running')).toEqual([
      { key: 'image', value: 'postgres' },
      { key: 'host', value: 'prod' },
      { value: 'running' },
    ]);
  });

  it('matches updateable containers by keyed filters', () => {
    const host = makeHost();
    const container = makeContainer();

    expect(containerMatchesDockerSearch('image:postgres', host, container)).toBe(true);
    expect(containerMatchesDockerSearch('host:prod', host, container)).toBe(true);
    expect(containerMatchesDockerSearch('state:running', host, container)).toBe(true);
    expect(containerMatchesDockerSearch('has:update', host, container)).toBe(true);
    expect(containerMatchesDockerSearch('image:redis', host, container)).toBe(false);
  });

  it('uses the same host display name fallback as the docker table', () => {
    expect(getDockerHostDisplayName(makeHost({ customDisplayName: 'Friendly Host' }))).toBe('Friendly Host');
    expect(getDockerHostDisplayName(makeHost({ customDisplayName: '', displayName: 'Display Host' }))).toBe('Display Host');
    expect(getDockerHostDisplayName(makeHost({ customDisplayName: '', displayName: '', hostname: 'raw-host' }))).toBe('raw-host');
  });
});
