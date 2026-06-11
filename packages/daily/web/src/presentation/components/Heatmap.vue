<script setup lang="ts">
import { computed } from 'vue';
import type { DailyStatDTO } from '@/application/ports/dto/Stats';
import { Tooltip as TTooltip } from 'tdesign-vue-next';

const props = defineProps<{
  data: DailyStatDTO[];
}>();

// 16 weeks = 112 days for a complete 7-column grid
const DAYS = 112;
const days = computed(() => {
  const result = [];
  const now = new Date();
  for (let i = DAYS - 1; i >= 0; i--) {
    const d = new Date();
    d.setDate(now.getDate() - i);
    const dateStr = d.toISOString().split('T')[0];
    const stat = props.data.find((s) => s.date === dateStr);
    result.push({
      date: dateStr,
      count: stat ? stat.count : 0,
    });
  }
  return result;
});

const getColor = (count: number) => {
  if (count === 0) return 'heat-level-0';
  if (count < 2) return 'heat-level-1';
  if (count < 5) return 'heat-level-2';
  return 'heat-level-3';
};
</script>

<template>
  <div class="flex flex-col gap-2">
    <div class="grid grid-flow-col grid-rows-7 gap-1 w-fit mx-auto">
      <TTooltip
        v-for="day in days"
        :key="day.date"
        :content="`${day.date}: ${day.count} 条记录`"
        placement="top"
      >
        <div
          class="w-3 h-3 rounded-[2px] transition-all hover:ring-1 hover:ring-primary cursor-crosshair"
          :class="getColor(day.count)"
        ></div>
      </TTooltip>
    </div>
    <div
      class="flex justify-between text-tiny text-muted px-1 uppercase tracking-tighter font-bold"
    >
      <span>3 months ago</span>
      <span>Today</span>
    </div>
  </div>
</template>

<style scoped>
.heat-level-0 {
  background: var(--daily-heat-0);
}

.heat-level-1 {
  background: var(--daily-heat-1);
}

.heat-level-2 {
  background: var(--daily-heat-2);
}

.heat-level-3 {
  background: var(--daily-heat-3);
}
</style>
