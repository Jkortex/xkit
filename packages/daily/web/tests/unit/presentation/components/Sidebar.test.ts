// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, ref } from 'vue';
import { mount } from '@vue/test-utils';
import Sidebar from '@/presentation/components/Sidebar.vue';
import type { ImportReportDTO } from '@/application/ports/dto/Resource';

import { setActivePinia, createPinia } from 'pinia';

const {
  handleImportMock,
  handleExportMock,
  fetchStatsMock,
  fetchTagsMock,
  humanizeReasonMock,
  routerPushMock,
  authLogoutMock,
  statsGatewayMock,
  memoGatewayMock,
  tagSetGatewayMock,
} = vi.hoisted(() => ({
  handleImportMock: vi.fn(),
  handleExportMock: vi.fn(),
  fetchStatsMock: vi.fn(),
  fetchTagsMock: vi.fn(),
  humanizeReasonMock: vi.fn((reason: string) => reason),
  routerPushMock: vi.fn(),
  authLogoutMock: vi.fn(),
  statsGatewayMock: { getStats: vi.fn() },
  memoGatewayMock: {
    getTags: vi.fn().mockResolvedValue({ kind: 'success' as const, value: [] }),
    renameTag: vi.fn(),
    mergeTags: vi.fn(),
  },
  tagSetGatewayMock: {
    listGroups: vi
      .fn()
      .mockResolvedValue({ kind: 'success' as const, value: [] }),
    listTagSets: vi
      .fn()
      .mockResolvedValue({ kind: 'success' as const, value: [] }),
  },
}));

const importingRef = ref(false);
const exportingRef = ref(false);
const showImportReportRef = ref(false);
const latestImportReportRef = ref<ImportReportDTO | null>(null);

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: routerPushMock,
    replace: vi.fn(),
  }),
}));

vi.mock('@/infra/stores/useAuthStore', () => ({
  useAuthStore: () => ({
    isAdmin: false,
    user: { username: 'light' },
    logout: authLogoutMock,
  }),
}));

vi.mock('@/infra/gateway/HttpStatsGateway', () => ({
  statsGateway: statsGatewayMock,
}));

vi.mock('@/infra/gateway/HttpMemoGateway', () => ({
  memoGateway: memoGatewayMock,
}));

vi.mock('@/infra/gateway/HttpTagSetGateway', () => ({
  tagSetGateway: tagSetGatewayMock,
}));

vi.mock('@/presentation/composables/useStats', () => ({
  useStats: () => ({
    stats: ref({
      memosTotal: 1,
      tagsTotal: 1,
      resourcesTotal: 1,
      heatmap: [],
    }),
    fetchStats: fetchStatsMock,
  }),
}));

vi.mock('@/presentation/composables/useTheme', () => ({
  useTheme: () => ({
    isDark: ref(false),
    toggleTheme: vi.fn(),
  }),
}));

vi.mock('@/presentation/composables/useTags', () => ({
  useTags: () => ({
    tags: ref([{ name: 'Ops', count: 2 }]),
    fetchTags: fetchTagsMock,
  }),
}));

vi.mock('@/presentation/composables/useBackup', () => ({
  useBackup: () => ({
    fileInput: ref<HTMLInputElement | null>(null),
    importing: importingRef,
    exporting: exportingRef,
    showImportReport: showImportReportRef,
    latestImportReport: latestImportReportRef,
    humanizeReason: humanizeReasonMock,
    handleImport: handleImportMock,
    handleExport: handleExportMock,
  }),
}));

vi.mock('lucide-vue-next', () => {
  const Icon = defineComponent({
    name: 'IconStub',
    setup: () => () => h('span'),
  });
  return {
    Hash: Icon,
    BarChart2: Icon,
    Download: Icon,
    Upload: Icon,
    FileText: Icon,
    Tag: Icon,
    Sparkles: Icon,
    Shield: Icon,
    LogOut: Icon,
    ChevronDown: Icon,
    ChevronRight: Icon,
    Settings: Icon,
    Sun: Icon,
    Moon: Icon,
    PencilLine: Icon,
    Bookmark: Icon,
    Plus: Icon,
    Key: Icon,
  };
});

