import { PRIORITIES, ROLES, STATUSES, TOKEN_KEY } from '../constants'
import { nextStatuses } from '../utils/status'

const DB_KEY = 'helpdesk_dummy_db'
const DB_VERSION = 1

const USER = {
  alex: '00000000-0000-4000-8000-000000000001',
  sarah: '00000000-0000-4000-8000-000000000002',
  lisa: '00000000-0000-4000-8000-000000000003',
  sam: '00000000-0000-4000-8000-000000000004',
  jamesDoe: '00000000-0000-4000-8000-000000000005',
  mike: '00000000-0000-4000-8000-000000000006',
  taylor: '00000000-0000-4000-8000-000000000007',
  jamesCarter: '00000000-0000-4000-8000-000000000008',
  riley: '00000000-0000-4000-8000-000000000009',
}

export class ApiError extends Error {
  constructor(message, status = 400, options = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.redirect = Boolean(options.redirect)
    this.response = { status, data: { error: message } }
  }
}

function user(id, name, email, role) {
  return {
    id,
    name,
    email,
    password: 'password123',
    role,
    created_at: '2024-01-08T09:00:00',
    updated_at: '2024-01-08T09:00:00',
  }
}

function ticket(code, title, description, category, priority, status, requesterId, assigneeId, created) {
  const number = String(code).replace(/\D/g, '').padStart(12, '0')
  return {
    id: `00000000-0000-4000-8000-${number}`,
    code,
    title,
    description,
    category,
    priority,
    status,
    requester_id: requesterId,
    assignee_id: assigneeId,
    created_at: created,
    updated_at: created,
  }
}

function comment(n, ticketCode, authorId, body, created) {
  const number = String(800000 + n).padStart(12, '0')
  const ticketNumber = String(ticketCode).replace(/\D/g, '').padStart(12, '0')
  return {
    id: `00000000-0000-4000-8000-${number}`,
    ticket_id: `00000000-0000-4000-8000-${ticketNumber}`,
    author_id: authorId,
    body,
    is_internal: false,
    created_at: created,
  }
}

