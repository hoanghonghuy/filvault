import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    return { left: 0, top: 0 }
  },
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { guest: true } },
    { path: '/register', name: 'register', component: () => import('@/views/RegisterView.vue'), meta: { guest: true } },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('@/views/VerifyEmailView.vue'),
      meta: { auth: true, unverifiedOnly: true },
    },
    { path: '/', name: 'home', redirect: '/chat' },
    { path: '/overview', name: 'overview', component: () => import('@/views/OverviewView.vue'), meta: { auth: true, verified: true } },
    { path: '/files', name: 'files', component: () => import('@/views/FilesView.vue'), meta: { auth: true, verified: true } },
    { path: '/photos', name: 'photos', component: () => import('@/views/PhotosView.vue'), meta: { auth: true, verified: true } },
    { path: '/chat', name: 'chat', component: () => import('@/views/ChatView.vue'), meta: { auth: true, verified: true, bare: true } },
    {
      path: '/photos/albums/:id',
      name: 'album',
      component: () => import('@/views/AlbumView.vue'),
      meta: { auth: true, verified: true },
    },
    { path: '/trash', name: 'trash', component: () => import('@/views/TrashView.vue'), meta: { auth: true, verified: true } },
    {
      path: '/shared',
      name: 'shared',
      component: () => import('@/views/SharedWithMeView.vue'),
      meta: { auth: true, verified: true },
    },
    { path: '/settings', name: 'settings', component: () => import('@/views/SettingsView.vue'), meta: { auth: true, verified: true } },
    { path: '/profile', name: 'profile', component: () => import('@/views/ProfileView.vue'), meta: { auth: true, verified: true } },
    {
      path: '/s/:token',
      name: 'public-share',
      component: () => import('@/views/PublicShareView.vue'),
    },
    { path: '/:pathMatch(.*)*', redirect: '/chat' },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.user && localStorage.getItem('filvault.accessToken')) {
    await auth.bootstrap()
  }
  if (to.meta.guest && auth.isAuthenticated) {
    return auth.isVerified ? '/chat' : '/verify-email'
  }
  if (to.meta.auth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if ((to.name === 'verify-email' || to.meta.unverifiedOnly) && auth.isAuthenticated && auth.isVerified) {
    return '/chat'
  }
  if (to.meta.verified && auth.isAuthenticated && !auth.isVerified) {
    return '/verify-email'
  }
  return true
})

export default router
