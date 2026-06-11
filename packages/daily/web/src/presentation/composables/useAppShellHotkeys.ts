import { computed, onMounted, onUnmounted } from 'vue';
import { useHotkeyRuntime } from '@xkit/hotkeys';
import type { RouteLocationNormalizedLoaded } from 'vue-router';
import { isTypingTarget } from '@/presentation/composables/hotkeys/keyboardScope';

interface UseAppShellHotkeysOptions {
  route: RouteLocationNormalizedLoaded;
}

export const useAppShellHotkeys = ({ route }: UseAppShellHotkeysOptions) => {
  const runtime = useHotkeyRuntime();
  const isShellRoute = computed(() => {
    return route.name !== 'login' && route.name !== 'invite-register';
  });
  const isShellHotkeyRoute = computed(() => {
    return isShellRoute.value && route.name !== 'home';
  });

  const syncTypingState = (target: EventTarget | null): void => {
    const typing = isTypingTarget(target);
    runtime.setFlag('isTyping', typing);
    runtime.setMode(typing ? 'insert' : 'normal');
  };

  const handleKeydown = (event: KeyboardEvent): void => {
    if (!isShellHotkeyRoute.value) {
      return;
    }

    syncTypingState(event.target);
    void runtime.handleKeydown(event);
  };

  onMounted(() => {
    document.addEventListener('keydown', handleKeydown, true);
  });

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown, true);
  });
};
