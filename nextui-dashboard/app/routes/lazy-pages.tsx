import { lazy, Suspense, type PropsWithChildren } from 'react';
import { LoaderCircle } from 'lucide-react';

export const LoginPage = lazy(() => import('@/app/(auth)/layout'));
export const LoginError = lazy(() => import('@/app/(auth)/layout').then((module) => ({ default: module.LoginError })));
export const LanguagePage = lazy(() => import('@/app/(app)/lang/page'));
export const HomePage = lazy(() => import('@/app/(app)/home/page'));
export const SessionPage = lazy(() => import('@/app/(app)/session/page'));
export const SessionWorkspaceRedirectPage = lazy(() => import('@/app/(app)/session/workspace/page'));
export const TopologyPageView = lazy(() => import('@/app/(app)/topology/page'));
export const LabCatalogPage = lazy(() => import('@/app/(app)/labs/page'));

export const AccountsPage = lazy(() =>
  import('@/app/(app)/accounts/page').then((module) => ({ default: module.AccountsPage }))
);
export const AccountsPageEdit = lazy(() =>
  import('@/app/(app)/accounts/page').then((module) => ({ default: module.AccountsPageEdit }))
);
export const ProfilePagePasswordChange = lazy(() =>
  import('@/app/(app)/accounts/page').then((module) => ({ default: module.ProfilePagePasswordChange }))
);
export const RolesListPage = lazy(() =>
  import('@/app/(app)/roles/page').then((module) => ({ default: module.RolesListPage }))
);
export const RolesPageEdit = lazy(() =>
  import('@/app/(app)/roles/page').then((module) => ({ default: module.RolesPageEdit }))
);
export const AuthProvidersPage = lazy(() =>
  import('@/app/(app)/auth-providers/page').then((module) => ({ default: module.AuthProvidersPage }))
);
export const AuthProvidersPageEdit = lazy(() =>
  import('@/app/(app)/auth-providers/page').then((module) => ({ default: module.AuthProvidersPageEdit }))
);
export const ServersPage = lazy(() =>
  import('@/app/(app)/servers/page').then((module) => ({ default: module.ServersPage }))
);
export const ServersPageEdit = lazy(() =>
  import('@/app/(app)/servers/page').then((module) => ({ default: module.ServersPageEdit }))
);
export const LTIRoutingPage = lazy(() =>
  import('@/app/(app)/lti-routings/page').then((module) => ({ default: module.LTIRoutingPage }))
);
export const LTIRoutingPageEdit = lazy(() =>
  import('@/app/(app)/lti-routings/page').then((module) => ({ default: module.LTIRoutingPageEdit }))
);
export const LTIAttemptPageConfirm = lazy(() =>
  import('@/app/(app)/lti-attempts/page').then((module) => ({ default: module.LTIAttemptPageConfirm }))
);
export const LTIAttemptsPage = lazy(() =>
  import('@/app/(app)/lti-attempts/page').then((module) => ({ default: module.LTIAttemptsPage }))
);
export const LTIAttemptsPageCreate = lazy(() =>
  import('@/app/(app)/lti-attempts/page').then((module) => ({ default: module.LTIAttemptsPageCreate }))
);
export const LTIAttemptsPageEdit = lazy(() =>
  import('@/app/(app)/lti-attempts/page').then((module) => ({ default: module.LTIAttemptsPageEdit }))
);
export const ServiceCardsPage = lazy(() =>
  import('@/app/(app)/service-cards/page').then((module) => ({ default: module.ServiceCardsPage }))
);
export const ServiceCardsPageEdit = lazy(() =>
  import('@/app/(app)/service-cards/page').then((module) => ({ default: module.ServiceCardsPageEdit }))
);
export const TargetsPage = lazy(() =>
  import('@/app/(app)/targets/page').then((module) => ({ default: module.TargetsPage }))
);
export const TargetsPageEdit = lazy(() =>
  import('@/app/(app)/targets/page').then((module) => ({ default: module.TargetsPageEdit }))
);

export const RouteSuspense = ({ children }: PropsWithChildren) => (
  <Suspense
    fallback={
      <div className='flex min-h-screen w-full items-center justify-center bg-background text-foreground'>
        <LoaderCircle className='h-7 w-7 animate-spin text-primary' aria-label='Loading' />
      </div>
    }
  >
    {children}
  </Suspense>
);
