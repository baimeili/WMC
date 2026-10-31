// 运营后台路由与访问守卫。
// 未登录跳转 /login；已登录但角色不足（非审核员/管理员/超级管理员）同样拒绝进入。
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { public: true, title: '登录' }
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { title: '数据看板' }
      },
      {
        path: 'funnel',
        name: 'funnel',
        component: () => import('@/views/FunnelView.vue'),
        meta: { title: '核心经营漏斗' }
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/UsersView.vue'),
        meta: { title: '用户管理' }
      },
      {
        path: 'users/:id',
        name: 'user-detail',
        component: () => import('@/views/UserDetailView.vue'),
        meta: { title: '用户详情', activeMenu: '/users' }
      },
      {
        path: 'reports',
        name: 'reports',
        component: () => import('@/views/ReportsView.vue'),
        meta: { title: '举报处理' }
      },
      {
        path: 'photos',
        name: 'photos',
        component: () => import('@/views/PhotosView.vue'),
        meta: { title: '照片审核' }
      },
      {
        path: 'baseline',
        name: 'baseline',
        component: () => import('@/views/BaselineView.vue'),
        meta: { title: '基准数据集' }
      },
      {
        path: 'baseline/:id',
        name: 'baseline-detail',
        component: () => import('@/views/BaselineDetailView.vue'),
        meta: { title: '数据集详情', activeMenu: '/baseline' }
      },
      {
        path: 'questions',
        name: 'questions',
        component: () => import('@/views/QuestionsView.vue'),
        meta: { title: '价值观问题' }
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/SettingsView.vue'),
        meta: { title: '系统设置' }
      },
      {
        path: 'audit-logs',
        name: 'audit-logs',
        component: () => import('@/views/AuditLogsView.vue'),
        meta: { title: '审计日志' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 全局守卫：校验登录态与后台角色。
router.beforeEach((to) => {
  const auth = useAuthStore()
  const isPublic = Boolean(to.meta.public)
  if (isPublic) {
    // 已登录且有权限时，访问登录页直接回到看板。
    if (auth.isLoggedIn && auth.canAccessAdmin) {
      return '/dashboard'
    }
    return true
  }
  if (!auth.isLoggedIn) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  if (!auth.canAccessAdmin) {
    auth.clear()
    return { path: '/login', query: { redirect: to.fullPath } }
  }
  return true
})

// 同步页面标题。
router.afterEach((to) => {
  const title = typeof to.meta.title === 'string' ? to.meta.title : ''
  document.title = title ? `${title} · 婚恋市场计算器运营后台` : '婚恋市场计算器 · 运营后台'
})

export default router