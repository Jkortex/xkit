<script setup lang="ts">
import { onMounted, ref, computed } from 'vue';
import {
  Button as TButton,
  Dialog as TDialog,
  Input as TInput,
  MessagePlugin,
} from 'tdesign-vue-next';
import { Plus, Trash2, PencilLine } from 'lucide-vue-next';
import { useTagSetStore } from '@/infra/stores/useTagSetStore';
import type { TagSetGroupDTO, TagSetDTO } from '@/application/ports/dto/TagSet';

const store = useTagSetStore();

const selectedGroupID = ref<string | undefined>();

const selectedGroupSets = computed(() =>
  selectedGroupID.value
    ? store.tagSets.filter((s) => s.group_id === selectedGroupID.value)
    : store.tagSets.filter((s) => !s.group_id),
);

const selectedGroupName = computed(
  () =>
    store.groups.find((g) => g.id === selectedGroupID.value)?.name || '未分组',
);

// --- Group dialogs ---
const showGroupDialog = ref(false);
const groupDialogMode = ref<'create' | 'edit'>('create');
const groupFormName = ref('');
const editingGroupID = ref<string | null>(null);

function openCreateGroup() {
  groupDialogMode.value = 'create';
  groupFormName.value = '';
  showGroupDialog.value = true;
}

function openEditGroup(group: TagSetGroupDTO) {
  groupDialogMode.value = 'edit';
  editingGroupID.value = group.id;
  groupFormName.value = group.name;
  showGroupDialog.value = true;
}

async function submitGroup() {
  const name = groupFormName.value.trim();
  if (!name) return;
  if (groupDialogMode.value === 'create') {
    const r = await store.createGroup(name);
    if (r.kind === 'failure') {
      await MessagePlugin.error(r.error.message);
    }
  } else if (editingGroupID.value) {
    const r = await store.updateGroup(editingGroupID.value, name);
    if (r.kind === 'failure') {
      await MessagePlugin.error(r.error.message);
    }
  }
  showGroupDialog.value = false;
}

async function deleteGroup(id: string) {
  const r = await store.deleteGroup(id);
  if (r.kind === 'failure') {
    await MessagePlugin.error(r.error.message);
    return;
  }
  if (selectedGroupID.value === id) selectedGroupID.value = undefined;
}

// --- TagSet dialogs ---
const showSetDialog = ref(false);
const setDialogMode = ref<'create' | 'edit'>('create');
const setForm = ref({
  name: '',
  tags_any: '',
  tags_all: '',
  tags_exclude: '',
});
const editingFormSetID = ref<string | null>(null);

function openCreateSet() {
  setDialogMode.value = 'create';
  setForm.value = { name: '', tags_any: '', tags_all: '', tags_exclude: '' };
  showSetDialog.value = true;
}

function openEditSet(set: TagSetDTO) {
  setDialogMode.value = 'edit';
  editingFormSetID.value = set.id;
  setForm.value = {
    name: set.name,
    tags_any: set.tags_any.join(', '),
    tags_all: set.tags_all.join(', '),
    tags_exclude: set.tags_exclude.join(', '),
  };
  showSetDialog.value = true;
}

function parseTags(raw: string): string[] {
  return raw
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);
}

async function submitSet() {
  const name = setForm.value.name.trim();
  if (!name) return;
  const tagsAny = parseTags(setForm.value.tags_any);
  const tagsAll = parseTags(setForm.value.tags_all);
  const tagsExclude = parseTags(setForm.value.tags_exclude);

  if (setDialogMode.value === 'create') {
    const r = await store.createTagSet({
      name,
      group_id: selectedGroupID.value,
      tags_any: tagsAny,
      tags_all: tagsAll,
      tags_exclude: tagsExclude,
    });
    if (r.kind === 'failure') {
      await MessagePlugin.error(r.error.message);
    }
  } else if (editingFormSetID.value) {
    const r = await store.updateTagSet(editingFormSetID.value, {
      name,
      tags_any: tagsAny,
      tags_all: tagsAll,
      tags_exclude: tagsExclude,
    });
    if (r.kind === 'failure') {
      await MessagePlugin.error(r.error.message);
    }
  }
  showSetDialog.value = false;
}

async function deleteSet(id: string) {
  const r = await store.deleteTagSet(id);
  if (r.kind === 'failure') {
    await MessagePlugin.error(r.error.message);
  }
}

async function load() {
  await Promise.all([store.fetchGroups(), store.fetchTagSets()]);
}

onMounted(load);
</script>

