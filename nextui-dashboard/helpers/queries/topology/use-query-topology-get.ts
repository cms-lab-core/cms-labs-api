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
import { createTopologyRefetchInterval } from './topology-polling';

type Params = CamelCasedPropertiesDeep<UsecasesTopologiesGetRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesTopologiesGetResponse['result']>;

const refetchInterval = createTopologyRefetchInterval();

export const useQueryTopologyGet = (params: Params, enabled = true) => {
  return useQuery(
    transportWithAuth.getQueryOptions<Response, Params>(ClabgateJsonRpcPath, 'topology.get', params, {
      retry: 3,
      refetchOnWindowFocus: false,
      refetchOnMount: false,
      refetchInterval,
      enabled: enabled && !!params.sessionId
    })
  );
};
