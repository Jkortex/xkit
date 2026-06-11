<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { Bookmark, Plus, ChevronRight } from 'lucide-vue-next';
import { Input as TInput } from 'tdesign-vue-next';
import { useTagSetStore } from '@/infra/stores/useTagSetStore';
import type { TagSetGroupDTO, TagSetDTO } from '@/application/ports/dto/TagSet';

const router = useRouter();
const store = useTagSetStore();

const search = ref('');
const normalizedSearch = computed(() => search.value.trim().toLowerCase());

const MAX_VISIBLE_GROUPS = 6;
const showAllGroups = ref(false);

const ungrouped = computed(() =>
  store.tagSets.filter((s) => !s.group_id && matchesSearch(s)),
);

const groupedSets = computed(() =>
  store.groups
    .map((g) => ({
      group: g,
      sets: store.tagSets.filter(
        (s) => s.group_id === g.id && matchesSearch(s),
      ),
    }))
    .filter((g) => g.sets.length > 0),
);

function matchesSearch(set: TagSetDTO): boolean {
  if (!normalizedSearch.value) return true;
  return set.name.toLowerCase().includes(normalizedSearch.value);
}

async function load() {
  await Promise.all([store.fetchGroups(), store.fetchTagSets()]);
}

function applyTagSet(set: TagSetDTO) {
  const query: Record<string, string> = {};
  if (set.tags_any.length > 0) query.tagAny = set.tags_any.join(',');
  if (set.tags_all.length > 0) query.tagAll = set.tags_all.join(',');
  if (set.tags_exclude.length > 0)
    query.tagExclude = set.tags_exclude.join(',');
  router.push({ path: '/', query });
}

function openManage() {
  router.push('/tag-sets');
}

function groupMatchesSearch(g: {
  group: TagSetGroupDTO;
  sets: TagSetDTO[];
}): boolean {
  if (!normalizedSearch.value) return true;
  const kw = normalizedSearch.value;
  if (g.group.name.toLowerCase().includes(kw)) return true;
  return g.sets.some((s) => s.name.toLowerCase().includes(kw));
}

const filteredGroupedSets = computed(() =>
  groupedSets.value.filter(groupMatchesSearch),
);

const filteredUngrouped = computed(() => ungrouped.value.filter(matchesSearch));

const hasFilteredSets = computed(
  () =>
    filteredUngrouped.value.length > 0 || filteredGroupedSets.value.length > 0,
);

const visibleGrouped = computed(() => {
  if (
    showAllGroups.value ||
    filteredGroupedSets.value.length <= MAX_VISIBLE_GROUPS
  )
    return filteredGroupedSets.value;
  return filteredGroupedSets.value.slice(0, MAX_VISIBLE_GROUPS);
});

const hiddenGroupCount = computed(() => {
  return Math.max(0, filteredGroupedSets.value.length - MAX_VISIBLE_GROUPS);
});

onMounted(load);
</script>

