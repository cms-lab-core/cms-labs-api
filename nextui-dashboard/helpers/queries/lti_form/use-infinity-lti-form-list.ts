import { useInfiniteQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  UsecasesAuthProviderListRequest,
  UsecasesAuthProviderListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { RpcTransport } from '@/helpers/api/core/request';

type Params = CamelCasedPropertiesDeep<UsecasesAuthProviderListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesAuthProviderListResponse['result']>;

export const useInfinityAuthProviderList = (params?: Params) => {
  return useInfiniteQuery({
    queryKey: RpcTransport.getInfiniteQueryKey(CoreJsonRpcPath, 'auth_provider.list', params ?? {}),
    queryFn: ({ pageParam }) => {
      return transportWithAuth.rpc<Response>(CoreJsonRpcPath, {
        method: 'auth_provider.list',
        params: {
          ...params,
          limit: params?.limit ?? 100,
          offset: pageParam,
          search: params?.search?.trim() || ''
        } as Params
      });
    },
    refetchOnWindowFocus: false,
    refetchInterval: false,
    gcTime: 0,
    getNextPageParam: (response, pages) => {
      const totalCount = response?.totalCount || 0;
      const count = pages.flatMap((p) => p?.model ?? []).length;
      return response?.model?.length && totalCount > count ? count : undefined;
    },
    initialPageParam: 0,
    retry: 3
  });
};
