import { CamelCasedPropertiesDeep } from 'type-fest';
import { UsecasesTargetItem, UsecasesTargetRelationItem } from '@/helpers/api';

export type TargetNodeType = CamelCasedPropertiesDeep<UsecasesTargetItem>;
export type TargetRelationType = CamelCasedPropertiesDeep<UsecasesTargetRelationItem>;

export interface TargetMapFilter {
  tags?: string[];
  layers?: string[];
  mine?: boolean;
  search?: string;
}

function createTagFilter(filterTags: string[]) {
  return (target: TargetNodeType) => {
    const targetTags = target?.taget?.tags ?? [];
    const targetInternalTags = target?.taget?.internalTags ?? [];
    const hasIntersectionWithTags = filterTags.some((tag) =>
      targetTags.some((value) => value.toLowerCase() === tag.toLowerCase())
    );
    const hasIntersectionWithInternal = filterTags.some((tag) =>
      targetInternalTags.some((value) => value.toLowerCase() === tag.toLowerCase())
    );
    return hasIntersectionWithTags || hasIntersectionWithInternal;
  };
}

export const filterTargets = (
  targets: TargetNodeType[],
  { tags, layers, mine, search }: TargetMapFilter
): TargetNodeType[] => {
  if (!targets) return [];
  const searchQuery = search?.trim().toLowerCase() || '';

  // Фильтрация по поисковому запросу
  let searchFiltered = searchQuery.trim()
    ? targets.filter((target) =>
        [
          target.taget.name?.toLowerCase(),
          target.taget.id?.toLowerCase(),
          target.taget.description?.toLowerCase(),
          ...(target.taget.tags ?? []).map((tag) => tag.toLowerCase()),
          ...(target.taget.internalTags ?? []).map((tag) => tag.toLowerCase())
        ].some((field) => field?.includes(searchQuery))
      )
    : targets;

  if (mine) {
    searchFiltered = searchFiltered.filter((target) => target?.isMine);
  }
  if (layers?.length) {
    searchFiltered = searchFiltered.filter(createTagFilter(layers));
  }

  if (tags?.length) {
    searchFiltered = searchFiltered.filter(createTagFilter(tags));
  }
  return searchFiltered;
};
