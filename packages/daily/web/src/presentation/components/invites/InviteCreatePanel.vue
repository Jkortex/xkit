<script setup lang="ts">
import {
  Button as TButton,
  InputNumber as TInputNumber,
  Select as TSelect,
} from 'tdesign-vue-next';
import { UserRoundPlus } from 'lucide-vue-next';

interface RoleOption {
  label: string;
  value: 'member' | 'admin';
}

const props = defineProps<{
  role: 'member' | 'admin';
  ttlHours: number;
  creating: boolean;
  error: string;
  roleOptions: RoleOption[];
  variant?: 'card' | 'compact' | 'toolbar';
}>();

const emit = defineEmits<{
  (event: 'update:role', value: 'member' | 'admin'): void;
  (event: 'update:ttlHours', value: number): void;
  (event: 'create'): void;
}>();

const onRoleChange = (value: string | number | Array<string | number>) => {
  if (value === 'member' || value === 'admin') {
    emit('update:role', value);
  }
};

const onTTLChange = (value: string | number | undefined) => {
  if (typeof value === 'number') {
    emit('update:ttlHours', value);
  }
};
</script>

<template>
  <div
    :class="
      props.variant === 'toolbar'
        ? 'invite-create-toolbar'
        : props.variant === 'compact'
          ? 'invite-create-compact invite-compact-shell'
          : 'ui-surface-card p-5 space-y-4'
    "
  >
    <div
      class="flex items-center gap-2 text-sm text-secondary"
      :class="{ 'sr-only': props.variant === 'toolbar' }"
    >
      <UserRoundPlus :size="16" />
      创建邀请码
    </div>
    <div
      :class="
        props.variant === 'compact'
          ? 'grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] md:items-end'
          : props.variant === 'toolbar'
            ? 'grid gap-2 sm:grid-cols-[minmax(140px,1fr)_minmax(130px,1fr)_auto] sm:items-end'
            : 'grid gap-4 md:grid-cols-2 md:items-end'
      "
    >
      <label class="invite-field">
        <span class="invite-label">目标角色</span>
        <TSelect
          :model-value="props.role"
          :options="props.roleOptions"
          :size="props.variant === 'card' ? 'large' : 'medium'"
          @update:model-value="onRoleChange"
        />
      </label>
      <label class="invite-field">
        <span class="invite-label">有效期（小时）</span>
        <TInputNumber
          :model-value="props.ttlHours"
          :size="props.variant === 'card' ? 'large' : 'medium'"
          :min="1"
          class="w-full"
          @update:model-value="onTTLChange"
        />
      </label>
      <TButton
        :size="props.variant === 'card' ? 'large' : 'medium'"
        theme="primary"
        :loading="props.creating"
        @click="emit('create')"
      >
        {{ props.creating ? '创建中...' : '新建' }}
      </TButton>
    </div>
    <p
      v-if="props.error"
      class="text-sm text-error"
      :class="{ 'text-xs': props.variant === 'toolbar' }"
    >
      {{ props.error }}
    </p>
  </div>
</template>

<style scoped>
.invite-field {
  display: grid;
  gap: 0.5rem;
}

.invite-create-compact {
  display: grid;
  gap: 0.75rem;
  padding: 0.65rem 0.75rem;
}

.invite-create-toolbar {
  display: grid;
  gap: 0.35rem;
}

.invite-compact-shell {
  border: 1px solid color-mix(in oklab, var(--color-border) 82%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in oklab, var(--color-surface) 88%, transparent);
}

.invite-label {
  min-height: 1.25rem;
  display: inline-flex;
  align-items: center;
  font-size: 0.875rem;
  color: var(--color-secondary);
}
</style>
