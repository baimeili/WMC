<!-- 运营后台布局：左侧菜单 + 顶栏（角色信息、退出登录）+ 内容区。 -->
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { labelOf, ROLE_LABELS } from '@/utils/enums'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

// 侧边栏菜单，与 router 中定义的后台路由一一对应。
const menus: Array<{ path: string; label: string }> = [
  { path: '/dashboard', label: '数据看板' },
  { path: '/funnel', label: '核心经营漏斗' },
  { path: '/users', label: '用户管理' },
  { path: '/reports', label: '举报处理' },
  { path: '/photos', label: '照片审核' },
  { path: '/baseline', label: '基准数据集' },
  { path: '/questions', label: '价值观问题' },
  { path: '/settings', label: '系统设置' },
  { path: '/audit-logs', label: '审计日志' }
]

// 详情页高亮所属列表菜单。
const activeMenu = computed<string>(() => {
  const configured = route.meta.activeMenu
  if (typeof configured === 'string') {
    return configured
  }
  return route.path
})

const pageTitle = computed<string>(() => (typeof route.meta.title === 'string' ? route.meta.title : '运营后台'))
const roleLabel = computed<string>(() => labelOf(ROLE_LABELS, auth.role, '访客'))
const userSuffix = computed<string>(() => (auth.userId ? auth.userId.slice(0, 8) : ''))

async function handleLogout(): Promise<void> {
  try {
    await ElMessageBox.confirm('退出后需要重新登录才能继续管理运营数据，确定退出吗？', '退出登录', {
      type: 'warning',
      confirmButtonText: '退出登录',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  await auth.logout()
  ElMessage.success('已退出登录')
  await router.replace('/login')
}
</script>

<template>
  <el-container class="admin-layout">
    <el-aside width="224px" class="admin-side">
      <div class="brand">
        <span class="brand-dot"></span>
        <div>
          <p class="brand-title">婚恋市场计算器</p>
          <p class="brand-sub">运营后台</p>
        </div>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="admin-menu"
        background-color="#1f2d3d"
        text-color="#bfcbd9"
        active-text-color="#ffffff"
      >
        <el-menu-item v-for="item in menus" :key="item.path" :index="item.path">
          {{ item.label }}
        </el-menu-item>
      </el-menu>
      <p class="side-note">理论市场数据为统计估算，禁止对外宣称精确人数或双向匹配率。</p>
    </el-aside>

    <el-container class="admin-body">
      <el-header height="60px" class="admin-header">
        <span class="header-title">{{ pageTitle }}</span>
        <div class="header-right">
          <span class="header-role">{{ roleLabel }}</span>
          <span v-if="userSuffix" class="header-user">{{ userSuffix }}</span>
          <el-button link type="primary" @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main class="admin-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  height: 100vh;
}

.admin-side {
  display: flex;
  flex-direction: column;
  background: #1f2d3d;
  color: #fff;
  overflow-y: auto;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.brand-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #409eff;
  flex: none;
}

.brand-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
}

.brand-sub {
  margin: 2px 0 0;
  color: #8a97a8;
  font-size: 12px;
  letter-spacing: 2px;
}

.admin-menu {
  border-right: none;
  padding: 8px 0;
}

.side-note {
  margin: auto 16px 18px;
  padding-top: 16px;
  color: #7d8a9c;
  font-size: 12px;
  line-height: 1.6;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.admin-body {
  min-width: 0;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  border-bottom: 1px solid var(--wmc-border);
}

.header-title {
  font-size: 16px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 14px;
  color: var(--wmc-text-regular);
  font-size: 13px;
}

.header-role {
  padding: 2px 10px;
  border-radius: 10px;
  background: #ecf5ff;
  color: #409eff;
}

.header-user {
  color: var(--wmc-text-secondary);
  font-family: Consolas, Monaco, monospace;
}

.admin-main {
  padding: 0;
  background: var(--wmc-bg);
  overflow-y: auto;
}
</style>
