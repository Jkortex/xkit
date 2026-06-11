import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useBackup } from '@/presentation/composables/useBackup';
import { MessagePlugin } from 'tdesign-vue-next';

const importBackupMock = vi.fn();
const exportBackupMock = vi.fn();
const errorRef = { value: null as string | null };
const importingRef = { value: false };
const exportingRef = { value: false };

vi.mock('@/presentation/composables/useResource', () => ({
  useResource: () => ({
    importBackup: importBackupMock,
    exportBackup: exportBackupMock,
    error: errorRef,
    importing: importingRef,
    exporting: exportingRef,
  }),
}));

vi.mock('tdesign-vue-next', () => ({
  MessagePlugin: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('useBackup', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    importBackupMock.mockReset();
    exportBackupMock.mockReset();
    errorRef.value = null;
  });

  it('shows import error message when import fails', async () => {
    importBackupMock.mockResolvedValue(null);
    errorRef.value = '导入失败';
    const backup = useBackup();
    const file = new File(['x'], 'backup.zip');
    const event = { target: { files: [file] } } as unknown as Event;

    await backup.handleImport(event);

    expect(MessagePlugin.error).toHaveBeenCalledWith('导入失败');
    expect(backup.showImportReport.value).toBe(false);
  });

  it('shows export error message when export fails', async () => {
    exportBackupMock.mockResolvedValue(null);
    errorRef.value = '导出失败';
    const backup = useBackup();

    await backup.handleExport();

    expect(MessagePlugin.error).toHaveBeenCalledWith('导出失败');
  });

  it('opens report and triggers refresh callback when import succeeds', async () => {
    importBackupMock.mockResolvedValue({
      message: 'ok',
      memosImported: 2,
      resourcesImported: 1,
      memosSkipped: 0,
      resourcesSkipped: 0,
      report: {
        memos: { imported: 2, skipped: 0, details: [] },
        resources: { imported: 1, skipped: 0, details: [] },
      },
    });
    const onImported = vi.fn();
    const backup = useBackup({ onImported });
    const file = new File(['x'], 'backup.zip');
    const event = { target: { files: [file] } } as unknown as Event;

    await backup.handleImport(event);

    expect(MessagePlugin.success).toHaveBeenCalled();
    expect(onImported).toHaveBeenCalled();
    expect(backup.showImportReport.value).toBe(true);
  });
});
