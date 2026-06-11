// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import CommandPalette from '@/presentation/components/CommandPalette.vue';

describe('CommandPalette', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('scrolls active item into view when selection moves', async () => {
    const scrollSpy = vi
      .spyOn(HTMLElement.prototype, 'scrollIntoView')
      .mockImplementation(() => undefined);

    const wrapper = mount(CommandPalette, {
      props: {
        visible: true,
        query: '',
        activeIndex: 0,
        items: [
          { id: 'a', title: 'A', category: '导航' },
          { id: 'b', title: 'B', category: '导航' },
        ],
        'onUpdate:visible': () => undefined,
        'onUpdate:query': () => undefined,
      },
    });

    await wrapper.setProps({ activeIndex: 1 });
    await wrapper.vm.$nextTick();

    expect(scrollSpy).toHaveBeenCalled();
  });
});
