const AVATAR_COLORS = ['#7c3aed', '#2563eb', '#db2777', '#059669', '#d97706', '#0891b2', '#4f46e5']

export function formatDate(value) {
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  }).format(new Date(value))
}

export function formatTime(value) {
  return new Intl.DateTimeFormat('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  }).format(new Date(value))
}

export function initials(name) {
  return String(name || '?')
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join('')
}

export function avatarColor(name) {
  let hash = 0
  for (const char of String(name || '')) hash += char.charCodeAt(0)
  return AVATAR_COLORS[hash % AVATAR_COLORS.length]
}

export function statusLabel(status) {
  return (
    {
      open: 'Open',
      in_progress: 'In Progress',
      resolved: 'Resolved',
      closed: 'Closed',
    }[status] || status
  )
}

export function priorityLabel(priority) {
  if (!priority) return ''
  return priority.charAt(0).toUpperCase() + priority.slice(1)
}

export function roleLabel(role) {
  return (
    {
      admin: 'Admin',
      agent: 'Agent',
      requester: 'Requester',
    }[role] || role
  )
}

export function ticketCode(ticket) {
  if (ticket?.code) return ticket.code
  const compact = String(ticket?.id || '').replace(/-/g, '').slice(0, 4).toUpperCase()
  return compact ? `HD-${compact}` : 'Ticket'
}

export function staffLabel(role) {
  return (
    {
      admin: 'Administrator',
      agent: 'Helpdesk Agent',
      requester: 'Requester',
    }[role] || role
  )
}
