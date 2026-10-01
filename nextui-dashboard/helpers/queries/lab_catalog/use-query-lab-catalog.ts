import { useQuery } from '@tanstack/react-query';
import { CoreJsonRpcPath, transportWithAuth } from '@/helpers/api';
import type { LabCatalogListResponse } from './types';

export const useQueryLabCatalog = (search: string) =>
  useQuery(
    transportWithAuth.getQueryOptions<LabCatalogListResponse, { search: string }>(
      CoreJsonRpcPath,
      'lab_catalog.list',
      { search },
      { retry: 2 }
    )
  );
