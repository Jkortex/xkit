<script setup lang="ts">
import {
  computed,
  nextTick,
  ref,
  watch,
  type ComponentPublicInstance,
} from 'vue';

interface CommandItem {
  id: string;
  title: string;
  category: string;
  shortcut?: string;
}

const visible = defineModel<boolean>('visible', { required: true });
const query = defineModel<string>('query', { required: true });

const props = defineProps<{
  items: readonly CommandItem[];
  activeIndex: number;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'execute-active'): void;
  (e: 'move', delta: number): void;
  (e: 'execute-item', id: string): void;
}>();

const hasItems = computed(() => props.items.length > 0);
const listRef = ref<HTMLElement | null>(null);
const itemRefs = ref<Record<string, HTMLElement | null>>({});
const inputRef = ref<HTMLInputElement | null>(null);

const resolveElement = (
  target: Element | ComponentPublicInstance | null,
): HTMLElement | null => {
  if (!target) return null;
  if (target instanceof HTMLElement) return target;
  if (!('$el' in target)) return null;
  const host = target.$el;
  return host instanceof HTMLElement ? host : null;
};

const bindItemRef =
  (id: string) => (target: Element | ComponentPublicInstance | null) => {
    const element = resolveElement(target);
    if (element) {
      itemRefs.value[id] = element;
      return;
    }
    itemRefs.value[id] = null;
  };

const scrollActiveItemIntoView = () => {
  if (!visible.value) return;
  const current = props.items[props.activeIndex];
  if (!current) return;
  const node = itemRefs.value[current.id];
  if (!node || !listRef.value) return;
  node.scrollIntoView({ block: 'nearest' });
};

watch(
  () => [visible.value, props.activeIndex, props.items.length],
  () => {
    void nextTick(() => {
      if (visible.value) {
        inputRef.value?.focus();
      }
      scrollActiveItemIntoView();
    });
  },
);

const onInputKeydown = (event: KeyboardEvent): void => {
  if (event.key === 'ArrowDown') {
    event.preventDefault();
    emit('move', 1);
    return;
  }
  if (event.key === 'ArrowUp') {
    event.preventDefault();
    emit('move', -1);
    return;
  }
  if (event.key === 'Enter') {
    event.preventDefault();
    emit('execute-active');
    return;
  }
  if (event.key === 'Escape') {
    event.preventDefault();
    emit('close');
  }
};
</script>

<template>
  <div
    v-if="visible"
    class="ui-command-palette-mask"
    @click.self="emit('close')"
  >
    <div class="ui-command-palette">
      <input
        ref="inputRef"
        v-model="query"
        data-command-input="true"
        class="ui-input-md ui-command-palette-input"
        placeholder="输入命令，例如 new / search / stats"
        @keydown="onInputKeydown"
      />

      <div v-if="!hasItems" class="ui-command-palette-empty">未找到命令</div>

      <div v-else ref="listRef" class="ui-command-palette-list">
        <button
          v-for="(item, index) in items"
          :key="item.id"
          :ref="bindItemRef(item.id)"
          class="ui-command-palette-item"
          :class="{ 'is-active': index === activeIndex }"
          @click="emit('execute-item', item.id)"
        >
          <span class="ui-command-palette-title-wrap">
            <span class="ui-command-palette-title">{{ item.title }}</span>
            <span class="ui-command-palette-category">{{ item.category }}</span>
          </span>
          <span class="ui-command-palette-meta">
            <span v-if="item.shortcut" class="ui-command-palette-shortcut">{{
              item.shortcut
            }}</span>
            <span class="ui-command-palette-id">{{ item.id }}</span>
          </span>
        </button>
      </div>
    </div>
  </div>
</template>
