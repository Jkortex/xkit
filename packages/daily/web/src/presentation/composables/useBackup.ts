import { ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { ImportReportDTO } from '@/application/ports/dto/Resource';
import { useResource } from './useResource';

const skipReasonLabels: Record<string, string> = {
  duplicate_by_id: '重复 ID',
  duplicate_by_hash: '重复哈希',
  duplicate_by_path: '重复路径',
  duplicate_by_fingerprint: '重复指纹',
  invalid_metadata: '元数据无效',
};

interface UseBackupOptions {
  onImported?: () => void;
}

const buildExportFilename = (): string =>
  `daily_export_${new Date().toISOString().replace(/[:.]/g, '-')}.zip`;

const triggerDownload = (blob: Blob, filename: string) => {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
};

const getFileFromInput = (event: Event): File | null => {
  const files = (event.target as HTMLInputElement).files;
  if (!files || files.length === 0) return null;
  return files[0] as File;
};

export function useBackup(options: UseBackupOptions = {}) {
  const { importBackup, exportBackup, importing, exporting, error } =
    useResource();
  const fileInput = ref<HTMLInputElement | null>(null);
  const showImportReport = ref(false);
  const latestImportReport = ref<ImportReportDTO | null>(null);

  const resetImportInput = () => {
    if (fileInput.value) fileInput.value.value = '';
  };

  const humanizeReason = (reason: string) => skipReasonLabels[reason] || reason;

  const handleImport = async (event: Event) => {
    const file = getFileFromInput(event);
    if (!file) return;

    const payload = await importBackup(file);
    if (!payload) {
      MessagePlugin.error(error.value || '导入失败');
      resetImportInput();
      return;
    }

    MessagePlugin.success(
      `导入完成: 笔记 +${payload.memosImported}，资源 +${payload.resourcesImported}`,
    );
    latestImportReport.value = payload;
    showImportReport.value = true;
    options.onImported?.();
    resetImportInput();
  };

  const handleExport = async () => {
    const blob = await exportBackup();
    if (!blob) {
      MessagePlugin.error(error.value || '导出失败');
      return;
    }

    triggerDownload(blob, buildExportFilename());
    MessagePlugin.success('导出成功');
  };

  return {
    fileInput,
    importing,
    exporting,
    showImportReport,
    latestImportReport,
    humanizeReason,
    handleImport,
    handleExport,
  };
}
