import type { DockerContainer, DockerHost, DockerService } from '@/types/api';

type SearchToken = { key?: string; value: string };

const toLower = (value?: string | null) => value?.toLowerCase() ?? '';

export const getDockerHostDisplayName = (host: DockerHost): string =>
  host.customDisplayName || host.displayName || host.hostname || host.id || '';

export const parseDockerSearchTerm = (term?: string): SearchToken[] => {
  if (!term) return [];
  return term
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .map((token) => {
      const [rawKey, ...rest] = token.split(':');
      if (rest.length === 0) {
        return { value: token.toLowerCase() };
      }
      return { key: rawKey.toLowerCase(), value: rest.join(':').toLowerCase() };
    });
};

const containerMatchesSearchToken = (
  token: SearchToken,
  host: DockerHost,
  container: DockerContainer,
) => {
  const state = toLower(container.state);
  const health = toLower(container.health);
  const hostName = toLower(getDockerHostDisplayName(host));

  if (token.key === 'name') {
    return (
      toLower(container.name).includes(token.value) ||
      toLower(container.id).includes(token.value)
    );
  }

  if (token.key === 'image') {
    return toLower(container.image).includes(token.value);
  }

  if (token.key === 'host') {
    return hostName.includes(token.value);
  }

  if (token.key === 'pod') {
    const pod = container.podman?.podName?.toLowerCase() ?? '';
    return pod.includes(token.value);
  }

  if (token.key === 'compose') {
    const project = container.podman?.composeProject?.toLowerCase() ?? '';
    const service = container.podman?.composeService?.toLowerCase() ?? '';
    return project.includes(token.value) || service.includes(token.value);
  }

  if (token.key === 'state') {
    return state.includes(token.value) || health.includes(token.value);
  }

  if (token.key === 'has' && token.value === 'update') {
    return container.updateStatus?.updateAvailable === true;
  }

  const fields: string[] = [
    container.name,
    container.id,
    container.image,
    container.status,
    container.state,
    container.health,
    host.displayName,
    host.hostname,
    host.id,
  ]
    .filter(Boolean)
    .map((value) => value!.toLowerCase());

  if (container.podman) {
    [
      container.podman.podName,
      container.podman.podId,
      container.podman.composeProject,
      container.podman.composeService,
      container.podman.autoUpdatePolicy,
      container.podman.userNamespace,
    ]
      .filter(Boolean)
      .forEach((value) => fields.push(value!.toLowerCase()));
  }

  if (container.labels) {
    Object.entries(container.labels).forEach(([key, value]) => {
      fields.push(key.toLowerCase());
      if (value) fields.push(value.toLowerCase());
    });
  }

  if (container.ports) {
    container.ports.forEach((port) => {
      const parts = [port.privatePort, port.publicPort, port.protocol, port.ip]
        .filter(Boolean)
        .map(String)
        .join(':')
        .toLowerCase();
      if (parts) fields.push(parts);
    });
  }

  return fields.some((field) => field.includes(token.value));
};

const serviceMatchesSearchToken = (token: SearchToken, host: DockerHost, service: DockerService) => {
  const hostName = toLower(getDockerHostDisplayName(host));
  const serviceName = toLower(service.name ?? service.id);
  const image = toLower(service.image);

  if (token.key === 'name') {
    return serviceName.includes(token.value);
  }

  if (token.key === 'image') {
    return image.includes(token.value);
  }

  if (token.key === 'host') {
    return hostName.includes(token.value);
  }

  if (token.key === 'state') {
    const desired = service.desiredTasks ?? 0;
    const running = service.runningTasks ?? 0;
    const status = desired > 0 && running >= desired ? 'healthy' : 'degraded';
    return status.includes(token.value);
  }

  const fields: string[] = [
    service.name,
    service.id,
    service.image,
    service.stack,
    service.mode,
    host.displayName,
    host.hostname,
    host.id,
  ]
    .filter(Boolean)
    .map((value) => value!.toLowerCase());

  if (service.labels) {
    Object.entries(service.labels).forEach(([key, value]) => {
      fields.push(key.toLowerCase());
      if (value) fields.push(value.toLowerCase());
    });
  }

  return fields.some((field) => field.includes(token.value));
};

export const containerMatchesDockerSearch = (
  term: string | undefined,
  host: DockerHost,
  container: DockerContainer,
) => {
  const tokens = parseDockerSearchTerm(term);
  return tokens.every((token) => containerMatchesSearchToken(token, host, container));
};

export const serviceMatchesDockerSearch = (
  term: string | undefined,
  host: DockerHost,
  service: DockerService,
) => {
  const tokens = parseDockerSearchTerm(term);
  return tokens.every((token) => serviceMatchesSearchToken(token, host, service));
};
