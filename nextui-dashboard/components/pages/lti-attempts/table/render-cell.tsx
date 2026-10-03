import React from 'react';
import { ModelsLTIAttemptListItem } from '@/helpers/api';
import { Link } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { CheckCircle2, CircleDashed, LogIn, SquarePen } from 'lucide-react';
import { Button } from '@heroui/react';
import { formatCheckScore, parseAttemptCheckResult } from './check-result';

interface Props {
  item: CamelCasedPropertiesDeep<ModelsLTIAttemptListItem>;
  columnKey: string | React.Key;
  locale: any;
  connectLabel: string;
  editLabel: string;
}

export const RenderCell = ({ item, columnKey, locale, connectLabel, editLabel }: Props) => {
  switch (columnKey) {
    case 'id':
      return (
        <div>
          <div>
            <span>#{item.id ?? 0}</span>
          </div>
        </div>
      );
    case 'user':
      return (
        <div>
          <div>
            <span>{item.userName}</span>
          </div>
          <div>
            <Link to={RoutesLocation.accountsEdit(item.userId?.toString())}>
              <span>{item.userEmail}</span>
            </Link>
          </div>
        </div>
      );
    case 'server':
      return (
        <Link to={RoutesLocation.serversEdit(item.serverId?.toString())}>
          <div>
            <div>
              <span>{item.serverName}</span>
            </div>
          </div>
        </Link>
      );
    case 'name':
      return (
        <Link to={RoutesLocation.ltiRoutingEdit(item.ltiRoutingId?.toString())}>
          <div>
            <span>{item.ltiRoutingName}</span>
          </div>
        </Link>
      );
    case 'status':
      return (
        <div>
          <span>{locale?.[item.status] || item.status}</span>
        </div>
      );
    case 'checks': {
      const result = parseAttemptCheckResult(item.result);
      if (!result) return <span className='text-default-400'>—</span>;
      const hasTasks = result.tasks.length > 0;
      const hasScore = result.currentScore !== undefined && result.maxScore !== undefined;
      return (
        <div
          className='min-w-32 space-y-1.5'
          title={result.tasks.map((task) => `${task.complete ? '✓' : '○'} ${task.title}`).join('\n')}
        >
          <div className='flex items-center gap-2 text-xs'>
            {result.progress === 100 ? (
              <CheckCircle2 className='h-4 w-4 shrink-0 text-success' />
            ) : (
              <CircleDashed className='h-4 w-4 shrink-0 text-warning' />
            )}
            <span className='font-medium text-foreground'>
              {hasScore
                ? `${formatCheckScore(result.currentScore!)} / ${formatCheckScore(result.maxScore!)}`
                : hasTasks
                  ? `${result.passedTasks} / ${result.tasks.length}`
                  : result.resultDisplay || '✓'}
            </span>
          </div>
          <div className='h-1.5 overflow-hidden rounded-full bg-default-200'>
            <div
              className={`h-full rounded-full ${result.progress === 100 ? 'bg-success' : 'bg-warning'}`}
              style={{ width: `${result.progress}%` }}
            />
          </div>
        </div>
      );
    }
    case 'actions':
      return (
        <div className='flex items-center justify-end gap-2'>
          {(item.status === 'pending' || item.status === 'active') && item.attemptId && (
            <Button
              as={Link}
              to={RoutesLocation.session(item.attemptId)}
              size='sm'
              color='success'
              variant='flat'
              startContent={<LogIn className='h-4 w-4' />}
            >
              {connectLabel}
            </Button>
          )}
          <Button
            as={Link}
            to={RoutesLocation.ltiAttemptEdit(item.id?.toString())}
            isIconOnly
            size='sm'
            variant='light'
            aria-label={editLabel}
          >
            <SquarePen className='h-5 w-5 stroke-[#969696]' />
          </Button>
        </div>
      );
    default:
      return '';
  }
};
