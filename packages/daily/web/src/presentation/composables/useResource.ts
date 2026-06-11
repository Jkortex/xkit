import { ref } from 'vue';
import { resourceGateway } from '@/infra/gateway/HttpResourceGateway';
import type { ResourceDTO } from '@/application/ports/dto/Memo';
import type { ImportReportDTO } from '@/application/ports/dto/Resource';
import { ErrorPresenter } from '@/presentation/presenters/ErrorPresenter';

export function useResource() {
  const uploading = ref(false);
  const importing = ref(false);
  const exporting = ref(false);
  const error = ref<string | null>(null);

  const uploadFile = async (file: File): Promise<ResourceDTO | null> => {
    uploading.value = true;
    error.value = null;

    const result = await resourceGateway.upload(file);
    uploading.value = false;

    if (result.kind === 'success') {
      return result.value;
    } else {
      error.value = ErrorPresenter.toMessage(result.error);
      return null;
    }
  };

  const importBackup = async (file: File): Promise<ImportReportDTO | null> => {
    importing.value = true;
    error.value = null;
    const result = await resourceGateway.importData(file);
    importing.value = false;

    if (result.kind === 'success') {
      return result.value;
    }
    error.value = ErrorPresenter.toMessage(result.error);
    return null;
  };

  const exportBackup = async (): Promise<Blob | null> => {
    exporting.value = true;
    error.value = null;
    const result = await resourceGateway.exportData();
    exporting.value = false;

    if (result.kind === 'success') {
      return result.value;
    }
    error.value = ErrorPresenter.toMessage(result.error);
    return null;
  };

  return {
    uploadFile,
    uploading,
    importBackup,
    importing,
    exportBackup,
    exporting,
    error,
  };
}
