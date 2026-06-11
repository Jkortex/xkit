// @vitest-environment happy-dom

import { describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
import HomeFilterBar from '@/presentation/components/HomeFilterBar.vue';

// Mock TDesign components as they might be complex to render in unit tests
vi.mock('tdesign-vue-next', () => ({
  Tag: defineComponent({
    name: 'TTag',
    props: ['closable'],
    emits: ['close'],
    setup(props, { slots, emit }) {
      return () =>
        h('div', { class: 't-tag' }, [
          slots.default?.(),
          props.closable
            ? h(
                'button',
                { class: 't-tag-close', onClick: () => emit('close') },
                'x',
              )
            : null,
        ]);
    },
  }),
  Dropdown: defineComponent({
    name: 'TDropdown',
    props: ['options'],
    emits: ['click'],
    setup(_, { slots }) {
      return () => h('div', { class: 't-dropdown' }, slots.default?.());
    },
  }),
}));

// Mock lucide-vue-next icons
vi.mock('lucide-vue-next', () => {
  const MockIcon = defineComponent({
    name: 'MockIcon',
    setup() {
      return () => h('span', { class: 'mock-icon' });
    },
  });
  return {
    Search: MockIcon,
    ListFilter: MockIcon,
    X: MockIcon,
    Calendar: MockIcon,
    Hash: MockIcon,
    Paperclip: MockIcon,
    ArrowDownWideNarrow: MockIcon,
  };
});

describe('HomeFilterBar', () => {
  const mountComponent = (props = {}) => {
    return mount(HomeFilterBar, {
      props: {
        tokens: [],
        inputText: '',
        'onUpdate:inputText': (val: string) =>
          wrapper.setProps({ inputText: val }),
        ...props,
      },
    });
  };

  let wrapper: any;

  it('shows suggestions when typing "?"', async () => {
    wrapper = mountComponent({ inputText: '?' });
    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestions = wrapper.findAll('.ui-suggestion-item');
    expect(suggestions.length).toBeGreaterThan(0);
    expect(suggestions.some((s: any) => s.text().includes('全文搜索'))).toBe(
      true,
    );
    expect(suggestions.some((s: any) => s.text().includes('标签'))).toBe(true);
  });

  it('shows suggestions when typing "？" (full-width)', async () => {
    wrapper = mountComponent({ inputText: '？' });
    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestions = wrapper.findAll('.ui-suggestion-item');
    expect(suggestions.length).toBeGreaterThan(0);
  });

  it('shows all 9 default suggestions when typing "?"', async () => {
    wrapper = mountComponent({ inputText: '?' });
    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestions = wrapper.findAll('.ui-suggestion-item');
    expect(suggestions.length).toBe(9);

    const labels = suggestions.map((s: any) => s.find('.font-medium').text());
    expect(labels).toContain('全文搜索');
    expect(labels).toContain('标签 (包含任一)');
    expect(labels).toContain('标签 (必须包含)');
    expect(labels).toContain('排除标签');
    expect(labels).toContain('开始日期');
    expect(labels).toContain('结束日期');
    expect(labels).toContain('包含附件');
    expect(labels).toContain('无附件');
    expect(labels).toContain('排序');
  });

  it('does not show suggestions when typing regular keywords without "?"', async () => {
    wrapper = mountComponent({ inputText: 'hello' });
    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestions = wrapper.findAll('.ui-suggestion-item');
    expect(suggestions.length).toBe(0);
  });

  it('selects a suggestion and clears "?"', async () => {
    let capturedTokens: any[] = [];
    wrapper = mount(HomeFilterBar, {
      props: {
        tokens: [],
        inputText: '?',
        'onUpdate:inputText': async (val: string) => {
          await wrapper.setProps({ inputText: val });
        },
        onAddToken: (token: any) => {
          capturedTokens.push(token);
        },
      },
    });

    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestion = wrapper.find('.ui-suggestion-item'); // First one is 'text:'
    await suggestion.trigger('mousedown');

    // For 'text:', it just updates inputText to 'text:' and keeps focus
    expect(wrapper.props('inputText')).toBe('text:');
  });

  it('selects a "has:resource" suggestion and adds token after typing "?"', async () => {
    let capturedTokens: any[] = [];
    wrapper = mount(HomeFilterBar, {
      props: {
        tokens: [],
        inputText: '?',
        'onUpdate:inputText': async (val: string) => {
          await wrapper.setProps({ inputText: val });
        },
        onAddToken: (token: any) => {
          capturedTokens.push(token);
        },
      },
    });

    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    const suggestions = wrapper.findAll('.ui-suggestion-item');
    const hasResSuggestion = suggestions.find((s: any) =>
      s.text().includes('包含附件'),
    );
    await hasResSuggestion?.trigger('mousedown');

    expect(capturedTokens.length).toBe(1);
    expect(capturedTokens[0].type).toBe('has_resource');
    expect(capturedTokens[0].value).toBe(true);
    expect(wrapper.props('inputText')).toBe('');
  });

  it('selects active suggestion with Enter after typing "?"', async () => {
    let capturedTokens: any[] = [];
    wrapper = mount(HomeFilterBar, {
      props: {
        tokens: [],
        inputText: '?',
        'onUpdate:inputText': async (val: string) => {
          await wrapper.setProps({ inputText: val });
        },
        onAddToken: (token: any) => {
          capturedTokens.push(token);
        },
      },
    });

    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    // First suggestion is 'text:'
    await input.trigger('keydown', { key: 'Enter' });

    expect(wrapper.props('inputText')).toBe('text:');
  });

  it('navigates suggestions with arrow keys after typing "?"', async () => {
    wrapper = mountComponent({ inputText: '?' });
    const input = wrapper.find('input[type="text"]');
    await input.trigger('focus');

    expect(wrapper.vm.activeSuggestionIndex).toBe(0);

    await input.trigger('keydown', { key: 'ArrowDown' });
    expect(wrapper.vm.activeSuggestionIndex).toBe(1);

    await input.trigger('keydown', { key: 'ArrowUp' });
    expect(wrapper.vm.activeSuggestionIndex).toBe(0);
  });
});
