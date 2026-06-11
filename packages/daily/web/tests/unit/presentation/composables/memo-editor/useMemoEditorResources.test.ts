// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ref } from 'vue';
import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

const { uploadFileMock, uploadingRef } = vi.hoisted(() => ({
  uploadFileMock: vi.fn(),
  uploadingRef: { value: false },
}));

vi.mock('@/presentation/composables/useResource', () => ({
  useResource: () => ({
    uploadFile: uploadFileMock,
    uploading: uploadingRef,
  }),
}));

import { useMemoEditorResources } from '@/presentation/composables/memo-editor/useMemoEditorResources';

describe('useMemoEditorResources', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uploads selected files and clears file input value', async () => {
    uploadFileMock.mockResolvedValue({
      id: 'r1',
      url: '/r1',
      filename: 'a.png',
      isImage: true,
    });
    const attachedResources = ref<ResourceVM[]>([]);
    const resources = useMemoEditorResources({ attachedResources });
    const input = document.createElement('input');
    input.value = 'fake';
    resources.fileInput.value = input;

    const file = new File(['abc'], 'a.png', { type: 'image/png' });
    const event = { target: { files: [file] } } as unknown as Event;

    await resources.onFileChange(event);

    expect(uploadFileMock).toHaveBeenCalledTimes(1);
    expect(attachedResources.value).toHaveLength(1);
    expect(resources.fileInput.value?.value).toBe('');
  });

  it('uploads pasted images only', async () => {
    uploadFileMock.mockResolvedValue({
      id: 'r2',
      url: '/r2',
      filename: 'b.png',
      isImage: true,
    });
    const attachedResources = ref<ResourceVM[]>([]);
    const resources = useMemoEditorResources({ attachedResources });

    const imageFile = new File(['img'], 'b.png', { type: 'image/png' });
    const event = {
      clipboardData: {
        items: [
          { type: 'text/plain', getAsFile: () => null },
          { type: 'image/png', getAsFile: () => imageFile },
        ],
      },
    } as unknown as ClipboardEvent;

    await resources.handlePaste(event);

    expect(uploadFileMock).toHaveBeenCalledTimes(1);
    expect(attachedResources.value).toHaveLength(1);
  });

  it('removes resource by id', () => {
    const attachedResources = ref([
      { id: 'a', url: '/a', filename: 'a', isImage: false },
      { id: 'b', url: '/b', filename: 'b', isImage: false },
    ]);
    const resources = useMemoEditorResources({ attachedResources });

    resources.removeResource('a');

    expect(attachedResources.value).toEqual([
      { id: 'b', url: '/b', filename: 'b', isImage: false },
    ]);
  });
});
