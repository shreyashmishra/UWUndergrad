const hasWindow = (): boolean => typeof window !== "undefined";

export function readStorageValue<T>(key: string, fallback: T): T {
  if (!hasWindow()) {
    return fallback;
  }

  const value = window.localStorage.getItem(key);
  if (!value) {
    return fallback;
  }

  try {
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

export function writeStorageValue<T>(key: string, value: T): void {
  if (!hasWindow()) {
    return;
  }

  window.localStorage.setItem(key, JSON.stringify(value));
}
