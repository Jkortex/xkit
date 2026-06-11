export interface TagSetGroupDTO {
  readonly id: string;
  readonly name: string;
  readonly weight: number;
  readonly created_at: string;
  readonly updated_at: string;
}

export interface TagSetDTO {
  readonly id: string;
  readonly name: string;
  readonly group_id: string | null;
  readonly tags_any: string[];
  readonly tags_all: string[];
  readonly tags_exclude: string[];
  readonly weight: number;
  readonly last_used_at: string | null;
  readonly created_at: string;
  readonly updated_at: string;
}
