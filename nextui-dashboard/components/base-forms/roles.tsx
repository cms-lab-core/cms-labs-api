import React, { ReactNode } from 'react';
import { Chip, Select, SelectItem, Tooltip } from '@heroui/react';
import { Key } from '@react-types/shared';
import { SharedSelection } from '@heroui/system';
import { Loading } from '@/components/scroll/loader';
import { useQueryRoleList } from '@/helpers/queries/role/use-query-role-list';

interface RolesSelectorProps {
  label?: ReactNode;
  description?: ReactNode;
  selectedKeys?: Key[];
  onSelectionChange?: (keys: SharedSelection) => void;
}

export const RolesSelector = (props: RolesSelectorProps) => {
  const queryRoles = useQueryRoleList({});
  const roles = queryRoles.data?.model || [];
  const selectedKeys = props.selectedKeys?.map((item) => item.toString());
  return (
    <Select
      variant='bordered'
      label={props.label}
      description={props.description}
      selectionMode='multiple'
      selectedKeys={selectedKeys}
      onSelectionChange={props.onSelectionChange}
      isLoading={queryRoles.isLoading}
    >
      {roles.map((role) => (
        <SelectItem key={role.id?.toString()}>{role.name}</SelectItem>
      ))}
    </Select>
  );
};

interface RolesChipProps {
  maxRoles: number;
  roles?: Array<number>;
}

export const RolesChip = ({ roles, maxRoles }: RolesChipProps) => {
  const queryRoles = useQueryRoleList({});
  if (queryRoles.isLoading) return <Loading size='sm' />;

  const filteredRoles = queryRoles.data?.model?.filter((role) => (roles || []).includes(role.id || 0));

  if (!filteredRoles?.length) return null;

  // Показываем первые maxRoles роли, остальные в тултипе
  const visibleRoles = filteredRoles.slice(0, maxRoles);
  const hiddenRoles = filteredRoles.slice(maxRoles);

  return (
    <div className='flex items-center gap-2'>
      <div className='flex gap-2 overflow-x-auto scrollbar-thin max-w-[200px]'>
        {visibleRoles.map((role, index) => (
          <Chip size='sm' key={index}>
            {role.name ?? role.code ?? role.id}
          </Chip>
        ))}
      </div>

      {hiddenRoles.length > 0 && (
        <Tooltip
          placement='top-start'
          className='bg-transparent p-0 shadow-none'
          content={
            <div className='flex flex-wrap gap-1 max-w-xs'>
              {hiddenRoles.map((role, index) => (
                <Chip size='sm' key={index}>
                  {role.name ?? role.code ?? role.id}
                </Chip>
              ))}
            </div>
          }
        >
          <Chip size='sm' variant='flat'>
            +{hiddenRoles.length}
          </Chip>
        </Tooltip>
      )}
    </div>
  );
};
