<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Button as TButton, Input as TInput } from 'tdesign-vue-next';
import { useAuthStore } from '@/infra/stores/useAuthStore';

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const code = computed(() => String(route.query.code ?? ''));
const username = ref('');
const password = ref('');
const verifyLoading = ref(false);
const validInvite = ref(false);
const inviteRole = ref('');
const inviteExpiresAt = ref('');
const error = ref('');

const verify = async () => {
  if (!code.value) {
    error.value = '邀请码缺失';
    return;
  }
  verifyLoading.value = true;
  const res = await auth.verifyInvite(code.value);
  verifyLoading.value = false;
  if (res.error) {
    error.value = res.error;
    return;
  }
  validInvite.value = res.valid;
  inviteRole.value = res.role;
  inviteExpiresAt.value = res.expiresAt;
  if (!res.valid) error.value = '邀请码无效或已过期';
};

const handleSubmit = async () => {
  error.value = '';
  const err = await auth.registerByInvite({
    code: code.value,
    username: username.value,
    password: password.value,
  });
  if (err) {
    error.value = err;
    return;
  }
  await router.replace('/');
};

onMounted(() => {
  void verify();
});
</script>

<template>
  <main
    class="min-h-screen grid place-items-center px-6 py-10 bg-(--daily-bg-page)"
  >
    <div
      class="w-full max-w-md rounded-xl border border-(--daily-border) bg-(--daily-bg-surface) p-6 space-y-5"
    >
      <header class="space-y-1">
        <h1 class="text-2xl font-semibold text-(--daily-text-primary)">
          邀请注册
        </h1>
        <p class="text-sm text-(--daily-text-secondary)">通过邀请码创建账号</p>
      </header>

      <p v-if="verifyLoading" class="text-sm text-(--daily-text-secondary)">
        正在校验邀请码...
      </p>
      <p
        v-else-if="validInvite"
        class="text-sm text-(--daily-success) break-all whitespace-pre-wrap"
      >
        邀请有效，角色：{{ inviteRole }}，到期：{{ inviteExpiresAt }}
      </p>
      <p v-if="error" class="text-sm text-(--daily-error)">{{ error }}</p>

      <form v-if="validInvite" class="space-y-4" @submit.prevent="handleSubmit">
        <label class="block space-y-2">
          <span class="text-sm text-(--daily-text-secondary)">用户名</span>
          <TInput
            v-model="username"
            size="large"
            clearable
            placeholder="请输入用户名"
          />
        </label>
        <label class="block space-y-2">
          <span class="text-sm text-(--daily-text-secondary)">密码</span>
          <TInput
            v-model="password"
            type="password"
            size="large"
            clearable
            placeholder="请输入密码"
          />
        </label>

        <TButton class="w-full" size="large" theme="primary" type="submit">
          注册并登录
        </TButton>
      </form>
    </div>
  </main>
</template>
