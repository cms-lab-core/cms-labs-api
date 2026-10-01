import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { RoutesLocation } from '@/components/routes';
import { LanguagePage, LoginPage, RouteSuspense } from '@/app/routes/lazy-pages';

const LoginRedirect = () => {
  const location = useLocation();
  const returnTo = `${location.pathname}${location.search}${location.hash}`;
  const loginSearch = new URLSearchParams({ return_to: returnTo });
  return <Navigate to={`${RoutesLocation.login()}?${loginSearch.toString()}`} replace />;
};

const RoutesUnknown = () => {
  return (
    <RouteSuspense>
      <Routes>
        <Route path={RoutesLocation.login()} element={<LoginPage />}></Route>
        <Route path={RoutesLocation.language()} element={<LanguagePage />}></Route>
        <Route path='*' element={<LoginRedirect />}></Route>
      </Routes>
    </RouteSuspense>
  );
};

export default RoutesUnknown;
