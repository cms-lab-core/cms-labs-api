import ru from './locales/ru';
import en from './locales/en';
import { usePreferences } from '@/helpers/preferences/context';

export type LanguageType = 'ru' | 'en';
const useLanguageBrowser = () => {
  const { values, update } = usePreferences();
  const lang: LanguageType =
    values.lang === 'ru' || values.lang === 'en'
      ? values.lang
      : navigator.language.toLowerCase().startsWith('en')
        ? 'en'
        : 'ru';
  return { locale: lang === 'en' ? en : ru, lang, setLang: (lang: LanguageType) => update({ lang }) };
};
export default useLanguageBrowser;
