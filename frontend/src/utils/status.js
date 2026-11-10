const TRANSITIONS = {
  open: {
    in_progress: ['agent', 'admin'],
    closed: ['requester', 'agent', 'admin'],
  },
  in_progress: {
    open: ['agent', 'admin'],
    resolved: ['agent', 'admin'],
  },
  resolved: {
    in_progress: ['agent', 'admin'],
    closed: ['requester', 'agent', 'admin'],
  },
  closed: {
    open: ['requester', 'agent', 'admin'],
  },
}

const ACTION_LABELS = {
  open: 'Set open',
  in_progress: 'Start progress',
  resolved: 'Mark resolved',
  closed: 'Close ticket',
}

export function nextStatuses(current, role) {
  const next = TRANSITIONS[current] || {}
  return Object.entries(next)
    .filter(([, roles]) => roles.includes(role))
    .map(([status]) => status)
}

export function statusActionLabel(status) {
  return ACTION_LABELS[status] || status
}
