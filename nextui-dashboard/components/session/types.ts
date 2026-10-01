import type { QueriesSessionRecord, UsecasesTopologiesGetOutputDTO } from '@/helpers/api';
import type { CamelCasedPropertiesDeep } from 'type-fest';

export type LabSession = CamelCasedPropertiesDeep<QueriesSessionRecord>;
export type LabTopology = CamelCasedPropertiesDeep<UsecasesTopologiesGetOutputDTO>;
export type WorkspaceTab = 'overview' | 'topology' | 'assignment';

export const isWorkspaceTab = (value: string | null): value is WorkspaceTab =>
  value === 'overview' || value === 'topology' || value === 'assignment';

export const formatDuration = (startedAt?: string, endedAt?: string): string => {
  if (!startedAt) return '—';
  const start = new Date(startedAt).getTime();
  const end = endedAt ? new Date(endedAt).getTime() : Date.now();
  if (!Number.isFinite(start) || !Number.isFinite(end)) return '—';
  const seconds = Math.max(0, Math.floor((end - start) / 1000));
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainder = seconds % 60;
  return [hours, minutes, remainder].map((part) => part.toString().padStart(2, '0')).join(':');
};

export const phaseTone = (phase?: string): 'success' | 'danger' | 'warning' | 'info' => {
  if (phase === 'ready' || phase === 'active') return 'success';
  if (phase === 'failed') return 'danger';
  if (phase === 'degraded' || phase === 'stopping') return 'warning';
  return 'info';
};
