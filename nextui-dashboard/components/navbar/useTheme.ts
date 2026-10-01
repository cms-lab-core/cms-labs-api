import { useTheme as useThemeNext } from 'next-themes';
import { useEffect, useState } from 'react';
import { useQueryUserGlobalStoreGet } from '@/helpers/queries/user/use-query-user-global-store-get';
import { useMutationUserGlobalStoreSet } from '@/helpers/queries/user/use-mutation-user-global-store-set';

export type ThemeType = 'dark' | 'light';

const useThemeBrowser = () => {
  const globalStoreQuery = useQueryUserGlobalStoreGet({});
  const { mutate } = useMutationUserGlobalStoreSet();
  const { setTheme: setBrowserTheme } = useThemeNext();
  const globalStore = globalStoreQuery.data as Record<string, unknown> | undefined;
  const storedTheme = globalStore?.theme === 'dark' ? 'dark' : 'light';
  const [localTheme, setLocalTheme] = useState<ThemeType>();
  const theme = localTheme || storedTheme;
  useEffect(() => {
    setBrowserTheme(theme);
  }, [setBrowserTheme, theme]);
  return {
    theme,
    setTheme: (theme: ThemeType) => {
      setLocalTheme(theme);
      setBrowserTheme(theme);
      mutate({ ...(globalStore || {}), theme });
    }
  };
};
export default useThemeBrowser;
