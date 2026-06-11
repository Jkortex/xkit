import { computed, onMounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import {
  useAuthStore,
  type InviteSummaryVM,
  type InviteVM,
} from '@/infra/stores/useAuthStore';

export const useAdminInvites = () => {
  const auth = useAuthStore();

  const role = ref<'member' | 'admin'>('member');
  const ttlHours = ref(48);
  const statusFilter = ref<'all' | 'active' | 'used' | 'revoked' | 'expired'>(
    'all',
  );
  const actionError = ref('');
  const loadError = ref('');
  const initialized = ref(false);
  const creating = ref(false);
  const loadingInvites = ref(false);
  const invites = ref<InviteVM[]>([]);
  const inviteSummary = ref<InviteSummaryVM>({
    activeMemberCount: 0,
    activeAdminCount: 0,
    memberLimit: 20,
    adminLimit: 5,
  });
  const revokingIds = ref<Set<string>>(new Set());
  const inviteCodeById = ref<Record<string, string>>({});

  const roleOptions: Array<{ label: string; value: 'member' | 'admin' }> = [
    { label: '成员（member）', value: 'member' },
    { label: '管理员（admin）', value: 'admin' },
  ];

  const statusOptions: Array<{
    label: string;
    value: 'all' | 'active' | 'used' | 'revoked' | 'expired';
  }> = [
    { label: '全部状态', value: 'all' },
    { label: '活跃', value: 'active' },
    { label: '已使用', value: 'used' },
    { label: '已撤销', value: 'revoked' },
    { label: '已过期', value: 'expired' },
  ];

  const fetchInvites = async () => {
    loadingInvites.value = true;
    loadError.value = '';
    const result = await auth.listInvites();
    loadingInvites.value = false;
    if (result.error) {
      loadError.value = result.error;
      return;
    }
    invites.value = result.invites;
    initialized.value = true;
  };

  const fetchInviteSummary = async () => {
    const result = await auth.getInviteSummary();
    if (result.summary) {
      inviteSummary.value = result.summary;
      return;
    }
    const activeMemberCount = invites.value.filter(
      (item) => item.status === 'active' && item.role === 'member',
    ).length;
    const activeAdminCount = invites.value.filter(
      (item) => item.status === 'active' && item.role === 'admin',
    ).length;
    inviteSummary.value = {
      activeMemberCount,
      activeAdminCount,
      memberLimit: inviteSummary.value.memberLimit,
      adminLimit: inviteSummary.value.adminLimit,
    };
  };

  const init = async () => {
    if (initialized.value || loadingInvites.value) {
      return;
    }
    await fetchInvites();
    await fetchInviteSummary();
  };

  const retryLoad = async () => {
    await fetchInvites();
    await fetchInviteSummary();
  };

  const createInvite = async () => {
    actionError.value = '';
    creating.value = true;
    const result = await auth.createInvite(role.value, ttlHours.value);
    creating.value = false;
    if (result.error) {
      actionError.value = result.error;
      return;
    }
    if (result.invite) {
      if (result.invite.code) {
        inviteCodeById.value[result.invite.id] = result.invite.code;
      }
      await fetchInvites();
      await fetchInviteSummary();
      MessagePlugin.success('邀请码已创建');
    }
  };

  const revokeInvite = async (id: string) => {
    revokingIds.value.add(id);
    const err = await auth.revokeInvite(id);
    revokingIds.value.delete(id);
    if (err) {
      actionError.value = err;
      return;
    }
    await fetchInvites();
    await fetchInviteSummary();
    MessagePlugin.success('邀请码已撤销');
  };

  const copyInviteCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      MessagePlugin.success('邀请码已复制');
    } catch {
      MessagePlugin.error('复制失败，请手动复制');
    }
  };

  const codeForInvite = (invite: InviteVM): string => {
    return inviteCodeById.value[invite.id] || invite.code || invite.id;
  };

  const visibleInvites = computed(() => {
    if (statusFilter.value === 'all') {
      return invites.value;
    }
    return invites.value.filter((item) => item.status === statusFilter.value);
  });

  const inviteCountLabel = computed(() => {
    return `活跃配额：member ${inviteSummary.value.activeMemberCount}/${inviteSummary.value.memberLimit}，admin ${inviteSummary.value.activeAdminCount}/${inviteSummary.value.adminLimit}`;
  });

  onMounted(() => {
    void init();
  });

  return {
    role,
    ttlHours,
    statusFilter,
    actionError,
    loadError,
    initialized,
    creating,
    loadingInvites,
    invites,
    inviteSummary,
    revokingIds,
    roleOptions,
    statusOptions,
    visibleInvites,
    inviteCountLabel,
    init,
    retryLoad,
    createInvite,
    revokeInvite,
    copyInviteCode,
    codeForInvite,
  };
};
