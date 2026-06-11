export type SortMode = 'created_at_desc' | 'created_at_asc' | 'updated_at_desc';

export interface MemoListQuery {
  search?: string;
  tag?: string;
  from?: string;
  to?: string;
  hasResource?: boolean;
  tagsAny?: string[];
  tagsAll?: string[];
  tagsExclude?: string[];
  sort?: SortMode;
  includeResources?: boolean;
  limit?: number;
  offset?: number;
}
