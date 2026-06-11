<script setup lang="ts">
import { ref } from 'vue';
import { Dialog as TDialog } from 'tdesign-vue-next';
import { Sparkles } from 'lucide-vue-next';
import MemoCard from './MemoCard.vue';
import { useRandomMemo } from '../composables/useRandomMemo';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';

const { randomMemo, fetchRandom, loading } = useRandomMemo();
const visible = ref(false);

useDialogOpenFlag(visible);

const open = () => {
  visible.value = true;
  fetchRandom();
};

defineExpose({ open });
</script>

<template>
  <TDialog
    v-model:visible="visible"
    attach="body"
    header="灵感偶遇"
    :footer="false"
    width="600px"
    destroy-on-close
  >
    <div class="py-4">
      <div
        v-if="loading"
        class="h-32 bg-surface rounded-xl animate-pulse"
      ></div>
      <div v-else-if="randomMemo">
        <MemoCard :memo="randomMemo" class="shadow-none! border-none!" />
        <div class="mt-6 flex justify-center">
          <button
            class="text-sm text-accent hover:underline flex items-center gap-1"
            @click="fetchRandom"
          >
            <Sparkles :size="14" /> 换一条试试
          </button>
        </div>
      </div>
      <div v-else class="text-center text-muted">暂无笔记可供漫步</div>
    </div>
  </TDialog>
</template>
