'use client';
import React from 'react';
import { CardLastAttempt } from './card-last-attempt';
import HomeUsersWidget from '@/app/(app)/home/users-table';
import { CardServers } from '@/components/home/card-servers';
import { ContentCardWrapperMain } from '@/components/home/card-wrapper';
import useLanguageBrowser from '@/helpers/locale';
import { RoutesLocation } from '@/components/routes';
import CardServersDistribute from '@/components/home/card-servers-distribute';
import { HomeSummaryCards } from '@/components/home/summary-cards';

export const Content = () => {
  const {
    locale: { Servers, ServersQueue }
  } = useLanguageBrowser();
  return (
    <div className='h-full w-full min-w-0 px-4 lg:px-6'>
      <div className='mx-auto w-full min-w-0 max-w-[90rem] pt-6'>
        <HomeSummaryCards />

        <div className='mt-6'>
          <CardLastAttempt />
        </div>

        <div className='mt-6 grid min-w-0 gap-6 xl:grid-cols-2'>
          <ContentCardWrapperMain title={Servers.StatsChart.Title} link={RoutesLocation.servers()} wrapChildren={true}>
            <CardServers />
          </ContentCardWrapperMain>
          <ContentCardWrapperMain title={ServersQueue.RoundRobinChart.Title} wrapChildren={true}>
            <CardServersDistribute />
          </ContentCardWrapperMain>
        </div>
      </div>
      <div className='mx-auto w-full max-w-[90rem] gap-4 pt-6 lg:px-0 xl:flex-nowrap xl:gap-6'>
        <HomeUsersWidget />
      </div>
    </div>
  );
};
