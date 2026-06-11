import { describe, expect, it } from 'vitest';
import { useTagGovernanceState } from '@/presentation/composables/tag-governance/tagGovernanceState';

describe('useTagGovernanceState', () => {
  it('creates default state', () => {
    const state = useTagGovernanceState();

    expect(state.showTagManageDialog.value).toBe(false);
    expect(state.renameFrom.value).toBe('');
    expect(state.renameTo.value).toBe('');
    expect(state.mergeSources.value).toBe('');
    expect(state.mergeTarget.value).toBe('');
    expect(state.aliasInput.value).toBe('');
    expect(state.aliasCanonical.value).toBe('');
    expect(state.tagAliases.value).toEqual([]);
    expect(state.tagAudits.value).toEqual([]);
    expect(state.auditAction.value).toBe('');
  });
});
