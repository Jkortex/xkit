<script setup lang="ts">
import { computed } from 'vue';
import type { MemoVM } from '../view-models/MemoVM';
import {
  Tag as TTag,
  Popconfirm as TPopconfirm,
  Button as TButton,
} from 'tdesign-vue-next';
import { Clock, Edit3, Trash2, History } from 'lucide-vue-next';
import { renderMarkdown } from '@/utils/markdown';

const props = defineProps<{
  memo: MemoVM;
  active?: boolean;
}>();

const displayContent = computed(() => {
  const lines = props.memo.content.split('\n');
  if (lines.length === 0) return props.memo.content;
  const lastLine = lines[lines.length - 1]?.trim() || '';
  const isTagOnlyLine = /^(#[^\s#]+)(\s+#[^\s#]+)*$/.test(lastLine);
  if (!isTagOnlyLine) return props.memo.content;
  const next = lines.slice(0, -1).join('\n').trimEnd();
  return next;
});

const previewResources = computed(() => props.memo.resources.slice(0, 2));
const hiddenResourceCount = computed(() =>
  Math.max(props.memo.resources.length - previewResources.value.length, 0),
);

const emit = defineEmits<{
  (e: 'edit', memo: MemoVM): void;
  (e: 'delete', id: string): void;
  (e: 'view-history', id: string): void;
  (e: 'select-tag', tag: string): void;
}>();
</script>

<template>
  <div
    class="ui-surface-card ui-surface-card-hover p-5 shadow-sm group relative"
    :class="{
      'ring-2 ring-accent border-accent': active,
    }"
  >
    <!-- Action Toolbar (Top Right) -->
    <div
      class="absolute top-4 right-4 flex gap-1 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity"
    >
      <TButton
        variant="text"
        shape="circle"
        size="small"
        title="查看历史"
        @click="emit('view-history', memo.uuid)"
      >
        <template #icon
          ><History :size="14" class="text-muted hover:text-accent"
        /></template>
      </TButton>
      <TButton
        variant="text"
        shape="circle"
        size="small"
        title="编辑"
        @click="emit('edit', memo)"
      >
        <template #icon
          ><Edit3 :size="14" class="text-muted hover:text-accent"
        /></template>
      </TButton>
      <TPopconfirm
        content="确定彻底删除这条笔记吗？"
        @confirm="emit('delete', memo.uuid)"
      >
        <TButton variant="text" shape="circle" size="small" title="删除">
          <template #icon
            ><Trash2 :size="14" class="text-muted hover:text-red-500"
          /></template>
        </TButton>
      </TPopconfirm>
    </div>

    <!-- Header: Time -->
    <div class="flex items-center gap-1 text-xs text-muted mb-3">
      <Clock :size="12" />
      <span>{{ memo.relativeTime }}</span>
    </div>

    <!-- Content (Markdown) -->
    <div
      class="prose prose-sm max-w-none text-primary-text leading-relaxed mb-4 markdown-body"
      v-html="renderMarkdown(displayContent)"
    ></div>

    <!-- Tags -->
    <div v-if="memo.tags.length" class="flex flex-wrap gap-2">
      <TTag
        v-for="tag in memo.tags"
        :key="tag"
        size="small"
        variant="light"
        shape="round"
        class="cursor-pointer hover:bg-accent-soft"
        @click="emit('select-tag', tag)"
      >
        #{{ tag }}
      </TTag>
    </div>

    <div
      v-if="memo.resources.length > 0"
      class="mt-4 flex items-center gap-2 text-muted"
    >
      <div class="flex items-center gap-2">
        <div
          v-for="resource in previewResources"
          :key="resource.id"
          class="h-10 w-10 overflow-hidden rounded-md border border-border"
        >
          <img
            v-if="resource.isImage"
            :src="resource.url"
            :alt="resource.filename"
            class="h-full w-full object-cover"
          />
          <div
            v-else
            class="flex h-full w-full items-center justify-center bg-surface p-1 text-tiny"
          >
            {{ resource.filename.slice(0, 2).toUpperCase() }}
          </div>
        </div>
      </div>
      <span class="text-xs">
        {{ memo.resources.length }} 个附件
        <span v-if="hiddenResourceCount > 0">(+{{ hiddenResourceCount }})</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
/* 简单的 Markdown 样式覆盖 */
.markdown-body :deep(p) {
  margin-bottom: 0.5rem;
}
.markdown-body :deep(ul) {
  list-style-type: disc;
  padding-left: 1.25rem;
  margin-bottom: 0.5rem;
}
.markdown-body :deep(code) {
  background: var(--daily-md-inline-code-bg);
  color: var(--daily-md-inline-code-text);
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
}

/* Custom Scrollbar for Memo Content */
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--daily-border);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--daily-text-muted);
}
</style>
