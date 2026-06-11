<script setup lang="ts">
import { onMounted } from 'vue';
import { Layout as TLayout, Content as TContent } from 'tdesign-vue-next';
import { useStats } from '@/presentation/composables/useStats';
import Heatmap from '@/presentation/components/Heatmap.vue';

const { stats, fetchStats, loading } = useStats();

onMounted(() => {
  fetchStats();
});
</script>

<template>
  <TLayout>
    <TContent class="p-8 flex justify-center">
      <div class="max-w-4xl w-full space-y-10">
        <h1 class="text-3xl font-bold">统计回顾</h1>

        <div v-if="loading" class="animate-pulse space-y-4">
          <div class="h-40 ui-surface-card"></div>
        </div>

        <div v-else-if="stats" class="space-y-8">
          <!-- Big Heatmap -->
          <div class="ui-surface-card p-8 shadow-sm">
            <h2 class="ui-section-title mb-6">年度活跃度</h2>
            <Heatmap :data="stats.heatmap" class="scale-110 origin-left" />
          </div>

          <!-- Grid Stats -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div class="ui-stat-card">
              <div class="ui-stat-label">累计记录笔记</div>
              <div class="text-4xl font-black">{{ stats.memosTotal }}</div>
            </div>
            <div class="ui-stat-card">
              <div class="ui-stat-label">活跃标签数</div>
              <div class="text-4xl font-black">{{ stats.tagsTotal }}</div>
            </div>
            <div class="ui-stat-card">
              <div class="ui-stat-label">存储附件</div>
              <div class="text-4xl font-black">{{ stats.resourcesTotal }}</div>
            </div>
          </div>
        </div>
      </div>
    </TContent>
  </TLayout>
</template>
