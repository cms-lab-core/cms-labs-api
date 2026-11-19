import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useSyncExternalStore,
  type ReactNode
} from 'react';
import { addToast } from '@heroui/react';
import { UserProfileContext } from '@/components/providers/auth-jwt/context';
import { CoreJsonRpcPath, transportWithAuth } from '@/helpers/api';
import { createPreferencesStore, type Preferences } from './store';

const PreferencesContext = createContext<ReturnType<typeof createPreferencesStore> | undefined>(undefined);

export function PreferencesProvider({ children }: { children: ReactNode }) {
  const { profile } = useContext(UserProfileContext);
  const userId = profile?.sub;
  const activeUser = useRef(userId);
  activeUser.current = userId;
  const store = useMemo(() => {
    let guest: Preferences = {};
    const assertCurrentUser = () => {
      if (activeUser.current !== userId) throw new Error('User changed while saving preferences');
    };
    return createPreferencesStore(
      async () => {
        if (!userId) return guest;
        assertCurrentUser();
        const existing =
          (await transportWithAuth.rpc<Preferences | null>(CoreJsonRpcPath, {
            method: 'user.global_store_get',
            params: {}
          })) || {};
        assertCurrentUser();
        return existing;
      },
      async (values) => {
        assertCurrentUser();
        if (!userId) {
          guest = values;
          return;
        }
        await transportWithAuth.rpc(CoreJsonRpcPath, { method: 'user.global_store_set', params: values });
      }
    );
  }, [userId]);
  const snapshot = useSyncExternalStore(store.subscribe, store.getSnapshot);
  useEffect(() => {
    void store.load().catch(() => {});
  }, [store]);
  useEffect(() => {
    const media = window.matchMedia('(prefers-color-scheme: dark)');
    const apply = () => {
      const theme =
        snapshot.values.theme === 'dark' || snapshot.values.theme === 'light'
          ? snapshot.values.theme
          : media.matches
            ? 'dark'
            : 'light';
      document.documentElement.classList.toggle('dark', theme === 'dark');
      document.documentElement.classList.toggle('light', theme === 'light');
      document.documentElement.style.colorScheme = theme;
    };
    apply();
    media.addEventListener('change', apply);
    return () => media.removeEventListener('change', apply);
  }, [snapshot.values.theme]);
  return <PreferencesContext.Provider value={store}>{children}</PreferencesContext.Provider>;
}

export function usePreferences() {
  const store = useContext(PreferencesContext);
  if (!store) throw new Error('PreferencesProvider is required');
  const snapshot = useSyncExternalStore(store.subscribe, store.getSnapshot);
  const update = useCallback(
    (patch: Preferences) => {
      void store.update(patch).catch(() => addToast({ title: 'Не удалось сохранить настройки', color: 'danger' }));
    },
    [store]
  );
  return { ...snapshot, update };
}
