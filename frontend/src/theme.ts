import { theme } from 'antd';

export type ThemeMode = 'system' | 'light' | 'dark';

export function systemPrefersDark(): boolean {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false;
}

export function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
  if (mode === 'system') {
    return systemPrefersDark() ? 'dark' : 'light';
  }
  return mode;
}

export function themeAlgorithm(mode: 'light' | 'dark') {
  return mode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm;
}
