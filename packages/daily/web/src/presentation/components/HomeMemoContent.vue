<script setup lang="ts">
import type { ComponentPublicInstance } from 'vue';
import { Content as TContent, Loading as TLoading } from 'tdesign-vue-next';
import MemoCard from '@/presentation/components/MemoCard.vue';
import type { MemoVM } from '@/presentation/view-models/MemoVM';

interface HomeMemoContentProps {
  readonly memos: MemoVM[];
  readonly loading: boolean;
  readonly loadingMore: boolean;
  readonly hasMore: boolean;
  readonly activeIndex: number;
  readonly bindMemoAnchor: (
    memoId: string,
  ) => (target: Element | ComponentPublicInstance | null) => void;
}

const props = defineProps<HomeMemoContentProps>();

const emit = defineEmits<{
  (e: 'edit', memo: MemoVM): void;
  (e: 'delete', id: string): void;
  (e: 'select-tag', tag: string): void;
  (e: 'view-history', id: string): void;
}>();
</script>

<template>
  <TContent class="p-4 md:p-8 flex justify-center">
    <div class="max-w-3xl w-full space-y-4">
      <div v-if="loading && props.memos.length === 0" class="space-y-4">
        <div v-for="i in 3" :key="i" class="h-28 ui-loading-card"></div>
      </div>

      <div v-else-if="props.memos.length === 0" class="ui-empty-board">
        没有命中结果，试试放宽标签条件或调整日期范围。
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="(memo, index) in props.memos"
          :key="memo.uuid"
          :ref="props.bindMemoAnchor(memo.uuid)"
        >
          <MemoCard
            :memo="memo"
            :active="index === props.activeIndex"
            @edit="emit('edit', $event)"
            @delete="emit('delete', $event)"
            @select-tag="emit('select-tag', $event)"
            @view-history="emit('view-history', $event)"
          />
        </div>

        <div class="py-8 flex justify-center">
          <TLoading v-if="loadingMore" size="small" text="加载更多中..." />
          <div v-else-if="!hasMore" class="text-xs text-muted">
            已加载完全部结果
          </div>
        </div>
      </div>
    </div>
  </TContent>
</template>
