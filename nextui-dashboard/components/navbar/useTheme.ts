import { useEffect, useState } from 'react';
import { usePreferences } from '@/helpers/preferences/context';
export type ThemeType = 'dark' | 'light';
const useThemeBrowser = () => {
  const { values, update } = usePreferences();
  const [systemDark, setSystemDark] = useState(() => window.matchMedia('(prefers-color-scheme: dark)').matches);
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const changed = () => setSystemDark(media.matches);
    media.addEventListener('change', changed);
    return () => media.removeEventListener('change', changed);
  }, []);
  const theme: ThemeType =
    values.theme === 'dark' || values.theme === 'light' ? values.theme : systemDark ? 'dark' : 'light';
  return { theme, setTheme: (theme: ThemeType) => update({ theme }) };
};
export default useThemeBrowser;
