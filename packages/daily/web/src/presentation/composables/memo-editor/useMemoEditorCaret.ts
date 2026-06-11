import { nextTick, ref } from 'vue';
import { Textarea as TTextarea } from 'tdesign-vue-next';

interface UseMemoEditorCaretResult {
  bindTextareaRootRef: (instance: unknown) => void;
  focusEditor: () => void;
  insertText: (text: string) => void;
}

const isTagOnlyLine = (line: string): boolean =>
  /^(#[^\s#]+)(\s+#[^\s#]+)*$/.test(line.trim());

const resolveCaretPosition = (text: string): number => {
  const lines = text.split('\n');
  const lastLine = lines[lines.length - 1] ?? '';
  if (!isTagOnlyLine(lastLine) || lines.length <= 1) return text.length;
  const previousLineEnd = text.lastIndexOf('\n');
  return previousLineEnd >= 0 ? previousLineEnd : text.length;
};

export function useMemoEditorCaret(): UseMemoEditorCaretResult {
  const textareaRootRef = ref<InstanceType<typeof TTextarea> | null>(null);

  const bindTextareaRootRef = (instance: unknown): void => {
    textareaRootRef.value =
      (instance as InstanceType<typeof TTextarea>) ?? null;
  };

  const focusEditor = (): void => {
    nextTick(() => {
      const node = textareaRootRef.value?.$el as HTMLElement | undefined;
      const textarea = node?.querySelector('textarea');
      textarea?.focus();
      if (!textarea) return;
      const caret = resolveCaretPosition(textarea.value);
      textarea.setSelectionRange(caret, caret);
    });
  };

  const insertText = (text: string): void => {
    const node = textareaRootRef.value?.$el as HTMLElement | undefined;
    const textarea = node?.querySelector('textarea');
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const value = textarea.value;

    textarea.value = value.substring(0, start) + text + value.substring(end);
    textarea.selectionStart = textarea.selectionEnd = start + text.length;

    // Trigger input event for v-model
    textarea.dispatchEvent(new Event('input', { bubbles: true }));
    textarea.focus();
  };

  return {
    bindTextareaRootRef,
    focusEditor,
    insertText,
  };
}
