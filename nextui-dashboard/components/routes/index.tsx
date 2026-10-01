export const RoutesLocation = {
  login: () => '/login',
  language: () => '/lang',
  accounts: () => '/accounts',
  accountsEdit: (id = ':id') => `/accounts/edit/${id}`,
  accountsCreate: () => `/accounts/create`,
  roles: () => '/roles',
  rolesEdit: (id = ':id') => `/roles/edit/${id}`,
  rolesCreate: () => `/roles/create`,
  authProviders: () => '/auth-providers',
  authProvidersEdit: (id = ':id') => `/auth-providers/edit/${id}`,
  authProvidersCreate: () => `/auth-providers/create`,
  ltiRouting: () => '/lti-routing',
  ltiRoutingEdit: (id = ':id') => `/lti-routing/edit/${id}`,
  ltiRoutingCreate: () => `/lti-routing/create`,
  servers: () => '/servers',
  serversEdit: (id = ':id') => `/servers/edit/${id}`,
  serversCreate: () => `/servers/create`,
  serviceCards: () => '/service-cards',
  serviceCardsEdit: (id = ':id') => `/service-cards/edit/${id}`,
  serviceCardsCreate: () => `/service-cards/create`,
  profileChangePassword: () => `/profile/password`,
  targets: () => '/targets',
  targetsEdit: (id = ':id') => `/targets/edit/${id}`,
  targetsCreate: () => '/targets/create',
  labs: () => '/labs',
  ltiRedirect: () => '/lti-redirect',
  ltiRedirectCreate: () => '/lti-redirect/create',
  ltiAttemptEdit: (id = ':id') => `/lti-attempt/edit/${id}`,
  ltiAttempts: (params?: { userId?: number | number[]; statuses?: string[]; serverClientId?: string | string[] }) => {
    const url = '/lti-attempts';
    const searchParams = new URLSearchParams();
    if (params?.userId) {
      const ids = Array.isArray(params.userId) ? params.userId : [params.userId];
      ids.forEach((id) => searchParams.append('userId', id.toString()));
    }
    if (params?.statuses?.length) {
      params.statuses.forEach((s) => searchParams.append('statuses', s));
    }
    if (params?.serverClientId) {
      const ids = Array.isArray(params.serverClientId) ? params.serverClientId : [params.serverClientId];
      ids.forEach((id) => searchParams.append('serverClientId', id));
    }
    const query = searchParams.toString();
    return query ? `${url}?${query}` : url;
  },
  session: (sessionId = ':sessionId') => `/session/${sessionId}`,
  sessionWorkspace: (sessionId = ':sessionId') => `/session/${sessionId}/workspace`,
  sessionTopology: (sessionId = ':sessionId') => `/session/${sessionId}?tab=topology`,
  sessionTopologyLegacy: (sessionId = ':sessionId') => `/session/${sessionId}/topology`,
  home: () => '/'
};
