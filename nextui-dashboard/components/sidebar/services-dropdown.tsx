'use client';
import { Dropdown, DropdownItem, DropdownMenu, DropdownSection, DropdownTrigger, Image } from '@heroui/react';
import React from 'react';
import { CmsIcon } from '../icons/cms-icon';
import useLanguageBrowser from '@/helpers/locale';
import { useQueryServiceCardList } from '@/helpers/queries/service_card/use-query-service-card-list';
import { ChevronDown } from 'lucide-react';
import { useSidebarContext } from '@/components/layout/layout-context';

export const ServicesDropdown = () => {
  const { locale, lang } = useLanguageBrowser();
  const response = useQueryServiceCardList({}, true);
  const rows = (response.data?.model || []).filter((service) => service.isActive && service.url);
  const emptyMessage = response.isLoading
    ? lang === 'ru'
      ? 'Загрузка сервисов…'
      : 'Loading services…'
    : response.isError
      ? lang === 'ru'
        ? 'Повторить загрузку сервисов'
        : 'Retry loading services'
      : lang === 'ru'
        ? 'Нет доступных сервисов'
        : 'No services available';
  const { collapsed } = useSidebarContext();
  const company = {
    title: locale.CompaniesDropdown.Title,
    description: locale.CompaniesDropdown.Description,
    link: '/',
    logo: <CmsIcon />
  };
  return (
    <Dropdown
      classNames={{
        base: 'w-full min-w-[240px]'
      }}
    >
      <DropdownTrigger className='cursor-pointer'>
        <div className='flex items-center gap-2'>
          {company.logo}
          {collapsed ? null : (
            <>
              <div className='flex flex-col gap-4'>
                <h3 className='text-xl font-medium m-0 text-default-900 -mb-4 whitespace-nowrap'>{company.title}</h3>
                <span className='text-xs font-medium text-default-500'>{company.description}</span>
              </div>
              <ChevronDown className='w-3 h-3 stroke-[#969696]' />
            </>
          )}
        </div>
      </DropdownTrigger>
      <DropdownMenu aria-label='Avatar Actions'>
        <DropdownSection title={locale.Tables.ServiceCardsTable.Title}>
          {rows.length > 0 ? (
            rows.map((service) => (
              <DropdownItem
                key={String(service.id)}
                href={service.url}
                target='_blank'
                rel='noopener noreferrer'
                startContent={service.imageUrl ? <Image src={service.imageUrl} width={30} alt='' /> : undefined}
                description={service.description}
                textValue={service.name}
                classNames={{ base: 'py-4', title: 'text-base font-semibold' }}
              >
                {service.name}
              </DropdownItem>
            ))
          ) : (
            <DropdownItem
              key='empty-services'
              isDisabled={!response.isError}
              onPress={() => {
                void response.refetch();
              }}
            >
              {emptyMessage}
            </DropdownItem>
          )}
        </DropdownSection>
      </DropdownMenu>
    </Dropdown>
  );
};
