import { Route, Routes } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import {
  LabCatalogPage,
  LanguagePage,
  LoginPage,
  LTIAttemptsPageCreate,
  RouteSuspense,
  SessionPage,
  SessionWorkspaceRedirectPage,
  TargetsPage,
  TargetsPageEdit,
  TopologyPageView
} from '@/app/routes/lazy-pages';

const RoutesStudent = () => {
  return (
    <RouteSuspense>
      <Routes>
        <Route path={RoutesLocation.language()} element={<LanguagePage />} />
        <Route path={RoutesLocation.login()} element={<LoginPage />} />
        <Route path={RoutesLocation.home()} element={<LabCatalogPage />} />
        <Route path={RoutesLocation.labs()} element={<LabCatalogPage />} />
        <Route path={RoutesLocation.targets()} element={<TargetsPage />} />
        <Route path={RoutesLocation.targetsEdit()} element={<TargetsPageEdit />} />
        <Route path={RoutesLocation.targetsCreate()} element={<TargetsPageEdit />} />
        <Route path={RoutesLocation.ltiRedirect()} element={<LTIAttemptsPageCreate />} />
        <Route path={RoutesLocation.ltiRedirectCreate()} element={<LTIAttemptsPageCreate />} />
        <Route path={RoutesLocation.session()} element={<SessionPage />} />
        <Route path={RoutesLocation.sessionWorkspace()} element={<SessionWorkspaceRedirectPage />} />
        <Route path={RoutesLocation.sessionTopologyLegacy()} element={<TopologyPageView />} />
      </Routes>
    </RouteSuspense>
  );
};

export default RoutesStudent;
