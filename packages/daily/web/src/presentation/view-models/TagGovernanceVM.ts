export interface TagAliasVM {
  readonly alias: string;
  readonly canonical: string;
}

export interface TagAuditVM {
  readonly action: string;
  readonly summary: string;
  readonly affectedMemos: number;
  readonly createdAt: string;
}
