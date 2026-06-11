export interface MemoHistoryVM {
  readonly id: string;
  readonly content: string;
  readonly tags: string[];
  readonly resourceIds: string[];
  readonly relativeTime: string;
  readonly absoluteTime: string;
}
