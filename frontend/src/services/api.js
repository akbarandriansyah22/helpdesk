import axios from 'axios'
import { TOKEN_KEY, USER_KEY } from '../constants'
import * as dummy from '../data/dummy'

export { TOKEN_KEY, USER_KEY }

export const useDummy = import.meta.env.VITE_USE_DUMMY !== 'false'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const url = error.config?.url || ''
    const isAuthAttempt = url.includes('/auth/login') || url.includes('/auth/register')
    if (status === 401 && !isAuthAttempt) {
      redirectToLogin()
    }
    return Promise.reject(error)
  },
)

function redirectToLogin() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  if (!window.location.pathname.startsWith('/login')) {
    window.location.assign('/login')
  }
}

export function getErrorMessage(error) {
  return error?.response?.data?.error || error?.message || 'Something went wrong'
}

async function call(dummyFn, request) {
  if (useDummy) {
    await new Promise((resolve) => setTimeout(resolve, 280))
    try {
      return await dummyFn()
    } catch (error) {
      if (error?.redirect) redirectToLogin()
      throw error
    }
  }

  const response = await request()
  return response.data?.data
}

export function loginRequest(email, password) {
  return call(
    () => dummy.login(email, password),
    () => api.post('/auth/login', { email, password }),
  )
}

export function registerRequest(payload) {
  return call(
    () => dummy.register(payload),
    () => api.post('/auth/register', payload),
  )
}

export function fetchTickets() {
  return call(
    () => dummy.listTickets(),
    () => api.get('/tickets'),
  )
}

export function fetchTicket(id) {
  return call(
    () => dummy.getTicket(id),
    () => api.get(`/tickets/${id}`),
  )
}

export function createTicket(payload) {
  return call(
    () => dummy.createTicket(payload),
    () => api.post('/tickets', payload),
  )
}

export function updateTicketStatus(id, status) {
  return call(
    () => dummy.updateStatus(id, status),
    () => api.patch(`/tickets/${id}/status`, { status }),
  )
}

export function fetchComments(ticketId) {
  return call(
    () => dummy.listComments(ticketId),
    () => api.get(`/tickets/${ticketId}/comments`),
  )
}

export function createComment(ticketId, body) {
  return call(
    () => dummy.createComment(ticketId, body),
    () => api.post(`/tickets/${ticketId}/comments`, { body, is_internal: false }),
  )
}

export function fetchUsers() {
  return call(
    () => dummy.listUsers(),
    () => api.get('/users'),
  )
}

export function updateUserRole(id, role) {
  return call(
    () => dummy.updateRole(id, role),
    () => api.patch(`/users/${id}/role`, { role }),
  )
}
