import { ref, type Ref } from 'vue';
import { useResource } from '@/presentation/composables/useResource';
import { ResourcePresenter } from '@/presentation/presenters/ResourcePresenter';
import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

interface UseMemoEditorResourcesOptions {
  attachedResources: Ref<ResourceVM[]>;
  onUploadSuccess?: (res: ResourceVM) => void;
}

interface UseMemoEditorResourcesResult {
  uploading: Ref<boolean>;
  error: Ref<string | null>;
  fileInput: Ref<HTMLInputElement | null>;
  onFileChange: (event: Event) => Promise<void>;
  handlePaste: (event: ClipboardEvent) => Promise<void>;
  removeResource: (id: string) => void;
  triggerUpload: () => void;
}

export function useMemoEditorResources(
  options: UseMemoEditorResourcesOptions,
): UseMemoEditorResourcesResult {
  const { uploadFile, uploading, error } = useResource();
  const fileInput = ref<HTMLInputElement | null>(null);

  const pushUploadedFile = async (file: File): Promise<void> => {
    const uploaded = await uploadFile(file);
    if (uploaded) {
      const vm = ResourcePresenter.toViewModel(uploaded);
      options.attachedResources.value.push(vm);
      options.onUploadSuccess?.(vm);
    }
  };

  const onFileChange = async (event: Event): Promise<void> => {
    const files = (event.target as HTMLInputElement).files;
    if (!files || files.length === 0) return;

    for (const file of Array.from(files)) {
      await pushUploadedFile(file);
    }

    if (fileInput.value) fileInput.value.value = '';
  };

  const handlePaste = async (event: ClipboardEvent): Promise<void> => {
    const items = event.clipboardData?.items;
    if (!items) return;

    for (const item of Array.from(items)) {
      if (!item.type.startsWith('image/') && !item.type.startsWith('video/'))
        continue;
      const file = item.getAsFile();
      if (!file) continue;
      await pushUploadedFile(file);
    }
  };

  const removeResource = (id: string): void => {
    options.attachedResources.value = options.attachedResources.value.filter(
      (item) => item.id !== id,
    );
  };

  const triggerUpload = (): void => {
    fileInput.value?.click();
  };

  return {
    uploading,
    error,
    fileInput,
    onFileChange,
    handlePaste,
    removeResource,
    triggerUpload,
  };
}
