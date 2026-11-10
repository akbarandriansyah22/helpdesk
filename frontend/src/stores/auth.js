import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { TOKEN_KEY, USER_KEY } from '../constants'
import { fetchMe, loginRequest, registerRequest } from '../services/api'

function readUser() {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem(TOKEN_KEY) || '')
  const user = ref(readUser())

  const isAuthenticated = computed(() => Boolean(token.value && user.value))
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isStaff = computed(() => user.value?.role === 'agent' || user.value?.role === 'admin')

  function persist() {
    if (token.value) localStorage.setItem(TOKEN_KEY, token.value)
    else localStorage.removeItem(TOKEN_KEY)

    if (user.value) localStorage.setItem(USER_KEY, JSON.stringify(user.value))
    else localStorage.removeItem(USER_KEY)
  }

  async function login(email, password) {
    const result = await loginRequest(email, password)
    token.value = result.token
    user.value = result.user
    persist()
    return result.user
  }

  async function register(payload) {
    return registerRequest(payload)
  }

  async function refresh() {
    if (!token.value) return null
    try {
      user.value = await fetchMe()
      persist()
      return user.value
    } catch {
      return user.value
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    persist()
  }

  return { token, user, isAuthenticated, isAdmin, isStaff, login, register, refresh, logout }
})
