<script setup lang="ts">
import { ref } from 'vue';
import { Button as TButton } from 'tdesign-vue-next';
import type { HotkeyMode } from '@xkit/hotkeys';
import { Keyboard, Plus } from 'lucide-vue-next';
import HomeFilterBar from './HomeFilterBar.vue';
import type { FilterToken } from '@/presentation/filters/types';

const inputText = defineModel<string>('inputText', { required: true });
const tokens = defineModel<FilterToken[]>('tokens', { required: true });
const showShortcuts = defineModel<boolean>('showShortcuts', { required: true });

const props = defineProps<{
  currentMode: HotkeyMode;
}>();

const filterBarRef = ref<InstanceType<typeof HomeFilterBar> | null>(null);

const emit = defineEmits<{
  (e: 'open-create'): void;
  (e: 'submit-search'): void;
  (e: 'add-token', token: Omit<FilterToken, 'id'>): void;
  (e: 'remove-token', id: string): void;
  (e: 'clear-all'): void;
}>();

const focusSearchInput = () => {
  filterBarRef.value?.focusInput();
};

defineExpose({
  focusSearchInput,
});
</script>

<template>
  <header class="ui-command-deck ui-layer-sticky sticky top-0 border-b">
    <div class="mx-auto max-w-5xl px-4 py-4 md:px-8">
      <div class="flex items-center gap-3">
        <div>
          <div class="ui-command-title">先记录，再检索</div>
          <div class="text-xs text-muted">支持 tag: syntax 或点击右侧筛选</div>
        </div>
        <span class="ui-command-pill hidden sm:inline-flex">
          <Keyboard :size="14" />
          键盘优先
        </span>
        <span class="ui-command-pill hidden sm:inline-flex">
          MODE: {{ props.currentMode.toUpperCase() }}
        </span>
        <div class="ml-auto hidden items-center gap-2 md:flex">
          <button
            class="ui-btn-icon-sm"
            @click="showShortcuts = !showShortcuts"
          >
            ?
          </button>
          <button class="ui-btn-icon-sm" @click="focusSearchInput">⌘K</button>
          <TButton
            size="medium"
            theme="primary"
            class="rounded-full!"
            @click="emit('open-create')"
          >
            <template #icon>
              <Plus :size="16" />
            </template>
            新建
          </TButton>
        </div>
      </div>

      <div class="mt-3">
        <HomeFilterBar
          ref="filterBarRef"
          v-model:input-text="inputText"
          :tokens="tokens"
          placeholder="输入关键词、标签 (tag:ops)，按 Enter 搜索"
          @submit="emit('submit-search')"
          @add-token="(t) => emit('add-token', t)"
          @remove-token="(id) => emit('remove-token', id)"
          @clear-all="emit('clear-all')"
        />
      </div>

      <div v-if="showShortcuts" class="ui-shortcut-board mt-3">
        <span><kbd>/</kbd> 聚焦搜索</span>
        <span><kbd>J</kbd>/<kbd>K</kbd> 上下选择</span>
        <span><kbd>N</kbd> 新建</span>
        <span><kbd>:</kbd> / <kbd>Ctrl/Cmd+K</kbd> 命令面板</span>
        <span
          ><kbd>G H</kbd>/<kbd>G S</kbd>/<kbd>G I</kbd>/<kbd>G T</kbd>
          页面跳转/治理</span
        >
        <span><kbd>G R</kbd> 随机漫步</span>
        <span><kbd>G U</kbd> 切换账号</span>
        <span><kbd>Alt+M</kbd> 打开账号菜单</span>
        <span><kbd>E</kbd> 编辑选中</span>
        <span><kbd>D</kbd> 删除选中</span>
        <span><kbd>Ctrl/Cmd+Shift+F</kbd> 扩展编辑区</span>
        <span><kbd>Esc</kbd> 关闭编辑并回到 Normal</span>
      </div>
    </div>
  </header>
</template>
