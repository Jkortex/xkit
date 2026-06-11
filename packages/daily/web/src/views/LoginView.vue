<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { Button as TButton, Input as TInput } from 'tdesign-vue-next';
import { useAuthStore } from '@/infra/stores/useAuthStore';

const router = useRouter();
const auth = useAuthStore();

const username = ref('');
const password = ref('');
const error = ref('');
const passwordInputRef = ref<InstanceType<typeof TInput> | null>(null);

const handleSubmit = async () => {
  error.value = '';
  const err = await auth.login({
    username: username.value,
    password: password.value,
  });
  if (err) {
    error.value = err;
    return;
  }
  await router.replace('/');
};

const pickRecentUsername = (value: string) => {
  username.value = value;
  const node = passwordInputRef.value?.$el as HTMLElement | undefined;
  const input = node?.querySelector('input');
  input?.focus();
};
</script>

<template>
  <main class="min-h-screen grid place-items-center px-6 py-10 bg-page">
    <div
      class="w-full max-w-md rounded-xl border border-border bg-surface p-6 space-y-5"
    >
      <header class="space-y-1">
        <h1 class="text-2xl font-semibold text-primary-text">登录</h1>
      </header>

      <form class="space-y-4" @submit.prevent="handleSubmit">
        <div v-if="auth.recentUsernames.length > 0" class="space-y-2">
          <span class="text-sm text-secondary">最近账号</span>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="item in auth.recentUsernames"
              :key="item"
              type="button"
              class="ui-btn-chip-sm"
              @click="pickRecentUsername(item)"
            >
              {{ item }}
            </button>
          </div>
        </div>

        <label class="block space-y-2">
          <span class="text-sm text-secondary">用户名</span>
          <TInput
            v-model="username"
            autocomplete="username"
            placeholder="请输入用户名"
            size="large"
            clearable
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm text-secondary">密码</span>
          <TInput
            ref="passwordInputRef"
            v-model="password"
            type="password"
            autocomplete="current-password"
            placeholder="请输入密码"
            size="large"
            clearable
          />
        </label>

        <p v-if="error" class="text-sm text-error">
          {{ error }}
        </p>

        <TButton
          class="w-full"
          size="large"
          theme="primary"
          type="submit"
          :disabled="auth.loading"
        >
          {{ auth.loading ? '登录中...' : '登录' }}
        </TButton>
      </form>
    </div>
  </main>
</template>
