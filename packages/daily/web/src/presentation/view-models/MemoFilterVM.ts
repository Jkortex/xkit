export interface MemoFilterVM {
  search?: string;
  tag?: string;
  from?: string;
  to?: string;
  hasResource?: boolean;
  tagsAny?: string[];
  tagsAll?: string[];
  tagsExclude?: string[];
  sort?: 'created_at_desc' | 'created_at_asc' | 'updated_at_desc';
  includeResources?: boolean;
}
