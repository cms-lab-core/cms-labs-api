/*
Template auto generated with params from openapi.json
*/
import { useQuery } from '@tanstack/react-query';
import {
  CoreJsonRpcPath,
  transportWithAuth,
  type UsecasesLTIAttemptListRequest,
  type UsecasesLTIAttemptListResponse
} from '@/helpers/api';
import { CamelCasedPropertiesDeep } from 'type-fest';

type Params = CamelCasedPropertiesDeep<UsecasesLTIAttemptListRequest['params']> & object;
type Response = CamelCasedPropertiesDeep<UsecasesLTIAttemptListResponse['result']>;
type Options = {
  refetchInterval?: number | false;
};

export const useQueryLtiAttemptList = (params: Params, options: Options = {}) => {
  return useQuery(
    transportWithAuth.getQueryOptions<Response, Params>(CoreJsonRpcPath, 'lti_attempt.list', params, {
      retry: 3,
      ...options
    })
  );
};
