<script setup lang="ts">
import { ref } from 'vue';
import { Button as TButton, Dialog as TDialog } from 'tdesign-vue-next';
import InviteCard from '@/presentation/components/invites/InviteCard.vue';
import InviteCreatePanel from '@/presentation/components/invites/InviteCreatePanel.vue';
import InviteFilterPanel from '@/presentation/components/invites/InviteFilterPanel.vue';
import { useDialogOpenFlag } from '@/presentation/composables/hotkeys/useDialogOpenFlag';
import { useAdminInvites } from '@/presentation/composables/useAdminInvites';

const {
  role,
  ttlHours,
  statusFilter,
  actionError,
  loadError,
  creating,
  loadingInvites,
  revokingIds,
  roleOptions,
  statusOptions,
  visibleInvites,
  inviteCountLabel,
  retryLoad,
  createInvite,
  revokeInvite,
  copyInviteCode,
  codeForInvite,
} = useAdminInvites();

const showCreateDialog = ref(false);

useDialogOpenFlag(showCreateDialog);

const openCreateDialog = () => {
  showCreateDialog.value = true;
};

const submitCreateInvite = async () => {
  await createInvite();
  if (actionError.value === '') {
    showCreateDialog.value = false;
  }
};
</script>

<template>
  <div>
    <div class="admin-invites-sticky ui-layer-sticky">
      <div class="admin-invites-shell">
        <div class="admin-invites-toolbar">
          <header class="admin-invites-head">
            <h1 class="admin-invites-title">邀请管理</h1>
            <p class="admin-invites-desc">仅管理员可创建和撤销邀请码</p>
            <div class="admin-invites-quota">{{ inviteCountLabel }}</div>
          </header>

          <div class="admin-invites-actions">
            <InviteFilterPanel
              :status="statusFilter"
              :status-options="statusOptions"
              variant="toolbar"
              class="admin-invites-filter"
              @update:status="statusFilter = $event"
            />

            <TButton
              theme="primary"
              size="medium"
              class="admin-invites-create-btn"
              @click="openCreateDialog"
            >
              新建邀请码
            </TButton>
          </div>
        </div>
      </div>
    </div>

    <section class="mx-auto w-full max-w-6xl p-4 pt-6 md:p-8 md:pt-6">
      <div v-if="loadingInvites" class="ui-empty-board py-8">加载中...</div>
      <div v-else-if="loadError" class="ui-empty-board py-8 space-y-3">
        <p>加载邀请码失败：{{ loadError }}</p>
        <TButton theme="primary" variant="outline" @click="retryLoad">
          重试
        </TButton>
      </div>
      <div v-else-if="visibleInvites.length === 0" class="ui-empty-board py-8">
        当前条件下没有邀请码记录。
      </div>

      <div v-else class="grid gap-3 md:grid-cols-2">
        <InviteCard
          v-for="invite in visibleInvites"
          :key="invite.id"
          :invite="invite"
          :code="codeForInvite(invite)"
          :revoking="revokingIds.has(invite.id)"
          @copy="copyInviteCode"
          @revoke="revokeInvite"
        />
      </div>
    </section>

    <TDialog
      v-model:visible="showCreateDialog"
      header="新建邀请码"
      :footer="false"
      width="560px"
      destroy-on-close
    >
      <InviteCreatePanel
        :role="role"
        :ttl-hours="ttlHours"
        :creating="creating"
        :error="actionError"
        :role-options="roleOptions"
        variant="card"
        @update:role="role = $event"
        @update:ttl-hours="ttlHours = $event"
        @create="submitCreateInvite"
      />
    </TDialog>
  </div>
</template>

<style scoped>
.admin-invites-sticky {
  position: sticky;
  top: 0;
  width: 100%;
  border-bottom: 1px solid
    color-mix(in oklab, var(--color-border) 86%, transparent);
  background: color-mix(in oklab, var(--color-surface) 92%, var(--color-page));
  backdrop-filter: saturate(108%) blur(5px);
}

.admin-invites-shell {
  width: 100%;
  max-width: 72rem;
  margin: 0 auto;
}

.admin-invites-toolbar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 1rem;
  align-items: center;
  justify-content: space-between;
  min-height: 4.4rem;
  padding: 0.55rem 1.1rem;
}

.admin-invites-head {
  display: grid;
  gap: 0.1rem;
  min-width: 0;
}

.admin-invites-title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-primary-text);
}

.admin-invites-desc {
  margin: 0;
  font-size: 0.78rem;
  color: var(--color-secondary);
}

.admin-invites-quota {
  margin-top: 0.2rem;
  width: fit-content;
  border-radius: 999px;
  border: 1px solid color-mix(in oklab, var(--color-border) 82%, transparent);
  background: color-mix(in oklab, var(--color-surface) 82%, transparent);
  padding: 0.16rem 0.55rem;
  font-size: 0.72rem;
  color: var(--color-secondary);
}

.admin-invites-actions {
  display: flex;
  align-items: end;
  gap: 0.5rem;
}

.admin-invites-filter {
  min-width: 180px;
}

.admin-invites-create-btn {
  white-space: nowrap;
}

@media (max-width: 1100px) {
  .admin-invites-toolbar {
    grid-template-columns: 1fr;
    align-items: stretch;
    gap: 0.7rem;
    padding: 0.55rem 0.8rem;
  }

  .admin-invites-actions {
    justify-content: space-between;
  }
}

@media (max-width: 680px) {
  .admin-invites-actions {
    flex-direction: column;
    align-items: stretch;
    gap: 0.45rem;
  }

  .admin-invites-filter,
  .admin-invites-create-btn {
    width: 100%;
  }
}
</style>
