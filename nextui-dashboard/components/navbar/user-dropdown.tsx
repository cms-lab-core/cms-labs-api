import { Dropdown, DropdownItem, DropdownMenu, DropdownTrigger } from '@heroui/react';
import React, { useCallback } from 'react';
import { useMutationUserLogout } from '@/helpers/queries/user/use-mutation-user-logout';
import useLanguageBrowser from '@/helpers/locale';
import { useNavigate } from 'react-router-dom';
import { useUserProfile } from '@/components/providers/auth-jwt/hooks';
import CustomAvatar from '@/components/sidebar/avatar';
import { RoutesLocation } from '@/components/routes';
import { useSidebarContext } from '@/components/layout/layout-context';

export const UserDropdown = () => {
  const navigate = useNavigate();
  const { collapsed } = useSidebarContext();
  const { locale } = useLanguageBrowser();
  const user = useUserProfile();

  const { mutateAsync } = useMutationUserLogout();

  const handleUser = useCallback(() => {
    navigate(RoutesLocation.accountsEdit(user.sub));
  }, [navigate, user.sub]);

  const handleLogout = useCallback(async () => {
    await mutateAsync({});
    window.location.replace(RoutesLocation.login());
  }, [mutateAsync]);

  const handleChangeLanguage = useCallback(() => {
    navigate(RoutesLocation.language());
  }, [navigate]);

  const handleToServicesCards = useCallback(() => {
    navigate(RoutesLocation.serviceCards());
  }, [navigate]);

  const handleToPasswordChange = useCallback(() => {
    navigate(RoutesLocation.profileChangePassword());
  }, [navigate]);

  return (
    <Dropdown>
      <DropdownTrigger>
        <div className='hover:bg-default-100 flex gap-2 w-full min-h-[44px] h-full items-center px-3.5 rounded-xl cursor-pointer transition-all duration-150 active:scale-[0.98]'>
          {CustomAvatar({
            tooltip: false,
            as: 'button',
            size: 'sm',
            username: user.name,
            name: user.name,
            email: user.email
          })}
          {collapsed ? null : <span className='text-default-900'>{user.name || user.email}</span>}
        </div>
      </DropdownTrigger>
      <DropdownMenu aria-label='User menu actions'>
        <DropdownItem key='profile' className='flex flex-col justify-start w-full items-start' onPress={handleUser}>
          <p>{locale.UserNavBar.SignedAs}</p>
          <p>{user.email}</p>
        </DropdownItem>
        <DropdownItem key='password' onPress={handleToPasswordChange}>
          {locale.UserNavBar.PasswordChange}
        </DropdownItem>
        <DropdownItem key='language' onPress={handleChangeLanguage}>
          {locale.UserNavBar.LanguageChange}
        </DropdownItem>
        <DropdownItem key='services' onPress={handleToServicesCards}>
          {locale.Sidebar.ServiceCards}
        </DropdownItem>
        <DropdownItem key='logout' color='danger' className='text-danger' onPress={handleLogout}>
          {locale.UserNavBar.Logout}
        </DropdownItem>
      </DropdownMenu>
    </Dropdown>
  );
};
