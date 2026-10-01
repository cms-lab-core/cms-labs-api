import { Route, Routes } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import {
  AccountsPage,
  AccountsPageEdit,
  AuthProvidersPage,
  AuthProvidersPageEdit,
  HomePage,
  LabCatalogPage,
  LanguagePage,
  LoginPage,
  LTIAttemptPageConfirm,
  LTIAttemptsPage,
  LTIAttemptsPageCreate,
  LTIAttemptsPageEdit,
  LTIRoutingPage,
  LTIRoutingPageEdit,
  ProfilePagePasswordChange,
  RolesListPage,
  RolesPageEdit,
  RouteSuspense,
  ServersPage,
  ServersPageEdit,
  ServiceCardsPage,
  ServiceCardsPageEdit,
  SessionPage,
  SessionWorkspaceRedirectPage,
  TargetsPage,
  TargetsPageEdit,
  TopologyPageView
} from '@/app/routes/lazy-pages';

const RoutesAdmin = () => {
  return (
    <RouteSuspense>
      <Routes>
        <Route path={RoutesLocation.accounts()} element={<AccountsPage />} />
        <Route path={RoutesLocation.accountsEdit()} element={<AccountsPageEdit />} />
        <Route path={RoutesLocation.accountsCreate()} element={<AccountsPageEdit />} />
        <Route path={RoutesLocation.profileChangePassword()} element={<ProfilePagePasswordChange />} />
        <Route path={RoutesLocation.roles()} element={<RolesListPage />} />
        <Route path={RoutesLocation.rolesEdit()} element={<RolesPageEdit />} />
        <Route path={RoutesLocation.rolesCreate()} element={<RolesPageEdit />} />
        <Route path={RoutesLocation.authProviders()} element={<AuthProvidersPage />} />
        <Route path={RoutesLocation.authProvidersEdit()} element={<AuthProvidersPageEdit />} />
        <Route path={RoutesLocation.authProvidersCreate()} element={<AuthProvidersPageEdit />} />
        <Route path={RoutesLocation.ltiRouting()} element={<LTIRoutingPage />} />
        <Route path={RoutesLocation.labs()} element={<LabCatalogPage />} />
        <Route path={RoutesLocation.ltiRoutingEdit()} element={<LTIRoutingPageEdit />} />
        <Route path={RoutesLocation.ltiRoutingCreate()} element={<LTIRoutingPageEdit />} />
        <Route path={RoutesLocation.ltiRedirect()} element={<LTIAttemptPageConfirm />} />
        <Route path={RoutesLocation.ltiRedirectCreate()} element={<LTIAttemptsPageCreate />} />
        <Route path={RoutesLocation.ltiAttemptEdit()} element={<LTIAttemptsPageEdit />} />
        <Route path={RoutesLocation.ltiAttempts()} element={<LTIAttemptsPage />} />
        <Route path={RoutesLocation.servers()} element={<ServersPage />} />
        <Route path={RoutesLocation.serversEdit()} element={<ServersPageEdit />} />
        <Route path={RoutesLocation.serversCreate()} element={<ServersPageEdit />} />
        <Route path={RoutesLocation.serviceCards()} element={<ServiceCardsPage />} />
        <Route path={RoutesLocation.serviceCardsEdit()} element={<ServiceCardsPageEdit />} />
        <Route path={RoutesLocation.serviceCardsCreate()} element={<ServiceCardsPageEdit />} />
        <Route path={RoutesLocation.targets()} element={<TargetsPage />} />
        <Route path={RoutesLocation.targetsEdit()} element={<TargetsPageEdit />} />
        <Route path={RoutesLocation.targetsCreate()} element={<TargetsPageEdit />} />
        <Route path={RoutesLocation.session()} element={<SessionPage />} />
        <Route path={RoutesLocation.sessionWorkspace()} element={<SessionWorkspaceRedirectPage />} />
        <Route path={RoutesLocation.sessionTopologyLegacy()} element={<TopologyPageView />} />
        <Route path={RoutesLocation.login()} element={<LoginPage />} />
        <Route path={RoutesLocation.language()} element={<LanguagePage />} />
        <Route path={RoutesLocation.home()} element={<HomePage />} />
      </Routes>
    </RouteSuspense>
  );
};

export default RoutesAdmin;
