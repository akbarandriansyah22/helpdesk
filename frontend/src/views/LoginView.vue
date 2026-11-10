<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getErrorMessage, useDummy } from '../services/api'
import AppIcon from '../components/AppIcon.vue'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const showPassword = ref(false)
const error = ref('')
const submitting = ref(false)

const demos = [
  { label: 'Admin', email: 'alex.kim@acme.com' },
  { label: 'Agent', email: 'sarah.mitchell@acme.com' },
  { label: 'Requester', email: 'james.carter@acme.com' },
]

function fillDemo(value) {
  email.value = value
  password.value = 'password123'
  error.value = ''
}

function safeRedirect(value) {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//')) {
    return '/tickets'
  }
  return value
}

async function submit() {
  error.value = ''
  if (!email.value.trim() || !password.value) {
    error.value = 'Enter your email and password.'
    return
  }

  submitting.value = true
  try {
    await auth.login(email.value, password.value)
    await router.push(safeRedirect(route.query.redirect))
  } catch (err) {
    error.value = getErrorMessage(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <form class="auth-card" novalidate @submit.prevent="submit">
      <div class="auth-logo" aria-hidden="true">
        <span class="brand-mark"><AppIcon name="headset" /></span>
      </div>
      <h1>Helpdesk Login</h1>
      <p class="sub">Sign in to access the internal helpdesk</p>

      <p v-if="error" class="alert error" role="alert">{{ error }}</p>

      <label class="field" for="email">
        Email
        <span class="input-wrap">
          <AppIcon name="mail" />
          <input
            id="email"
            v-model="email"
            type="email"
            autocomplete="username"
            placeholder="you@company.com"
          />
        </span>
      </label>

      <label class="field" for="password">
        Password
        <span class="input-wrap">
          <AppIcon name="lock" />
          <input
            id="password"
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="current-password"
            placeholder="Enter your password"
          />
          <button
            class="eye"
            type="button"
            :aria-label="showPassword ? 'Hide password' : 'Show password'"
            @click="showPassword = !showPassword"
          >
            <AppIcon :name="showPassword ? 'eye-off' : 'eye'" />
          </button>
        </span>
      </label>

      <button class="btn btn-navy auth-submit" type="submit" :disabled="submitting">
        {{ submitting ? 'Signing in...' : 'Sign in' }}
      </button>

      <p class="switch">
        Need an account?
        <RouterLink to="/register">Create one</RouterLink>
      </p>

      <div v-if="useDummy" class="demo">
        <p>Demo password: password123</p>
        <div class="demo-list">
          <button
            v-for="demo in demos"
            :key="demo.email"
            type="button"
            class="demo-chip"
            @click="fillDemo(demo.email)"
          >
            <strong>{{ demo.label }}</strong>
            {{ demo.email }}
          </button>
        </div>
      </div>
    </form>
  </div>
</template>

<style scoped>
.auth-logo {
  display: grid;
  place-items: center;
}

.auth-logo .brand-mark {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  font-size: 26px;
}

.auth-submit {
  width: 100%;
  height: 46px;
  margin-top: 6px;
}

.eye {
  border: 0;
  background: transparent;
  color: #9ca3af;
  display: grid;
  place-items: center;
  padding: 4px;
}

.switch {
  text-align: center;
  margin: 16px 0 0;
  color: var(--muted);
  font-size: 14px;
}

.demo {
  margin-top: 22px;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}

.demo p {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 12px;
  text-align: center;
}

.demo-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.demo-chip {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  border: 1px solid var(--border);
  background: #fafafa;
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 12px;
  color: #4b5563;
  text-align: left;
}

.demo-chip strong {
  color: var(--text);
  font-weight: 650;
}

.demo-chip:hover {
  border-color: #c4b5fd;
  background: #faf5ff;
}
</style>
