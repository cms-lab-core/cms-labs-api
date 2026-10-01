import { useCallback, useEffect, useMemo, useRef } from 'react';
import { Button } from '@heroui/react';
import { AlertCircle, LoaderCircle, NotebookTabs } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import { labCopy } from '@/components/session/copy';
import useLanguageBrowser from '@/helpers/locale';
import { useMutationSessionOpen } from '@/helpers/queries/session/use-mutation-session-open';

const errorMessage = (error: unknown): string => {
  if (typeof error === 'object' && error && 'data' in error) {
    const data = (error as { data?: { message?: string } }).data;
    if (data?.message) return data.message;
  }
  return error instanceof Error ? error.message : '';
};

export default function SessionWorkspaceRedirectPage() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const { lang } = useLanguageBrowser();
  const copy = labCopy(lang);
  const open = useMutationSessionOpen();
  const started = useRef(false);

  const requestAccess = useCallback(() => {
    if (!sessionId || open.isPending) return;
    started.current = true;
    open.mutate(
      { sessionId },
      {
        onSuccess: ({ url }) => {
          if (!url) return;
          window.location.replace(url);
        }
      }
    );
  }, [open, sessionId]);

  useEffect(() => {
    if (started.current) return;
    requestAccess();
  }, [requestAccess]);

  const message = useMemo(
    () => (open.isError ? errorMessage(open.error) || copy.workspaceRedirectError : undefined),
    [copy.workspaceRedirectError, open.error, open.isError]
  );

  return (
    <main className='flex min-h-screen w-full items-center justify-center bg-[#111827] px-6 text-slate-100'>
      <section className='w-full max-w-md rounded-lg border border-slate-700 bg-[#182235] p-7 text-center shadow-xl'>
        {message ? (
          <AlertCircle className='mx-auto mb-4 h-10 w-10 text-red-400' />
        ) : (
          <div className='relative mx-auto mb-4 w-fit'>
            <NotebookTabs className='h-11 w-11 text-orange-300' />
            <LoaderCircle className='absolute -bottom-2 -right-3 h-5 w-5 animate-spin text-emerald-400' />
          </div>
        )}
        <h1 className='text-lg font-semibold'>{message ? copy.workspaceRedirectError : copy.workspaceRedirecting}</h1>
        <p className={`mt-2 text-sm ${message ? 'text-red-300' : 'text-slate-400'}`}>
          {message || copy.workspaceRedirectingDescription}
        </p>
        {message && (
          <div className='mt-6 flex justify-center gap-3'>
            <Button
              variant='flat'
              onPress={() => navigate(sessionId ? RoutesLocation.session(sessionId) : RoutesLocation.home())}
            >
              {copy.backHome}
            </Button>
            <Button
              color='primary'
              onPress={() => {
                open.reset();
                started.current = false;
                requestAccess();
              }}
            >
              {copy.retry}
            </Button>
          </div>
        )}
      </section>
    </main>
  );
}
