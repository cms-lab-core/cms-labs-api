import { addToast, Button, Select, SelectItem } from '@heroui/react';
import React, { useMemo, useState } from 'react';
import { BookPlus, House, Trash2 } from 'lucide-react';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { CrumbsLayout } from '@/components/layout/crumbs';
import { LTIAttemptTableWrapper } from '@/components/pages/lti-attempts/table/table';
import { useInfinityLtiAttemptList } from '@/helpers/queries/lti_attempt/use-infinity-lti-attempt-list';
import { AutoCompleteFull } from '@/components/base-forms/autocomplete';
import { useQueryUserAutoComplete } from '@/helpers/queries/user/use-query-user-auto-complete';
import { useQueryServerAutoComplete } from '@/helpers/queries/server/use-query-server-auto-complete';
import { useMutationLtiAttemptDelete } from '@/helpers/queries/lti_attempt/use-mutation-lti-attempt-delete';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { ModelsLTIAttemptListItem } from '@/helpers/api';
import { useSearchParams } from 'react-router-dom';
import { useMutationLtiAttemptUpdate } from '@/helpers/queries/lti_attempt/use-mutation-lti-attempt-update';

interface QueryParams {
  statuses: string[];
  userId?: string;
  serverClientId?: string;
}

function getQueryParams(searchParams: URLSearchParams): QueryParams {
  return {
    statuses: searchParams.getAll('statuses'),
    userId: searchParams.get('userId') || undefined,
    serverClientId: searchParams.get('serverClientId') || undefined
  };
}

