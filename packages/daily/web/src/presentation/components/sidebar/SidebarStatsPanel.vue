<script setup lang="ts">
import Heatmap from '@/presentation/components/Heatmap.vue';
import type { StatsDTO } from '@/application/ports/dto/Stats';

defineProps<{
  stats: StatsDTO;
}>();
</script>

<template>
  <div class="space-y-2">
    <div class="grid grid-cols-2 gap-2">
      <div class="bg-page py-2 px-3 rounded-lg border border-border">
        <div class="text-muted text-tiny font-bold uppercase tracking-tight">
          Memos
        </div>
        <div class="text-base font-bold leading-tight">
          {{ stats.memosTotal }}
        </div>
      </div>
      <div class="bg-page py-2 px-3 rounded-lg border border-border">
        <div class="text-muted text-tiny font-bold uppercase tracking-tight">
          Tags
        </div>
        <div class="text-base font-bold leading-tight">
          {{ stats.tagsTotal }}
        </div>
      </div>
    </div>

    <Heatmap :data="stats.heatmap" />
  </div>
</template>
