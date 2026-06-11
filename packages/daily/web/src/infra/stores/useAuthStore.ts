import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import { authGateway } from '@/infra/gateway/HttpAuthGateway';
import { useMemoStore } from '@/infra/stores/useMemoStore';
import { useTagStore } from '@/infra/stores/useTagStore';
import { useStatsStore } from '@/infra/stores/useStatsStore';
import type {
  AuthUserDTO,
  InviteDTO,
  InviteSummaryDTO,
} from '@/application/ports/dto/Auth';

export type AuthUserVM = AuthUserDTO;
export type InviteVM = InviteDTO;
export type InviteSummaryVM = InviteSummaryDTO;

const RECENT_USERNAMES_KEY = 'daily_recent_usernames';
const MAX_RECENT_USERNAMES = 5;

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUserVM | null>(null);
  const loading = ref(false);
  const initialized = ref(false);
  const recentUsernames = ref<string[]>(loadRecentUsernames());

  const isAuthenticated = computed(() => !!user.value);
  const isAdmin = computed(() => user.value?.role === 'admin');

  const bootstrap = async () => {
    if (initialized.value) return;
    initialized.value = true;
    loading.value = true;
    const res = await authGateway.me();
    if (res.kind === 'success') {
      user.value = res.value;
    } else {
      user.value = null;
    }
    loading.value = false;
  };

  const login = async (payload: {
    username: string;
    password: string;
  }): Promise<string | null> => {
    loading.value = true;
    const res = await authGateway.login(payload);
    loading.value = false;
    if (res.kind === 'failure') return res.error.message;
    user.value = res.value;
    rememberUsername(res.value.username);
    return null;
  };

  const logout = async () => {
    await authGateway.logout();
    user.value = null;
    useMemoStore().clear();
    useTagStore().setTags([]);
    useStatsStore().setStats(null);
  };

  const createInvite = async (
    role: 'admin' | 'member',
    ttlHours: number,
  ): Promise<{ invite: InviteVM | null; error: string | null }> => {
    const res = await authGateway.createInvite(role, ttlHours);
    if (res.kind === 'failure') {
      return { invite: null, error: res.error.message };
    }
    return { invite: res.value, error: null };
  };

  const listInvites = async (
    status = '',
  ): Promise<{ invites: InviteVM[]; error: string | null }> => {
    const res = await authGateway.listInvites(status || undefined);
    if (res.kind === 'failure') {
      return { invites: [], error: res.error.message };
    }
    return { invites: res.value, error: null };
  };

  const revokeInvite = async (id: string): Promise<string | null> => {
    const res = await authGateway.revokeInvite(id);
    if (res.kind === 'failure') return res.error.message;
    return null;
  };

  const getInviteSummary = async (): Promise<{
    summary: InviteSummaryVM | null;
    error: string | null;
  }> => {
    const res = await authGateway.getInviteSummary();
    if (res.kind === 'failure') {
      return { summary: null, error: res.error.message };
    }
    return { summary: res.value, error: null };
  };

  const verifyInvite = async (
    code: string,
  ): Promise<{
    valid: boolean;
    role: string;
    expiresAt: string;
    error: string | null;
  }> => {
    const res = await authGateway.verifyInvite(code);
    if (res.kind === 'failure') {
      return {
        valid: false,
        role: '',
        expiresAt: '',
        error: res.error.message,
      };
    }
    return {
      valid: res.value.valid,
      role: res.value.role,
      expiresAt: res.value.expiresAt,
      error: null,
    };
  };

  const registerByInvite = async (payload: {
    username: string;
    password: string;
    code: string;
  }): Promise<string | null> => {
    const res = await authGateway.registerByInvite(payload);
    if (res.kind === 'failure') return res.error.message;
    user.value = res.value;
    rememberUsername(res.value.username);
    return null;
  };

  const rememberUsername = (username: string) => {
    const normalized = username.trim();
    if (!normalized) return;
    const next = [
      normalized,
      ...recentUsernames.value.filter((name) => name !== normalized),
    ].slice(0, MAX_RECENT_USERNAMES);
    recentUsernames.value = next;
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(RECENT_USERNAMES_KEY, JSON.stringify(next));
    }
  };

  return {
    user,
    loading,
    initialized,
    recentUsernames,
    isAuthenticated,
    isAdmin,
    bootstrap,
    login,
    logout,
    createInvite,
    listInvites,
    getInviteSummary,
    revokeInvite,
    verifyInvite,
    registerByInvite,
    rememberUsername,
  };
});

function loadRecentUsernames(): string[] {
  if (typeof localStorage === 'undefined') return [];
  const raw = localStorage.getItem(RECENT_USERNAMES_KEY);
  if (!raw) return [];
  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter((item): item is string => typeof item === 'string')
      .map((item) => item.trim())
      .filter((item) => item.length > 0)
      .slice(0, MAX_RECENT_USERNAMES);
  } catch {
    return [];
  }
}
