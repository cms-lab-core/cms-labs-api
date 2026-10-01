import { useCallback, useEffect, useMemo, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react';
import { addToast, Button } from '@heroui/react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { useConfirmPopup } from '@/components/hooks/useDeletePopup';
import useThemeBrowser from '@/components/navbar/useTheme';
import { ErrorModal } from '@/components/pages/auth/error';
import { RoutesLocation } from '@/components/routes';
import { useUserProfile } from '@/components/providers/auth-jwt/hooks';
import { parseTopology } from '@/components/topology/objectTypes/parse';
import { WorkspaceHeader } from '@/components/session/header';
import { JupyterWorkspace } from '@/components/session/jupyter';
import { OverviewPanel } from '@/components/session/overview';
import { StopSessionModal } from '@/components/session/stop-session-modal';
import { SessionTopology } from '@/components/session/topology';
import { WorkspaceTour } from '@/components/session/tour';
import { ValidationPanel } from '@/components/session/validation';
import { labCopy } from '@/components/session/copy';
import { isWorkspaceTab, type WorkspaceTab } from '@/components/session/types';
import useLanguageBrowser from '@/helpers/locale';
import { useMutationSessionEnsure } from '@/helpers/queries/session/use-mutation-session-ensure';
import { useQuerySessionGet } from '@/helpers/queries/session/use-query-session-get';
import { useMutationSessionStop } from '@/helpers/queries/session/use-mutation-session-stop';
import { useMutationSessionCheck } from '@/helpers/queries/session/use-mutation-session-check';
import { useMutationSessionOpen } from '@/helpers/queries/session/use-mutation-session-open';
import { useMutationNodeAction } from '@/helpers/queries/node/use-mutation-node-action';
import { useQueryTopologyGet } from '@/helpers/queries/topology/use-query-topology-get';

const errorMessage = (error: unknown): string => {
  if (typeof error === 'object' && error && 'data' in error) {
    const data = (error as { data?: { message?: string } }).data;
    if (data?.message) return data.message;
  }
  return error instanceof Error ? error.message : '';
};

const checksWidthKey = 'cms-labs-checks-width';
const checksCollapsedKey = 'cms-labs-checks-collapsed';
const tourSeenKey = 'cms-labs-workspace-tour-v1';

export default function SessionPage() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { lang } = useLanguageBrowser();
  const { theme, setTheme } = useThemeBrowser();
  const copy = labCopy(lang);
  const user = useUserProfile();
  const ensureStarted = useRef(false);
  const [checkerPolling, setCheckerPolling] = useState(false);
  const [checkerJobName, setCheckerJobName] = useState<string>();
  const [workspaceURL, setWorkspaceURL] = useState<string>();
  const [workspaceVisited, setWorkspaceVisited] = useState(() => searchParams.get('tab') === 'assignment');
  const [topologyVisited, setTopologyVisited] = useState(() => searchParams.get('tab') === 'topology');
  const [highlightedNode, setHighlightedNode] = useState<string>();
  const [stopModalOpen, setStopModalOpen] = useState(false);
  const [tourOpen, setTourOpen] = useState(false);
  const [checksWidth, setChecksWidth] = useState(() => {
    const stored = Number(localStorage.getItem(checksWidthKey));
    return Number.isFinite(stored) && stored >= 320 && stored <= 600 ? stored : 400;
  });
  const [checksCollapsed, setChecksCollapsed] = useState(() => localStorage.getItem(checksCollapsedKey) === 'true');

  const requestedTab = searchParams.get('tab');
  const activeTab: WorkspaceTab = isWorkspaceTab(requestedTab) ? requestedTab : 'overview';
  const ensure = useMutationSessionEnsure();
  const sessionQuery = useQuerySessionGet({ sessionId }, ensure.isSuccess, checkerPolling);
  const stop = useMutationSessionStop();
  const check = useMutationSessionCheck();
  const open = useMutationSessionOpen();
  const session = sessionQuery.data?.session ?? ensure.data?.session;
  const topologyEnabled = !!session?.topologyReady || session?.phase === 'degraded' || session?.phase === 'failed';
  const topologyQuery = useQueryTopologyGet({ sessionId: sessionId || '' }, topologyEnabled);
  const topologyNodeNames = useMemo(() => {
    if (!topologyQuery.data?.topology) return [];
    try {
      return parseTopology(topologyQuery.data).nodes.map((node) => node.id);
    } catch {
      return [];
    }
  }, [topologyQuery.data]);
  const topologyAction = useMutationNodeAction({
    onSuccess: () => {
      addToast({ title: copy.actionSuccess, color: 'success' });
      topologyQuery.refetch();
    },
    onError: (error) => {
      addToast({
        title: copy.actionError,
        description: errorMessage(error),
        color: 'danger'
      });
    }
  });

  const runTopologyAction = useCallback(
    (action: 'restart' | 'wipe') => {
      if (!sessionId || topologyNodeNames.length === 0) return;
      topologyAction.mutate({
        sessionId,
        actions: topologyNodeNames.map((node) => ({ node, action }))
      });
    },
    [sessionId, topologyAction, topologyNodeNames]
  );

  const restartPopup = useConfirmPopup({
    title: copy.restartTopology,
    description: copy.restartTopologyDescription,
    onConfirm: () => runTopologyAction('restart')
  });
  const wipePopup = useConfirmPopup({
    title: copy.wipeTopology,
    description: copy.wipeTopologyDescription,
    onConfirm: () => runTopologyAction('wipe')
  });

  useEffect(() => {
    if (!sessionId || ensureStarted.current) return;
    ensureStarted.current = true;
    ensure.mutate({ attemptId: sessionId });
  }, [ensure, sessionId]);

  useEffect(() => {
    if (!checkerPolling || !checkerJobName) return;
    const checker = session?.checker;
    if (checker?.jobName === checkerJobName && (checker.status === 'passed' || checker.status === 'failed')) {
      setCheckerPolling(false);
    }
  }, [checkerJobName, checkerPolling, session?.checker]);

  useEffect(() => {
    window.document.title = session?.title ? `${session.title} · CMS Labs` : 'CMS Labs';
    return () => {
      window.document.title = 'CMS LABS';
    };
  }, [session?.title]);

  useEffect(() => {
    if (!session || localStorage.getItem(tourSeenKey)) return;
    const timer = window.setTimeout(() => setTourOpen(true), 500);
    return () => window.clearTimeout(timer);
  }, [session?.id]);

  const changeTab = useCallback(
    (tab: WorkspaceTab) => {
      const next = new URLSearchParams(searchParams);
      if (tab === 'overview') next.delete('tab');
      else next.set('tab', tab);
      setSearchParams(next, { replace: true });
      if (tab === 'topology') setTopologyVisited(true);
      if (tab === 'assignment') setWorkspaceVisited(true);
    },
    [searchParams, setSearchParams]
  );

  const requestWorkspace = useCallback(() => {
    if (!sessionId || !session?.workspaceReady || open.isPending) return;
    open.mutate(
      { sessionId },
      {
        onSuccess: (result) => {
          if (result.url) setWorkspaceURL(result.url);
        }
      }
    );
  }, [open, session?.workspaceReady, sessionId]);

  useEffect(() => {
    if (activeTab === 'assignment' && workspaceVisited && session?.workspaceReady && !workspaceURL && !open.isError) {
      requestWorkspace();
    }
  }, [activeTab, open.isError, requestWorkspace, session?.workspaceReady, workspaceURL, workspaceVisited]);

  const refresh = useCallback(() => {
    sessionQuery.refetch();
    if (topologyEnabled) topologyQuery.refetch();
  }, [sessionQuery, topologyEnabled, topologyQuery]);

  const runCheck = useCallback(() => {
    if (!sessionId) return;
    check.mutate(
      { sessionId },
      {
        onSuccess: (result) => {
          setCheckerJobName(result.jobName);
          setCheckerPolling(true);
          sessionQuery.refetch();
        }
      }
    );
  }, [check, sessionId, sessionQuery]);

  const beginResize = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    event.preventDefault();
    const onMove = (moveEvent: PointerEvent) => {
      const width = Math.min(600, Math.max(320, window.innerWidth - moveEvent.clientX));
      setChecksWidth(width);
    };
    const onEnd = () => {
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerup', onEnd);
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      setChecksWidth((width) => {
        localStorage.setItem(checksWidthKey, width.toString());
        return width;
      });
    };
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    window.addEventListener('pointermove', onMove);
    window.addEventListener('pointerup', onEnd);
  }, []);

  const toggleChecks = useCallback(() => {
    setChecksCollapsed((value) => {
      localStorage.setItem(checksCollapsedKey, (!value).toString());
      return !value;
    });
  }, []);

  const workspaceError = useMemo(
    () => (open.isError ? errorMessage(open.error) || copy.workspaceError : undefined),
    [copy.workspaceError, open.error, open.isError]
  );
  const checkerError = useMemo(
    () => (check.isError ? errorMessage(check.error) || copy.checkFailed : undefined),
    [check.error, check.isError, copy.checkFailed]
  );
  const stopError = useMemo(
    () => (stop.isError ? errorMessage(stop.error) || copy.sessionError : undefined),
    [copy.sessionError, stop.error, stop.isError]
  );

  const completeTour = useCallback(() => {
    localStorage.setItem(tourSeenKey, 'true');
    setTourOpen(false);
  }, []);

  if (!sessionId) {
    return <ErrorModal title={copy.sessionError} description={copy.invalidSession} />;
  }

  if (ensure.isError) {
    return (
      <ErrorModal title={copy.sessionError} description={errorMessage(ensure.error) || copy.sessionError}>
        <Button
          color='primary'
          variant='light'
          onPress={() => {
            ensureStarted.current = true;
            ensure.mutate({ attemptId: sessionId });
          }}
        >
          {copy.retry}
        </Button>
      </ErrorModal>
    );
  }

  return (
    <main
      data-theme={theme}
      className='lab-workspace flex h-screen min-h-0 w-screen flex-col overflow-hidden bg-[#111827] text-slate-100'
    >
      <WorkspaceHeader
        copy={copy}
        session={session}
        activeTab={activeTab}
        userName={user.name || user.email}
        refreshing={sessionQuery.isFetching || topologyQuery.isFetching}
        stopping={stop.isPending}
        topologyActionPending={topologyAction.isPending}
        topologyActionsAvailable={topologyNodeNames.length > 0}
        language={lang}
        theme={theme}
        onTabChange={changeTab}
        onRefresh={refresh}
        onHome={() => navigate(RoutesLocation.home())}
        onStop={() => setStopModalOpen(true)}
        onToggleLanguage={() => navigate(RoutesLocation.language())}
        onToggleTheme={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        onRestartTopology={restartPopup.onOpen}
        onWipeTopology={wipePopup.onOpen}
        onStartTour={() => setTourOpen(true)}
      />

      <div className='flex min-h-0 flex-1'>
        <section data-tour='workspace' className='relative min-w-0 flex-1 overflow-hidden'>
          <div className={activeTab === 'overview' ? 'h-full' : 'hidden'}>
            <OverviewPanel
              copy={copy}
              session={session}
              topology={topologyQuery.data}
              topologyLoading={topologyQuery.isLoading || (!session?.topologyReady && !session?.message)}
              onOpenTopology={() => changeTab('topology')}
            />
          </div>
          {topologyVisited && (
            <div className={activeTab === 'topology' ? 'h-full' : 'hidden'}>
              <SessionTopology
                copy={copy}
                ready={!!session?.topologyReady}
                highlightedNode={highlightedNode}
                active={activeTab === 'topology'}
              />
            </div>
          )}
          {workspaceVisited && (
            <div className={activeTab === 'assignment' ? 'h-full' : 'hidden'}>
              <JupyterWorkspace
                copy={copy}
                ready={!!session?.workspaceReady}
                url={workspaceURL}
                loading={open.isPending}
                error={workspaceError}
                onOpen={() => {
                  open.reset();
                  requestWorkspace();
                }}
                onOpenExternal={() => {
                  if (sessionId) {
                    window.open(RoutesLocation.sessionWorkspace(sessionId), '_blank', 'noopener,noreferrer');
                  }
                }}
              />
            </div>
          )}
          {(session?.message || sessionQuery.isError || topologyQuery.isError) && (
            <div className='pointer-events-none absolute bottom-4 left-1/2 z-30 w-[min(620px,calc(100%-32px))] -translate-x-1/2 rounded-md border border-red-500/25 bg-red-950/90 px-4 py-3 text-sm text-red-200 shadow-xl backdrop-blur'>
              {session?.message || errorMessage(sessionQuery.error) || errorMessage(topologyQuery.error)}
            </div>
          )}
        </section>

        {!checksCollapsed && (
          <div
            role='separator'
            aria-orientation='vertical'
            onPointerDown={beginResize}
            className='group relative z-40 w-1 shrink-0 cursor-col-resize bg-slate-700/80 transition hover:bg-emerald-500/70'
          >
            <span className='absolute inset-y-0 -left-1 -right-1' />
          </div>
        )}
        <div
          data-tour='checks'
          className='h-full shrink-0 overflow-hidden'
          style={{ width: checksCollapsed ? 48 : `min(${checksWidth}px, 42vw)` }}
        >
          <ValidationPanel
            copy={copy}
            session={session}
            collapsed={checksCollapsed}
            checking={check.isPending}
            error={checkerError}
            onToggle={toggleChecks}
            onRun={runCheck}
            onRefresh={refresh}
            onShowTopology={(node) => {
              setHighlightedNode(node);
              setTopologyVisited(true);
              changeTab('topology');
            }}
          />
        </div>
      </div>
      <WorkspaceTour copy={copy} open={tourOpen} onComplete={completeTour} />
      <StopSessionModal
        copy={copy}
        open={stopModalOpen}
        stopping={stop.isPending}
        error={stopError}
        onOpenChange={(openState) => {
          if (!stop.isPending) {
            setStopModalOpen(openState);
            if (!openState) stop.reset();
          }
        }}
        onConfirm={() =>
          stop.mutate(
            { sessionId },
            {
              onSuccess: () => navigate(RoutesLocation.home())
            }
          )
        }
      />
      {restartPopup.component({})}
      {wipePopup.component({})}
    </main>
  );
}
