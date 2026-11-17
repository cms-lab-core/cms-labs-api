import { useCallback, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react';
import { TerminalActionFunc, TerminalClient } from './service/context';
import { SmartLink } from '@/components/navbar/smartLink';
import { Chip, Tooltip } from '@heroui/react';
import useLanguageBrowser from '@/helpers/locale';
import { useReactFlow } from '@xyflow/react';
import { RFNodeTopology } from '@/components/topology/objectTypes/types';
import copy from 'copy-to-clipboard';
import { fitTerminalRectangle, resizeTerminalRectangle, type TerminalBounds, type TerminalRectangle } from './geometry';

interface KubernetesTerminalProps {
  node: TerminalClient;
  dispatch?: TerminalActionFunc;
  bounds: TerminalBounds;
}

export const KubernetesTerminal = ({ node, dispatch, bounds }: KubernetesTerminalProps) => {
  const {
    locale: {
      Topology: { Terminal }
    }
  } = useLanguageBrowser();
  const [rectangle, setRectangle] = useState<TerminalRectangle>({
    x: 100 + node.index * 20,
    y: 100 + node.index * 20,
    width: 400,
    height: 300
  });
  const fitted = fitTerminalRectangle(rectangle, bounds);
  const headerRef = useRef<HTMLDivElement>(null);
  const [isInteracting, setIsInteracting] = useState(false);
  const interaction = useRef<{
    pointerId: number;
    mode: 'move' | 'resize';
    startX: number;
    startY: number;
    rectangle: TerminalRectangle;
  } | null>(null);
  const { getNode } = useReactFlow();
  const nodeRF = getNode(node.id) as RFNodeTopology | undefined;

  const beginInteraction = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>, mode: 'move' | 'resize') => {
      if (event.button !== 0 || (mode === 'move' && event.target !== headerRef.current)) return;
      event.preventDefault();
      event.stopPropagation();
      event.currentTarget.setPointerCapture(event.pointerId);
      interaction.current = {
        pointerId: event.pointerId,
        mode,
        startX: event.clientX,
        startY: event.clientY,
        rectangle: fitted
      };
      setIsInteracting(true);
      dispatch?.({ type: 'BRING_TO_FRONT', payload: { id: node.id } });
    },
    [dispatch, fitted, node.id]
  );

  const moveInteraction = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const start = interaction.current;
      if (!start || start.pointerId !== event.pointerId) return;
      const deltaX = event.clientX - start.startX;
      const deltaY = event.clientY - start.startY;
      const proposed =
        start.mode === 'move'
          ? fitTerminalRectangle(
              { ...start.rectangle, x: start.rectangle.x + deltaX, y: start.rectangle.y + deltaY },
              bounds
            )
          : resizeTerminalRectangle(start.rectangle, { width: deltaX, height: deltaY }, bounds);
      setRectangle(proposed);
    },
    [bounds]
  );

  const endInteraction = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    if (interaction.current?.pointerId !== event.pointerId) return;
    interaction.current = null;
    setIsInteracting(false);
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
  }, []);

  if (!node.isVisible) return null;

  const externalIp = nodeRF?.data?.data?.serviceExternalIp?.[0];
  const shellUrl = nodeRF?.data?.data?.shellUrl;

  return (
    <div
      className='pointer-events-auto absolute rounded-lg overflow-hidden shadow-xl border border-slate-600'
      style={{
        width: `${fitted.width}px`,
        height: `${fitted.height}px`,
        left: `${fitted.x}px`,
        top: `${fitted.y}px`,
        zIndex: node.zIndex,
        opacity: node.isOpacity ? 0.6 : 1,
        backdropFilter: 'blur(4px)',
        backgroundColor: 'rgba(30, 41, 59, 0.7)',
        transform: 'translate3d(0,0,0)'
      }}
      onMouseDown={() => dispatch?.({ type: 'BRING_TO_FRONT', payload: { id: node.id } })}
    >
      <div
        ref={headerRef}
        className='flex touch-none items-center justify-between bg-slate-800 bg-opacity-70 px-3 py-2 cursor-move'
        onPointerDown={(event) => beginInteraction(event, 'move')}
        onPointerMove={moveInteraction}
        onPointerUp={endInteraction}
        onPointerCancel={endInteraction}
        onLostPointerCapture={endInteraction}
      >
        <Tooltip content={<p>{nodeRF?.data.data?.serviceHealthy ? Terminal.ReadyStatus : Terminal.NotReadyStatus}</p>}>
          <Chip
            color={nodeRF?.data.data?.serviceHealthy ? 'success' : 'warning'}
            variant='dot'
            radius='none'
            size='sm'
            classNames={{
              base: 'border-none text-white'
            }}
            onClick={() => {
              externalIp && copy(externalIp);
            }}
          >
            {nodeRF?.data.label || node.label || node.id}&nbsp;{externalIp}
          </Chip>
        </Tooltip>
        <div className='flex space-x-2'>
          <button
            onClick={() => dispatch?.({ type: 'TOGGLE_OPACITY', payload: { id: node.id } })}
            className='text-slate-300 hover:text-blue-400'
            title='Toggle opacity'
          >
            {node.isOpacity ? '◉' : '◎'}
          </button>
          <SmartLink
            to={shellUrl}
            className='text-slate-300 hover:text-blue-400 cursor-pointer'
            title='Open in new window'
          >
            ↗
          </SmartLink>
          <button
            onClick={() => dispatch?.({ type: 'TOGGLE_VISIBILITY', payload: { id: node.id } })}
            className='text-slate-300 hover:text-red-400'
            title='Close'
          >
            ✕
          </button>
        </div>
      </div>
      {shellUrl && (
        <iframe
          className='w-full h-full'
          style={{
            height: `calc(100% - 40px)`,
            backgroundColor: 'white',
            pointerEvents: isInteracting ? 'none' : 'auto'
          }}
          src={shellUrl}
        ></iframe>
      )}
      <div
        className='absolute bottom-0 right-0 w-4 h-4 touch-none cursor-se-resize bg-blue-500 opacity-0 hover:opacity-100'
        onPointerDown={(event) => beginInteraction(event, 'resize')}
        onPointerMove={moveInteraction}
        onPointerUp={endInteraction}
        onPointerCancel={endInteraction}
        onLostPointerCapture={endInteraction}
      />
    </div>
  );
};
