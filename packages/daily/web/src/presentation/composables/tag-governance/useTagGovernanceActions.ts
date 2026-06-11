import type { TagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';
import { useTagGovernanceDeps } from '@/presentation/composables/tag-governance/actions/tagGovernanceDeps';
import { createTagGovernanceQueryActions } from '@/presentation/composables/tag-governance/actions/tagGovernanceQueryActions';
import { createTagGovernanceMutationActions } from '@/presentation/composables/tag-governance/actions/tagGovernanceMutationActions';

interface UseTagGovernanceActionsOptions {
  onChanged: () => void;
}

export const useTagGovernanceActions = (
  state: TagGovernanceState,
  options: UseTagGovernanceActionsOptions,
) => {
  const deps = useTagGovernanceDeps();
  const queryActions = createTagGovernanceQueryActions(state, deps);
  const mutationActions = createTagGovernanceMutationActions({
    state,
    deps,
    onChanged: options.onChanged,
    refreshAliasAndAudit: () => {
      void queryActions.fetchTagAliases();
      void queryActions.fetchTagAudits();
    },
  });

  return {
    ...queryActions,
    ...mutationActions,
  };
};