function seed() {
  const carter = USER.jamesCarter
  const riley = USER.riley

  return {
    version: DB_VERSION,
    users: [
      user(USER.alex, 'Alex Kim', 'alex.kim@acme.com', 'admin'),
      user(USER.sarah, 'Sarah Mitchell', 'sarah.mitchell@acme.com', 'agent'),
      user(USER.lisa, 'Lisa Patel', 'lisa.patel@acme.com', 'agent'),
      user(USER.sam, 'Sam Wilson', 'sam.wilson@acme.com', 'agent'),
      user(USER.jamesDoe, 'James Doe', 'james.doe@acme.com', 'agent'),
      user(USER.mike, 'Mike Brown', 'mike.brown@acme.com', 'agent'),
      user(USER.taylor, 'Taylor Chen', 'taylor.chen@acme.com', 'agent'),
      user(USER.jamesCarter, 'James Carter', 'james.carter@acme.com', 'requester'),
      user(USER.riley, 'Riley Nguyen', 'riley.nguyen@acme.com', 'requester'),
    ],
    tickets: [
      ticket(
        'HD-1287',
        'Cannot access billing dashboard',
        'The billing dashboard returns a blank page after I sign in. Invoices for this month are due today and I cannot download them.',
        'Billing',
        'high',
        'open',
        carter,
        USER.jamesDoe,
        '2024-05-16T10:24:00',
      ),
      ticket(
        'HD-1286',
        'Error when exporting report to CSV',
        'Exporting the weekly operations report to CSV fails with "unexpected token" after about 30 seconds. The on-screen report still loads.',
        'Reports',
        'medium',
        'in_progress',
        riley,
        USER.lisa,
        '2024-05-16T09:15:00',
      ),
      ticket(
        'HD-1285',
        'Login page not loading on Safari',
        'Safari 17 on macOS shows a spinning loader on the login page and never renders the form. Chrome on the same machine works.',
        'Authentication',
        'high',
        'open',
        carter,
        USER.mike,
        '2024-05-16T16:42:00',
      ),
      ticket(
        'HD-1284',
        'Feature request: dark mode support',
        'Our night-shift team has asked for a dark theme in the internal tools. Happy to help test it if you pick this up.',
        'Feature Request',
        'low',
        'open',
        riley,
        null,
        '2024-05-15T14:31:00',
      ),
      ticket(
        'HD-1283',
        'Unable to update profile information',
        'Saving my profile returns "something went wrong" when I change my phone number. Name and email updates still work.',
        'Accounts',
        'medium',
        'in_progress',
        carter,
        USER.alex,
        '2024-05-15T11:07:00',
      ),
      ticket(
        'HD-1282',
        'Payment failed but charged',
        'A card payment for the May invoice was declined in the portal, but the bank shows a completed charge. Please confirm whether it settled.',
        'Billing',
        'high',
        'open',
        riley,
        null,
        '2024-05-14T15:50:00',
      ),
      ticket(
        'HD-1281',
        'API rate limit error on file upload',
        'Uploads over 20 MB started failing this morning with HTTP 429. Smaller files still go through.',
        'Integrations',
        'medium',
        'resolved',
        carter,
        USER.sam,
        '2024-05-14T13:22:00',
      ),
      ticket(
        'HD-1280',
        'How to reset my MFA device?',
        'I replaced my phone and no longer have the authenticator app. I need the MFA device reset so I can sign in tomorrow.',
        'Accounts',
        'low',
        'closed',
        riley,
        USER.taylor,
        '2024-05-13T17:18:00',
      ),
      ticket(
        'HD-1279',
        'VPN drops every few minutes',
        'The office VPN disconnects during calls and I have to reconnect. It started after yesterday\'s laptop update.',
        'Access',
        'urgent',
        'in_progress',
        carter,
        USER.sarah,
        '2024-05-18T08:05:00',
      ),
      ticket(
        'HD-1278',
        'Laptop request for new hire',
        'We have a new hire starting Monday and still need a standard laptop, charger, and docking station.',
        'Hardware',
        'low',
        'open',
        riley,
        null,
        '2024-05-17T15:10:00',
      ),
      ticket(
        'HD-1277',
        'Slack alerts arriving late',
        'Incident alerts in the #ops channel are landing 15 to 20 minutes late. Email alerts are on time.',
        'Integrations',
        'medium',
        'resolved',
        riley,
        USER.lisa,
        '2024-05-12T11:40:00',
      ),
      ticket(
        'HD-1276',
        'Printer offline on the 3rd floor',
        'The 3rd floor printer shows offline for everyone. The queue on the 2nd floor still prints.',
        'Hardware',
        'medium',
        'open',
        riley,
        USER.sam,
        '2024-05-12T09:20:00',
      ),
      ticket(
        'HD-1275',
        'Expense report stuck on submit',
        'My May expense report stays on "submitting" and never leaves draft. The receipt files are under 2 MB each.',
        'Billing',
        'high',
        'in_progress',
        carter,
        USER.alex,
        '2024-05-11T16:05:00',
      ),
      ticket(
        'HD-1274',
        'Password reset email never arrives',
        'I requested three password resets today and nothing reached my inbox or spam folder.',
        'Authentication',
        'urgent',
        'open',
        riley,
        USER.sarah,
        '2024-05-11T08:48:00',
      ),
      ticket(
        'HD-1273',
        'Shared calendar missing meetings',
        'Meetings from last week disappeared from the team calendar. They still show on my personal calendar.',
        'Accounts',
        'low',
        'closed',
        riley,
        USER.lisa,
        '2024-05-10T13:15:00',
      ),
      ticket(
        'HD-1272',
        'Build server disk almost full',
        'The CI builder is at 96% disk and new pipelines are failing to check out. Please clear old artifacts or expand the volume.',
        'Infrastructure',
        'high',
        'in_progress',
        riley,
        USER.sam,
        '2024-05-10T10:02:00',
      ),
      ticket(
        'HD-1271',
        'Customer export times out',
        'Exporting all customers from Reports times out after two minutes. Filtering to a single month works.',
        'Reports',
        'high',
        'open',
        carter,
        null,
        '2024-05-09T14:27:00',
      ),
      ticket(
        'HD-1270',
        'Meeting room badge not working',
        'My badge opens the lobby but not meeting room 4B. I have a client workshop there on Thursday.',
        'Access',
        'low',
        'resolved',
        riley,
        USER.sarah,
        '2024-05-09T09:36:00',
      ),
      ticket(
        'HD-1269',
        'External monitor flickers',
        'The monitor on my desk flickers whenever I share my screen. A spare cable did not change it.',
        'Hardware',
        'medium',
        'closed',
        riley,
        USER.sam,
        '2024-05-08T15:55:00',
      ),
      ticket(
        'HD-1268',
        'Outlook SSO redirect loop',
        'Outlook on the web bounces between the login page and a blank redirect. The desktop app still opens mail.',
        'Authentication',
        'urgent',
        'in_progress',
        riley,
        USER.lisa,
        '2024-05-08T08:18:00',
      ),
      ticket(
        'HD-1267',
        'Wrong department on my profile',
        'My profile still says Finance after I moved to IT last month. I cannot edit the department field.',
        'Accounts',
        'low',
        'open',
        riley,
        null,
        '2024-05-07T11:11:00',
      ),
      ticket(
        'HD-1266',
        'Invoice PDF prints blank pages',
        'Downloaded invoice PDFs open with the right number of pages, but every page is blank when printed.',
        'Billing',
        'medium',
        'resolved',
        carter,
        USER.alex,
        '2024-05-07T09:44:00',
      ),
      ticket(
        'HD-1265',
        'Webhook deliveries keep failing',
        'Our staging webhook endpoint returns 200, but the delivery log shows repeated timeouts since Monday.',
        'Integrations',
        'high',
        'open',
        riley,
        USER.sarah,
        '2024-05-06T16:30:00',
      ),
      ticket(
        'HD-1048',
        'Unable to access shared drive after password reset',
        'After resetting my password this morning, I can no longer access the shared drive (S:). I get an error saying my credentials are invalid. This is affecting several files I need for a report due today.\n\nSteps tried:\n• Restarted my laptop\n• Cleared cached credentials\n• Verified I can log in to other systems\n\nLet me know if you need any additional info.',
        'Access',
        'high',
        'open',
        carter,
        USER.sarah,
        '2024-05-16T09:14:00',
      ),
    ],
    comments: [
      comment(
        1,
        'HD-1048',
        USER.jamesCarter,
        "I reset my password using the self-service portal and immediately noticed I couldn't connect to the shared drive.",
        '2024-05-16T09:16:00',
      ),
      comment(
        2,
        'HD-1048',
        USER.sarah,
        'Thanks James. Can you try mapping the drive again using your username in the format acme\\james.carter ?',
        '2024-05-16T09:28:00',
      ),
      comment(
        3,
        'HD-1048',
        USER.jamesCarter,
        'Just tried that—still getting the same credentials error.',
        '2024-05-16T09:35:00',
      ),
      comment(
        4,
        'HD-1048',
        USER.sarah,
        "I've checked the account and permissions look good. I'll escalate this to our infrastructure team for further investigation.",
        '2024-05-16T09:47:00',
      ),
      comment(
        5,
        'HD-1287',
        USER.jamesDoe,
        'I can reproduce the blank billing page. Checking the service logs next.',
        '2024-05-16T11:05:00',
      ),
      comment(
        6,
        'HD-1281',
        USER.sam,
        'Raised the upload limit and confirmed a 40 MB file goes through. Closing this out as resolved.',
        '2024-05-14T16:10:00',
      ),
    ],
  }
}

