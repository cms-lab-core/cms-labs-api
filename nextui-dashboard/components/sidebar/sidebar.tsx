import React from 'react';
import { Sidebar } from './sidebar.styles';
import { ServicesDropdown } from './services-dropdown';
import { SidebarItem } from './sidebar-item';
import { SidebarMenu } from './sidebar-menu';
import { useSidebarContext } from '../layout/layout-context';
import { useLocation } from 'react-router-dom';
import useLanguageBrowser from '@/helpers/locale';
import { RoutesLocation } from '@/components/routes';
import { BookOpen, ChevronLeft, ChevronRight, House, KeyRound, Map, Server, Split, Users } from 'lucide-react';
import { UserDropdown } from '@/components/navbar/user-dropdown';
import { useUserProfile } from '@/components/providers/auth-jwt/hooks';
import { useIsLabCatalogVisible } from '@/helpers/queries/lab_catalog/use-is-lab-catalog-visible';
import { UserRoleBase } from '@/helpers/queries/sso/auth';

export const SidebarWrapper = () => {
  const { pathname } = useLocation();
  const { locale } = useLanguageBrowser();
  const { collapsed, setCollapsed } = useSidebarContext();
  const user = useUserProfile();
  const isLabCatalogVisible = useIsLabCatalogVisible();
  const isOperator = user.roles?.some((role) => role === UserRoleBase.Admin || role === UserRoleBase.Instructor);

  return (
    <aside className='h-screen z-[20] sticky top-0'>
      <div>
        <div className={Sidebar.Header({ collapsed })}>
          <ServicesDropdown />
        </div>
        <div className='flex flex-col justify-between h-full'>
          <div className={Sidebar.Body()}>
            <SidebarItem
              title={locale.Sidebar.Home}
              icon={<House className='w-5 h-5 stroke-[#969696]' />}
              isActive={pathname === '/'}
              href='/'
            />
            {isLabCatalogVisible && (
              <SidebarItem
                isActive={pathname === RoutesLocation.labs()}
                title={locale.Sidebar.Labs}
                icon={<BookOpen className='w-5 h-5 stroke-[#969696]' />}
                href={RoutesLocation.labs()}
              />
            )}
            {isOperator && (
              <SidebarMenu title={locale.Sidebar.MainMenu}>
                <SidebarItem
                  isActive={pathname === RoutesLocation.ltiRouting()}
                  title={locale.Sidebar.LTIRouting}
                  icon={<Split className='w-5 h-5 stroke-[#969696]' />}
                  href={RoutesLocation.ltiRouting()}
                />
                <SidebarItem
                  isActive={pathname === RoutesLocation.authProviders()}
                  title={locale.Sidebar.AuthProviders}
                  icon={<KeyRound className='w-5 h-5 stroke-[#969696]' />}
                  href={RoutesLocation.authProviders()}
                />
                <SidebarItem
                  isActive={pathname === RoutesLocation.accounts()}
                  title={locale.Sidebar.Users}
                  icon={<Users className='w-5 h-5 stroke-[#969696]' />}
                  href={RoutesLocation.accounts()}
                />
                <SidebarItem
                  isActive={pathname === RoutesLocation.servers()}
                  title={locale.Sidebar.Servers}
                  icon={<Server className='w-5 h-5 stroke-[#969696]' />}
                  href={RoutesLocation.servers()}
                />
                <SidebarItem
                  isActive={pathname === RoutesLocation.targets()}
                  title={locale.Sidebar.Targets}
                  icon={<Map className='w-5 h-5 stroke-[#969696]' />}
                  href={RoutesLocation.targets()}
                />
              </SidebarMenu>
            )}
            <SidebarMenu title={locale.Sidebar.Profile}>
              <UserDropdown />
              <SidebarItem
                isActive={false}
                title={locale.Sidebar.Collapse}
                icon={collapsed ? <ChevronRight className='w-4 h-4' /> : <ChevronLeft className='w-4 h-4' />}
                onPress={() => setCollapsed(!collapsed)}
              />
            </SidebarMenu>
          </div>
        </div>
      </div>
    </aside>
  );
};
