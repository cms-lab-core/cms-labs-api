import React, { ChangeEvent, useState } from 'react';
import useDebounce from '@/components/sidebar/search-debounce-input';
import { InputProps } from '@heroui/input/dist/input';
import { Autocomplete, AutocompleteItem } from '@heroui/react';

interface CallbackProps {
  isLoading: boolean;
  items: { key: string | number; value: string }[];
}

export interface AutoCompleteFullProps extends InputProps {
  fetchData?: (search: string, props: InputProps) => CallbackProps;
  defaultItems?: { key: string | number; value: string }[];
}

export const AutoCompleteFull = (props: AutoCompleteFullProps) => {
  const [inputChangeValue, onInputChange] = useState('');
  const debounceChangeValue = useDebounce(inputChangeValue, 300);
  const responseSearch = props?.fetchData?.(debounceChangeValue, props as InputProps);
  const items = [...(props.defaultItems ?? []), ...(responseSearch?.items ?? [])];

  return (
    <Autocomplete
      variant={props.variant}
      className={props.className}
      items={items}
      selectedKey={props.value ? String(props.value) : null}
      label={props.label}
      labelPlacement='inside'
      description={props.description}
      placeholder={props.placeholder}
      onInputChange={onInputChange}
      onSelectionChange={(key) => {
        const event = {
          target: {
            value: key === null ? '' : String(key)
          }
        } as ChangeEvent<HTMLInputElement>;
        props.onChange?.(event);
      }}
      isLoading={responseSearch?.isLoading}
      aria-label={props.label?.toString()}
      isReadOnly={props.isReadOnly}
      isDisabled={props.isDisabled}
    >
      {(item) => <AutocompleteItem key={item.key}>{item.value}</AutocompleteItem>}
    </Autocomplete>
  );
};
