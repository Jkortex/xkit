export type SortMode = 'created_at_desc' | 'created_at_asc' | 'updated_at_desc';

export type FilterTokenType =
  | 'search'
  | 'text'
  | 'tag'
  | 'tags_any'
  | 'tags_all'
  | 'tags_exclude'
  | 'from'
  | 'to'
  | 'has_resource'
  | 'sort';

export interface FilterToken {
  id: string;
  type: FilterTokenType;
  value: any;
  label: string;
}

export interface HomeFilterRouteFields {
  searchText: string;
  tagAny: string;
  tagAll: string;
  tagExclude: string;
  fromDate: string;
  toDate: string;
  sortMode: SortMode;
  hasResource: boolean | undefined;
}
