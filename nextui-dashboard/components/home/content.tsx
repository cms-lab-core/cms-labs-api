'use client';
import React from 'react';
import { CardLastAttempt } from './card-last-attempt';
import HomeUsersWidget from '@/app/(app)/home/users-table';
import { CardServers } from '@/components/home/card-servers';
import { ContentCardWrapperMain } from '@/components/home/card-wrapper';
import useLanguageBrowser from '@/helpers/locale';
import { RoutesLocation } from '@/components/routes';
import CardServersDistribute from '@/components/home/card-servers-distribute';

export const Content = () => {
  const {
    locale: { Servers, ServersQueue }
  } = useLanguageBrowser();
  return (
    <div className='h-full w-full min-w-0 px-4 lg:px-6'>
      <div className='mx-auto flex w-full min-w-0 max-w-[90rem] flex-wrap justify-center gap-4 lg:px-0 xl:flex-nowrap xl:gap-6'>
        <div className='mt-6 flex w-full min-w-0 flex-col gap-6'>
          {/* ServersCharts */}
          <ContentCardWrapperMain title={Servers.StatsChart.Title} link={RoutesLocation.servers()} wrapChildren={true}>
            <CardServers />
          </ContentCardWrapperMain>
          <ContentCardWrapperMain title={ServersQueue.RoundRobinChart.Title} wrapChildren={true}>
            <CardServersDistribute />
          </ContentCardWrapperMain>
        </div>

        {/* Left Section */}
        <div className='mt-4 flex w-full min-w-0 flex-col gap-2 xl:max-w-md'>
          <div>
            <CardLastAttempt />
          </div>
        </div>
      </div>
      <div className='gap-4 xl:gap-6 pt-3 px-4 lg:px-0 xl:flex-nowrap sm:pt-10 max-w-[90rem] mx-auto w-full'>
        <HomeUsersWidget />
      </div>
    </div>
  );
};
