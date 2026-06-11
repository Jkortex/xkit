import { useTagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';
import { useTagGovernanceActions } from '@/presentation/composables/tag-governance/useTagGovernanceActions';
import { useTagGovernanceAudit } from '@/presentation/composables/tag-governance/tagGovernanceAuditUtils';

interface UseTagGovernanceOptions {
  onChanged: () => void;
}

export function useTagGovernance(options: UseTagGovernanceOptions) {
  const state = useTagGovernanceState();
  const handlers = useTagGovernanceActions(state, {
    onChanged: options.onChanged,
  });
  const audit = useTagGovernanceAudit();

  return {
    ...state,
    ...handlers,
    ...audit,
  };
}
