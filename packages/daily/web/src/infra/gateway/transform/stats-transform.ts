import type { StatsDTO } from '@/application/ports/dto/Stats';
import {
  BackendDailyStatDTO,
  BackendStatsDTO,
} from '@/infra/gateway/dto/BackendMemoDTO';

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

const isValidCount = (value: number): boolean =>
  Number.isFinite(value) && value >= 0;

const transformDailyStat = (raw: BackendDailyStatDTO) => {
  if (!DATE_PATTERN.test(raw.date)) {
    throw new Error(`invalid heatmap date: ${raw.date}`);
  }
  if (!isValidCount(raw.count)) {
    throw new Error(`invalid heatmap count: ${raw.count}`);
  }
  return {
    date: raw.date,
    count: raw.count,
  };
};

export function transformStats(raw: BackendStatsDTO): StatsDTO {
  if (!isValidCount(raw.memos_total)) {
    throw new Error(`invalid memos_total: ${raw.memos_total}`);
  }
  if (!isValidCount(raw.tags_total)) {
    throw new Error(`invalid tags_total: ${raw.tags_total}`);
  }
  if (!isValidCount(raw.resources_total)) {
    throw new Error(`invalid resources_total: ${raw.resources_total}`);
  }

  return {
    memosTotal: raw.memos_total,
    tagsTotal: raw.tags_total,
    resourcesTotal: raw.resources_total,
    heatmap: (raw.heatmap || []).map(transformDailyStat),
  };
}