<template>
  <div class="py-1 space-y-2">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-1.5">
        <Bookmark :size="13" class="text-muted" />
        <span class="text-tiny font-bold text-muted uppercase tracking-widest"
          >标签预设</span
        >
      </div>
      <button
        class="flex items-center justify-center w-5 h-5 rounded text-secondary hover:text-accent hover:bg-bg-color-container-hover"
        @click="openManage"
        title="管理预设"
      >
        <Plus :size="14" />
      </button>
    </div>

    <TInput v-model="search" size="small" clearable placeholder="搜索预设" />

    <div v-if="store.loading" class="py-2">
      <div class="h-4 bg-muted rounded animate-pulse" />
    </div>

    <div
      v-else-if="!hasFilteredSets"
      class="py-2 text-tiny text-muted text-center border border-dashed border-border rounded-lg"
    >
      {{ normalizedSearch ? '未匹配到预设' : '暂无预设' }}
    </div>

    <div v-else>
      <div class="grid grid-cols-2 gap-1">
        <div v-for="g in visibleGrouped" :key="g.group.id" class="group-row">
          <ChevronRight
            :size="11"
            class="shrink-0 text-muted group-row-chevron"
          />
          <span class="truncate text-xs font-medium">{{ g.group.name }}</span>
          <div class="group-float">
            <div
              v-for="set in g.sets"
              :key="set.id"
              class="group-float-item"
              @click.stop="applyTagSet(set)"
            >
              <div class="group-float-item-name">{{ set.name }}</div>
              <div v-if="set.tags_any.length > 0" class="group-float-chip-row">
                <span
                  v-for="tag in set.tags_any"
                  :key="tag"
                  class="ff-chip ff-chip-any"
                  >{{ tag }}</span
                >
              </div>
              <div v-if="set.tags_all.length > 0" class="group-float-chip-row">
                <span
                  v-for="tag in set.tags_all"
                  :key="tag"
                  class="ff-chip ff-chip-all"
                  >+{{ tag }}</span
                >
              </div>
              <div
                v-if="set.tags_exclude.length > 0"
                class="group-float-chip-row"
              >
                <span
                  v-for="tag in set.tags_exclude"
                  :key="tag"
                  class="ff-chip ff-chip-exclude"
                  >-{{ tag }}</span
                >
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="filteredUngrouped.length > 0" class="mt-1 space-y-0.5">
        <div
          v-for="set in filteredUngrouped"
          :key="set.id"
          class="preset-item"
          @click="applyTagSet(set)"
        >
          <span class="truncate text-xs font-medium">{{ set.name }}</span>
          <div class="preset-float">
            <div class="preset-float-name">{{ set.name }}</div>
            <div v-if="set.tags_any.length > 0" class="preset-float-section">
              <span class="preset-float-label">包含任一</span>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tag in set.tags_any"
                  :key="tag"
                  class="ff-chip ff-chip-any"
                  >{{ tag }}</span
                >
              </div>
            </div>
            <div v-if="set.tags_all.length > 0" class="preset-float-section">
              <span class="preset-float-label">必须包含</span>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tag in set.tags_all"
                  :key="tag"
                  class="ff-chip ff-chip-all"
                  >+{{ tag }}</span
                >
              </div>
            </div>
            <div
              v-if="set.tags_exclude.length > 0"
              class="preset-float-section"
            >
              <span class="preset-float-label">排除</span>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tag in set.tags_exclude"
                  :key="tag"
                  class="ff-chip ff-chip-exclude"
                  >-{{ tag }}</span
                >
              </div>
            </div>
          </div>
        </div>
      </div>

      <button
        v-if="hiddenGroupCount > 0"
        class="w-full text-tiny text-accent hover:underline text-center pt-0.5"
        @click="showAllGroups = true"
      >
        展开全部（+{{ hiddenGroupCount }}组）
      </button>
      <button
        v-else-if="
          showAllGroups && filteredGroupedSets.length > MAX_VISIBLE_GROUPS
        "
        class="w-full text-tiny text-accent hover:underline text-center pt-0.5"
        @click="showAllGroups = false"
      >
        收起
      </button>
    </div>
  </div>
</template>

<style scoped>
.group-row {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 8px;
  border-radius: var(--td-radius-small);
  cursor: default;
  color: var(--td-text-color-secondary);
  transition: all 0.15s;
}

.group-row:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-text-color-primary);
}

.group-row:hover .group-row-chevron {
  transform: rotate(90deg);
}

.group-row-chevron {
  transition: transform 0.15s;
}

.group-float {
  opacity: 0;
  visibility: hidden;
  transition:
    visibility 0s linear 0.15s,
    opacity 0.12s;
  position: absolute;
  left: calc(100% - 4px);
  top: -8px;
  z-index: 100;
  min-width: 200px;
  max-width: 260px;
  padding: 6px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-border-level-2-color);
  border-radius: var(--td-radius-medium);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.group-float::before {
  content: '';
  position: absolute;
  right: calc(100% - 4px);
  top: -4px;
  width: 20px;
  height: calc(100% + 8px);
}

.group-row:hover .group-float {
  opacity: 1;
  visibility: visible;
  transition-delay: 0s;
}

.group-float-item {
  padding: 6px 8px;
  border-radius: var(--td-radius-small);
  cursor: pointer;
  transition: background 0.12s;
}

.group-float-item:hover {
  background: var(--td-bg-color-container-hover);
}

.group-float-item + .group-float-item {
  border-top: 1px solid var(--td-border-level-1-color);
  margin-top: 2px;
  padding-top: 8px;
}

.group-float-item-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin-bottom: 4px;
}

.group-float-chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 3px;
  margin-top: 2px;
}

.preset-item {
  position: relative;
  padding: 5px 8px;
  border-radius: var(--td-radius-small);
  cursor: pointer;
  color: var(--td-text-color-secondary);
  transition: all 0.15s;
}

.preset-item:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-text-color-primary);
}

.preset-float {
  display: none;
  position: absolute;
  left: calc(100% + 8px);
  top: -8px;
  z-index: 100;
  min-width: 200px;
  max-width: 260px;
  padding: 10px 12px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-border-level-2-color);
  border-radius: var(--td-radius-medium);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.preset-item:hover .preset-float {
  display: block;
}

.preset-float-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--td-border-level-1-color);
}

.preset-float-section {
  margin-top: 6px;
}

.preset-float-section:first-of-type {
  margin-top: 0;
}

.preset-float-label {
  display: block;
  font-size: 9px;
  font-weight: 600;
  color: var(--td-text-color-placeholder);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 3px;
}

.ff-chip {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 10px;
  line-height: 18px;
}

.ff-chip-any {
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-secondary);
}

.ff-chip-all {
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
}

.ff-chip-exclude {
  background: var(--td-error-color-light);
  color: var(--td-error-color);
}
</style>
