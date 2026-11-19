/*
Template auto generated with params from openapi.json
*/
import { useQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  type UsecasesUserListRequest,
  type UsecasesUserListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<UsecasesUserListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesUserListResponse['result']>;

export const useQueryUserList = (params: Params, enabled = true) => {
  return useQuery(
    transportWithAuth.getQueryOptions<Response, Params>(CoreJsonRpcPath, 'user.list', params, { retry: 3, enabled })
  );
};
