import { useMutation } from '@tanstack/react-query';
import { CoreJsonRpcPath, transportWithAuth } from '@/helpers/api';
import type { LabCatalogStartResponse } from './types';

export const useMutationLabCatalogStart = () =>
  useMutation<LabCatalogStartResponse, unknown, { labId: number }>({
    mutationFn: (params) => transportWithAuth.rpc(CoreJsonRpcPath, { method: 'lab_catalog.start', params })
  });
