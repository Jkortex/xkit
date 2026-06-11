<script setup lang="ts">
import { computed, ref } from 'vue';
import { Input as TInput } from 'tdesign-vue-next';
import { Hash, PencilLine } from 'lucide-vue-next';
import type { TagStatDTO } from '@/infra/gateway/HttpMemoGateway';
import {
  LOW_FREQ_THRESHOLD,
  TAG_CLOUD_DEFAULT_LIMIT,
} from '@/presentation/config/tagCloud';
import TagGovernanceDialog from './TagGovernanceDialog.vue';
import type { TagGovernanceDialogExpose } from './TagGovernanceDialog.vue';

interface TagCloudPanelProps {
  tags: TagStatDTO[];
}

const props = defineProps<TagCloudPanelProps>();

const emit = defineEmits<{
  (e: 'select-tag', tagName: string): void;
  (e: 'changed'): void;
}>();

const governanceDialogRef = ref<TagGovernanceDialogExpose | null>(null);
const search = ref('');
const showAll = ref(false);
const includeLowFreq = ref(false);

function openTagGovernance() {
  governanceDialogRef.value?.open();
}

const normalizedSearch = computed(() => search.value.trim().toLowerCase());
const sortedTags = computed(() =>
  [...props.tags].sort((a, b) => {
    if (b.count !== a.count) return b.count - a.count;
    return a.name.localeCompare(b.name);
  }),
);
const effectiveTags = computed(() =>
  includeLowFreq.value
    ? sortedTags.value
    : sortedTags.value.filter((item) => item.count >= LOW_FREQ_THRESHOLD),
);
const filteredTags = computed(() => {
  const keyword = normalizedSearch.value;
  if (!keyword) return effectiveTags.value;
  return effectiveTags.value.filter((item) =>
    item.name.toLowerCase().includes(keyword),
  );
});
const visibleTags = computed(() =>
  showAll.value
    ? filteredTags.value
    : filteredTags.value.slice(0, TAG_CLOUD_DEFAULT_LIMIT),
);
const hiddenCount = computed(() =>
  Math.max(filteredTags.value.length - visibleTags.value.length, 0),
);
const lowFreqTotalCount = computed(
  () =>
    sortedTags.value.filter((item) => item.count < LOW_FREQ_THRESHOLD).length,
);

const maxCount = computed(() => {
  const first = filteredTags.value[0];
  return first ? first.count : 0;
});

function tagFontSizeClass(count: number): string {
  if (maxCount.value === 0) return 'text-xs';
  const ratio = count / maxCount.value;
  if (ratio >= 0.8) return 'text-sm font-semibold';
  if (ratio >= 0.5) return 'text-xs-plus font-medium';
  if (ratio >= 0.25) return 'text-xs font-medium';
  return 'text-tiny';
}

function tagOpacity(count: number): string {
  if (maxCount.value === 0) return 'opacity-100';
  const ratio = count / maxCount.value;
  if (ratio >= 0.6) return 'opacity-100';
  if (ratio >= 0.3) return 'opacity-75';
  return 'opacity-60';
}
</script>

<template>
  <div v-if="props.tags.length > 0" class="py-2 space-y-3">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-1.5">
        <Hash :size="13" class="text-muted" />
        <span class="text-tiny font-bold text-muted uppercase tracking-widest"
          >标签云</span
        >
      </div>
      <button
        class="flex items-center justify-center w-5 h-5 rounded text-secondary hover:text-accent hover:bg-bg-color-container-hover"
        @click="openTagGovernance"
        title="管理标签"
      >
        <PencilLine :size="13" />
      </button>
    </div>

    <TInput
      v-model="search"
      size="small"
      clearable
      placeholder="搜索标签"
      @clear="showAll = false"
    />

    <div class="flex items-center gap-2 text-tiny text-muted">
      <span>{{ visibleTags.length }} / {{ filteredTags.length }} 个标签</span>
      <span class="text-border">|</span>
      <button
        v-if="!includeLowFreq && lowFreqTotalCount > 0"
        class="text-accent hover:underline"
        @click="includeLowFreq = true"
      >
        显示低频（+{{ lowFreqTotalCount }}）
      </button>
      <button
        v-else-if="includeLowFreq && lowFreqTotalCount > 0"
        class="text-accent hover:underline"
        @click="includeLowFreq = false"
      >
        隐藏低频
      </button>
      <span v-else class="text-muted">全部</span>
    </div>

    <div class="pt-1">
      <div class="flex flex-wrap gap-x-2 gap-y-1.5">
        <button
          v-for="t in visibleTags"
          :key="t.name"
          class="tag-cloud-item"
          :class="[tagFontSizeClass(t.count), tagOpacity(t.count)]"
          @click="emit('select-tag', t.name)"
        >
          <span>{{ t.name }}</span>
          <span class="tag-count-badge">{{ t.count }}</span>
        </button>
      </div>
    </div>

    <div
      v-if="filteredTags.length === 0"
      class="text-xs text-muted text-center py-2"
    >
      未匹配到标签
    </div>

    <div v-if="hiddenCount > 0" class="text-center">
      <button
        class="text-tiny text-accent hover:underline"
        @click="showAll = true"
      >
        显示全部（+{{ hiddenCount }}）
      </button>
    </div>
    <div
      v-else-if="showAll && filteredTags.length > TAG_CLOUD_DEFAULT_LIMIT"
      class="text-center"
    >
      <button
        class="text-tiny text-accent hover:underline"
        @click="showAll = false"
      >
        收起（Top {{ TAG_CLOUD_DEFAULT_LIMIT }}）
      </button>
    </div>
  </div>

  <TagGovernanceDialog ref="governanceDialogRef" @changed="emit('changed')" />
</template>

<style scoped>
.tag-cloud-item {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 2px 0;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--td-text-color-secondary);
  transition: color 0.15s;
  line-height: 1.4;
}

.tag-cloud-item:hover {
  color: var(--td-brand-color);
}

.tag-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 14px;
  height: 14px;
  padding: 0 3px;
  border-radius: 7px;
  font-size: 9px;
  font-weight: 600;
  line-height: 1;
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-placeholder);
}
</style>
