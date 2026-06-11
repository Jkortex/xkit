/**
 * 后端 API 返回的原始笔记数据结构
 * 严格映射 server/internal/application/dto/memo_dto.go
 */
export interface BackendMemoDTO {
  uuid: string;
  content: string;
  row_status: string;
  tags: string[] | null;
  resources: BackendResourceDTO[] | null;
  expires_at: string | null;
  headline?: string;
  created_at: string; // ISO 8601
  updated_at: string;
}

export interface BackendResourceDTO {
  id: string;
  filename: string;
  size: number;
  mime_type: string;
  created_at: string;
}

export interface BackendImportSkipDetailDTO {
  entity: string;
  key: string;
  reason: string;
}

export interface BackendImportSectionReportDTO {
  imported: number;
  skipped: number;
  details: BackendImportSkipDetailDTO[];
}

export interface BackendImportReportDTO {
  message: string;
  memos_imported: number;
  resources_imported: number;
  memos_skipped: number;
  resources_skipped: number;
  report: {
    memos: BackendImportSectionReportDTO;
    resources: BackendImportSectionReportDTO;
  };
}

export interface BackendDailyStatDTO {
  date: string;
  count: number;
}

export interface BackendStatsDTO {
  memos_total: number;
  tags_total: number;
  resources_total: number;
  heatmap: BackendDailyStatDTO[] | null;
}
