import { ExternalLink, LoaderCircle, NotebookTabs, RefreshCw } from 'lucide-react';
import type { LabCopy } from './copy';

interface JupyterWorkspaceProps {
  copy: LabCopy;
  ready: boolean;
  url?: string;
  loading: boolean;
  error?: string;
  onOpen: () => void;
  onOpenExternal: () => void;
}

export const JupyterWorkspace = ({
  copy,
  ready,
  url,
  loading,
  error,
  onOpen,
  onOpenExternal
}: JupyterWorkspaceProps) => {
  if (!ready) {
    return (
      <div className='flex h-full flex-col items-center justify-center bg-[#111827] px-6 text-center text-slate-200'>
        <div className='relative mb-4'>
          <NotebookTabs className='h-10 w-10 text-slate-500' />
          <LoaderCircle className='absolute -bottom-2 -right-3 h-5 w-5 animate-spin text-emerald-400' />
        </div>
        <h2 className='text-base font-semibold'>{copy.workspaceLoading}</h2>
        <p className='mt-1 max-w-md text-sm text-slate-500'>{copy.workspaceLoadingDescription}</p>
      </div>
    );
  }

  if (!url) {
    return (
      <div className='flex h-full flex-col items-center justify-center bg-[#111827] px-6 text-center text-slate-200'>
        {loading ? (
          <LoaderCircle className='mb-4 h-9 w-9 animate-spin text-emerald-400' />
        ) : (
          <NotebookTabs className='mb-4 h-10 w-10 text-slate-500' />
        )}
        <h2 className='text-base font-semibold'>{error ? copy.workspaceError : copy.workspaceLoading}</h2>
        <p className={`mt-2 max-w-lg text-sm ${error ? 'text-red-300' : 'text-slate-500'}`}>
          {error || copy.workspaceLoadingDescription}
        </p>
        {!loading && (
          <button
            type='button'
            onClick={onOpen}
            className='mt-5 inline-flex items-center gap-2 rounded-md border border-emerald-500/40 bg-emerald-500/10 px-4 py-2 text-sm font-medium text-emerald-300 transition hover:bg-emerald-500/20'
          >
            <RefreshCw className='h-4 w-4' />
            {copy.retry}
          </button>
        )}
      </div>
    );
  }

  return (
    <div className='flex h-full min-h-0 flex-col bg-[#111827]'>
      <div className='flex h-10 shrink-0 items-center justify-between border-b border-slate-700/80 bg-[#182235] px-3'>
        <div className='flex min-w-0 items-center gap-2 text-xs text-slate-400'>
          <NotebookTabs className='h-4 w-4 text-orange-300' />
          <span className='truncate font-medium text-slate-300'>JupyterLab</span>
        </div>
        <button
          type='button'
          onClick={onOpenExternal}
          className='inline-flex items-center gap-1.5 rounded px-2 py-1 text-xs text-slate-400 transition hover:bg-slate-700/60 hover:text-slate-100'
        >
          <ExternalLink className='h-3.5 w-3.5' />
          {copy.openExternal}
        </button>
      </div>
      <iframe
        title='JupyterLab'
        src={url}
        className='min-h-0 flex-1 border-0 bg-white'
        allow='clipboard-read; clipboard-write'
      />
    </div>
  );
};
