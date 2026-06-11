import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/infra/stores/useAuthStore';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { publicOnly: true },
    },
    {
      path: '/register/invite',
      name: 'invite-register',
      component: () => import('@/views/InviteRegisterView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/stats',
      name: 'stats',
      component: () => import('@/views/StatsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin/invites',
      name: 'admin-invites',
      component: () => import('@/views/AdminInvitesView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
    },
    {
      path: '/tag-sets',
      name: 'tag-sets',
      component: () => import('@/views/TagSetManageView.vue'),
      meta: { requiresAuth: true },
    },
  ],
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.initialized) {
    await auth.bootstrap();
  }
  const isPublic = to.meta.public === true;
  const publicOnly = to.meta.publicOnly === true;
  const requiresAuth = to.meta.requiresAuth === true;
  const requiresAdmin = to.meta.requiresAdmin === true;

  if (publicOnly && auth.isAuthenticated) return '/';
  if (isPublic) return true;
  if (requiresAuth && !auth.isAuthenticated) return '/login';
  if (requiresAdmin && !auth.isAdmin) return '/';
  return true;
});

export default router;
