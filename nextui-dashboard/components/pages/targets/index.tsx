import { Button } from '@heroui/react';
import React, { useMemo, useState } from 'react';
import { House, Layers, Map, Tags, User } from 'lucide-react';
import { RoutesLocation } from '@/components/routes';
import useLanguageBrowser from '@/helpers/locale';
import { CrumbsLayout } from '@/components/layout/crumbs';
import { Link } from 'react-router-dom';
import SearchInput from '@/components/sidebar/search-input';
import { TargetsMap } from '@/app/(app)/targets/TargetsMap';
import { FilterSelector, FilterSelectorItem } from '@/components/sidebar/filter-selector';
import { useQueryTargetList } from '@/helpers/queries/target/use-query-target-list';

const LayersDefault = [
  { key: 'prod', value: 'PROD' },
  { key: 'pre', value: 'PRE' },
  { key: 'temp', value: 'TEMP' }
];

export const TargetsList = () => {
  const { locale } = useLanguageBrowser();
  const {
    locale: { Target }
  } = useLanguageBrowser();
  const crumbs = [
    {
      icon: <House className='w-5 h-5 stroke-[#969696]' />,
      name: locale.Sidebar.Home,
      href: RoutesLocation.home()
    },
    {
      icon: <Map className='w-5 h-5 stroke-[#969696]' />,
      name: Target.Title,
      href: RoutesLocation.targets()
    }
  ];
  const { data, isLoading } = useQueryTargetList({});
  const defaultTags = useMemo(() => {
    const tagsValues: Set<string> = new Set();
    data?.model?.targets?.forEach((item) => {
      item?.taget?.tags?.forEach((tag) => {
        tagsValues.add(tag);
      });
      item?.taget?.internalTags?.forEach((tag) => {
        tagsValues.add(tag);
      });
    });
    LayersDefault.forEach((tag) => {
      tagsValues.delete(tag.key);
      tagsValues.delete(tag.value);
    });
    const tags: FilterSelectorItem[] = [];
    tagsValues.forEach((tag) => {
      tags.push({ key: tag, value: tag });
    });
    return tags;
  }, [data]);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [mineFilter, setMineFilter] = useState<FilterSelectorItem[]>([]);
  const [layers, setLayers] = useState<FilterSelectorItem[]>([]);
  const [tags, setTags] = useState<FilterSelectorItem[]>([]);
  return (
    <CrumbsLayout crumbs={crumbs}>
      <>
        <div className='flex justify-between flex-wrap gap-4 items-center'>
          <div className='flex items-center gap-3 flex-wrap w-full'>
            {/* Мои / Все */}
            <FilterSelector
              defaultValues={[{ key: 'mine', value: 'Мои' }]}
              values={mineFilter}
              onChange={setMineFilter}
              textAll='Все'
              multiSelect={false}
              startContent={<User size={16} />}
            />

            {/* Слои */}
            <FilterSelector
              defaultValues={LayersDefault}
              values={layers}
              onChange={setLayers}
              textAll='Все слои'
              textSelected='Слои'
              multiSelect={true}
              startContent={<Layers size={16} />}
            />

            {/* Теги */}
            <FilterSelector
              defaultValues={defaultTags}
              values={tags}
              onChange={setTags}
              showSearch={true}
              textAll='Все теги'
              textSelected='Теги'
              multiSelect={true}
              startContent={<Tags size={16} />}
            />

            <div className='flex-1 min-w-[200px]'>
              <SearchInput placeholder={Target.SearchBar} setValue={setSearchTerm} />
            </div>

            {/* Кнопка добавления */}
            <Link to={RoutesLocation.targetsCreate()}>
              <Button color='primary'>{Target.ButtonAdd}</Button>
            </Link>
          </div>
        </div>
        <TargetsMap
          data={data}
          isLoading={isLoading}
          filter={{
            tags: tags.map((item) => item.key),
            layers: layers.map((item) => item.key),
            mine: !!mineFilter.find((value) => value.key === 'mine'),
            search: searchTerm
          }}
        />
      </>
    </CrumbsLayout>
  );
};
