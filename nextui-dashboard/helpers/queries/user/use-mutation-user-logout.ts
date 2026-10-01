/*
Template auto generated with params from openapi.json
*/
import { useMutation } from '@tanstack/react-query';
import { AuthUserLogoutRequest, CoreJsonRpcPath, transportWithAuth, transportWithoutAuth } from '@/helpers/api';
import { TMutationCustomOptions } from '@/helpers/queries/types';
import queryClient from '@/helpers/queries/base';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<AuthUserLogoutRequest['params']>;
type Response = boolean;

export const useMutationUserLogout = (options: TMutationCustomOptions<Response, Params> = {}) => {
  return useMutation<Response, unknown, Params>({
    mutationFn: (params: Params) => {
      return transportWithAuth.rpc(CoreJsonRpcPath, {
        method: 'user.logout',
        params
      });
    },
    ...options,
    async onSuccess(...args) {
      if (options.onSuccess) {
        options.onSuccess(...args);
      }
      transportWithoutAuth.userTokenClear();
      queryClient.clear();
    }
  });
};
