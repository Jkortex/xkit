// @vitest-environment happy-dom

import { describe, expect, it } from 'vitest';
import { defineComponent, h, ref } from 'vue';
import { mount } from '@vue/test-utils';
import { useMemoEditorCaret } from '@/presentation/composables/memo-editor/useMemoEditorCaret';

const TTextarea = defineComponent({
  name: 'TTextarea',
  setup(_, { expose }) {
    const el = ref<HTMLElement | null>(null);
    expose({ $el: el });
    return () => h('div', { ref: el, class: 't-textarea' }, [h('textarea')]);
  },
});

describe('useMemoEditorCaret', () => {
  it('inserts text at selection and triggers input event', async () => {
    const component = defineComponent({
      components: { TTextarea },
      setup() {
        const { bindTextareaRootRef, insertText } = useMemoEditorCaret();
        return { bindTextareaRootRef, insertText };
      },
      template: '<TTextarea :ref="bindTextareaRootRef" />',
    });

    const wrapper = mount(component);
    const textarea = wrapper.find('textarea').element as HTMLTextAreaElement;
    textarea.value = 'Hello world';
    textarea.selectionStart = 6;
    textarea.selectionEnd = 6;

    let inputTriggered = false;
    textarea.addEventListener('input', () => {
      inputTriggered = true;
    });

    wrapper.vm.insertText('awesome ');

    expect(textarea.value).toBe('Hello awesome world');
    expect(textarea.selectionStart).toBe(14);
    expect(inputTriggered).toBe(true);
  });
});
