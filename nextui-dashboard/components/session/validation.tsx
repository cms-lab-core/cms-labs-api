import { useEffect, useMemo, useState } from 'react';
import {
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  Circle,
  LoaderCircle,
  PanelRightClose,
  PanelRightOpen,
  Play,
  RefreshCw,
  XCircle
} from 'lucide-react';
import { Button } from '@heroui/react';
import type { LabCopy } from './copy';
import { formatDuration, type LabSession } from './types';

interface ValidationPanelProps {
  copy: LabCopy;
  session?: LabSession;
  collapsed: boolean;
  checking: boolean;
  error?: string;
  onToggle: () => void;
  onRun: () => void;
  onRefresh: () => void;
  onShowTopology: (node?: string) => void;
}

type CheckStatus = 'passed' | 'failed' | 'running' | 'pending' | 'warning';

interface ValidationItemData {
  id: string;
  title: string;
  description?: string;
  status: CheckStatus;
  output: Array<{ node?: string; namespace?: string; message?: string }>;
  rawOutput?: string;
}

const statusStyle: Record<CheckStatus, string> = {
  passed: 'border-emerald-500/25 bg-emerald-500/[0.06]',
  failed: 'border-red-500/30 bg-red-500/[0.06]',
  running: 'border-blue-500/30 bg-blue-500/[0.06]',
  pending: 'border-slate-700 bg-slate-800/30',
  warning: 'border-amber-500/30 bg-amber-500/[0.06]'
};

const StatusIcon = ({ status }: { status: CheckStatus }) => {
  if (status === 'passed') return <CheckCircle2 className='h-4 w-4 text-emerald-400' />;
  if (status === 'failed') return <XCircle className='h-4 w-4 text-red-400' />;
  if (status === 'running') return <LoaderCircle className='h-4 w-4 animate-spin text-blue-400' />;
  if (status === 'warning') return <AlertTriangle className='h-4 w-4 text-amber-400' />;
  return <Circle className='h-4 w-4 text-slate-500' />;
};

