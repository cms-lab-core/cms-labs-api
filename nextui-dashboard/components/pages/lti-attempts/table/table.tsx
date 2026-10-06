import { Table, TableBody, TableCell, TableColumn, TableHeader, TableRow } from '@heroui/react';
import React from 'react';
import { RenderCell } from './render-cell';
import useLanguageBrowser from '@/helpers/locale';
import { ModelsLTIAttemptListItem } from '@/helpers/api';
import InfiniteScroll from '@/components/scroll/infinity-scroll';
import { CamelCasedPropertiesDeep } from 'type-fest';

interface Props {
  loadMore?: () => void;
  isLoading?: boolean;
  isInitialLoading?: boolean;
  rows?: CamelCasedPropertiesDeep<ModelsLTIAttemptListItem>[];
  selectedKeys?: Set<string>;
  onSelectionChange?: (keys: Set<string>) => void;
}

export const LTIAttemptTableWrapper = ({ rows, isLoading, loadMore, selectedKeys, onSelectionChange }: Props) => {
  const {
    locale: {
      Tables: { LTIAttemptsTable },
      AuthProviderAttempt
    }
  } = useLanguageBrowser();
  const locale = Object.fromEntries(
    AuthProviderAttempt.FieldStatusValues.map((item) => {
      return [item.key, item.value];
    })
  );

  const handleSelectionChange = (keys: 'all' | Set<React.Key>) => {
    if (keys === 'all') {
      const allKeys = new Set<string>();
      rows?.forEach((row) => {
        if (row.id !== undefined) allKeys.add(String(row.id));
      });
      onSelectionChange?.(allKeys);
    } else {
      onSelectionChange?.(new Set(Array.from(keys, String)));
    }
  };

  return (
    <InfiniteScroll loadMore={loadMore} isLoading={isLoading}>
      <div className=' w-full flex flex-col gap-4'>
        <Table
          aria-label='LTI attempts table'
          selectionMode='multiple'
          selectionBehavior='toggle'
          selectedKeys={selectedKeys}
          onSelectionChange={handleSelectionChange}
        >
          <TableHeader columns={LTIAttemptsTable.Columns}>
            {(column) => (
              <TableColumn
                key={column.uid}
                hideHeader={column.uid === 'actions'}
                align={column.uid === 'actions' ? 'center' : 'start'}
              >
                {column.name}
              </TableColumn>
            )}
          </TableHeader>
          <TableBody items={rows ?? []}>
            {(item) => (
              <TableRow key={String(item.id)}>
                {(columnKey) => (
                  <TableCell>
                    {RenderCell({
                      item,
                      columnKey,
                      locale,
                      connectLabel: LTIAttemptsTable.Connect,
                      editLabel: LTIAttemptsTable.ButtonEdit
                    })}
                  </TableCell>
                )}
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </InfiniteScroll>
  );
};
