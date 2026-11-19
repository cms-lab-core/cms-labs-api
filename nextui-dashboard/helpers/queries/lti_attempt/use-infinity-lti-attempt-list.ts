/*
Template auto generated with params from openapi.json
*/
import { useInfiniteQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  type UsecasesLTIAttemptListRequest,
  type UsecasesLTIAttemptListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';
import { RpcTransport } from '@/helpers/api/core/request';

type Params = CamelCasedPropertiesDeep<UsecasesLTIAttemptListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesLTIAttemptListResponse['result']>;

export const useInfinityLtiAttemptList = (params?: Params) => {
  return useInfiniteQuery({
    queryKey: RpcTransport.getInfiniteQueryKey(CoreJsonRpcPath, 'lti_attempt.list', params ?? {}),
    queryFn: ({ pageParam }) => {
      return transportWithAuth.rpc<Response>(CoreJsonRpcPath, {
        method: 'lti_attempt.list',
        params: {
          ...params,
          limit: params?.limit ?? 100,
          offset: pageParam,
          statuses: params?.statuses || [],
          userIds: params?.userIds || [],
          serverClientIds: params?.serverClientIds || [],
          serverIds: params?.serverIds || []
        } as Params
      });
    },
    refetchOnWindowFocus: false,
    refetchInterval: false,
    gcTime: 0,
    getNextPageParam: (response, pages) => {
      const totalCount = (response as any)?.totalCount || 0;
      const count = pages.flatMap((p) => p?.model ?? []).length;
      return response?.model?.length && totalCount > count ? count : undefined;
    },
    initialPageParam: 0,
    retry: 3
  });
};
