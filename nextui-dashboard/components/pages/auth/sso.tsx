import { AuthSSOAuthorizeInputDTO, AuthSSOAuthorizeOutputDTO } from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';

const generateCodeVerifier = (): string => {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return base64URLEncode(array);
};

const base64URLEncode = (buffer: Uint8Array): string => {
  let str = '';
  buffer.forEach((byte) => {
    str += String.fromCharCode(byte);
  });
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
};

const sha256 = async (plain: string): Promise<string> => {
  const encoder = new TextEncoder();
  const data = encoder.encode(plain);
  const hash = await crypto.subtle.digest('SHA-256', data);
  return base64URLEncode(new Uint8Array(hash));
};

const generateCodeChallenge = async (verifier: string): Promise<string> => {
  return sha256(verifier);
};

export const getCodeVerifier = (): string | null => {
  return sessionStorage.getItem('pkce_code_verifier');
};

export const setCodeVerifier = (verifier: string): void => {
  sessionStorage.setItem('pkce_code_verifier', verifier);
};

export const SSOAuthorizationParams = () => {
  const currentURI = new URL(window.location.href);
  const codeChallenge = currentURI.searchParams.get('code_challenge') || '';
  const codeChallengeMethod = currentURI.searchParams.get('code_challenge_method') || '';

  let codeVerifier = getCodeVerifier();
  if (!codeVerifier && codeChallenge) {
    codeVerifier = generateCodeVerifier();
    setCodeVerifier(codeVerifier);
  }

  const params: CamelCasedPropertiesDeep<AuthSSOAuthorizeInputDTO> = {
    clientId: currentURI.searchParams.get('client_id') || '',
    path: currentURI.searchParams.get('path') || '',
    redirectUri: currentURI.searchParams.get('redirect_uri') || '',
    responseType: currentURI.searchParams.get('response_type') || '',
    scope: currentURI.searchParams.get('scope') || '',
    state: currentURI.searchParams.get('state') || '',
    nonce: currentURI.searchParams.get('nonce') || '',
    extra: currentURI.searchParams.get('extra') || '',
    codeChallenge: codeChallenge,
    codeChallengeMethod: codeChallengeMethod || 'S256'
  };
  return params;
};

export const SSOAuthorizationComplete = (params?: CamelCasedPropertiesDeep<AuthSSOAuthorizeOutputDTO>) => {
  if (!params) return '';
  const codeVerifier = getCodeVerifier();
  let url = `${params.redirectUri || '/'}?state=${params.state || ''}&path=${params.path || ''}&code=${params.code || ''}&application=${params.application || ''}&extra=${params.extra || ''}&nonce=${params.nonce || ''}&client_id=${params.clientId || ''}&scope=${params.scope}`;
  if (codeVerifier) {
    url += `&code_verifier=${codeVerifier}`;
  }
  return url;
};