function loadDb() {
  try {
    const raw = localStorage.getItem(DB_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (parsed?.version === DB_VERSION && Array.isArray(parsed.users)) return parsed
    }
  } catch {
    // Ignore a broken cache and reseed below.
  }
  const fresh = seed()
  localStorage.setItem(DB_KEY, JSON.stringify(fresh))
  return fresh
}

let db = loadDb()

function save() {
  localStorage.setItem(DB_KEY, JSON.stringify(db))
}

function publicUser(record) {
  if (!record) return null
  return {
    id: record.id,
    name: record.name,
    email: record.email,
    role: record.role,
    created_at: record.created_at,
    updated_at: record.updated_at,
  }
}

function codeNumber(code) {
  const value = Number(String(code).replace(/\D/g, ''))
  return Number.isFinite(value) ? value : 0
}

function hydrateTicket(record) {
  return {
    ...record,
    requester: publicUser(db.users.find((item) => item.id === record.requester_id)),
    assignee: record.assignee_id
      ? publicUser(db.users.find((item) => item.id === record.assignee_id))
      : null,
  }
}

function hydrateComment(record) {
  return {
    ...record,
    author: publicUser(db.users.find((item) => item.id === record.author_id)),
  }
}

function actor() {
  const token = localStorage.getItem(TOKEN_KEY) || ''
  if (!token.startsWith('dummy.')) {
    throw new ApiError('invalid or expired token', 401, { redirect: true })
  }
  const id = token.slice('dummy.'.length)
  const record = db.users.find((item) => item.id === id)
  if (!record) throw new ApiError('invalid or expired token', 401, { redirect: true })
  return record
}

