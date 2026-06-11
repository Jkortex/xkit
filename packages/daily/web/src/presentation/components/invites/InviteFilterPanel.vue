<script setup lang="ts">
import { Select as TSelect } from 'tdesign-vue-next';

interface StatusOption {
  label: string;
  value: 'all' | 'active' | 'used' | 'revoked' | 'expired';
}

const props = defineProps<{
  status: 'all' | 'active' | 'used' | 'revoked' | 'expired';
  statusOptions: StatusOption[];
  variant?: 'card' | 'compact' | 'toolbar';
}>();

const emit = defineEmits<{
  (
    event: 'update:status',
    value: 'all' | 'active' | 'used' | 'revoked' | 'expired',
  ): void;
}>();

const onStatusChange = (value: string | number | Array<string | number>) => {
  if (
    value === 'all' ||
    value === 'active' ||
    value === 'used' ||
    value === 'revoked' ||
    value === 'expired'
  ) {
    emit('update:status', value);
  }
};
</script>

<template>
  <div
    :class="
      props.variant === 'compact'
        ? 'invite-filter-compact invite-compact-shell'
        : props.variant === 'toolbar'
          ? 'invite-filter-toolbar'
          : 'ui-surface-card p-4'
    "
  >
    <label class="invite-field">
      <span class="invite-label">列表筛选</span>
      <TSelect
        :model-value="props.status"
        :options="props.statusOptions"
        :size="props.variant === 'card' ? 'large' : 'medium'"
        @update:model-value="onStatusChange"
      />
    </label>
  </div>
</template>

<style scoped>
.invite-field {
  display: grid;
  gap: 0.5rem;
}

.invite-filter-compact {
  padding: 0.65rem 0.75rem;
}

.invite-filter-toolbar {
  min-width: 160px;
}

.invite-compact-shell {
  border: 1px solid color-mix(in oklab, var(--daily-border) 82%, transparent);
  border-radius: 0.75rem;
  background: color-mix(in oklab, var(--daily-bg-surface) 88%, transparent);
}

.invite-label {
  min-height: 1.25rem;
  display: inline-flex;
  align-items: center;
  font-size: 0.875rem;
  color: var(--daily-text-secondary);
}
</style>
