import { ref } from 'vue';
import type { Ref } from 'vue';
import type {
  TagAliasVM,
  TagAuditVM,
} from '@/presentation/view-models/TagGovernanceVM';

export interface TagGovernanceState {
  showTagManageDialog: Ref<boolean>;
  renameFrom: Ref<string>;
  renameTo: Ref<string>;
  renaming: Ref<boolean>;
  mergeSources: Ref<string>;
  mergeTarget: Ref<string>;
  merging: Ref<boolean>;
  aliasInput: Ref<string>;
  aliasCanonical: Ref<string>;
  aliasing: Ref<boolean>;
  tagAliases: Ref<TagAliasVM[]>;
  deletingAlias: Ref<string>;
  tagAudits: Ref<TagAuditVM[]>;
  auditAction: Ref<string>;
}

export const useTagGovernanceState = (): TagGovernanceState => {
  return {
    showTagManageDialog: ref(false),
    renameFrom: ref(''),
    renameTo: ref(''),
    renaming: ref(false),
    mergeSources: ref(''),
    mergeTarget: ref(''),
    merging: ref(false),
    aliasInput: ref(''),
    aliasCanonical: ref(''),
    aliasing: ref(false),
    tagAliases: ref<TagAliasVM[]>([]),
    deletingAlias: ref(''),
    tagAudits: ref<TagAuditVM[]>([]),
    auditAction: ref(''),
  };
};
