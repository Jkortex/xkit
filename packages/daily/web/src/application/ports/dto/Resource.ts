export interface ImportSkipDetailDTO {
  readonly entity: string;
  readonly key: string;
  readonly reason: string;
}

export interface ImportSectionReportDTO {
  readonly imported: number;
  readonly skipped: number;
  readonly details: ImportSkipDetailDTO[];
}

export interface ImportReportDTO {
  readonly message: string;
  readonly memosImported: number;
  readonly resourcesImported: number;
  readonly memosSkipped: number;
  readonly resourcesSkipped: number;
  readonly report: {
    readonly memos: ImportSectionReportDTO;
    readonly resources: ImportSectionReportDTO;
  };
}
