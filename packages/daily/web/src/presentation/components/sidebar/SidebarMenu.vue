<script setup lang="ts">
import { BarChart2, Hash, Sparkles } from 'lucide-vue-next';

const emit = defineEmits<{
  (e: 'random-walk'): void;
}>();
</script>

<template>
  <nav class="space-y-1 mb-4">
    <RouterLink
      to="/"
      class="ui-nav-link"
      active-class="bg-accent-soft text-accent font-medium"
    >
      <Hash :size="18" /> 全部笔记
    </RouterLink>
    <RouterLink
      to="/stats"
      class="ui-nav-link"
      active-class="bg-accent-soft text-accent font-medium"
    >
      <BarChart2 :size="18" /> 统计回顾
    </RouterLink>
    <button
      class="ui-nav-link text-left cursor-pointer w-full text-secondary"
      @click="emit('random-walk')"
    >
      <Sparkles :size="18" class="text-amber-500" /> 随机漫步
    </button>
  </nav>
</template>
