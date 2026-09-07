import { useSyncExternalStore } from "react";

export interface ExternalStore<T> {
  useValue(): T;
  get(): T;
  set(next: T): void;
}

// A minimal useSyncExternalStore-backed store, shared module-wide so every
export function createExternalStore<T>(initialValue: T, init: () => void): ExternalStore<T> {
  let value = initialValue;
  const listeners = new Set<() => void>();
  let initialized = false;

  function get(): T {
    return value;
  }

  function set(next: T): void {
    value = next;
    listeners.forEach((l) => l());
  }

  function subscribe(listener: () => void): () => void {
    if (!initialized) {
      initialized = true;
      init();
    }
    listeners.add(listener);
    return () => listeners.delete(listener);
  }

  function useValue(): T {
    return useSyncExternalStore(subscribe, get);
  }

  return { useValue, get, set };
}
