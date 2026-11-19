import { Button, Input, Listbox, ListboxItem, Popover, PopoverContent, PopoverTrigger } from '@heroui/react';
import { ChevronDown, Search } from 'lucide-react';
import { type Key, type ReactNode, useCallback, useMemo, useState } from 'react';

export interface FilterSelectorItem {
  key: string;
  value: string;
}

interface FilterSelectorProps {
  /** Все возможные варианты */
  defaultValues: FilterSelectorItem[];
  /** Текущие выбранные */
  values: FilterSelectorItem[];
  /** Колбэк при изменении */
  onChange: (values: FilterSelectorItem[]) => void;
  /** Режим множественного выбора */
  multiSelect: boolean;
  /** Текст для варианта «Все» (пустая выборка) */
  textAll?: string;
  /** Лейбл на кнопке при множественном выборе, напр. "Слои" → "Слои (2)" */
  textSelected?: string;
  /** Показать поле поиска в выпадашке */
  showSearch?: boolean;
  /** Иконка слева на кнопке */
  startContent?: ReactNode;
  /** className для кнопки-триггера */
  className?: string;
}

const ALL_KEY = '__all__';

export const FilterSelector = ({
  defaultValues,
  values,
  onChange,
  multiSelect,
  textAll = 'Все',
  textSelected,
  showSearch = false,
  startContent,
  className
}: FilterSelectorProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');

  // --- Элементы списка (с фильтрацией и «Все» наверху) ---
  const listItems = useMemo(() => {
    let items = defaultValues;

    if (search.trim()) {
      const q = search.trim().toLowerCase();
      items = items.filter((item) => item.value.toLowerCase().includes(q));
      // при поиске не показываем «Все»
      return items;
    }

    return [{ key: ALL_KEY, value: textAll }, ...items];
  }, [defaultValues, search, textAll]);

  // --- Текст кнопки ---
  const label = useMemo(() => {
    if (values.length === 0) return textAll;
    if (values.length === 1) return values[0].value;
    if (textSelected) return `${textSelected} (${values.length})`;
    return values.map((v) => v.value).join(', ');
  }, [values, textAll, textSelected, multiSelect]);

  // --- Ключи для Listbox ---
  const selectedKeys = useMemo((): Set<Key> => {
    if (values.length === 0) return new Set([ALL_KEY]);
    return new Set(values.map((v) => v.key));
  }, [values]);

  // --- Обработчик выбора ---
  const handleSelectionChange = useCallback(
    (keys: 'all' | Set<Key>) => {
      // «select all» шорткат Listbox — выбрать все реальные элементы
      if (keys === 'all') {
        onChange([...defaultValues]);
        return;
      }

      const newKeys = new Set(keys);
      const wasAllSelected = values.length === 0;
      const nowHasAll = newKeys.has(ALL_KEY);

      // Кликнули «Все» при наличии выбранных → сброс
      if (nowHasAll && !wasAllSelected) {
        onChange([]);
        if (!multiSelect) setIsOpen(false);
        return;
      }

      newKeys.delete(ALL_KEY);

      if (newKeys.size === 0) {
        onChange([]);
        if (!multiSelect) setIsOpen(false);
        return;
      }

      const selected = defaultValues.filter((item) => newKeys.has(item.key));

      if (multiSelect) {
        onChange(selected);
      } else {
        onChange(selected.slice(0, 1));
        setIsOpen(false);
      }
    },
    [defaultValues, multiSelect, onChange, values]
  );

  return (
    <Popover
      isOpen={isOpen}
      onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) setSearch('');
      }}
      placement='bottom-start'
      offset={4}
    >
      <PopoverTrigger>
        <Button
          variant='flat'
          size='sm'
          startContent={startContent}
          endContent={
            <ChevronDown
              className={`w-3.5 h-3.5 shrink-0 opacity-50 transition-transform duration-200 ${
                isOpen ? 'rotate-180' : ''
              }`}
            />
          }
          className={className}
        >
          <span className='truncate'>{label}</span>
        </Button>
      </PopoverTrigger>

      <PopoverContent className='p-0 min-w-[50px]'>
        <div
          className='flex flex-col'
          onKeyDownCapture={(event) => {
            if (event.key === 'Escape') {
              event.stopPropagation();
              setIsOpen(false);
              setSearch('');
            }
          }}
        >
          {/* --- Поиск --- */}
          {showSearch && (
            <div className='p-2'>
              <Input
                size='sm'
                variant='flat'
                placeholder='Поиск...'
                value={search}
                onValueChange={setSearch}
                startContent={<Search className='w-3.5 h-3.5 opacity-50' />}
                isClearable
                onClear={() => setSearch('')}
                classNames={{ inputWrapper: 'h-8' }}
              />
            </div>
          )}

          {/* --- Список --- */}
          <Listbox
            aria-label='Filter options'
            items={listItems}
            selectionMode={multiSelect ? 'multiple' : 'single'}
            /*@ts-ignore*/
            selectedKeys={selectedKeys}
            onSelectionChange={handleSelectionChange}
            emptyContent='Ничего не найдено'
            classNames={{
              list: 'max-h-[240px] overflow-auto py-1'
            }}
          >
            {(item) => (
              <ListboxItem key={item.key} textValue={item.value}>
                {item.value}
              </ListboxItem>
            )}
          </Listbox>
        </div>
      </PopoverContent>
    </Popover>
  );
};
