<script setup lang="ts">
import { Button as TButton, Input as TInput } from 'tdesign-vue-next';

defineProps<{
  renaming: boolean;
}>();

const renameFrom = defineModel<string>('renameFrom', { required: true });
const renameTo = defineModel<string>('renameTo', { required: true });

const emit = defineEmits<{
  (e: 'submit'): void;
}>();
</script>

<template>
  <div class="ui-dialog-section space-y-2">
    <div class="space-y-2">
      <label class="ui-dialog-label">来源标签</label>
      <TInput v-model="renameFrom" placeholder="例如: Legacy" />
    </div>
    <div class="space-y-2">
      <label class="ui-dialog-label">目标标签</label>
      <TInput v-model="renameTo" placeholder="例如: Ops" />
    </div>
    <div class="ui-dialog-actions">
      <TButton
        size="small"
        theme="primary"
        :loading="renaming"
        @click="emit('submit')"
      >
        执行重命名
      </TButton>
    </div>
  </div>
</template>
