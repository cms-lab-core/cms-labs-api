export type Preferences = Record<string, unknown>;
export interface PreferencesSnapshot {
  values: Preferences;
  ready: boolean;
  error?: unknown;
}

export function createPreferencesStore(
  read: () => Promise<Preferences>,
  write: (value: Preferences) => Promise<unknown>
) {
  let saved: Preferences = {};
  let snapshot: PreferencesSnapshot = { values: {}, ready: false };
  let initial: Promise<void> | undefined;
  let queue = Promise.resolve();
  const pending: { patch: Preferences }[] = [];
  const listeners = new Set<() => void>();
  const emit = (ready = snapshot.ready, error?: unknown) => {
    let values = { ...saved };
    for (const item of pending) values = { ...values, ...item.patch };
    snapshot = { values, ready, error };
    listeners.forEach((listener) => listener());
  };
  const load = () => {
    if (!initial) {
      initial = read()
        .then((value) => {
          saved = value || {};
          emit(true);
        })
        .catch((error) => {
          emit(true, error);
          throw error;
        });
    }
    return initial;
  };
  return {
    load,
    getSnapshot: () => snapshot,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    update: (patch: Preferences) => {
      const item = { patch: { ...patch } };
      pending.push(item);
      emit();
      const operation = queue.then(async () => {
        await load();
        const next = { ...saved, ...item.patch };
        await write(next);
        saved = next;
      });
      queue = operation.catch(() => {});
      return operation.then(
        () => {
          pending.splice(pending.indexOf(item), 1);
          emit();
        },
        (error) => {
          pending.splice(pending.indexOf(item), 1);
          emit(snapshot.ready, error);
          throw error;
        }
      );
    }
  };
}
