import { useEffect, useMemo, useState, type ReactNode } from 'react';
import { addToast } from '@heroui/react';
import copyToClipboard from 'copy-to-clipboard';
import { Check, Circle, Clock3, GitCommitHorizontal, Network, Server, UserRound } from 'lucide-react';
import { parseTopology } from '@/components/topology/objectTypes/parse';
import type { LabCopy } from './copy';
import { formatDuration, phaseTone, type LabSession, type LabTopology } from './types';

interface OverviewProps {
  copy: LabCopy;
  session?: LabSession;
  topology?: LabTopology;
  topologyLoading: boolean;
  onOpenTopology: () => void;
}

const statusClasses = {
  success: 'border-emerald-500/25 bg-emerald-500/10 text-emerald-300',
  danger: 'border-red-500/25 bg-red-500/10 text-red-300',
  warning: 'border-amber-500/25 bg-amber-500/10 text-amber-300',
  info: 'border-blue-500/25 bg-blue-500/10 text-blue-300'
};

interface CopyValueProps {
  copy: LabCopy;
  label: string;
  value: string;
  className?: string;
  children?: ReactNode;
}

const CopyValue = ({ copy, label, value, className = '', children }: CopyValueProps) => (
  <button
    type='button'
    disabled={!value || value === '—'}
    title={`${copy.copyValueHint}: ${label}`}
    onClick={() => {
      if (!copyToClipboard(value)) return;
      addToast({ title: copy.valueCopied, description: `${label}: ${value}`, color: 'success' });
    }}
    className={`cursor-copy bg-transparent p-0 text-inherit focus-visible:outline focus-visible:outline-1 focus-visible:outline-emerald-500 disabled:cursor-default ${className}`}
  >
    {children ?? value}
  </button>
);

