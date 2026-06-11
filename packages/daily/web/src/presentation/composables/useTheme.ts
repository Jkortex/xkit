import { computed } from 'vue';
import {
  currentTheme,
  getThemeMode,
  toggleTheme as toggleThemeInternal,
} from '@/presentation/theme/themeManager';

export function useTheme() {
  const isDark = computed(() => getThemeMode(currentTheme.value) === 'dark');

  return {
    themeId: currentTheme,
    isDark,
    toggleTheme: toggleThemeInternal,
  };
}
