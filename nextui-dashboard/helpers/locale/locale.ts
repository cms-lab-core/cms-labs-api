import ru from './locales/ru';
import en from './locales/en';
import { useQueryUserGlobalStoreGet } from '@/helpers/queries/user/use-query-user-global-store-get';
import { useMutationUserGlobalStoreSet } from '@/helpers/queries/user/use-mutation-user-global-store-set';
import { useState } from 'react';
import { transportWithoutAuth } from '@/helpers/api';

export type LanguageType = 'ru' | 'en';

const languageStorageKey = 'cms-labs-language';

const isLanguage = (value: unknown): value is LanguageType => value === 'ru' || value === 'en';

const storedLanguage = (): LanguageType | undefined => {
  const value = localStorage.getItem(languageStorageKey);
  return isLanguage(value) ? value : undefined;
};

const browserLanguage = (): LanguageType => {
  const language = storedLanguage();
  if (language) return language;
  return navigator.language.toLowerCase().startsWith('en') ? 'en' : 'ru';
};

const languageResource = (lang: LanguageType) => {
  return lang === 'en' ? en : ru;
};

const useLanguageBrowser = () => {
  const globalStoreQuery = useQueryUserGlobalStoreGet({});
  const { mutate } = useMutationUserGlobalStoreSet();
  const [localLanguage, setLocalLanguage] = useState<LanguageType>(browserLanguage);
  const [hasLocalLanguage, setHasLocalLanguage] = useState(() => storedLanguage() !== undefined);
  const globalStore = globalStoreQuery.data as Record<string, unknown> | undefined;
  const serverLanguage = globalStore?.lang;
  const lang = hasLocalLanguage || !isLanguage(serverLanguage) ? localLanguage : serverLanguage;
  return {
    locale: languageResource(lang),
    lang,
    setLang: (lang: LanguageType) => {
      localStorage.setItem(languageStorageKey, lang);
      setLocalLanguage(lang);
      setHasLocalLanguage(true);
      if (transportWithoutAuth.userTokenAccess()?.accessToken) {
        mutate({ ...(globalStore || {}), lang });
      }
    }
  };
};
export default useLanguageBrowser;
