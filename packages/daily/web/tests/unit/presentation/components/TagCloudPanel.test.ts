// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import { mount } from '@vue/test-utils';
import TagCloudPanel from '@/presentation/components/TagCloudPanel.vue';

vi.mock('tdesign-vue-next', () => {
  const Tag = defineComponent({
    name: 'TTag',
    emits: ['click'],
    setup(_, { slots, emit }) {
      return () =>
        h(
          'button',
          {
            'data-testid': 'tag-item',
            onClick: () => emit('click'),
          },
          slots.default?.(),
        );
    },
  });
  const Input = defineComponent({
    name: 'TInput',
    props: {
      modelValue: { type: String, default: '' },
    },
    emits: ['update:modelValue', 'clear'],
    setup(props, { emit }) {
      return () =>
        h('input', {
          'data-testid': 'tag-search',
          value: props.modelValue,
          onInput: (event: Event) => {
            const target = event.target as HTMLInputElement;
            emit('update:modelValue', target.value);
          },
        });
    },
  });
  return { Tag, Input };
});

vi.mock('lucide-vue-next', () => {
  const stub = defineComponent({
    name: 'IconStub',
    setup: () => () => h('span'),
  });
  return {
    PencilLine: stub,
    Hash: stub,
  };
});

vi.mock('@/presentation/components/TagGovernanceDialog.vue', () => ({
  default: defineComponent({
    name: 'TagGovernanceDialog',
    setup: () => () => h('div'),
  }),
}));

describe('TagCloudPanel', () => {
  const tags = Array.from({ length: 30 }, (_, index) => ({
    name: `tag-${index + 1}`,
    count: 30 - index,
  }));

  beforeEach(() => {
    localStorage.clear();
  });

  it('shows top N tags by default and supports expand', async () => {
    const wrapper = mount(TagCloudPanel, {
      props: { tags },
    });

    expect(wrapper.text()).toContain('16 / 29 个标签');
    expect(wrapper.text()).toContain('显示全部（+13）');

    const expandBtn = wrapper
      .findAll('button')
      .find((node) => node.text().includes('显示全部'));
    expect(expandBtn).toBeTruthy();
    await expandBtn!.trigger('click');

    expect(wrapper.text()).toContain('29 / 29 个标签');
    expect(wrapper.text()).toContain('收起（Top 16）');
  });

  it('filters tags by keyword', async () => {
    const wrapper = mount(TagCloudPanel, {
      props: {
        tags: [
          { name: 'infra', count: 5 },
          { name: 'ops', count: 4 },
          { name: 'frontend', count: 3 },
        ],
      },
    });

    const input = wrapper.get('[data-testid="tag-search"]');
    await input.setValue('op');

    expect(wrapper.text()).toContain('1 / 1 个标签');
    expect(wrapper.text()).toContain('ops');
    expect(wrapper.text()).not.toContain('infra');
  });

  it('hides low-frequency tags by default and can include them', async () => {
    const wrapper = mount(TagCloudPanel, {
      props: {
        tags: [
          { name: 'ops', count: 3 },
          { name: 'infra', count: 2 },
          { name: 'rare', count: 1 },
        ],
      },
    });

    expect(wrapper.text()).toContain('2 / 2 个标签');
    expect(wrapper.text()).not.toContain('rare');
    expect(wrapper.text()).toContain('显示低频（+1）');

    const includeBtn = wrapper
      .findAll('button')
      .find((node) => node.text().includes('显示低频'));
    expect(includeBtn).toBeTruthy();
    await includeBtn!.trigger('click');

    expect(wrapper.text()).toContain('3 / 3 个标签');
    expect(wrapper.text()).toContain('rare');
    expect(wrapper.text()).toContain('隐藏低频');
  });
});
