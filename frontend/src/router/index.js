import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/', redirect: '/tickets' },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { bare: true, guest: true, title: 'Sign in' },
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { bare: true, guest: true, title: 'Create account' },
  },
  {
    path: '/tickets',
    name: 'tickets',
    component: () => import('../views/TicketListView.vue'),
    meta: { requiresAuth: true, title: 'Tickets' },
  },
  {
    path: '/tickets/create',
    name: 'ticket-create',
    component: () => import('../views/TicketCreateView.vue'),
    meta: { requiresAuth: true, title: 'New ticket' },
  },
  {
    path: '/tickets/:id',
    name: 'ticket-detail',
    component: () => import('../views/TicketDetailView.vue'),
    meta: { requiresAuth: true, title: 'Ticket' },
  },
  {
    path: '/users',
    name: 'users',
    component: () => import('../views/UserManagementView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, title: 'Users' },
  },
  { path: '/:pathMatch(.*)*', redirect: '/tickets' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

let sessionChecked = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!auth.isAuthenticated) {
    sessionChecked = false
  } else if (!sessionChecked) {
    sessionChecked = true
    await auth.refresh()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requiresAdmin && auth.user?.role !== 'admin') {
    return { name: 'tickets' }
  }

  if (to.meta.guest && auth.isAuthenticated) {
    return { name: 'tickets' }
  }

  return true
})

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · Helpdesk` : 'Helpdesk'
})

export default router
