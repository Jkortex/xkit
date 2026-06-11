<script setup lang="ts">
import { Button as TButton, Input as TInput } from 'tdesign-vue-next';
import type { TagAliasVM } from '@/presentation/view-models/TagGovernanceVM';

defineProps<{
  aliasing: boolean;
  tagAliases: TagAliasVM[];
  deletingAlias: string;
}>();

const aliasInput = defineModel<string>('aliasInput', { required: true });
const aliasCanonical = defineModel<string>('aliasCanonical', {
  required: true,
});

const emit = defineEmits<{
  (e: 'submit'): void;
  (e: 'delete-alias', alias: string): void;
}>();
</script>

<template>
  <div class="ui-dialog-section space-y-2">
    <div class="space-y-2">
      <label class="ui-dialog-label">别名</label>
      <TInput v-model="aliasInput" placeholder="例如: SRE" />
    </div>
    <div class="space-y-2">
      <label class="ui-dialog-label">规范标签</label>
      <TInput v-model="aliasCanonical" placeholder="例如: Ops" />
    </div>
    <div class="ui-dialog-actions">
      <TButton
        size="small"
        theme="primary"
        :loading="aliasing"
        @click="emit('submit')"
      >
        保存别名
      </TButton>
    </div>
    <div class="ui-list-shell max-h-40">
      <div v-if="tagAliases.length === 0" class="px-3 py-2 text-xs text-muted">
        暂无别名
      </div>
      <div
        v-for="item in tagAliases"
        :key="`${item.alias}->${item.canonical}`"
        class="ui-list-row items-center"
      >
        <span class="text-xs text-primary-text">
          #{{ item.alias }} -> #{{ item.canonical }}
        </span>
        <button
          class="text-xs text-secondary hover:text-red-500"
          :disabled="deletingAlias === item.alias"
          @click="emit('delete-alias', item.alias)"
        >
          {{ deletingAlias === item.alias ? '删除中...' : '删除' }}
        </button>
      </div>
    </div>
  </div>
</template>
