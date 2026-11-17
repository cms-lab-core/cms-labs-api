import { TerminalActionFunc, TerminalClient } from '@/components/topology/terminal/service/context';
import { KubernetesTerminal } from '@/components/topology/terminal/terminal';
import type { TerminalBounds } from './geometry';

interface TerminalWindows {
  nodes: TerminalClient[];
  dispatch?: TerminalActionFunc;
  bounds: TerminalBounds;
}

export const TerminalWindows = ({ nodes, dispatch, bounds }: TerminalWindows) => {
  return (
    <div className='pointer-events-none absolute inset-0 z-30 overflow-hidden'>
      {nodes.map((node) => (
        <KubernetesTerminal key={node.id} node={node} dispatch={dispatch} bounds={bounds} />
      ))}
    </div>
  );
};
