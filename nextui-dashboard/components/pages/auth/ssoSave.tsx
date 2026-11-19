import { AuthSSOAuthorizeInputDTO } from '@/helpers/api';
import dayjs from 'dayjs';
import { SSOAuthorizationParams } from '@/components/pages/auth/sso';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<AuthSSOAuthorizeInputDTO>;

interface SSORedirectSave {
  params: Params;
  createdAt: string;
}

const SSO_REDIRECT_PARAMS_KEY = 'sso_redirect_params';

export const SSOAuthorizationSave = (params: Params) => {
  if (params?.redirectUri && params?.clientId) {
    const payload: SSORedirectSave = {
      params: params,
      createdAt: dayjs().toString()
    };
    sessionStorage.setItem(SSO_REDIRECT_PARAMS_KEY, JSON.stringify(payload));
  }
};

export const SSOAuthorizationReset = () => {
  sessionStorage.setItem(SSO_REDIRECT_PARAMS_KEY, '');
};

export const SSOAuthorizationGet = (): Params | null => {
  SSOAuthorizationSave(SSOAuthorizationParams());
  const payloadString = sessionStorage.getItem(SSO_REDIRECT_PARAMS_KEY);
  try {
    const payload = JSON.parse(payloadString || '') as SSORedirectSave;
    if (
      payload.params.redirectUri &&
      payload.params.clientId &&
      dayjs().diff(dayjs(payload.createdAt), 'minute') <= 3
    ) {
      return payload.params;
    }
  } catch (err) {
    return null;
  }
  return null;
};
