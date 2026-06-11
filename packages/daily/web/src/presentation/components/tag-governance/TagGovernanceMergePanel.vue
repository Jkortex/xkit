<script setup lang="ts">
import { Button as TButton, Input as TInput } from 'tdesign-vue-next';

defineProps<{
  merging: boolean;
}>();

const mergeSources = defineModel<string>('mergeSources', { required: true });
const mergeTarget = defineModel<string>('mergeTarget', { required: true });

const emit = defineEmits<{
  (e: 'submit'): void;
}>();
</script>

<template>
  <div class="ui-dialog-section space-y-2">
    <div class="space-y-2">
      <label class="ui-dialog-label">来源标签（逗号分隔）</label>
      <TInput
        v-model="mergeSources"
        placeholder="例如: Infra, Platform, LegacyOps"
      />
    </div>
    <div class="space-y-2">
      <label class="ui-dialog-label">目标标签</label>
      <TInput v-model="mergeTarget" placeholder="例如: Ops" />
    </div>
    <div class="ui-dialog-actions">
      <TButton
        size="small"
        theme="primary"
        :loading="merging"
        @click="emit('submit')"
      >
        执行批量合并
      </TButton>
    </div>
  </div>
</template>
