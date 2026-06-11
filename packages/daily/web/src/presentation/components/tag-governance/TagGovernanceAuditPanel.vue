<script setup lang="ts">
import type { TagAuditVM } from '@/presentation/view-models/TagGovernanceVM';

defineProps<{
  tagAudits: TagAuditVM[];
  formatRelativeAuditTime: (raw: string) => string;
  formatAbsoluteAuditTime: (raw: string) => string;
}>();

const auditAction = defineModel<string>('auditAction', { required: true });

const emit = defineEmits<{
  (e: 'fetch'): void;
  (e: 'copy', summary: string): void;
}>();
</script>

<template>
  <div class="ui-dialog-section space-y-2">
    <select
      v-model="auditAction"
      class="ui-dialog-select"
      @change="emit('fetch')"
    >
      <option value="">全部操作</option>
      <option value="rename">rename</option>
      <option value="merge">merge</option>
      <option value="alias_upsert">alias_upsert</option>
      <option value="alias_delete">alias_delete</option>
    </select>
    <div class="ui-list-shell max-h-52">
      <div v-if="tagAudits.length === 0" class="px-3 py-2 text-xs text-muted">
        暂无记录
      </div>
      <div
        v-for="item in tagAudits"
        :key="`${item.action}-${item.summary}-${item.createdAt}`"
        class="space-y-1 border-b border-border px-3 py-2 last:border-b-0"
      >
        <div class="flex items-center gap-2 text-xs text-primary-text">
          <span class="truncate">[{{ item.action }}] {{ item.summary }}</span>
          <button
            class="ml-auto text-xs-plus text-secondary hover:text-accent"
            @click="emit('copy', item.summary)"
          >
            复制
          </button>
        </div>
        <div class="text-xs-plus text-secondary">
          影响 {{ item.affectedMemos }} 条 ·
          <span :title="formatAbsoluteAuditTime(item.createdAt)">
            {{ formatRelativeAuditTime(item.createdAt) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
