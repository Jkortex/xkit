import { ref } from 'vue';

export const THEME_PRESETS = {
  'calm-light': { mode: 'light', partner: 'calm-dark' },
  'calm-dark': { mode: 'dark', partner: 'calm-light' },
} as const;

export type ThemeId = keyof typeof THEME_PRESETS;
export type ThemeMode = (typeof THEME_PRESETS)[ThemeId]['mode'];

const STORAGE_KEY = 'DAILY_THEME';
const DEFAULT_LIGHT_THEME: ThemeId = 'calm-light';
const DEFAULT_DARK_THEME: ThemeId = 'calm-dark';

const hasTheme = (value: string | null): value is ThemeId =>
  Boolean(value && value in THEME_PRESETS);

const resolveInitialTheme = (): ThemeId => {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (hasTheme(stored)) return stored;

  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  return prefersDark ? DEFAULT_DARK_THEME : DEFAULT_LIGHT_THEME;
};

export const currentTheme = ref<ThemeId>(DEFAULT_LIGHT_THEME);

export const getThemeMode = (themeId: ThemeId): ThemeMode =>
  THEME_PRESETS[themeId].mode;

export const applyTheme = (themeId: ThemeId): void => {
  const mode = getThemeMode(themeId);
  const root = document.documentElement;

  currentTheme.value = themeId;
  root.setAttribute('data-theme', themeId);
  root.setAttribute('theme-mode', mode);
  root.classList.toggle('dark', mode === 'dark');
  localStorage.setItem(STORAGE_KEY, themeId);
};

export const initTheme = (): void => {
  applyTheme(resolveInitialTheme());
};

export const toggleTheme = (): void => {
  const nextTheme = THEME_PRESETS[currentTheme.value].partner;
  applyTheme(nextTheme);
};
