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

function compactParams(params) {
  const query = {}
  for (const [key, value] of Object.entries(params || {})) {
    if (value !== '' && value !== null && value !== undefined) query[key] = value
  }
  return query
}

export async function fetchMe() {
  return call(
    () => dummy.me(),
    () => api.get('/me'),
  )
}

// List responses keep pagination beside the ticket array: { data, meta }.
export async function fetchTickets(params = {}) {
  const query = compactParams(params)
  if (useDummy) {
    await new Promise((resolve) => setTimeout(resolve, 280))
    try {
      return dummy.listTickets(query)
    } catch (error) {
      if (error?.redirect) redirectToLogin()
      throw error
    }
  }

  const response = await api.get('/tickets', { params: query })
  const tickets = response.data?.data || []
  const meta = response.data?.meta || {
    page: Number(query.page) || 1,
    limit: Number(query.limit) || 20,
    total: tickets.length,
    total_pages: tickets.length ? 1 : 0,
  }
  return { tickets, meta }
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

export function updateTicket(id, patch) {
  return call(
    () => dummy.updateTicket(id, patch),
    () => api.patch(`/tickets/${id}`, patch),
  )
}

export function updateTicketStatus(id, status) {
  return updateTicket(id, { status })
}

export function fetchComments(ticketId) {
  return call(
    () => dummy.listComments(ticketId),
    () => api.get(`/tickets/${ticketId}/comments`),
  )
}

export function createComment(ticketId, body, isInternal = false) {
  return call(
    () => dummy.createComment(ticketId, body, isInternal),
    () => api.post(`/tickets/${ticketId}/comments`, { body, is_internal: Boolean(isInternal) }),
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
