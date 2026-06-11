<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue';
import {
  Button as TButton,
  Dialog as TDialog,
  Tabs as TTabs,
  TabPanel as TTabPanel,
} from 'tdesign-vue-next';
import { useTagGovernance } from '@/presentation/composables/useTagGovernance';
import TagGovernanceRenamePanel from '@/presentation/components/tag-governance/TagGovernanceRenamePanel.vue';
import TagGovernanceMergePanel from '@/presentation/components/tag-governance/TagGovernanceMergePanel.vue';
import TagGovernanceAliasPanel from '@/presentation/components/tag-governance/TagGovernanceAliasPanel.vue';
import TagGovernanceAuditPanel from '@/presentation/components/tag-governance/TagGovernanceAuditPanel.vue';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';

export interface TagGovernanceDialogExpose {
  open: (tagName?: string) => void;
}

const emit = defineEmits<{
  (e: 'changed'): void;
}>();

type GovernanceTab = 'rename' | 'merge' | 'alias' | 'audit';
const activeTab = ref<GovernanceTab>('rename');

const {
  showTagManageDialog,
  renameFrom,
  renameTo,
  renaming,
  mergeSources,
  mergeTarget,
  merging,
  aliasInput,
  aliasCanonical,
  aliasing,
  tagAliases,
  deletingAlias,
  tagAudits,
  auditAction,
  openTagManage,
  fetchTagAudits,
  copyAuditSummary,
  formatRelativeAuditTime,
  formatAbsoluteAuditTime,
  handleRenameTag,
  handleMergeTags,
  handleUpsertAlias,
  handleDeleteAlias,
} = useTagGovernance({
  onChanged: () => {
    emit('changed');
  },
});

useDialogOpenFlag(showTagManageDialog);

const open = (tagName = '') => {
  openTagManage(tagName);
  activeTab.value = 'rename';
};

const closeDialog = (): void => {
  showTagManageDialog.value = false;
};

const executeActiveTab = async (): Promise<void> => {
  if (activeTab.value === 'rename') {
    await handleRenameTag();
    return;
  }
  if (activeTab.value === 'merge') {
    await handleMergeTags();
    return;
  }
  if (activeTab.value === 'alias') {
    await handleUpsertAlias();
    return;
  }
  await fetchTagAudits();
};

const isTypingTarget = (target: EventTarget | null): boolean => {
  const node = target as HTMLElement | null;
  if (!node) return false;
  const tag = node.tagName;
  return (
    tag === 'INPUT' ||
    tag === 'TEXTAREA' ||
    tag === 'SELECT' ||
    Boolean(node.closest('[contenteditable="true"]'))
  );
};

const tabByKey: Record<string, GovernanceTab> = {
  '1': 'rename',
  '2': 'merge',
  '3': 'alias',
  '4': 'audit',
};

const dialogHotkeys = computed(
  () => '快捷键: 1-4 切页签 · Ctrl/Cmd+Enter 执行当前页动作 · Esc 关闭',
);

const handleDialogKeydown = (event: KeyboardEvent): void => {
  if (!showTagManageDialog.value) return;
  if (event.key === 'Escape') {
    event.preventDefault();
    closeDialog();
    return;
  }
  if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
    event.preventDefault();
    void executeActiveTab();
    return;
  }
  if (event.metaKey || event.ctrlKey || event.altKey || event.shiftKey) return;
  if (isTypingTarget(event.target)) return;
  const tab = tabByKey[event.key];
  if (!tab) return;
  event.preventDefault();
  activeTab.value = tab;
};

watch(showTagManageDialog, (visible) => {
  if (visible) {
    window.addEventListener('keydown', handleDialogKeydown);
    return;
  }
  window.removeEventListener('keydown', handleDialogKeydown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleDialogKeydown);
});

defineExpose<TagGovernanceDialogExpose>({
  open,
});
</script>

<template>
  <TDialog
    v-model:visible="showTagManageDialog"
    header="标签治理"
    :footer="false"
    width="460px"
    attach="body"
    :z-index="3000"
    class="tag-governance-dialog"
    destroy-on-close
  >
    <div class="ui-dialog-body space-y-3 text-sm">
      <div class="ui-dialog-caption">
        可执行单标签重命名或批量合并。目标标签不存在时会自动创建。
      </div>
      <div class="text-xs-plus text-muted">标签治理仅影响当前账号数据。</div>
      <div class="text-xs-plus text-muted">
        {{ dialogHotkeys }}
      </div>

      <TTabs v-model="activeTab">
        <TTabPanel value="rename" label="重命名">
          <TagGovernanceRenamePanel
            v-model:rename-from="renameFrom"
            v-model:rename-to="renameTo"
            :renaming="renaming"
            @submit="handleRenameTag"
          />
        </TTabPanel>

        <TTabPanel value="merge" label="批量合并">
          <TagGovernanceMergePanel
            v-model:merge-sources="mergeSources"
            v-model:merge-target="mergeTarget"
            :merging="merging"
            @submit="handleMergeTags"
          />
        </TTabPanel>

        <TTabPanel value="alias" label="标签别名">
          <TagGovernanceAliasPanel
            v-model:alias-input="aliasInput"
            v-model:alias-canonical="aliasCanonical"
            :aliasing="aliasing"
            :tag-aliases="tagAliases"
            :deleting-alias="deletingAlias"
            @submit="handleUpsertAlias"
            @delete-alias="handleDeleteAlias"
          />
        </TTabPanel>

        <TTabPanel value="audit" label="最近治理">
          <TagGovernanceAuditPanel
            v-model:audit-action="auditAction"
            :tag-audits="tagAudits"
            :format-relative-audit-time="formatRelativeAuditTime"
            :format-absolute-audit-time="formatAbsoluteAuditTime"
            @fetch="fetchTagAudits"
            @copy="copyAuditSummary"
          />
        </TTabPanel>
      </TTabs>

      <div class="ui-dialog-actions">
        <TButton size="small" variant="outline" @click="closeDialog">
          关闭
        </TButton>
      </div>
    </div>
  </TDialog>
</template>