function requireTicket(id, user) {
  const record = db.tickets.find((item) => item.id === id)
  if (!record) throw new ApiError('not found', 404)
  if (user.role === 'requester' && record.requester_id !== user.id) {
    throw new ApiError('forbidden', 403)
  }
  return record
}

export function login(email, password) {
  const cleanEmail = String(email || '').trim().toLowerCase()
  const record = db.users.find((item) => item.email === cleanEmail)
  if (!record || record.password !== password) {
    throw new ApiError('invalid email or password', 401)
  }
  return { token: `dummy.${record.id}`, user: publicUser(record) }
}

export function register(payload) {
  const name = String(payload?.name || '').trim()
  const email = String(payload?.email || '').trim().toLowerCase()
  const password = String(payload?.password || '')

  if (name.length < 2 || name.length > 100) throw new ApiError('invalid input', 400)
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) throw new ApiError('invalid input', 400)
  if (password.length < 8 || password.length > 72) throw new ApiError('invalid input', 400)
  if (db.users.some((item) => item.email === email)) {
    throw new ApiError('email already registered', 409)
  }

  const now = new Date().toISOString()
  const record = {
    id: crypto.randomUUID(),
    name,
    email,
    password,
    role: 'requester',
    created_at: now,
    updated_at: now,
  }
  db.users.push(record)
  save()
  return publicUser(record)
}

export function me() {
  return publicUser(actor())
}

export function listTickets(params = {}) {
  const user = actor()
  let items = db.tickets.filter((item) => user.role !== 'requester' || item.requester_id === user.id)

  if (params.status) items = items.filter((item) => item.status === params.status)
  if (params.priority) items = items.filter((item) => item.priority === params.priority)
  if (params.assignee_id === 'unassigned') items = items.filter((item) => !item.assignee_id)
  else if (params.assignee_id) items = items.filter((item) => item.assignee_id === params.assignee_id)
  if (params.search) {
    const query = String(params.search).trim().toLowerCase()
    items = items.filter((item) => item.title.toLowerCase().includes(query))
  }

  const sortBy = params.sort_by || 'created_at'
  const direction = params.order === 'asc' ? 1 : -1
  items = items.slice().sort((a, b) => {
    const left = a[sortBy] ?? ''
    const right = b[sortBy] ?? ''
    if (left < right) return -1 * direction
    if (left > right) return 1 * direction
    return codeNumber(b.code) - codeNumber(a.code)
  })

  const page = Math.max(1, Number(params.page) || 1)
  const limit = Math.max(1, Number(params.limit) || 20)
  const total = items.length
  const start = (page - 1) * limit
  return {
    tickets: items.slice(start, start + limit).map(hydrateTicket),
    meta: {
      page,
      limit,
      total,
      total_pages: total === 0 ? 0 : Math.ceil(total / limit),
    },
  }
}

export function getTicket(id) {
  const user = actor()
  return hydrateTicket(requireTicket(id, user))
}

