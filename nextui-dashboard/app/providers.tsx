'use client';
import * as React from 'react';
import { HeroUIProvider } from '@heroui/system';
import { PreferencesProvider } from '@/helpers/preferences/context';
import { UserProfileProvider } from '@/components/providers/auth-jwt/context';
import { QueryClientProvider } from '@tanstack/react-query';
import queryClient from '@/helpers/queries/base';
import { BrowserRouter } from 'react-router-dom';
import RoutesDynamic from '@/app/routes';
import { ToastProvider } from '@heroui/toast';
import { zIndexClassToast } from '@/components/providers/const';

export function Providers() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename='/'>
        <UserProfileProvider>
          <HeroUIProvider>
            <PreferencesProvider>
              <ToastProvider
                regionProps={{
                  classNames: { base: zIndexClassToast }
                }}
              />
              <RoutesDynamic />
            </PreferencesProvider>
          </HeroUIProvider>
        </UserProfileProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