vi.mock('tdesign-vue-next', () => {
  const Aside = defineComponent({
    name: 'TAside',
    setup(_, { slots }) {
      return () => h('aside', slots.default?.());
    },
  });
  const Button = defineComponent({
    name: 'TButton',
    emits: ['click'],
    setup(_, { slots, emit }) {
      return () =>
        h(
          'button',
          {
            onClick: () => emit('click'),
          },
          slots.default?.(),
        );
    },
  });
  const Tag = defineComponent({
    name: 'TTag',
    setup(_, { slots }) {
      return () => h('span', slots.default?.());
    },
  });
  const Input = defineComponent({
    name: 'TInput',
    setup(_, { slots }) {
      return () => h('input', slots.default?.());
    },
  });
  const Dialog = defineComponent({
    name: 'TDialog',
    props: {
      visible: { type: Boolean, default: false },
      header: { type: String, default: '' },
    },
    setup(props, { slots }) {
      return () =>
        props.visible
          ? h('section', [h('h2', props.header), slots.default?.()])
          : null;
    },
  });
  const Tabs = defineComponent({
    name: 'TTabs',
    setup(_, { slots }) {
      return () => h('div', slots.default?.());
    },
  });
  const TabPanel = defineComponent({
    name: 'TTabPanel',
    setup(_, { slots }) {
      return () => h('section', slots.default?.());
    },
  });
  return {
    Aside,
    Button,
    Tag,
    Input,
    Dialog,
    Tabs,
    TabPanel,
    MessagePlugin: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
    },
  };
});

vi.mock('@/presentation/components/Heatmap.vue', () => ({
  default: defineComponent({
    name: 'HeatmapStub',
    setup: () => () => h('div', 'heatmap'),
  }),
}));

describe('Sidebar', () => {
  const mountSidebar = () =>
    mount(Sidebar, {
      global: {
        stubs: {
          RouterLink: defineComponent({
            name: 'RouterLink',
            setup(_, { slots }) {
              return () => h('a', slots.default?.());
            },
          }),
        },
      },
    });

  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
    showImportReportRef.value = false;
    latestImportReportRef.value = null;
    humanizeReasonMock.mockImplementation((reason: string) => reason);
  });

  it('delegates export action to useBackup handler', async () => {
    const wrapper = mountSidebar();

    await wrapper.get('[data-testid="account-menu-trigger"]').trigger('click');
    await wrapper.get('[data-testid="account-menu-export"]').trigger('click');

    expect(handleExportMock).toHaveBeenCalledTimes(1);
  });

  it('delegates import change event to useBackup handler', async () => {
    const wrapper = mountSidebar();

    const file = new File(['dummy'], 'backup.zip', { type: 'application/zip' });
    const input = wrapper.get('input[type="file"]');
    Object.defineProperty(input.element, 'files', {
      value: [file],
      configurable: true,
    });
    await input.trigger('change');

    expect(handleImportMock).toHaveBeenCalledTimes(1);
  });

  it('renders import report panel when backup report is visible', () => {
    showImportReportRef.value = true;
    latestImportReportRef.value = {
      message: 'ok',
      memosImported: 2,
      resourcesImported: 1,
      memosSkipped: 0,
      resourcesSkipped: 0,
      report: {
        memos: {
          imported: 2,
          skipped: 1,
          details: [{ key: 'memo-1', reason: 'duplicate_by_id' }],
        },
        resources: { imported: 1, skipped: 0, details: [] },
      },
    };
    humanizeReasonMock.mockImplementation(() => '重复 ID');

    const wrapper = mountSidebar();

    expect(wrapper.text()).toContain('导入报告');
    expect(wrapper.text()).toContain('memo-1');
    expect(wrapper.text()).toContain('重复 ID');
  });
});
