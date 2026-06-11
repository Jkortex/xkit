// @vitest-environment happy-dom

import { describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import { mount } from '@vue/test-utils';
import SidebarAccountMenu from '@/presentation/components/sidebar/SidebarAccountMenu.vue';

vi.mock('lucide-vue-next', () => {
  const Icon = defineComponent({
    name: 'IconStub',
    setup: () => () => h('span'),
  });
  return {
    ChevronDown: Icon,
    Download: Icon,
    Key: Icon,
    LogOut: Icon,
    Settings: Icon,
    Shield: Icon,
    Upload: Icon,
  };
});

describe('SidebarAccountMenu', () => {
  it('toggles menu with exposed shortcut toggle', async () => {
    const wrapper = mount(SidebarAccountMenu, {
      props: {
        username: 'light',
      },
      attachTo: document.body,
    });

    wrapper.vm.toggleFromShortcut();
    await wrapper.vm.$nextTick();
    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(true);

    wrapper.vm.toggleFromShortcut();
    await wrapper.vm.$nextTick();
    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(false);
  });

  it('closes when clicking outside', async () => {
    const wrapper = mount(SidebarAccountMenu, {
      props: {
        username: 'light',
      },
      attachTo: document.body,
    });

    await wrapper.get('[data-testid="account-menu-trigger"]').trigger('click');
    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(true);

    document.body.dispatchEvent(
      new PointerEvent('pointerdown', { bubbles: true }),
    );
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(false);
  });

  it('closes on Escape key', async () => {
    const wrapper = mount(SidebarAccountMenu, {
      props: {
        username: 'light',
      },
      attachTo: document.body,
    });

    await wrapper.get('[data-testid="account-menu-trigger"]').trigger('click');
    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(true);

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
    await wrapper.vm.$nextTick();

    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(false);
  });

  it('emits action and closes menu', async () => {
    const wrapper = mount(SidebarAccountMenu, {
      props: {
        username: 'light',
      },
      attachTo: document.body,
    });

    await wrapper.get('[data-testid="account-menu-trigger"]').trigger('click');
    await wrapper.get('[data-testid="account-menu-export"]').trigger('click');

    expect(wrapper.emitted('export')?.length).toBe(1);
    expect(wrapper.find('[data-testid="account-menu"]').exists()).toBe(false);
  });

  it('supports arrow key navigation in menu', async () => {
    const wrapper = mount(SidebarAccountMenu, {
      props: {
        username: 'light',
      },
      attachTo: document.body,
    });

    const trigger = wrapper.get('[data-testid="account-menu-trigger"]');
    await trigger.trigger('keydown', { key: 'ArrowDown' });
    await wrapper.vm.$nextTick();

    const firstAction = wrapper.get('[data-testid="account-menu-import"]');
    expect(document.activeElement).toBe(firstAction.element);

    await firstAction.trigger('keydown', { key: 'ArrowDown' });
    await wrapper.vm.$nextTick();
    const secondAction = wrapper.get('[data-testid="account-menu-export"]');
    expect(document.activeElement).toBe(secondAction.element);
  });
});
