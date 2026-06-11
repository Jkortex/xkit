<script setup lang="ts">
import {
  Button as TButton,
  Popconfirm as TPopconfirm,
  Tag as TTag,
} from 'tdesign-vue-next';
import { Copy, Shield, Timer } from 'lucide-vue-next';
import type { InviteVM } from '@/infra/stores/useAuthStore';

const props = defineProps<{
  invite: InviteVM;
  code: string;
  revoking: boolean;
}>();

const emit = defineEmits<{
  (event: 'copy', code: string): void;
  (event: 'revoke', id: string): void;
}>();

const statusTagTheme = (
  status: InviteVM['status'],
): 'success' | 'default' | 'danger' | 'warning' => {
  if (status === 'active') return 'success';
  if (status === 'used') return 'default';
  if (status === 'revoked') return 'danger';
  return 'warning';
};
</script>

<template>
  <article class="ui-surface-card space-y-3 p-4">
    <div class="flex flex-wrap items-center gap-2">
      <TTag theme="primary" variant="light">
        <span class="inline-flex items-center gap-1">
          <Shield :size="12" />
          {{ props.invite.role }}
        </span>
      </TTag>
      <TTag :theme="statusTagTheme(props.invite.status)" variant="light">
        {{ props.invite.status }}
      </TTag>
      <TTag theme="default" variant="light">
        <span class="inline-flex items-center gap-1">
          <Timer :size="12" />
          到期：{{ props.invite.expiresAt }}
        </span>
      </TTag>
    </div>

    <div
      class="flex items-center gap-2 rounded-lg border border-border bg-page px-3 py-2"
    >
      <code class="min-w-0 flex-1 truncate text-sm">{{
        props.code || '邀请码不可用'
      }}</code>
      <TButton
        size="small"
        variant="outline"
        theme="primary"
        class="invite-copy-btn"
        :disabled="!props.code"
        @click="emit('copy', props.code)"
      >
        <template #icon><Copy :size="14" /></template>
        复制
      </TButton>
    </div>

    <div v-if="props.invite.status === 'active'" class="flex justify-end">
      <TPopconfirm
        content="确认撤销该邀请码？"
        @confirm="emit('revoke', props.invite.id)"
      >
        <TButton
          size="small"
          variant="outline"
          theme="danger"
          :loading="props.revoking"
        >
          撤销
        </TButton>
      </TPopconfirm>
    </div>
  </article>
</template>

<style scoped>
.invite-copy-btn {
  display: inline-flex;
  align-items: center;
}

code {
  color: var(--color-secondary);
}
</style>