export const LTIAttemptsListPage = () => {
  const { locale } = useLanguageBrowser();
  const {
    locale: {
      Tables: { LTIAttemptsTable },
      AuthProviderAttempt: { FieldStatusValues }
    }
  } = useLanguageBrowser();
  const [searchParams, setSearchParams] = useSearchParams();

  const queryParams = useMemo(() => {
    const params = getQueryParams(searchParams);
    return {
      statuses: params.statuses || [],
      userId: params.userId || '',
      serverClientId: params.serverClientId || ''
    };
  }, [searchParams]);

  const [selectedAttempts, setSelectedAttempts] = useState<Set<string | number>>(new Set());
  const [bulkTerminating, setBulkTerminating] = useState(false);

  const selectedStatuses = queryParams.statuses;
  const selectedUserId = queryParams.userId;
  const selectedServerClientId = queryParams.serverClientId;

  const updateParam = (key: string, value: string | string[]) => {
    const newParams = new URLSearchParams(searchParams.toString());
    newParams.delete(key);
    if (Array.isArray(value)) {
      value.forEach((v) => newParams.append(key, v));
    } else if (value) {
      newParams.set(key, value);
    }
    setSearchParams(newParams);
  };

  const crumbs = [
    { icon: <House className='w-5 h-5 stroke-[#969696]' />, name: locale.Sidebar.Home, href: RoutesLocation.home() },
    {
      icon: <BookPlus className='w-5 h-5 stroke-[#969696]' />,
      name: locale.Sidebar.LTIAttempts,
      href: RoutesLocation.ltiAttempts()
    },
    { icon: undefined, name: locale.Sidebar.AnyList, href: '#' }
  ];

  const response = useInfinityLtiAttemptList({
    limit: 100,
    statuses: selectedStatuses,
    userIds: selectedUserId ? [Number(selectedUserId)] : [],
    serverClientIds: selectedServerClientId ? [selectedServerClientId] : []
  });

  const rows = response?.data?.pages.flatMap((p) => p?.model ?? []) || [];
  const totalCount = (response.data?.pages[0] as any)?.totalCount ?? 0;
  const hasSelection = selectedAttempts.size > 0;

  const bulkTerminatingMutation = useMutationLtiAttemptUpdate();

  const terminateSelected = async () => {
    const attempts = Array.from(selectedAttempts);
    if (attempts.length === 0 || bulkTerminating) return;

    setBulkTerminating(true);
    const failed = new Set<string | number>();
    let firstError = '';
    for (const selectedID of attempts) {
      try {
        await bulkTerminatingMutation.mutateAsync({ id: Number(selectedID), status: 'terminating' });
      } catch (error: any) {
        failed.add(selectedID);
        firstError ||= error?.data?.message || error?.message || locale.Forms.SaveError;
      }
    }
    setSelectedAttempts(failed);
    setBulkTerminating(false);

    const completed = attempts.length - failed.size;
    addToast({
      title: failed.size === 0 ? locale.Sidebar.Success : locale.Forms.SaveError,
      description:
        failed.size === 0 ? `${completed} / ${attempts.length}` : `${completed} / ${attempts.length}. ${firstError}`,
      color: failed.size === 0 ? 'success' : completed > 0 ? 'warning' : 'danger'
    });
  };

  const bulkDeleteMutation = useMutationLtiAttemptDelete({
    onSuccess: () => {
      addToast({ title: locale.Sidebar.Success, color: 'success' });
      setSelectedAttempts(new Set());
    },
    onError: (error: any) =>
      addToast({ title: locale.Forms.DeleteError, description: error?.data?.message, color: 'danger' })
  });

  return (
    <CrumbsLayout name={`${LTIAttemptsTable.Title} (${totalCount})`} crumbs={crumbs}>
      <div className='flex flex-col gap-4 mb-4'>
        {hasSelection ? (
          <div className='flex items-center justify-between bg-warning-50 p-3 rounded-lg'>
            <div className='flex items-center gap-3'>
              <span className='text-sm font-medium'>
                {selectedAttempts.size} {locale.Sidebar.Selected}
              </span>
              <Button size='sm' variant='light' onPress={() => setSelectedAttempts(new Set())}>
                {locale.Sidebar.Clear}
              </Button>
            </div>
            <div className='flex items-center gap-2'>
              <Button size='sm' variant='flat' color='warning' onPress={terminateSelected} isLoading={bulkTerminating}>
                {LTIAttemptsTable.BulkTerminating}
              </Button>
              <Button
                size='sm'
                variant='flat'
                color='danger'
                onPress={() =>
                  Array.from(selectedAttempts)
                    .map(Number)
                    .forEach((id) => bulkDeleteMutation.mutate({ id }))
                }
                isLoading={bulkDeleteMutation.isPending}
                startContent={<Trash2 className='w-4 h-4' />}
              >
                {locale.Sidebar.Delete}
              </Button>
            </div>
          </div>
        ) : (
          <div className='flex items-center gap-3 flex-wrap'>
            <Select
              label={LTIAttemptsTable.FilterStatus || 'Status'}
              selectionMode='multiple'
              selectedKeys={new Set(selectedStatuses)}
              onSelectionChange={(keys) => updateParam('statuses', Array.from(keys) as string[])}
              className='w-48'
            >
              {FieldStatusValues.map((status) => (
                <SelectItem key={status.key}>{status.value}</SelectItem>
              ))}
            </Select>
            <AutoCompleteFull
              label={LTIAttemptsTable.FilterUser || 'User'}
              placeholder={LTIAttemptsTable.FilterUser || 'Select user'}
              value={selectedUserId}
              onChange={(e) => updateParam('userId', e.target.value)}
              fetchData={useQueryUserAutoComplete}
              className='w-64'
            />
            <AutoCompleteFull
              label={LTIAttemptsTable.FilterServer || 'Server'}
              placeholder={LTIAttemptsTable.FilterServer || 'Select server'}
              value={selectedServerClientId}
              onChange={(e) => updateParam('serverClientId', e.target.value)}
              fetchData={useQueryServerAutoComplete}
              className='w-64'
            />
          </div>
        )}
      </div>
      <div className='max-w-[95rem] mx-auto w-full'>
        <LTIAttemptTableWrapper
          rows={rows as CamelCasedPropertiesDeep<ModelsLTIAttemptListItem>[]}
          isLoading={response.isLoading}
          loadMore={response.fetchNextPage.bind(response.fetchNextPage)}
          selectedKeys={selectedAttempts}
          onSelectionChange={setSelectedAttempts}
        />
      </div>
    </CrumbsLayout>
  );
};
