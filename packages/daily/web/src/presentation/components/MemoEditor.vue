<script setup lang="ts">
import { computed, watch } from 'vue';
import {
  Button as TButton,
  Textarea as TTextarea,
  MessagePlugin,
  type TextareaValue,
} from 'tdesign-vue-next';
import { Paperclip, Send } from 'lucide-vue-next';
import type { ResourceVM } from '../view-models/ResourceVM';
import AttachedResources from './AttachedResources.vue';
import {
  useMemoEditorState,
  type EditableMemoInput,
} from '@/presentation/composables/memo-editor/useMemoEditorState';
import { useMemoActions } from '@/presentation/composables/useMemoActions';
import { useMemoEditorResources } from '@/presentation/composables/memo-editor/useMemoEditorResources';
import { useMemoEditorCaret } from '@/presentation/composables/memo-editor/useMemoEditorCaret';

const props = defineProps<{
  editMemo?: {
    uuid: string;
    content: string;
    tags: string[];
    resources: ResourceVM[];
  };
  expanded?: boolean;
}>();

const emit = defineEmits<{
  (e: 'cancel-edit'): void;
  (e: 'success'): void;
}>();

const { createMemo, updateMemo, loading: saving, error } = useMemoActions();
const editMemoRef = computed<EditableMemoInput | undefined>(
  () => props.editMemo,
);
const { content, attachedResources, clearAfterSuccess } = useMemoEditorState({
  editMemo: editMemoRef,
});

const { bindTextareaRootRef, focusEditor, insertText } = useMemoEditorCaret();

const handleUploadSuccess = (res: ResourceVM) => {
  const reference = res.isImage
    ? `![${res.filename}](${res.id})`
    : `[${res.filename}](${res.id})`;
  // Ensure we have a newline if the current line isn't empty
  const lines = content.value.split('\n');
  const lastLine = lines[lines.length - 1];
  const prefix = lastLine && lastLine.trim() !== '' ? '\n' : '';

  insertText(`${prefix}${reference}\n`);
};

const {
  uploading,
  error: uploadError,
  fileInput,
  onFileChange,
  handlePaste,
  removeResource,
  triggerUpload,
} = useMemoEditorResources({
  attachedResources,
  onUploadSuccess: handleUploadSuccess,
});

watch(uploadError, (err) => {
  if (err) {
    MessagePlugin.error(err);
  }
});

watch(error, (err) => {
  if (err) {
    MessagePlugin.error(err);
  }
});

const handleSave = async () => {
  if (!content.value.trim() && attachedResources.value.length === 0) return;

  let success = false;
  if (props.editMemo) {
    const resourceIds = attachedResources.value.map((r) => r.id);
    success = await updateMemo(props.editMemo.uuid, content.value, resourceIds);
  } else {
    const resourceIds = attachedResources.value.map((r) => r.id);
    success = await createMemo(content.value, resourceIds);
  }

  if (success) {
    clearAfterSuccess();
    MessagePlugin.success(props.editMemo ? '修改成功' : '记录成功');
    emit('success');
    if (props.editMemo) emit('cancel-edit');
  }
};

const handleKeydown = (
  _value: TextareaValue,
  context: { e: KeyboardEvent },
) => {
  const e = context.e;
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    handleSave();
    return;
  }
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'u') {
    e.preventDefault();
    triggerUpload();
  }
};

defineExpose({
  focusEditor,
});
</script>

<template>
  <div
    class="bg-surface border rounded-xl p-4 shadow-sm transition-all"
    :class="
      editMemo ? 'border-accent ring-2 ring-accent-soft' : 'border-border'
    "
  >
    <div v-if="editMemo" class="flex justify-between items-center mb-2">
      <span class="text-xs font-bold text-accent">正在编辑笔记</span>
      <button
        class="text-xs text-muted hover:text-primary-text"
        @click="$emit('cancel-edit')"
      >
        取消编辑
      </button>
    </div>

    <!-- Editor Area -->
    <TTextarea
      :ref="bindTextareaRootRef"
      v-model="content"
      :placeholder="
        editMemo ? '修改内容...' : '此刻在想什么？Ctrl + Enter 快速保存'
      "
      :autosize="{
        minRows: props.expanded ? 12 : 3,
        maxRows: props.expanded ? 28 : 10,
      }"
      class="border-none! shadow-none! p-0! focus:ring-0! text-sm"
      @paste="handlePaste"
      @keydown="handleKeydown"
    />

    <AttachedResources
      :resources="attachedResources"
      :uploading="uploading"
      @remove="removeResource"
    />

    <!-- Footer Toolbar -->
    <div
      class="flex items-center justify-between mt-3 pt-3 border-t border-border"
    >
      <div class="flex gap-2">
        <TButton
          variant="text"
          shape="circle"
          size="small"
          @click="fileInput?.click()"
        >
          <template #icon>
            <Paperclip :size="18" class="text-muted hover:text-accent" />
          </template>
        </TButton>
        <input
          ref="fileInput"
          type="file"
          multiple
          class="hidden"
          @change="onFileChange"
        />
      </div>

      <TButton
        theme="primary"
        size="small"
        :loading="saving"
        @click="handleSave"
      >
        <template #icon><Send :size="14" /></template>
        {{ editMemo ? '更新' : '保存' }}
      </TButton>
    </div>
  </div>
</template>

<style scoped>
:deep(.t-textarea__inner) {
  border: none !important;
  background: transparent !important;
  padding: 0 !important;
}
</style>
