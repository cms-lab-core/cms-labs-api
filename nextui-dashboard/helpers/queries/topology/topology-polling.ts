import type { CamelCasedPropertiesDeep } from 'type-fest';
import type { UsecasesTopologiesGetOutputDTO } from '@/helpers/api';

type Response = CamelCasedPropertiesDeep<UsecasesTopologiesGetOutputDTO>;
type PollingQuery = {
  state: {
    data?: Response;
    status: string;
    dataUpdateCount: number;
    errorUpdateCount: number;
  };
};

const normalInterval = 30_000;
const readyInterval = (4 * 60 + 20) * 1000;

export const createTopologyRefetchInterval = (random = Math.random) => {
  // All observers of the same query share a stable timer for each response.
  // Sampling on every render would continuously postpone the next request.
  const intervals = new WeakMap<PollingQuery, { version: string; interval: number }>();

  return (query: PollingQuery): number => {
    const { data, status, dataUpdateCount, errorUpdateCount } = query.state;
    const ready =
      status === 'success' &&
      !!data?.topology?.trim() &&
      !!data.deployments?.length &&
      data.deployments.every((deployment) => deployment.status === 'Ready') &&
      !!data.ttyd?.length &&
      data.ttyd.every((terminal) => !!terminal.url?.trim());
    const version = `${status}:${dataUpdateCount}:${errorUpdateCount}:${ready}`;
    const cached = intervals.get(query);
    if (cached?.version === version) return cached.interval;

    const base = ready ? readyInterval : normalInterval;
    const interval = Math.round(base * (0.9 + random() * 0.2));
    intervals.set(query, { version, interval });
    return interval;
  };
};