export function createTicket(payload) {
  const user = actor()
  const title = String(payload?.title || '').trim()
  const description = String(payload?.description || '').trim()
  const category = String(payload?.category || '').trim()
  const priority = String(payload?.priority || '').trim()

  if (title.length < 3 || title.length > 200) throw new ApiError('invalid input', 400)
  if (!description || description.length > 5000) throw new ApiError('invalid input', 400)
  if (category.length < 2 || category.length > 50) throw new ApiError('invalid input', 400)
  if (!PRIORITIES.includes(priority)) throw new ApiError('invalid input', 400)

  const max = db.tickets.reduce((highest, item) => Math.max(highest, codeNumber(item.code)), 1000)
  const now = new Date().toISOString()
  const record = {
    id: crypto.randomUUID(),
    code: `HD-${max + 1}`,
    title,
    description,
    category,
    priority,
    status: 'open',
    requester_id: user.id,
    assignee_id: null,
    created_at: now,
    updated_at: now,
  }
  db.tickets.push(record)
  save()
  return hydrateTicket(record)
}

export function updateTicket(id, patch = {}) {
  const user = actor()
  const record = requireTicket(id, user)

  if (Object.prototype.hasOwnProperty.call(patch, 'assignee_id')) {
    if (user.role === 'requester') {
      throw new ApiError('forbidden: requesters cannot assign or unassign tickets', 403)
    }
    if (patch.assignee_id) {
      const assignee = db.users.find((item) => item.id === patch.assignee_id)
      if (!assignee) throw new ApiError('assignee not found', 404)
      if (assignee.role !== 'agent' && assignee.role !== 'admin') {
        throw new ApiError('assignee must be an agent or admin', 400)
      }
      record.assignee_id = assignee.id
    } else {
      record.assignee_id = null
    }
  }

  if (patch.priority) {
    if (!PRIORITIES.includes(patch.priority)) throw new ApiError('invalid input', 400)
    record.priority = patch.priority
  }

  if (patch.status && patch.status !== record.status) {
    if (!STATUSES.includes(patch.status)) throw new ApiError('invalid status', 400)
    if (!nextStatuses(record.status, user.role).includes(patch.status)) {
      throw new ApiError(
        `status transition is not allowed: cannot change status from ${record.status} to ${patch.status} as ${user.role}`,
        409,
      )
    }
    record.status = patch.status
  }

  record.updated_at = new Date().toISOString()
  save()
  return hydrateTicket(record)
}

export function updateStatus(id, status) {
  return updateTicket(id, { status })
}

export function listComments(ticketId) {
  const user = actor()
  requireTicket(ticketId, user)
  return db.comments
    .filter((item) => item.ticket_id === ticketId)
    .filter((item) => user.role !== 'requester' || !item.is_internal)
    .map(hydrateComment)
    .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
}

export function createComment(ticketId, body, isInternal = false) {
  const user = actor()
  requireTicket(ticketId, user)
  if (isInternal && user.role === 'requester') throw new ApiError('forbidden', 403)
  const text = String(body || '').trim()
  if (!text || text.length > 5000) throw new ApiError('invalid input', 400)

  const record = {
    id: crypto.randomUUID(),
    ticket_id: ticketId,
    author_id: user.id,
    body: text,
    is_internal: Boolean(isInternal),
    created_at: new Date().toISOString(),
  }
  db.comments.push(record)
  save()
  return hydrateComment(record)
}

export function listUsers() {
  const user = actor()
  if (user.role !== 'admin' && user.role !== 'agent') throw new ApiError('forbidden', 403)
  return db.users
    .map(publicUser)
    .sort((a, b) => a.name.localeCompare(b.name))
}

export function updateRole(id, role) {
  const user = actor()
  if (user.role !== 'admin') throw new ApiError('forbidden', 403)
  if (!ROLES.includes(role)) throw new ApiError('invalid role', 400)
  if (user.id === id) throw new ApiError('cannot change your own role', 403)

  const record = db.users.find((item) => item.id === id)
  if (!record) throw new ApiError('not found', 404)
  record.role = role
  record.updated_at = new Date().toISOString()
  save()
  return publicUser(record)
}
