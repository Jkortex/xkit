import type { TagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';
import type { TagGovernanceDeps } from '@/presentation/composables/tag-governance/actions/tagGovernanceDeps';
import { TagGovernancePresenter } from '@/presentation/presenters/TagGovernancePresenter';

export interface TagGovernanceQueryActions {
  fetchTagAliases: () => Promise<void>;
  fetchTagAudits: () => Promise<void>;
  openTagManage: (tagName?: string) => void;
}

export const createTagGovernanceQueryActions = (
  state: TagGovernanceState,
  deps: TagGovernanceDeps,
): TagGovernanceQueryActions => {
  const fetchTagAliases = async (): Promise<void> => {
    const result = await deps.memoGateway.listTagAliases();
    if (result.kind === 'failure') return;
    state.tagAliases.value = TagGovernancePresenter.toAliasViewModels(
      result.value,
    );
  };

  const fetchTagAudits = async (): Promise<void> => {
    const result = await deps.memoGateway.listTagAudits(
      12,
      state.auditAction.value,
    );
    if (result.kind === 'failure') return;
    state.tagAudits.value = TagGovernancePresenter.toAuditViewModels(
      result.value,
    );
  };

  const openTagManage = (tagName = ''): void => {
    state.renameFrom.value = tagName;
    state.renameTo.value = '';
    state.mergeSources.value = tagName;
    state.mergeTarget.value = '';
    state.aliasInput.value = tagName;
    state.aliasCanonical.value = '';
    void fetchTagAliases();
    void fetchTagAudits();
    state.showTagManageDialog.value = true;
  };

  return {
    fetchTagAliases,
    fetchTagAudits,
    openTagManage,
  };
};
