// @vitest-environment happy-dom

import { describe, expect, it, vi } from 'vitest';
import { defineComponent, h } from 'vue';
import { mount } from '@vue/test-utils';
import InviteCard from '@/presentation/components/invites/InviteCard.vue';
import InviteCreatePanel from '@/presentation/components/invites/InviteCreatePanel.vue';
import InviteFilterPanel from '@/presentation/components/invites/InviteFilterPanel.vue';

vi.mock('tdesign-vue-next', () => {
  const Button = defineComponent({
    name: 'TButton',
    props: {
      disabled: { type: Boolean, default: false },
      loading: { type: Boolean, default: false },
    },
    emits: ['click'],
    setup(props, { slots, emit }) {
      return () =>
        h(
          'button',
          {
            'data-testid': 't-button',
            disabled: props.disabled,
            'data-loading': props.loading,
            onClick: () => emit('click'),
          },
          slots.default?.(),
        );
    },
  });

  const Select = defineComponent({
    name: 'TSelect',
    props: {
      modelValue: { type: String, default: '' },
      options: { type: Array, default: () => [] },
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
      return () =>
        h(
          'select',
          {
            'data-testid': 't-select',
            value: props.modelValue,
            onChange: (event: Event) => {
              const target = event.target as HTMLSelectElement;
              emit('update:modelValue', target.value);
            },
          },
          (props.options as Array<{ value: string; label: string }>).map(
            (opt) =>
              h(
                'option',
                {
                  value: opt.value,
                },
                opt.label,
              ),
          ),
        );
    },
  });

  const InputNumber = defineComponent({
    name: 'TInputNumber',
    props: {
      modelValue: { type: Number, default: 0 },
    },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
      return () =>
        h('input', {
          'data-testid': 't-input-number',
          type: 'number',
          value: String(props.modelValue),
          onInput: (event: Event) => {
            const target = event.target as HTMLInputElement;
            emit('update:modelValue', Number(target.value));
          },
        });
    },
  });

  const Popconfirm = defineComponent({
    name: 'TPopconfirm',
    emits: ['confirm'],
    setup(_, { slots, emit }) {
      return () =>
        h('div', [
          slots.default?.(),
          h(
            'button',
            {
              'data-testid': 't-popconfirm-confirm',
              onClick: () => emit('confirm'),
            },
            'confirm',
          ),
        ]);
    },
  });

  const Tag = defineComponent({
    name: 'TTag',
    setup(_, { slots }) {
      return () => h('span', { 'data-testid': 't-tag' }, slots.default?.());
    },
  });

  return {
    Button,
    Select,
    InputNumber,
    Popconfirm,
    Tag,
  };
});

vi.mock('lucide-vue-next', () => {
  const Icon = defineComponent({
    name: 'IconStub',
    setup: () => () => h('span'),
  });
  return {
    Copy: Icon,
    Shield: Icon,
    Timer: Icon,
    UserRoundPlus: Icon,
  };
});

describe('Invite components', () => {
  it('InviteCreatePanel emits role/ttl/create events', async () => {
    const wrapper = mount(InviteCreatePanel, {
      props: {
        role: 'member',
        ttlHours: 48,
        creating: false,
        error: '',
        roleOptions: [
          { label: '成员（member）', value: 'member' },
          { label: '管理员（admin）', value: 'admin' },
        ],
      },
    });

    const selects = wrapper.findAll('[data-testid="t-select"]');
    await selects[0].setValue('admin');

    const input = wrapper.get('[data-testid="t-input-number"]');
    await input.setValue('72');

    const createBtn = wrapper.get('[data-testid="t-button"]');
    await createBtn.trigger('click');

    expect(wrapper.emitted('update:role')?.[0]).toEqual(['admin']);
    expect(wrapper.emitted('update:ttlHours')?.[0]).toEqual([72]);
    expect(wrapper.emitted('create')?.length).toBe(1);
  });

  it('InviteFilterPanel emits update:status', async () => {
    const wrapper = mount(InviteFilterPanel, {
      props: {
        status: 'all',
        statusOptions: [
          { label: '全部状态', value: 'all' },
          { label: '活跃', value: 'active' },
        ],
      },
    });

    const select = wrapper.get('[data-testid="t-select"]');
    await select.setValue('active');

    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active']);
  });

  it('InviteCard emits copy and revoke actions for active invite', async () => {
    const wrapper = mount(InviteCard, {
      props: {
        invite: {
          id: 'i-1',
          role: 'member',
          status: 'active',
          expiresAt: '2026-03-10T00:00:00Z',
        },
        code: 'abc123',
        revoking: false,
      },
    });

    const buttons = wrapper.findAll('[data-testid="t-button"]');
    expect(wrapper.text()).toContain('abc123');
    await buttons[0].trigger('click');
    await wrapper.get('[data-testid="t-popconfirm-confirm"]').trigger('click');

    expect(wrapper.emitted('copy')?.[0]).toEqual(['abc123']);
    expect(wrapper.emitted('revoke')?.[0]).toEqual(['i-1']);
  });

  it('InviteCard disables copy when code is empty and hides revoke for non-active', () => {
    const wrapper = mount(InviteCard, {
      props: {
        invite: {
          id: 'i-2',
          role: 'admin',
          status: 'used',
          expiresAt: '2026-03-10T00:00:00Z',
        },
        code: '',
        revoking: false,
      },
    });

    expect(wrapper.text()).toContain('邀请码不可用');
    expect(
      wrapper.get('[data-testid="t-button"]').attributes('disabled'),
    ).toBeDefined();
    expect(wrapper.find('[data-testid="t-popconfirm-confirm"]').exists()).toBe(
      false,
    );
  });
});
