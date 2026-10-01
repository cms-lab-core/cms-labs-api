import React from 'react';
import { Loading } from '@/components/scroll/loader';
import { useQueryServerQueueList } from '@/helpers/queries/server_queue/use-query-server-queue-list';

const chartColors = ['#10b981', '#3b82f6', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16'];
const chartWidth = 800;
const chartHeight = 220;
const chartPadding = 36;

export const CardServersDistribute = () => {
  const response = useQueryServerQueueList({});
  if (response.isLoading) return <Loading size='md' />;
  const servers = response.data?.model || [];
  const uniqueServers = Array.from(new Map(servers.map((server) => [server.id, server])).values());
  const colorByID = new Map(uniqueServers.map((server, index) => [server.id, chartColors[index % chartColors.length]]));
  const xRange = Math.max(servers.length - 1, 1);

  return (
    <div className='space-y-4'>
      <div className='w-full overflow-x-auto rounded-medium border border-divider bg-content1'>
        <svg
          viewBox={`0 0 ${chartWidth} ${chartHeight}`}
          className='min-h-52 min-w-[560px] text-default-400'
          role='img'
          aria-label={servers.map((server, index) => `${index}: ${server.name || `#${server.id}`}`).join(', ')}
        >
          {[0, 1, 2, 3, 4].map((line) => {
            const x = chartPadding + (line / 4) * (chartWidth - chartPadding * 2);
            return (
              <line
                key={line}
                x1={x}
                x2={x}
                y1={chartPadding}
                y2={chartHeight - chartPadding}
                stroke='currentColor'
                strokeOpacity='0.18'
              />
            );
          })}
          <line
            x1={chartPadding}
            x2={chartWidth - chartPadding}
            y1={chartHeight / 2}
            y2={chartHeight / 2}
            stroke='currentColor'
            strokeOpacity='0.3'
          />
          {servers.map((server, index) => {
            const x = chartPadding + (index / xRange) * (chartWidth - chartPadding * 2);
            const y = server.lastUsed ? chartHeight / 2 - 30 : chartHeight / 2;
            const color = colorByID.get(server.id) || chartColors[0];
            const label = server.name || `#${server.id}`;
            return (
              <g key={`${server.id}-${index}`}>
                <title>{`${index}: ${label}${server.lastUsed ? ' · last used' : ''}`}</title>
                {server.lastUsed ? (
                  <rect
                    x={x - 8}
                    y={y - 8}
                    width='16'
                    height='16'
                    rx='2'
                    fill={color}
                    stroke='hsl(var(--heroui-background))'
                    strokeWidth='3'
                    transform={`rotate(45 ${x} ${y})`}
                  />
                ) : (
                  <circle cx={x} cy={y} r='8' fill={color} stroke='hsl(var(--heroui-background))' strokeWidth='3' />
                )}
              </g>
            );
          })}
        </svg>
      </div>
      <div className='flex flex-wrap gap-x-5 gap-y-2'>
        {uniqueServers.map((server) => (
          <span key={server.id} className='inline-flex items-center gap-2 text-xs text-default-600'>
            <span
              className='h-2.5 w-2.5 rounded-full'
              style={{ backgroundColor: colorByID.get(server.id) || chartColors[0] }}
            />
            {server.name || `#${server.id}`}
          </span>
        ))}
      </div>
    </div>
  );
};

export default CardServersDistribute;
