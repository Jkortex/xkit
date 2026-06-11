import { MessagePlugin } from 'tdesign-vue-next';
import type { TagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';
import type { TagGovernanceDeps } from '@/presentation/composables/tag-governance/actions/tagGovernanceDeps';

export interface TagGovernanceMutationActions {
  handleRenameTag: () => Promise<void>;
  handleMergeTags: () => Promise<void>;
  handleUpsertAlias: () => Promise<void>;
  handleDeleteAlias: (alias: string) => Promise<void>;
}

interface MutationContext {
  state: TagGovernanceState;
  deps: TagGovernanceDeps;
  onChanged: () => void;
  refreshAliasAndAudit: () => void;
}

const parseMergeSources = (raw: string): string[] =>
  raw
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);

export const createTagGovernanceMutationActions = ({
  state,
  deps,
  onChanged,
  refreshAliasAndAudit,
}: MutationContext): TagGovernanceMutationActions => {
  const handleRenameTag = async (): Promise<void> => {
    const from = state.renameFrom.value.trim();
    const to = state.renameTo.value.trim();
    if (!from || !to) {
      MessagePlugin.warning('请填写来源标签和目标标签');
      return;
    }
    state.renaming.value = true;
    const result = await deps.memoGateway.renameTag(from, to);
    state.renaming.value = false;
    if (result.kind === 'failure') {
      MessagePlugin.error(result.error.message || '标签治理失败');
      return;
    }
    const data = result.value;
    MessagePlugin.success(
      data.merged
        ? `已合并为 #${data.to}，影响 ${data.affected_memos} 条笔记`
        : `已重命名为 #${data.to}，影响 ${data.affected_memos} 条笔记`,
    );
    state.showTagManageDialog.value = false;
    onChanged();
  };

  const handleMergeTags = async (): Promise<void> => {
    const sources = parseMergeSources(state.mergeSources.value);
    const target = state.mergeTarget.value.trim();
    if (sources.length === 0 || !target) {
      MessagePlugin.warning('请填写来源标签列表和目标标签');
      return;
    }
    state.merging.value = true;
    const result = await deps.memoGateway.mergeTags(sources, target);
    state.merging.value = false;
    if (result.kind === 'failure') {
      MessagePlugin.error(result.error.message || '批量合并失败');
      return;
    }
    const data = result.value;
    const skipped =
      data.skipped_sources.length > 0
        ? `，跳过 ${data.skipped_sources.length} 个不存在标签`
        : '';
    MessagePlugin.success(
      `已合并 ${data.merged_sources} 个标签到 #${data.target}，影响 ${data.affected_memos} 条笔记${skipped}`,
    );
    state.showTagManageDialog.value = false;
    onChanged();
    refreshAliasAndAudit();
  };

  const handleUpsertAlias = async (): Promise<void> => {
    const alias = state.aliasInput.value.trim();
    const canonical = state.aliasCanonical.value.trim();
    if (!alias || !canonical) {
      MessagePlugin.warning('请填写别名和规范标签');
      return;
    }
    state.aliasing.value = true;
    const result = await deps.memoGateway.upsertTagAlias(alias, canonical);
    state.aliasing.value = false;
    if (result.kind === 'failure') {
      MessagePlugin.error(result.error.message || '保存别名失败');
      return;
    }
    MessagePlugin.success(`别名 #${alias} 已指向 #${result.value.canonical}`);
    state.aliasInput.value = '';
    onChanged();
    refreshAliasAndAudit();
  };

  const handleDeleteAlias = async (alias: string): Promise<void> => {
    state.deletingAlias.value = alias;
    const result = await deps.memoGateway.deleteTagAlias(alias);
    state.deletingAlias.value = '';
    if (result.kind === 'failure') {
      MessagePlugin.error(result.error.message || '删除别名失败');
      return;
    }
    MessagePlugin.success(`已删除别名 #${alias}`);
    refreshAliasAndAudit();
  };

  return {
    handleRenameTag,
    handleMergeTags,
    handleUpsertAlias,
    handleDeleteAlias,
  };
};
