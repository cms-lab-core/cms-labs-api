/*
Template auto generated with params from openapi.json
*/
import { useInfiniteQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  type UsecasesServerListRequest,
  type UsecasesServerListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { RpcTransport } from '@/helpers/api/core/request';

type Params = CamelCasedPropertiesDeep<UsecasesServerListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesServerListResponse['result']>;

export const useInfinityServerList = (params?: Params) => {
  return useInfiniteQuery({
    queryKey: RpcTransport.getInfiniteQueryKey(CoreJsonRpcPath, 'server.list', params ?? {}),
    queryFn: ({ pageParam }) => {
      return transportWithAuth.rpc<Response>(CoreJsonRpcPath, {
        method: 'server.list',
        params: {
          limit: params?.limit ?? 100,
          offset: pageParam,
          search: params?.search?.trim() || '',
          types: params?.types || [],
          orderBy: params?.orderBy || '',
          status: params?.status || ''
        } as Params
      });
    },
    refetchOnWindowFocus: false,
    refetchInterval: false,
    gcTime: 0,
    getNextPageParam: (response, pages) => {
      const totalCount = response?.totalCount || 0;
      const count = pages.flatMap((p) => p?.model).length;
      return totalCount && totalCount > count ? count : undefined;
    },
    initialPageParam: 0,
    retry: 3
  });
};
