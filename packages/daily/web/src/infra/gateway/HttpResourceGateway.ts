import type { ResourceDTO } from '@/application/ports/dto/Memo';
import type { ImportReportDTO } from '@/application/ports/dto/Resource';
import {
  BackendImportReportDTO,
  BackendResourceDTO,
} from './dto/BackendMemoDTO';
import { transformResource } from './transform/memo-transform';
import { Result, success } from '@/utils/result';
import { httpClient } from '../http/FetchClient';

export class HttpResourceGateway {
  async upload(file: File): Promise<Result<ResourceDTO>> {
    const formData = new FormData();
    formData.append('file', file);

    const result = await httpClient.request<BackendResourceDTO>('/resources', {
      method: 'POST',
      body: formData,
    });
    if (result.kind === 'failure') return result;

    return success(transformResource(result.value));
  }

  async importData(file: File): Promise<Result<ImportReportDTO>> {
    const formData = new FormData();
    formData.append('file', file);

    const result = await httpClient.request<BackendImportReportDTO>(
      '/system/import',
      {
        method: 'POST',
        body: formData,
      },
    );
    if (result.kind === 'failure') return result;

    const raw = result.value;
    return success({
      message: raw.message,
      memosImported: raw.memos_imported,
      resourcesImported: raw.resources_imported,
      memosSkipped: raw.memos_skipped,
      resourcesSkipped: raw.resources_skipped,
      report: raw.report,
    });
  }

  async exportData(): Promise<Result<Blob>> {
    return await httpClient.requestBlob('/system/export');
  }
}

export const resourceGateway = new HttpResourceGateway();
