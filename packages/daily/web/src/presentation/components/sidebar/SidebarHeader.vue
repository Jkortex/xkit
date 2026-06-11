<script setup lang="ts">
import { Button as TButton } from 'tdesign-vue-next';
import { Moon, Sun } from 'lucide-vue-next';

defineProps<{
  isDark: boolean;
}>();

const emit = defineEmits<{
  (e: 'toggle-theme'): void;
}>();
</script>

<template>
  <div class="flex justify-between items-center mb-4">
    <div class="text-xl font-bold flex items-center gap-2">
      <div
        class="w-8 h-8 bg-black rounded-lg flex items-center justify-center dark:bg-white"
      >
        <span class="text-white text-sm font-black dark:text-black">D</span>
      </div>
      Daily
    </div>
    <div class="flex items-center gap-2">
      <TButton
        variant="text"
        shape="circle"
        size="small"
        @click="emit('toggle-theme')"
      >
        <template #icon>
          <Sun v-if="isDark" :size="18" class="text-amber-400" />
          <Moon v-else :size="18" class="text-slate-600" />
        </template>
      </TButton>
      <slot name="account" />
    </div>
  </div>
</template>