export const ValidationPanel = ({
  copy,
  session,
  collapsed,
  checking,
  error,
  onToggle,
  onRun,
  onRefresh,
  onShowTopology
}: ValidationPanelProps) => {
  const checker = session?.checker;
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const items = useMemo<ValidationItemData[]>(() => {
    if (!checker) return [];
    if (checker.tasks?.length) {
      return checker.tasks.map((task, index) => ({
        id: `${checker.checkId || checker.jobName || 'check'}-${index}`,
        title: task.title || `${copy.checkerJob} ${index + 1}`,
        description: task.description,
        status: task.complete ? 'passed' : 'failed',
        output: task.logs ?? []
      }));
    }
    const status = checker.status as CheckStatus | undefined;
    return [
      {
        id: checker.checkId || checker.jobName || 'checker',
        title: checker.resultDisplay || copy.checkerJob,
        description: checker.error || checker.report || copy.checkerUnavailable,
        status: status === 'passed' || status === 'failed' || status === 'running' ? status : 'pending',
        output: [],
        rawOutput: checker.logs
      }
    ];
  }, [checker, copy]);

  useEffect(() => {
    const failed = items.find((item) => item.status === 'failed');
    if (failed) setExpanded(new Set([failed.id]));
  }, [checker?.checkId]);

  if (collapsed) {
    return (
      <aside className='flex h-full w-12 shrink-0 flex-col items-center border-l border-slate-700/80 bg-[#151f30] py-3'>
        <button
          type='button'
          onClick={onToggle}
          title={copy.expandChecks}
          className='rounded-md p-2 text-slate-400 transition hover:bg-slate-700/60 hover:text-slate-100'
        >
          <PanelRightOpen className='h-4 w-4' />
        </button>
        <div className='mt-4 h-px w-6 bg-slate-700' />
        <span className='mt-4 [writing-mode:vertical-rl] text-xs font-semibold uppercase tracking-[0.16em] text-slate-500'>
          {copy.validations}
        </span>
        {session?.checkerRunning && <LoaderCircle className='mt-4 h-4 w-4 animate-spin text-blue-400' />}
      </aside>
    );
  }

  const passed = items.filter((item) => item.status === 'passed').length;
  const total = items.length;
  const percentage = total > 0 ? Math.round((passed / total) * 100) : 0;
  const canRun = session?.phase === 'ready' && !session?.checkerRunning && !checking;
  const statusLabel = (status: CheckStatus) => {
    if (status === 'passed') return copy.passed;
    if (status === 'failed') return copy.checkFailed;
    if (status === 'running') return copy.checkRunning;
    return copy.pending;
  };

  return (
    <aside className='flex h-full min-h-0 w-full flex-col bg-[#151f30] text-slate-100'>
      <div className='shrink-0 border-b border-slate-700/80 px-4 pb-4 pt-3'>
        <div className='flex items-center justify-between gap-3'>
          <div className='min-w-0'>
            <h2 className='truncate text-sm font-semibold'>{copy.validations}</h2>
            <p className='mt-1 text-xs text-slate-500'>
              {copy.completed} {passed} {copy.of} {total}
            </p>
          </div>
          <button
            type='button'
            onClick={onToggle}
            title={copy.collapseChecks}
            className='rounded-md p-2 text-slate-500 transition hover:bg-slate-700/60 hover:text-slate-200'
          >
            <PanelRightClose className='h-4 w-4' />
          </button>
        </div>
        <div className='mt-3 flex items-center gap-3'>
          <div className='h-1.5 flex-1 overflow-hidden rounded-full bg-slate-700'>
            <div
              className='h-full rounded-full bg-emerald-500 transition-[width] duration-500'
              style={{ width: `${percentage}%` }}
            />
          </div>
          <span className='w-9 text-right font-mono text-xs text-slate-400'>{percentage}%</span>
        </div>
        {checker?.maxScore ? (
          <div className='mt-2 text-right text-[11px] text-slate-500'>
            {copy.score}:{' '}
            <span className='font-mono text-slate-300'>
              {checker.currentScore ?? 0} / {checker.maxScore}
            </span>
          </div>
        ) : null}
      </div>

      <div className='min-h-0 flex-1 overflow-y-auto p-3'>
        {items.length === 0 ? (
          <div className='flex h-full min-h-52 flex-col items-center justify-center px-5 text-center'>
            <Circle className='mb-3 h-8 w-8 text-slate-600' />
            <p className='text-sm font-medium text-slate-300'>{copy.noChecks}</p>
            <p className='mt-2 text-xs leading-5 text-slate-500'>{copy.noChecksDescription}</p>
          </div>
        ) : (
          <div className='space-y-2'>
            {items.map((item) => {
              const isExpanded = expanded.has(item.id);
              const node = item.output.find((line) => line.node)?.node;
              return (
                <div key={item.id} className={`overflow-hidden rounded-md border ${statusStyle[item.status]}`}>
                  <button
                    type='button'
                    onClick={() =>
                      setExpanded((current) => {
                        const next = new Set(current);
                        if (next.has(item.id)) next.delete(item.id);
                        else next.add(item.id);
                        return next;
                      })
                    }
                    className='flex w-full items-start gap-3 px-3 py-3 text-left'
                  >
                    <span className='mt-0.5 shrink-0'>
                      <StatusIcon status={item.status} />
                    </span>
                    <span className='min-w-0 flex-1'>
                      <span className='block text-sm font-medium text-slate-200'>{item.title}</span>
                      {item.description && (
                        <span className='mt-1 line-clamp-2 block text-xs leading-5 text-slate-500'>
                          {item.description}
                        </span>
                      )}
                    </span>
                    <span className='flex shrink-0 items-center gap-1.5 text-[10px] font-semibold uppercase tracking-wider text-slate-500'>
                      {statusLabel(item.status)}
                      <ChevronDown className={`h-3.5 w-3.5 transition ${isExpanded ? 'rotate-180' : ''}`} />
                    </span>
                  </button>
                  {isExpanded && (
                    <div className='border-t border-slate-700/70 px-3 py-3'>
                      {item.output.length > 0 && (
                        <div className='space-y-2'>
                          {item.output.map((line, index) => (
                            <div
                              key={`${item.id}-line-${index}`}
                              className='rounded bg-slate-950/70 px-3 py-2 font-mono text-[11px] leading-5 text-slate-300'
                            >
                              {(line.node || line.namespace) && (
                                <div className='mb-1 text-emerald-400'>
                                  {[line.node, line.namespace].filter(Boolean).join(' · ')}
                                </div>
                              )}
                              {line.message || '—'}
                            </div>
                          ))}
                        </div>
                      )}
                      {item.rawOutput && (
                        <pre className='max-h-52 overflow-auto whitespace-pre-wrap rounded bg-slate-950/70 p-3 font-mono text-[11px] leading-5 text-slate-300'>
                          {item.rawOutput}
                        </pre>
                      )}
                      {node && (
                        <button
                          type='button'
                          onClick={() => onShowTopology(node)}
                          className='mt-3 text-xs font-medium text-emerald-400 transition hover:text-emerald-300'
                        >
                          {copy.showOnTopology} →
                        </button>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className='shrink-0 border-t border-slate-700/80 bg-[#151f30] p-3'>
        {error && (
          <p className='mb-2 rounded border border-red-500/20 bg-red-500/10 px-3 py-2 text-xs text-red-300'>{error}</p>
        )}
        <div className='flex gap-2'>
          <Button
            color='success'
            radius='sm'
            isDisabled={!canRun}
            onPress={onRun}
            className='h-10 flex-1 font-semibold'
            startContent={
              checking || session?.checkerRunning ? (
                <LoaderCircle className='h-4 w-4 animate-spin' />
              ) : (
                <Play className='h-4 w-4 fill-current' />
              )
            }
          >
            {checking || session?.checkerRunning ? copy.checking : copy.runCheck}
          </Button>
          <Button
            isIconOnly
            radius='sm'
            variant='bordered'
            onPress={onRefresh}
            className='h-10 text-slate-400'
            title={copy.refresh}
          >
            <RefreshCw className='h-4 w-4' />
          </Button>
        </div>
        <p className='mt-2 text-center text-[11px] text-slate-600'>
          {copy.lastCheck}: {checker?.completedAt ? new Date(checker.completedAt).toLocaleTimeString() : '—'}
          {checker?.startedAt ? ` · ${copy.duration} ${formatDuration(checker.startedAt, checker.completedAt)}` : ''}
        </p>
      </div>
    </aside>
  );
};
