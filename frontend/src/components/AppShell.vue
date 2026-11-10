<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { avatarColor, initials, staffLabel } from '../utils/format'
import AppIcon from './AppIcon.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const open = ref(false)

const nav = computed(() => {
  const items = [{ to: '/tickets', label: 'Tickets', icon: 'ticket', active: route.path.startsWith('/tickets') }]
  if (auth.isAdmin) {
    items.push({ to: '/users', label: 'Users', icon: 'users', active: route.path.startsWith('/users') })
  }
  return items
})

watch(
  () => route.fullPath,
  () => {
    open.value = false
  },
)

function onKeydown(event) {
  if (event.key === 'Escape') open.value = false
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="shell">
    <div v-if="open" class="backdrop" @click="open = false"></div>
    <aside class="sidebar" :class="{ open }">
      <RouterLink class="brand" to="/tickets">
        <span class="brand-mark"><AppIcon name="headset" /></span>
        <span>Helpdesk</span>
      </RouterLink>

      <nav class="nav" aria-label="Main">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="nav-link"
          :class="{ active: item.active }"
        >
          <AppIcon :name="item.icon" />
          {{ item.label }}
        </RouterLink>
      </nav>

      <div v-if="auth.user" class="user-card">
        <span class="avatar" :style="{ background: avatarColor(auth.user.name) }">
          {{ initials(auth.user.name) }}
        </span>
        <div class="user-copy">
          <div class="user-name">{{ auth.user.name }}</div>
          <div class="user-role">{{ staffLabel(auth.user.role) }}</div>
        </div>
        <button class="logout" type="button" aria-label="Log out" @click="logout">
          <AppIcon name="logout" />
        </button>
      </div>
    </aside>

    <div class="content">
      <header class="mobile-bar">
        <button
          class="icon-button"
          type="button"
          aria-label="Open menu"
          :aria-expanded="open"
          @click="open = true"
        >
          <AppIcon name="menu" />
        </button>
        <RouterLink class="mobile-brand" to="/tickets">Helpdesk</RouterLink>
      </header>
      <main class="main">
        <slot />
      </main>
    </div>
  </div>
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: flex;
  background: var(--bg);
}

.sidebar {
  width: var(--sidebar);
  flex: none;
  background: #fff;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 18px 14px 14px;
  position: sticky;
  top: 0;
  height: 100vh;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 8px 18px;
  color: var(--text);
  font-weight: 700;
  font-size: 18px;
  letter-spacing: -0.03em;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  color: #4b5563;
  font-weight: 550;
  font-size: 14px;
}

.nav-link.active {
  background: #f3e8ff;
  color: #7c3aed;
}

.user-card {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--border);
  border-radius: 14px;
}

.user-copy {
  min-width: 0;
  flex: 1;
}

.user-name {
  font-size: 14px;
  font-weight: 650;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  color: var(--muted);
  font-size: 12px;
  margin-top: 2px;
}

.logout {
  width: 34px;
  height: 34px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #6b7280;
  display: grid;
  place-items: center;
}

.logout:hover {
  background: #f3f4f6;
  color: var(--text);
}

.content {
  flex: 1;
  min-width: 0;
}

.mobile-bar {
  display: none;
}

.main {
  padding: 20px 28px 48px;
}

.backdrop {
  display: none;
}

@media (max-width: 900px) {
  .sidebar {
    position: fixed;
    z-index: 40;
    transform: translateX(-105%);
    transition: transform 0.2s ease;
  }

  .sidebar.open {
    transform: translateX(0);
    box-shadow: 0 12px 40px rgba(17, 24, 39, 0.16);
  }

  .backdrop {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(17, 24, 39, 0.35);
    z-index: 30;
  }

  .mobile-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    position: sticky;
    top: 0;
    z-index: 20;
    height: 56px;
    padding: 0 16px;
    background: #fff;
    border-bottom: 1px solid var(--border);
  }

  .mobile-brand {
    font-weight: 700;
    color: var(--text);
    letter-spacing: -0.03em;
  }

  .main {
    padding: 16px 16px 32px;
  }
}
</style>
