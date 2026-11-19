import { Button } from '@heroui/react';
import React, { useState } from 'react';
import { House, UserRoundKey, UsersRound } from 'lucide-react';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { CrumbsLayout } from '@/components/layout/crumbs';
import { Link } from 'react-router-dom';
import SearchInput from '@/components/sidebar/search-input';
import { RolesTableWrapper } from '@/components/pages/roles/table/table';
import { useQueryRoleList } from '@/helpers/queries/role/use-query-role-list';

export const RolesList = () => {
  const { locale } = useLanguageBrowser();
  const {
    locale: {
      Tables: { RoleTable }
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
      icon: <UserRoundKey className='w-5 h-5 stroke-[#969696]' />,
      name: locale.Sidebar.Roles,
      href: RoutesLocation.roles()
    },
    {
      icon: undefined,
      name: locale.Sidebar.AnyList,
      href: '#'
    }
  ];
  const [searchTerm, setSearchTerm] = useState<string>('');
  const queryRoles = useQueryRoleList({});
  const rows = queryRoles.data?.model || [];
  const search = searchTerm.trim().toLowerCase();
  const filteredRows = rows.filter((row) =>
    [row.name, row.code, String(row.id ?? '')].some((value) => value?.toLowerCase().includes(search))
  );
  const totalCount = filteredRows.length;
  return (
    <CrumbsLayout name={`${RoleTable.Title} (${totalCount})`} crumbs={crumbs}>
      <>
        <div className='flex justify-between flex-wrap gap-4 items-center'>
          <div className='flex items-center gap-3 flex-nowrap w-full'>
            <SearchInput placeholder={RoleTable.SearchBar} setValue={setSearchTerm} />
            <Link to={RoutesLocation.rolesCreate()}>
              <Button color='primary'>{RoleTable.ButtonAdd}</Button>
            </Link>
          </div>
        </div>
        <div className='max-w-[95rem] mx-auto w-full'>
          <RolesTableWrapper rows={filteredRows} isLoading={queryRoles.isLoading} />
        </div>
      </>
    </CrumbsLayout>
  );
};