export const OverviewPanel = ({ copy, session, topology, topologyLoading, onOpenTopology }: OverviewProps) => {
  const [, setClock] = useState(0);
  useEffect(() => {
    const timer = window.setInterval(() => setClock((value) => value + 1), 1000);
    return () => window.clearInterval(timer);
  }, []);

  const parsed = useMemo(() => {
    if (!topology?.topology) return undefined;
    try {
      return parseTopology(topology);
    } catch {
      return undefined;
    }
  }, [topology]);
  const deployments = topology?.deployments ?? [];
  const deploymentByName = useMemo(
    () => new Map(deployments.map((deployment) => [deployment.name, deployment])),
    [deployments]
  );
  const nodes = parsed?.nodes ?? [];
  const readyNodes = nodes.filter((node) => deploymentByName.get(node.id)?.status === 'Ready').length;
  const tone = phaseTone(session?.phase);
  const phaseLabel =
    tone === 'success'
      ? copy.running
      : tone === 'danger'
        ? copy.failed
        : tone === 'warning'
          ? session?.phase === 'stopping'
            ? copy.stopping
            : copy.degraded
          : copy.preparing;

  const cards = [
    { label: copy.standState, value: phaseLabel, icon: Network, tone },
    { label: copy.nodes, value: `${readyNodes} / ${nodes.length} ${copy.active}`, icon: Server, tone: 'info' as const },
    { label: copy.user, value: session?.username || session?.ownerId || '—', icon: UserRound, tone: 'info' as const },
    { label: copy.uptime, value: formatDuration(session?.createdAt), icon: Clock3, tone: 'info' as const }
  ];

  const progress = [
    { label: copy.sessionCreated, complete: !!session?.createdAt, value: session?.createdAt },
    { label: copy.desiredAccepted, complete: !!session?.desiredResourcesAccepted },
    { label: copy.topologyReady, complete: !!session?.topologyReady },
    { label: copy.workspaceReady, complete: !!session?.workspaceReady }
  ];

  return (
    <div className='h-full overflow-y-auto bg-[#111827] p-5 text-slate-100 lg:p-7'>
      <div className='mx-auto flex w-full max-w-[1280px] flex-col gap-6'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-emerald-400'>{copy.labStand}</p>
          <h1 className='mt-1 text-2xl font-semibold tracking-tight'>{session?.title || copy.labStand}</h1>
          <p className='mt-1 text-sm text-slate-400'>{copy.currentEnvironment}</p>
        </div>

        <div className='grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4'>
          {cards.map((card) => {
            const Icon = card.icon;
            return (
              <div key={card.label} className='rounded-lg border border-slate-700/80 bg-[#1b2638] p-4'>
                <div className='flex items-start justify-between gap-4'>
                  <div>
                    <p className='text-xs font-medium text-slate-400'>{card.label}</p>
                    <p className='mt-2 truncate text-base font-semibold text-slate-100'>{card.value}</p>
                  </div>
                  <div className={`rounded-md border p-2 ${statusClasses[card.tone]}`}>
                    <Icon className='h-4 w-4' />
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        <div className='grid gap-5 xl:grid-cols-[minmax(0,1.45fr)_minmax(280px,0.55fr)]'>
          <section className='overflow-hidden rounded-lg border border-slate-700/80 bg-[#182235]'>
            <div className='flex items-center justify-between border-b border-slate-700/80 px-4 py-3'>
              <div>
                <h2 className='text-sm font-semibold'>{copy.components}</h2>
                <p className='mt-0.5 text-xs text-slate-500'>{session?.namespace || '—'}</p>
              </div>
              <button
                type='button'
                onClick={onOpenTopology}
                className='rounded-md border border-slate-600 px-3 py-1.5 text-xs font-medium text-slate-300 transition hover:border-emerald-500/60 hover:bg-emerald-500/10 hover:text-emerald-300'
              >
                {copy.topology}
              </button>
            </div>
            {topologyLoading ? (
              <div className='space-y-2 p-4'>
                {[0, 1, 2].map((item) => (
                  <div key={item} className='h-12 animate-pulse rounded-md bg-slate-700/35' />
                ))}
              </div>
            ) : nodes.length === 0 ? (
              <div className='flex min-h-44 flex-col items-center justify-center px-6 text-center'>
                <Network className='mb-3 h-8 w-8 text-slate-600' />
                <p className='text-sm font-medium text-slate-300'>{copy.noTopology}</p>
                <p className='mt-1 text-xs text-slate-500'>{copy.noComponents}</p>
              </div>
            ) : (
              <div className='divide-y divide-slate-700/70'>
                <div className='hidden grid-cols-[minmax(140px,1fr)_120px_minmax(130px,0.8fr)_90px_70px] gap-3 px-4 py-2 text-[11px] font-semibold uppercase tracking-wider text-slate-500 md:grid'>
                  <span>{copy.component}</span>
                  <span>{copy.kind}</span>
                  <span>{copy.address}</span>
                  <span>{copy.state}</span>
                  <span className='text-right'>{copy.restarts}</span>
                </div>
                {nodes.map((node) => {
                  const data = node.data?.data;
                  const deployment = deploymentByName.get(node.id);
                  const ready = deployment?.status === 'Ready';
                  const address = data?.serviceExternalIp?.[0] || data?.serviceClusterIp || '—';
                  const kind = data?.kind || 'node';
                  const state = ready ? copy.ready : copy.notReady;
                  const restarts = deployment?.restarts ?? 0;
                  return (
                    <div
                      key={node.id}
                      className='grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 text-left md:grid-cols-[minmax(140px,1fr)_120px_minmax(130px,0.8fr)_90px_70px]'
                    >
                      <div className='min-w-0'>
                        <CopyValue
                          copy={copy}
                          label={copy.component}
                          value={node.id}
                          className='block max-w-full truncate text-left font-mono text-sm font-medium text-slate-200'
                        >
                          {node.id}
                        </CopyValue>
                        <div className='mt-1 flex min-w-0 items-center gap-2 md:hidden'>
                          <CopyValue
                            copy={copy}
                            label={copy.kind}
                            value={kind}
                            className='truncate text-xs text-slate-500'
                          >
                            {kind}
                          </CopyValue>
                          <CopyValue
                            copy={copy}
                            label={copy.address}
                            value={address}
                            className='truncate font-mono text-xs text-slate-500'
                          >
                            {address}
                          </CopyValue>
                        </div>
                      </div>
                      <CopyValue
                        copy={copy}
                        label={copy.kind}
                        value={kind}
                        className='hidden truncate text-left text-xs text-slate-400 md:block'
                      >
                        {kind}
                      </CopyValue>
                      <CopyValue
                        copy={copy}
                        label={copy.address}
                        value={address}
                        className='hidden truncate text-left font-mono text-xs text-slate-400 md:block'
                      >
                        {address}
                      </CopyValue>
                      <CopyValue
                        copy={copy}
                        label={copy.state}
                        value={state}
                        className={`inline-flex items-center gap-1.5 text-xs ${ready ? 'text-emerald-300' : 'text-amber-300'}`}
                      >
                        <span className={`h-1.5 w-1.5 rounded-full ${ready ? 'bg-emerald-400' : 'bg-amber-400'}`} />
                        {state}
                      </CopyValue>
                      <CopyValue
                        copy={copy}
                        label={copy.restarts}
                        value={String(restarts)}
                        className='hidden text-right font-mono text-xs text-slate-400 md:block'
                      >
                        {restarts}
                      </CopyValue>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          <div className='flex flex-col gap-5'>
            <section className='rounded-lg border border-slate-700/80 bg-[#182235] p-4'>
              <h2 className='text-sm font-semibold'>{copy.environment}</h2>
              <dl className='mt-4 space-y-3 text-xs'>
                {[
                  [copy.namespace, session?.namespace],
                  [copy.labPath, session?.labPath],
                  [copy.testPath, session?.testPath],
                  [copy.revision, session?.taskRevision]
                ].map(([label, value]) => (
                  <div key={label} className='flex items-start justify-between gap-4'>
                    <dt className='shrink-0 text-slate-500'>{label}</dt>
                    <dd className='min-w-0 break-all text-right font-mono text-slate-300'>{value || '—'}</dd>
                  </div>
                ))}
              </dl>
            </section>

            <section className='rounded-lg border border-slate-700/80 bg-[#182235] p-4'>
              <h2 className='text-sm font-semibold'>{copy.events}</h2>
              <ol className='mt-4 space-y-4'>
                {progress.map((event, index) => (
                  <li key={event.label} className='relative flex gap-3'>
                    {index < progress.length - 1 && (
                      <span className='absolute left-[7px] top-4 h-[calc(100%+4px)] w-px bg-slate-700' />
                    )}
                    {event.complete ? (
                      <span className='relative z-10 flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-emerald-500 text-[#111827]'>
                        <Check className='h-2.5 w-2.5 stroke-[3]' />
                      </span>
                    ) : (
                      <Circle className='relative z-10 h-4 w-4 shrink-0 fill-[#182235] text-slate-600' />
                    )}
                    <div className='min-w-0 -translate-y-0.5'>
                      <p className={`text-xs ${event.complete ? 'text-slate-300' : 'text-slate-500'}`}>{event.label}</p>
                      <p className='mt-0.5 font-mono text-[11px] text-slate-600'>
                        {event.value
                          ? new Date(event.value).toLocaleString()
                          : event.complete
                            ? copy.ready
                            : copy.waiting}
                      </p>
                    </div>
                  </li>
                ))}
              </ol>
            </section>

            {session?.taskRevision && (
              <div className='flex items-center gap-2 px-1 text-xs text-slate-600'>
                <GitCommitHorizontal className='h-3.5 w-3.5' />
                <span className='truncate font-mono'>{session.taskRevision}</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