<template>
  <div>
    <div class="ui-layer-sticky">
      <div class="mx-auto w-full max-w-6xl px-4 md:px-8">
        <div class="flex items-center justify-between py-4">
          <div>
            <h1 class="text-xl font-bold">标签预设管理</h1>
            <p class="text-sm text-muted">管理标签预设分组和预设组合</p>
          </div>
        </div>
      </div>
    </div>

    <div class="mx-auto w-full max-w-6xl p-4 md:p-8">
      <div
        v-if="store.loading || store.groupLoading"
        class="ui-empty-board py-8"
      >
        加载中...
      </div>

      <div v-else class="flex gap-6">
        <!-- Left: Groups -->
        <div class="w-64 shrink-0 space-y-2">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-medium">分组</h3>
            <button
              class="ui-btn-icon-xs"
              title="新建分组"
              @click="openCreateGroup"
            >
              <Plus :size="16" />
            </button>
          </div>

          <div
            class="group-item"
            :class="{ 'is-active': selectedGroupID === undefined }"
            @click="selectedGroupID = undefined"
          >
            未分组
          </div>

          <div
            v-for="g in store.groups"
            :key="g.id"
            class="group-item"
            :class="{ 'is-active': selectedGroupID === g.id }"
          >
            <span class="flex-1 truncate" @click="selectedGroupID = g.id">{{
              g.name
            }}</span>
            <button
              class="opacity-0 group-hover:opacity-100"
              @click.stop="openEditGroup(g)"
            >
              <PencilLine :size="13" />
            </button>
            <button
              class="opacity-0 group-hover:opacity-100 text-danger"
              @click.stop="deleteGroup(g.id)"
            >
              <Trash2 :size="13" />
            </button>
          </div>
        </div>

        <!-- Right: TagSets -->
        <div class="flex-1 space-y-3">
          <div class="flex items-center justify-between">
            <h3 class="text-sm font-medium">{{ selectedGroupName }}</h3>
            <TButton size="small" @click="openCreateSet">新建预设</TButton>
          </div>

          <div
            v-if="selectedGroupSets.length === 0"
            class="text-sm text-muted py-4"
          >
            暂无预设
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="set in selectedGroupSets"
              :key="set.id"
              class="set-card"
            >
              <div class="flex items-center justify-between">
                <span class="font-medium text-sm">{{ set.name }}</span>
                <div class="flex items-center gap-1">
                  <button class="ui-btn-icon-xs" @click="openEditSet(set)">
                    <PencilLine :size="14" />
                  </button>
                  <button
                    class="ui-btn-icon-xs text-danger"
                    @click="deleteSet(set.id)"
                  >
                    <Trash2 :size="14" />
                  </button>
                </div>
              </div>
              <div class="flex flex-wrap gap-1 mt-1">
                <span
                  v-for="tag in set.tags_any"
                  :key="tag"
                  class="tag-chip tag-chip-any"
                  >{{ tag }}</span
                >
                <span
                  v-for="tag in set.tags_all"
                  :key="tag"
                  class="tag-chip tag-chip-all"
                  >+{{ tag }}</span
                >
                <span
                  v-for="tag in set.tags_exclude"
                  :key="tag"
                  class="tag-chip tag-chip-exclude"
                  >-{{ tag }}</span
                >
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Group Dialog -->
    <TDialog
      :visible="showGroupDialog"
      :header="groupDialogMode === 'create' ? '新建分组' : '编辑分组'"
      :confirm-btn="{ content: '保存', loading: false }"
      :cancel-btn="'取消'"
      @confirm="submitGroup"
      @update:visible="showGroupDialog = $event"
    >
      <TInput v-model="groupFormName" placeholder="分组名称" />
    </TDialog>

    <!-- TagSet Dialog -->
    <TDialog
      :visible="showSetDialog"
      :header="setDialogMode === 'create' ? '新建预设' : '编辑预设'"
      :confirm-btn="{ content: '保存', loading: false }"
      :cancel-btn="'取消'"
      @confirm="submitSet"
      @update:visible="showSetDialog = $event"
    >
      <div class="space-y-3">
        <TInput v-model="setForm.name" placeholder="预设名称" />
        <TInput
          v-model="setForm.tags_any"
          placeholder="包含任一标签（逗号分隔）"
        />
        <TInput
          v-model="setForm.tags_all"
          placeholder="必须包含标签（逗号分隔）"
        />
        <TInput
          v-model="setForm.tags_exclude"
          placeholder="排除标签（逗号分隔）"
        />
      </div>
    </TDialog>
  </div>
</template>

<style scoped>
.group-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 8px;
  border-radius: var(--td-radius-small);
  cursor: pointer;
  font-size: 13px;
  color: var(--td-text-color-secondary);
  transition: all 0.15s;
}

.group-item:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-text-color-primary);
}

.group-item.is-active {
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
  font-weight: 500;
}

.set-card {
  padding: 10px 12px;
  background: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: var(--td-radius-medium);
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  line-height: 18px;
}

.tag-chip-any {
  background: var(--td-bg-color-secondarycontainer);
  color: var(--td-text-color-primary);
}

.tag-chip-all {
  background: var(--td-brand-color-light);
  color: var(--td-brand-color);
}

.tag-chip-exclude {
  background: var(--td-error-color-light);
  color: var(--td-error-color);
}

.ui-btn-icon-xs {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  color: var(--td-text-color-secondary);
  transition: all 0.15s;
}

.ui-btn-icon-xs:hover {
  background: var(--td-bg-color-container-hover);
  color: var(--td-text-color-primary);
}
</style>
