import { useQueryLabCatalog } from './use-query-lab-catalog';

// The catalog is empty while LAB_CATALOG_ENABLED is off or no laboratory has
// been imported, and the frontend hides the section in that case.
export const useIsLabCatalogVisible = () => {
  const query = useQueryLabCatalog('');
  return !query.isLoading && (query.data?.labs?.length ?? 0) > 0;
};
