import React, { useState } from 'react';
import { Select, SelectItem } from '@heroui/react';
import useLanguageBrowser from '@/helpers/locale';
import { Loading } from '@/components/scroll/loader';
import { useInfinityServerList } from '@/helpers/queries/server/use-infinity-server-list';

const chartColors = ['#10b981', '#3b82f6', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#f97316', '#84cc16'];

export const CardServers = () => {
  const {
    locale: {
      Servers: { StatsChart }
    }
  } = useLanguageBrowser();
  const [orderBy, setOrderBy] = useState('unitRate');
  const [status, setStatus] = useState('active');
  const response = useInfinityServerList({
    limit: 100,
    orderBy: orderBy,
    status: status
  });

  if (response.isLoading) return <Loading size='md' />;

  const rows = response?.data?.pages.flatMap((p) => p?.model ?? []) || [];

  const chartRows = rows.map((server, index) => ({
    label: server.model.name || `#${server.model.id}`,
    value: orderBy === 'lastCountUsers' ? server.model.lastCountUsers || 0 : server.model.unitRate || 0,
    color: chartColors[index % chartColors.length]
  }));
  const total = chartRows.reduce((sum, row) => sum + row.value, 0);
  let consumed = 0;
  const gradient =
    total > 0
      ? `conic-gradient(${chartRows
          .map((row) => {
            const start = (consumed / total) * 360;
            consumed += row.value;
            const end = (consumed / total) * 360;
            return `${row.color} ${start}deg ${end}deg`;
          })
          .join(', ')})`
      : 'conic-gradient(hsl(var(--heroui-default-200)) 0deg 360deg)';

  return (
    <div>
      <div className='flex justify-between gap-4 mb-4'>
        <Select
          variant='bordered'
          label={StatsChart.Status}
          selectedKeys={[status ?? '']}
          onSelectionChange={(keys) => setStatus(keys.currentKey || '')}
        >
          <SelectItem key='active'>{StatsChart.StatusValueActive}</SelectItem>
          <SelectItem key='all'>{StatsChart.StatusValueAll}</SelectItem>
        </Select>
        <Select
          variant='bordered'
          label={StatsChart.OrderBy}
          selectedKeys={[orderBy ?? '']}
          onSelectionChange={(keys) => setOrderBy(keys.currentKey || '')}
        >
          <SelectItem key='unitRate'>{StatsChart.OrderByValueUnitRate}</SelectItem>
          <SelectItem key='lastCountUsers'>{StatsChart.OrderByValueLastCountUsers}</SelectItem>
          <SelectItem key='createdAt'>{StatsChart.OrderByValueCreatedAt}</SelectItem>
        </Select>
      </div>
      <div className='grid min-h-72 items-center gap-8 md:grid-cols-[minmax(180px,280px)_minmax(0,1fr)]'>
        <div
          role='img'
          aria-label={`${StatsChart.Title}: ${chartRows.map((row) => `${row.label} ${row.value}`).join(', ')}`}
          className='relative mx-auto aspect-square w-full max-w-64 rounded-full shadow-inner'
          style={{ background: gradient }}
        >
          <div className='absolute inset-[28%] flex items-center justify-center rounded-full bg-content1 text-center shadow-sm'>
            <span>
              <span className='block text-xs text-default-500'>{StatsChart.Title}</span>
              <span className='block font-mono text-xl font-semibold text-foreground'>{total}</span>
            </span>
          </div>
        </div>
        <div className='max-h-64 space-y-2 overflow-y-auto pr-2'>
          {chartRows.length === 0 ? (
            <p className='text-sm text-default-500'>—</p>
          ) : (
            chartRows.map((row) => (
              <div key={row.label} className='flex items-center gap-3 text-sm'>
                <span className='h-2.5 w-2.5 shrink-0 rounded-full' style={{ backgroundColor: row.color }} />
                <span className='min-w-0 flex-1 truncate text-default-600'>{row.label}</span>
                <span className='font-mono font-medium text-foreground'>{row.value}</span>
                <span className='w-12 text-right font-mono text-xs text-default-400'>
                  {total > 0 ? `${Math.round((row.value / total) * 100)}%` : '0%'}
                </span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};

export default CardServers;
