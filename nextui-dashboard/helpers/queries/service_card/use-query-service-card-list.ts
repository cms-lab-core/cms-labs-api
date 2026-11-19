/*
Template auto generated with params from openapi.json
*/
import { useQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  type UsecasesServiceCardListRequest,
  type UsecasesServiceCardListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<UsecasesServiceCardListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesServiceCardListResponse['result']>;

export const useQueryServiceCardList = (params: Params, available = false) => {
  return useQuery(
    transportWithAuth.getQueryOptions<Response, Params>(
      CoreJsonRpcPath,
      available ? 'service_card.available' : 'service_card.list',
      { limit: 100, offset: 0, ...params },
      { retry: 3 }
    )
  );
};
