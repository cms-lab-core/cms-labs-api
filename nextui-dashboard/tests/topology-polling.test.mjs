import assert from 'node:assert/strict';
import test from 'node:test';
import { focusManager, QueryClient, QueryObserver } from '@tanstack/query-core';
import { createTopologyRefetchInterval } from '../helpers/queries/topology/topology-polling.ts';

const readyData = () => ({
  topology: 'topology:\n  nodes:\n    r1: {}',
  deployments: [{ name: 'r1', status: 'Ready' }],
  ttyd: [{ name: 'r1-terminal', url: '/clabgate/workspace/session/terminal/r1-terminal/' }]
});
const query = (data = readyData(), status = 'success') => ({
  state: { data, status, dataUpdateCount: 1, errorUpdateCount: 0 }
});

test('ready topology polls at 260 seconds with symmetric 10% jitter', () => {
  assert.equal(createTopologyRefetchInterval(() => 0)(query()), 234_000);
  assert.equal(createTopologyRefetchInterval(() => 0.5)(query()), 260_000);
  assert.equal(createTopologyRefetchInterval(() => 1)(query()), 286_000);
});

test('pending, empty, incomplete and failed responses poll at 30 seconds', () => {
  const cases = [
    { state: { ...query({}, 'pending').state, data: undefined } },
    query({}, 'success'),
    query({ ...readyData(), topology: '  ' }),
    query({ ...readyData(), deployments: [] }),
    query({ ...readyData(), deployments: [{ status: 'Ready' }, { status: 'Pending' }] }),
    query({ ...readyData(), ttyd: [] }),
    query({ ...readyData(), ttyd: [{ url: '' }] }),
    query(readyData(), 'error')
  ];
  const interval = createTopologyRefetchInterval(() => 0.5);
  for (const candidate of cases) assert.equal(interval(candidate), 30_000);
  assert.equal(createTopologyRefetchInterval(() => 0)(query({}, 'pending')), 27_000);
  assert.equal(createTopologyRefetchInterval(() => 1)(query({}, 'error')), 33_000);
});

test('renders and multiple observers do not resample the same query response', () => {
  let samples = 0;
  const interval = createTopologyRefetchInterval(() => ++samples / 10);
  const candidate = query();
  const first = interval(candidate);
  for (let observer = 0; observer < 10; observer++) assert.equal(interval(candidate), first);
  assert.equal(samples, 1);
  candidate.state.dataUpdateCount++;
  assert.notEqual(interval(candidate), first);
  assert.equal(samples, 2);
});

test('sessions receive independent jitter', () => {
  let samples = 0;
  const interval = createTopologyRefetchInterval(() => ++samples / 10);
  assert.notEqual(interval(query()), interval(query()));
});

test('readiness and errors switch interval immediately despite cached ready data', () => {
  const interval = createTopologyRefetchInterval(() => 0.5);
  const candidate = query({}, 'pending');
  assert.equal(interval(candidate), 30_000);
  candidate.state.data = readyData();
  candidate.state.status = 'success';
  candidate.state.dataUpdateCount++;
  assert.equal(interval(candidate), 260_000);
  candidate.state.status = 'error';
  candidate.state.errorUpdateCount++;
  assert.equal(interval(candidate), 30_000);
});

test('cached topology is not fetched on mount or window focus', async () => {
  const client = new QueryClient();
  const key = ['topology.get', 'cached-session'];
  client.setQueryData(key, readyData());
  client.mount();
  let requests = 0;
  const observer = new QueryObserver(client, {
    queryKey: key,
    queryFn: async () => {
      requests++;
      return readyData();
    },
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchInterval: createTopologyRefetchInterval(() => 0.5)
  });
  const unsubscribe = observer.subscribe(() => {});
  try {
    focusManager.setFocused(false);
    focusManager.setFocused(true);
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.equal(requests, 0);
  } finally {
    unsubscribe();
    client.unmount();
    client.clear();
    focusManager.setFocused(undefined);
  }
});

test('a new session still fetches initial data and retries polling faster after errors', async () => {
  const client = new QueryClient();
  const interval = createTopologyRefetchInterval(() => 0.5);
  const key = ['topology.get', 'new-session'];
  let requests = 0;
  const observer = new QueryObserver(client, {
    queryKey: key,
    queryFn: async () => {
      if (++requests > 1) throw new Error('unavailable');
      return readyData();
    },
    retry: false,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchInterval: interval
  });
  const unsubscribe = observer.subscribe(() => {});
  try {
    await new Promise((resolve) => setTimeout(resolve, 0));
    assert.equal(requests, 1);
    const currentQuery = client.getQueryCache().find({ queryKey: key });
    assert.equal(interval(currentQuery), 260_000);
    await observer.refetch();
    assert.equal(interval(currentQuery), 30_000);
  } finally {
    unsubscribe();
    client.clear();
  }
});
