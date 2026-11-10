// Mirrors the backend status graph. Anything absent here is rejected with 409.
const TRANSITIONS = {
  open: {
    in_progress: ['agent', 'admin'],
    closed: ['requester', 'admin'],
  },
  in_progress: {
    resolved: ['agent', 'admin'],
  },
  resolved: {
    closed: ['requester', 'agent', 'admin'],
    open: ['requester'],
  },
}

const ACTION_LABELS = {
  open: 'Reopen',
  in_progress: 'Start progress',
  resolved: 'Mark resolved',
  closed: 'Close ticket',
}

const ACTION_ORDER = ['in_progress', 'resolved', 'closed', 'open']

export function nextStatuses(current, role) {
  const next = TRANSITIONS[current] || {}
  return Object.entries(next)
    .filter(([, roles]) => roles.includes(role))
    .map(([status]) => status)
}

export function statusActions(current, role) {
  const allowed = new Set(nextStatuses(current, role))
  return ACTION_ORDER.filter((status) => allowed.has(status))
}

export function statusActionLabel(status) {
  return ACTION_LABELS[status] || status
}
