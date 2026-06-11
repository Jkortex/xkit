import type { TagAliasDTO, TagAuditDTO } from '@/infra/gateway/HttpMemoGateway';
import type {
  TagAliasVM,
  TagAuditVM,
} from '@/presentation/view-models/TagGovernanceVM';

export const TagGovernancePresenter = {
  toAliasViewModels(dtos: TagAliasDTO[]): TagAliasVM[] {
    return dtos.map((dto) => ({
      alias: dto.alias,
      canonical: dto.canonical,
    }));
  },
  toAuditViewModels(dtos: TagAuditDTO[]): TagAuditVM[] {
    return dtos.map((dto) => ({
      action: dto.action,
      summary: dto.summary,
      affectedMemos: dto.affected_memos,
      createdAt: dto.created_at,
    }));
  },
};
