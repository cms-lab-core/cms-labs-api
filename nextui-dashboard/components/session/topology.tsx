import { useEffect, useReducer, useRef, useState } from 'react';
import { LoaderCircle, Network } from 'lucide-react';
import { ReactFlowProvider } from '@xyflow/react';
import { TopologyFlowVisualization } from '@/components/topology/view';
import { TerminalWindows } from '@/components/topology/terminal/window';
import { terminalInitialState, terminalReducer } from '@/components/topology/terminal/service/context';
import type { LabCopy } from './copy';

interface SessionTopologyProps {
  copy: LabCopy;
  ready: boolean;
  highlightedNode?: string;
  active?: boolean;
}

export const SessionTopology = ({ copy, ready, highlightedNode, active = true }: SessionTopologyProps) => {
  const [terminals, dispatch] = useReducer(terminalReducer, terminalInitialState);
  const viewportRef = useRef<HTMLDivElement>(null);
  const [bounds, setBounds] = useState({ width: 0, height: 0 });

  useEffect(() => {
    const viewport = viewportRef.current;
    if (!viewport) return;
    const observer = new ResizeObserver(([entry]) => {
      const { width, height } = entry.contentRect;
      // Hidden tabs must not reset the saved terminal geometry to zero.
      if (width <= 0 || height <= 0) return;
      setBounds((previous) => (previous.width === width && previous.height === height ? previous : { width, height }));
    });
    observer.observe(viewport);
    return () => observer.disconnect();
  }, [ready]);

  if (!ready) {
    return (
      <div className='flex h-full flex-col items-center justify-center bg-[#111827] px-6 text-center text-slate-200'>
        <div className='relative mb-4'>
          <Network className='h-10 w-10 text-slate-500' />
          <LoaderCircle className='absolute -bottom-2 -right-3 h-5 w-5 animate-spin text-emerald-400' />
        </div>
        <h2 className='text-base font-semibold'>{copy.topologyLoading}</h2>
        <p className='mt-1 max-w-md text-sm text-slate-500'>{copy.topologyLoadingDescription}</p>
      </div>
    );
  }

  return (
    <ReactFlowProvider>
      <div ref={viewportRef} className='relative isolate h-full min-h-0 w-full overflow-hidden bg-[#111827]'>
        <div className='pointer-events-none absolute left-4 top-4 z-20 rounded-md border border-slate-700/80 bg-slate-900/85 px-3 py-2 text-xs text-slate-400 backdrop-blur'>
          {copy.topologyHint}
        </div>
        <TerminalWindows nodes={terminals.clients} dispatch={dispatch} bounds={bounds} />
        <TopologyFlowVisualization
          dispatch={dispatch}
          highlightedNodeId={highlightedNode}
          active={active}
          viewportSize={bounds}
        />
      </div>
    </ReactFlowProvider>
  );
};
