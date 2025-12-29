import { createRouter, createWebHistory } from 'vue-router'
import VariableList from '../views/VariableList.vue'
import ProviderList from '../views/ProviderList.vue'
import Provider from '../views/ProviderItem.vue'
import ServerList from '../views/ServerList.vue'
import ServerLog from '../views/ServerLog.vue'
import Server from '../views/ServerItem.vue'
import HomeDash from '../views/HomeDash.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: HomeDash,
    meta: { title: 'Dashboard' },
  },
  {
    path: '/variables',
    name: 'Variables',
    component: VariableList,
    meta: { title: 'Environment Variables' },
  },
  {
    path: '/providers',
    name: 'Providers',
    component: ProviderList,
    meta: { title: 'API Providers' },
  },
  {
    path: '/providers/:id',
    name: 'ProviderDetail',
    component: Provider,
    meta: { title: 'Provider Details' },
  },
  {
    path: '/servers',
    name: 'Servers',
    component: ServerList,
    meta: { title: 'MCP Servers' },
  },
  {
    path: '/servers/:id',
    name: 'ServerDetail',
    component: Server,
    meta: { title: 'MCP Server Details' },
  },
  {
    path: '/servers/new',
    name: 'ServerCreate',
    component: Server,
    meta: { title: 'Create New MCP Server' },
  },
  {
    path: '/servers/:id/logs',
    name: 'ServerLogs',
    component: ServerLog,
    meta: { title: 'MCP Server Logs' },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  // Use a default title if meta.title is missing
  const title = to.meta.title || 'HasMCP'
  document.title = title + ' | HasMCP'
  next()
})

export default router
