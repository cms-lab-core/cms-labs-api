import React from 'react';
import { ModelsLTIAttemptListItem } from '@/helpers/api';
import { Link } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { LogIn, SquarePen } from 'lucide-react';
import { Button } from '@heroui/react';

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
