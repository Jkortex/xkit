export interface DailyStatDTO {
  readonly date: string; // YYYY-MM-DD
  readonly count: number;
}

export interface StatsDTO {
  readonly memosTotal: number;
  readonly tagsTotal: number;
  readonly resourcesTotal: number;
  readonly heatmap: DailyStatDTO[];
}
