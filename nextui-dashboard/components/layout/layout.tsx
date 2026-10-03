'use client';

import React from 'react';
import { SidebarWrapper } from '../sidebar/sidebar';
import { SidebarContext } from './layout-context';
import useCollapseBrowser from '@/components/navbar/useCollapse';

interface Props {
  children: React.ReactNode;
}

export const Layout = ({ children }: Props) => {
  const sidebarState = useCollapseBrowser();

  return (
    <SidebarContext.Provider
      value={{
        collapsed: sidebarState.collapsed,
        setCollapsed: sidebarState.setCollapsed
      }}
    >
      <section className='flex h-screen w-full min-w-0 overflow-hidden'>
        <SidebarWrapper />
        <main className='relative flex min-w-0 flex-1 flex-col overflow-x-hidden overflow-y-auto'>{children}</main>
      </section>
    </SidebarContext.Provider>
  );
};
