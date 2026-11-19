import { Button } from '@heroui/react';
import React, { useState } from 'react';
import { House, UsersRound } from 'lucide-react';
import { UsersTableWrapper } from '@/components/pages/accounts/table/table';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { CrumbsLayout } from '@/components/layout/crumbs';
import SearchInput from '@/components/sidebar/search-input';
import { Link } from 'react-router-dom';
import { MapUserItem, useInfinityUserList } from '@/helpers/queries/user/use-infinity-user-list';

export const Accounts = () => {
  const { locale } = useLanguageBrowser();
  const {
    locale: {
      Tables: { UsersTable, RoleTable }
    }
  } = useLanguageBrowser();
  const crumbs = [
    {
      icon: <House className='w-5 h-5 stroke-[#969696]' />,
      name: locale.Sidebar.Home,
      href: RoutesLocation.home()
    },
    {
      icon: <UsersRound className='w-5 h-5 stroke-[#969696]' />,
      name: locale.Sidebar.Users,
      href: RoutesLocation.accounts()
    },
    {
      icon: undefined,
      name: locale.Sidebar.AnyList,
      href: '#'
    }
  ];
  const [searchTerm, setSearchTerm] = useState<string>('');
  const response = useInfinityUserList({ limit: 100, search: searchTerm });
  const users =
    response?.data?.pages.flatMap((p) => p?.model?.map((item) => MapUserItem(item.model, item.roles)) ?? []) || [];
  const totalCount = response.data?.pages[0]?.totalCount ?? 0;
  return (
    <CrumbsLayout name={`${UsersTable.Title} (${totalCount})`} crumbs={crumbs}>
      <>
        <div className='flex justify-between flex-wrap gap-4 items-center'>
          <div className='flex items-center gap-3 flex-nowrap w-full'>
            <SearchInput placeholder={UsersTable.SearchBar} setValue={setSearchTerm} />
            <Link to={RoutesLocation.roles()}>
              <Button color='default'>{RoleTable.Title}</Button>
            </Link>
            <Link to={RoutesLocation.accountsCreate()}>
              <Button color='primary'>{UsersTable.ButtonAdd}</Button>
            </Link>
          </div>
        </div>
        <div className='max-w-[95rem] mx-auto w-full'>
          <UsersTableWrapper
            users={users}
            isLoading={response.isLoading}
            loadMore={response.fetchNextPage.bind(response.fetchNextPage)}
          />
        </div>
      </>
    </CrumbsLayout>
  );
};
