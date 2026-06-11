import { nextTick, watch, type ComponentPublicInstance, type Ref } from 'vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';

interface UseActiveMemoViewportOptions {
  memos: Ref<MemoVM[]>;
  activeIndex: Ref<number>;
  topOffset?: number;
}

const VIEWPORT_BOTTOM_GAP = 16;

export function useActiveMemoViewport(options: UseActiveMemoViewportOptions) {
  const memoAnchors = new Map<string, HTMLElement>();
  const topOffset = options.topOffset ?? 108;

  const resolveElement = (
    target: Element | ComponentPublicInstance | null,
  ): HTMLElement | null => {
    if (!target) return null;
    if (target instanceof HTMLElement) return target;
    if (!('$el' in target)) return null;
    const host = target.$el;
    return host instanceof HTMLElement ? host : null;
  };

  const bindMemoAnchor =
    (memoId: string) =>
    (target: Element | ComponentPublicInstance | null): void => {
      const element = resolveElement(target);
      if (element) {
        memoAnchors.set(memoId, element);
        return;
      }
      memoAnchors.delete(memoId);
    };

  const scrollActiveMemoIntoView = () => {
    const targetMemo = options.memos.value[options.activeIndex.value];
    if (!targetMemo) return;
    const target = memoAnchors.get(targetMemo.uuid);
    if (!target) return;
    const rect = target.getBoundingClientRect();
    const maxBottom = window.innerHeight - VIEWPORT_BOTTOM_GAP;

    if (rect.top < topOffset) {
      window.scrollBy({
        top: rect.top - topOffset - 12,
        behavior: 'smooth',
      });
      return;
    }

    if (rect.bottom > maxBottom) {
      // Prioritize top visibility: don't scroll so far that the top goes above topOffset
      const scrollAmount = Math.min(
        rect.bottom - maxBottom + 12,
        rect.top - topOffset - 12,
      );
      if (scrollAmount > 0) {
        window.scrollBy({
          top: scrollAmount,
          behavior: 'smooth',
        });
      }
    }
  };

  watch(
    () => options.activeIndex.value,
    () => {
      void nextTick(() => {
        scrollActiveMemoIntoView();
      });
    },
  );

  return {
    bindMemoAnchor,
  };
}
