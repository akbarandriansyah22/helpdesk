<script setup>
import { reactive, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getErrorMessage } from '../services/api'
import AppIcon from '../components/AppIcon.vue'

const auth = useAuthStore()

const form = reactive({ name: '', email: '', password: '' })
const errors = reactive({ name: '', email: '', password: '' })
const formError = ref('')
const showPassword = ref(false)
const submitting = ref(false)
const success = ref(false)

function validate() {
  const name = form.name.trim()
  errors.name = ''
  if (name.length < 2) errors.name = 'Enter your name (at least 2 characters).'
  else if (name.length > 100) errors.name = 'Use 100 characters or fewer.'
  errors.email = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim()) ? '' : 'Enter a valid email.'
  errors.password = form.password.length >= 8 ? '' : 'Use at least 8 characters.'
  if (form.password.length > 72) errors.password = 'Use 72 characters or fewer.'
  return !errors.name && !errors.email && !errors.password
}

async function submit() {
  formError.value = ''
  success.value = false
  if (!validate()) return

  submitting.value = true
  try {
    await auth.register({
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password,
    })
    success.value = true
  } catch (err) {
    formError.value = getErrorMessage(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo" aria-hidden="true">
        <span class="brand-mark"><AppIcon name="headset" /></span>
      </div>
      <h1>Create account</h1>
      <p class="sub">Register for the internal helpdesk</p>

      <div v-if="success" class="done">
        <p class="alert success" role="status">Account created. You can sign in now.</p>
        <RouterLink class="btn btn-navy auth-submit" to="/login">Go to sign in</RouterLink>
      </div>

      <form v-else novalidate @submit.prevent="submit">
        <p v-if="formError" class="alert error" role="alert">{{ formError }}</p>

        <label class="field" for="name">
          Name
          <span class="input-wrap">
            <AppIcon name="users" />
            <input id="name" v-model="form.name" type="text" autocomplete="name" placeholder="Your name" />
          </span>
          <span v-if="errors.name" class="field-error">{{ errors.name }}</span>
        </label>

        <label class="field" for="email">
          Email
          <span class="input-wrap">
            <AppIcon name="mail" />
            <input
              id="email"
              v-model="form.email"
              type="email"
              autocomplete="email"
              placeholder="you@company.com"
            />
          </span>
          <span v-if="errors.email" class="field-error">{{ errors.email }}</span>
        </label>

        <label class="field" for="password">
          Password
          <span class="input-wrap">
            <AppIcon name="lock" />
            <input
              id="password"
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              autocomplete="new-password"
              placeholder="At least 8 characters"
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
          <span v-if="errors.password" class="field-error">{{ errors.password }}</span>
        </label>

        <button class="btn btn-navy auth-submit" type="submit" :disabled="submitting">
          {{ submitting ? 'Creating account...' : 'Create account' }}
        </button>
      </form>

      <p class="switch">
        Already have an account?
        <RouterLink to="/login">Sign in</RouterLink>
      </p>
    </div>
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

.done .auth-submit {
  margin-top: 14px;
}
</style>
