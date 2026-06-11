<script setup lang="ts">
import { ref, onUnmounted } from 'vue';
import {
  Dialog as TDialog,
  Table as TTable,
  Button as TButton,
  Input as TInput,
  Select as TSelect,
  MessagePlugin,
  Popconfirm as TPopconfirm,
} from 'tdesign-vue-next';
import { Copy, Trash2, Key } from 'lucide-vue-next';
import { useApiKeyStore } from '@/infra/stores/useApiKeyStore';
import { uiCommandBus } from '@/presentation/ui-command/uiCommandBus';

const store = useApiKeyStore();
const visible = ref(false);
const showCreateForm = ref(false);
const newKeyLabel = ref('');
const newKeyTtl = ref<number | undefined>(undefined);
const newlyCreatedKey = ref<string | null>(null);

const ttlOptions = [
  { label: '永不过期', value: undefined },
  { label: '1 小时', value: 1 },
  { label: '24 小时', value: 24 },
  { label: '7 天', value: 168 },
  { label: '30 天', value: 720 },
];

const columns = [
  { colKey: 'label', title: '名称', width: 120 },
  { colKey: 'createdAt', title: '创建于', width: 160 },
  { colKey: 'expiresAt', title: '过期时间', width: 160 },
  { colKey: 'lastUsedAt', title: '最后使用', width: 160 },
  { colKey: 'operation', title: '操作', width: 80, fixed: 'right' as const },
];

const open = () => {
  visible.value = true;
  newlyCreatedKey.value = null;
  void store.fetchKeys();
};

const handleCreate = async () => {
  if (!newKeyLabel.value.trim()) {
    void MessagePlugin.warning('请输入 Key 名称');
    return;
  }

  const { key, error } = await store.createKey(
    newKeyLabel.value,
    newKeyTtl.value,
  );
  if (error) {
    void MessagePlugin.error(`创建失败: ${error}`);
  } else {
    newlyCreatedKey.value = key?.key || null;
    newKeyLabel.value = '';
    showCreateForm.value = false;
    void MessagePlugin.success('创建成功');
  }
};

const handleDelete = async (id: string) => {
  const error = await store.deleteKey(id);
  if (error) {
    void MessagePlugin.error(`删除失败: ${error}`);
  } else {
    void MessagePlugin.success('已删除');
  }
};

const copyToClipboard = (text: string) => {
  void navigator.clipboard.writeText(text);
  void MessagePlugin.success('已复制到剪贴板');
};

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '无';
  return new Date(dateStr).toLocaleString();
};

const unsubscribe = uiCommandBus.on('OpenApiKeyManager', () => {
  open();
});

onUnmounted(() => {
  unsubscribe();
});

defineExpose({ open });
</script>

<template>
  <TDialog
    v-model:visible="visible"
    header="API Key 管理"
    :footer="false"
    width="800px"
    destroy-on-close
  >
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <p class="text-sm text-muted">
          API Key 用于 CLI 登录或第三方工具集成。
        </p>
        <TButton
          v-if="!showCreateForm"
          theme="primary"
          size="small"
          @click="showCreateForm = true"
        >
          新建 API Key
        </TButton>
      </div>

      <div
        v-if="showCreateForm"
        class="rounded-lg border border-border bg-page p-4 space-y-4"
      >
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-tiny font-bold text-muted uppercase"
              >名称 (Label)</label
            >
            <TInput v-model="newKeyLabel" placeholder="例如: Macbook-CLI" />
          </div>
          <div>
            <label class="mb-1 block text-tiny font-bold text-muted uppercase"
              >有效期 (TTL)</label
            >
            <TSelect v-model="newKeyTtl" :options="ttlOptions" />
          </div>
        </div>
        <div class="flex justify-end gap-2">
          <TButton
            variant="outline"
            size="small"
            @click="showCreateForm = false"
            >取消</TButton
          >
          <TButton theme="primary" size="small" @click="handleCreate"
            >确认创建</TButton
          >
        </div>
      </div>

      <div
        v-if="newlyCreatedKey"
        class="rounded-lg border border-accent bg-accent/5 p-4 space-y-2"
      >
        <div class="flex items-center gap-2 text-accent font-bold text-sm">
          <Key :size="16" /> 请保存您的 API Key
        </div>
        <p class="text-xs text-secondary">
          此 Key 仅显示一次，离开页面后将无法再次查看。
        </p>
        <div
          class="flex items-center gap-2 bg-surface border border-border rounded p-2"
        >
          <code class="flex-1 text-sm font-mono break-all">{{
            newlyCreatedKey
          }}</code>
          <TButton
            variant="text"
            size="small"
            @click="copyToClipboard(newlyCreatedKey!)"
          >
            <Copy :size="14" />
          </TButton>
        </div>
      </div>

      <TTable
        :data="store.keys"
        :columns="columns"
        row-key="id"
        size="small"
        :loading="store.loading"
        empty="暂无 API Key"
      >
        <template #createdAt="{ row }">
          <span class="text-xs">{{ formatDate(row.createdAt) }}</span>
        </template>
        <template #expiresAt="{ row }">
          <span
            class="text-xs"
            :class="row.expiresAt ? 'text-primary-text' : 'text-muted'"
          >
            {{ formatDate(row.expiresAt) }}
          </span>
        </template>
        <template #lastUsedAt="{ row }">
          <span class="text-xs">{{ formatDate(row.lastUsedAt) }}</span>
        </template>
        <template #operation="{ row }">
          <TPopconfirm
            content="确认删除此 Key？删除后关联工具将失效。"
            @confirm="handleDelete(row.id)"
          >
            <TButton variant="text" theme="danger" size="small">
              <Trash2 :size="14" />
            </TButton>
          </TPopconfirm>
        </template>
      </TTable>
    </div>
  </TDialog>
</template>
