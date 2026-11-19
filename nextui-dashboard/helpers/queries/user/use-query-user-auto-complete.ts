import { InputProps } from '@heroui/input/dist/input';
import { useQueryUserList } from '@/helpers/queries/user/use-query-user-list';

export const useQueryUserAutoComplete = (search: string, props: InputProps) => {
  const id = Number(props.value);
  const hasSelection = Number.isSafeInteger(id) && id > 0;
  const queryList = useQueryUserList({
    search: search,
    limit: 20,
    offset: 0
  });
  const queryListItem = useQueryUserList(
    {
      userIds: hasSelection ? [id] : [],
      limit: 1,
      offset: 0
    },
    hasSelection
  );
  const entitiesSearch = queryList?.data?.model || [];
  const entityCurrent = hasSelection ? queryListItem?.data?.model?.[0] : undefined;
  const entities =
    entityCurrent && !entitiesSearch.find((entity) => entity.model.id === entityCurrent.model.id)
      ? [...entitiesSearch, entityCurrent]
      : entitiesSearch;
  // Объединяем данные и статусы загрузки
  const isLoading = queryList.isLoading || queryListItem?.isLoading;
  const items = entities.map((entity) => {
    return {
      key: entity.model.id || 0,
      value: (entity.model.name || '') + (entity.model.deletedAt ? ' *' : '')
    };
  });

  return { isLoading, items };
};
