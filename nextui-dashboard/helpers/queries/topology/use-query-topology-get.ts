/*
Template auto generated with params from openapi.json
*/
import { useQuery } from '@tanstack/react-query';
import {
  ClabgateJsonRpcPath,
  transportWithAuth,
  type UsecasesTopologiesGetRequest,
  type UsecasesTopologiesGetResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<UsecasesTopologiesGetRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesTopologiesGetResponse['result']>;

export const useQueryTopologyGet = (params: Params, enabled = true) => {
  return useQuery(
    transportWithAuth.getQueryOptions<Response, Params>(ClabgateJsonRpcPath, 'topology.get', params, {
      retry: 3,
      refetchOnWindowFocus: true,
      refetchInterval: 30000, // 30 seconds
      enabled: enabled && !!params.sessionId
    })
  );
};
