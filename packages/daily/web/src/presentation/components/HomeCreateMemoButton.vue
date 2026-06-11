<script setup lang="ts">
import { Button as TButton } from 'tdesign-vue-next';
import { Plus } from 'lucide-vue-next';

const emit = defineEmits<{
  (e: 'open'): void;
}>();
</script>

<template>
  <TButton
    shape="circle"
    size="large"
    theme="primary"
    class="ui-btn-fab-lg"
    @click="emit('open')"
  >
    <template #icon>
      <Plus :size="24" />
    </template>
  </TButton>
</template>
