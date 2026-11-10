<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppIcon from '../components/AppIcon.vue'
import PriorityBadge from '../components/PriorityBadge.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { useDirectory } from '../composables/useDirectory'
import { createComment, fetchComments, fetchTicket, getErrorMessage, updateTicket } from '../services/api'
import { avatarColor, formatDate, formatTime, initials, roleLabel, ticketCode } from '../utils/format'
import { directoryPerson } from '../utils/people'
import { statusActionLabel, statusActions } from '../utils/status'

const route = useRoute()
const auth = useAuthStore()
const { people, staff, directoryError, loadDirectory } = useDirectory()

const ticket = ref(null)
const comments = ref([])
const loading = ref(true)
const error = ref('')
const draft = ref('')
const isInternal = ref(false)
const commentError = ref('')
const statusError = ref('')
const statusMessage = ref('')
const assignError = ref('')
const assignMessage = ref('')
const assigneeDraft = ref('')
const posting = ref(false)
const savingStatus = ref('')
const savingAssign = ref(false)

let requestId = 0

const shown = computed(() => {
  if (!ticket.value) return null
  return {
    ...ticket.value,
    requester: ticket.value.requester || directoryPerson(people.value, ticket.value.requester_id),
    assignee: ticket.value.assignee_id
      ? ticket.value.assignee || directoryPerson(people.value, ticket.value.assignee_id, 'Assigned')
      : null,
  }
})

const actions = computed(() => {
  if (!ticket.value || !auth.user) return []
  if (auth.user.role === 'requester' && ticket.value.requester_id !== auth.user.id) return []
  return statusActions(ticket.value.status, auth.user.role)
})

const visibleComments = computed(() =>
  comments.value.map((comment) => ({
    ...comment,
    author: comment.author || directoryPerson(people.value, comment.author_id),
  })),
)

async function load(id) {
  const current = ++requestId
  loading.value = true
  error.value = ''
  ticket.value = null
  comments.value = []
  statusMessage.value = ''
  statusError.value = ''
  try {
    const [ticketData, commentData] = await Promise.all([fetchTicket(id), fetchComments(id)])
    if (current !== requestId) return
    ticket.value = ticketData
    comments.value = commentData || []
    assigneeDraft.value = ticketData?.assignee_id || ''
    document.title = `${ticketCode(ticketData)} · Helpdesk`
  } catch (err) {
    if (current !== requestId) return
    error.value = getErrorMessage(err)
  } finally {
    if (current === requestId) loading.value = false
  }
}

async function changeStatus(status) {
  if (!ticket.value) return
  savingStatus.value = status
  statusError.value = ''
  statusMessage.value = ''
  try {
    ticket.value = await updateTicket(ticket.value.id, { status })
    assigneeDraft.value = ticket.value.assignee_id || ''
    statusMessage.value = 'Status updated.'
  } catch (err) {
    statusError.value = getErrorMessage(err)
  } finally {
    savingStatus.value = ''
  }
}

async function saveAssignment() {
  if (!ticket.value) return
  savingAssign.value = true
  assignError.value = ''
  assignMessage.value = ''
  statusMessage.value = ''
  try {
    ticket.value = await updateTicket(ticket.value.id, {
      assignee_id: assigneeDraft.value || null,
    })
    assigneeDraft.value = ticket.value.assignee_id || ''
    assignMessage.value = assigneeDraft.value ? 'Ticket assigned.' : 'Ticket unassigned.'
  } catch (err) {
    assignError.value = getErrorMessage(err)
  } finally {
    savingAssign.value = false
  }
}

async function submitComment() {
  commentError.value = ''
  const body = draft.value.trim()
  if (!body) {
    commentError.value = 'Write a comment before posting.'
    return
  }
  if (body.length > 5000) {
    commentError.value = 'Comment must be 5000 characters or fewer.'
    return
  }

  posting.value = true
  try {
    const created = await createComment(ticket.value.id, body, isInternal.value)
    comments.value = [...comments.value, created]
    draft.value = ''
    isInternal.value = false
  } catch (err) {
    commentError.value = getErrorMessage(err)
  } finally {
    posting.value = false
  }
}

watch(
  () => route.params.id,
  (id) => {
    loadDirectory()
    load(id)
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <div v-if="loading" class="state" role="status">
      <span class="spinner"></span>
      Loading ticket...
    </div>

    <div v-else-if="error" class="state">
      <h2>Ticket unavailable</h2>
      <p>{{ error }}</p>
      <RouterLink class="btn btn-ghost" to="/tickets">Back to tickets</RouterLink>
    </div>

    <template v-else-if="shown">
      <nav class="crumb" aria-label="Breadcrumb">
        <RouterLink to="/tickets">Tickets</RouterLink>
        <span>/</span>
        <span>#{{ ticketCode(shown) }}</span>
      </nav>

      <header class="detail-head">
        <div class="code-row">
          <h1>{{ ticketCode(shown) }}</h1>
          <StatusBadge :status="shown.status" dot />
        </div>
        <p>Created {{ formatDate(shown.created_at) }} {{ formatTime(shown.created_at) }}</p>
      </header>

      <div class="detail-grid">
        <section class="panel info">
          <p class="kicker">Title</p>
          <h2>{{ shown.title }}</h2>

          <p class="kicker">Description</p>
          <p class="description">{{ shown.description }}</p>

          <p class="kicker">Category</p>
          <p class="value">{{ shown.category }}</p>

          <p class="kicker">Status</p>
          <div class="status-row">
            <StatusBadge :status="shown.status" />
            <button
              v-for="status in actions"
              :key="status"
              class="btn btn-small"
              type="button"
              :disabled="Boolean(savingStatus)"
              @click="changeStatus(status)"
            >
              {{ savingStatus === status ? 'Saving...' : statusActionLabel(status) }}
            </button>
          </div>
          <p v-if="statusMessage" class="inline-success">{{ statusMessage }}</p>
          <p v-if="statusError" class="field-error">{{ statusError }}</p>

          <p class="kicker">Priority</p>
          <p><PriorityBadge :priority="shown.priority" /></p>

          <p class="kicker">Requester</p>
          <div class="person-block">
            <span class="avatar lg" :style="{ background: avatarColor(shown.requester?.name) }">
              {{ initials(shown.requester?.name) }}
            </span>
            <div>
              <div class="person-name">{{ shown.requester?.name || 'Unknown' }}</div>
              <div v-if="shown.requester?.email" class="person-email">{{ shown.requester.email }}</div>
              <span v-if="shown.requester?.role" class="role-pill">{{ roleLabel(shown.requester.role) }}</span>
            </div>
          </div>

          <p class="kicker">Assignee</p>
          <div class="person-block">
            <span
              class="avatar lg"
              :class="{ empty: !shown.assignee }"
              :style="shown.assignee ? { background: avatarColor(shown.assignee.name) } : null"
            >
              {{ shown.assignee ? initials(shown.assignee.name) : '—' }}
            </span>
            <div>
              <div class="person-name">{{ shown.assignee?.name || 'Unassigned' }}</div>
              <div v-if="shown.assignee?.email" class="person-email">{{ shown.assignee.email }}</div>
              <span v-if="shown.assignee?.role" class="role-pill">{{ roleLabel(shown.assignee.role) }}</span>
            </div>
          </div>
          <div v-if="auth.isStaff" class="assign-row">
            <select v-model="assigneeDraft" aria-label="Assign ticket" :disabled="savingAssign">
              <option value="">Unassigned</option>
              <option v-for="user in staff" :key="user.id" :value="user.id">
                {{ user.name }} · {{ roleLabel(user.role) }}
              </option>
            </select>
            <button
              class="btn btn-small"
              type="button"
              :disabled="savingAssign || assigneeDraft === (shown.assignee_id || '')"
              @click="saveAssignment"
            >
              {{ savingAssign ? 'Saving...' : assigneeDraft ? 'Assign' : 'Unassign' }}
            </button>
          </div>
          <p v-if="directoryError" class="field-error">{{ directoryError }}</p>
          <p v-if="assignMessage" class="inline-success">{{ assignMessage }}</p>
          <p v-if="assignError" class="field-error">{{ assignError }}</p>
        </section>

        <section class="panel comments">
          <header class="comments-head">
            <AppIcon name="chat" />
            <h2>Comments</h2>
            <span class="count">{{ visibleComments.length }}</span>
          </header>

          <p v-if="visibleComments.length === 0" class="empty-comments">No comments yet.</p>

          <ol v-else class="thread">
            <li v-for="comment in visibleComments" :key="comment.id" class="comment" :class="{ internal: comment.is_internal }">
              <span class="avatar" :style="{ background: avatarColor(comment.author?.name) }">
                {{ initials(comment.author?.name) }}
              </span>
              <div>
                <div class="comment-meta">
                  <strong>{{ comment.author?.name || 'Unknown' }}</strong>
                  <time>{{ formatDate(comment.created_at) }} {{ formatTime(comment.created_at) }}</time>
                  <span v-if="comment.is_internal" class="internal-badge">Internal</span>
                </div>
                <span v-if="comment.author?.role" class="role-pill">{{ roleLabel(comment.author.role) }}</span>
                <p>{{ comment.body }}</p>
              </div>
            </li>
          </ol>

          <form class="composer" novalidate @submit.prevent="submitComment">
            <label class="sr-only" for="comment">Add a comment</label>
            <textarea
              id="comment"
              v-model="draft"
              maxlength="5000"
              rows="3"
              :placeholder="isInternal ? 'Add an internal note...' : 'Add a comment...'"
            ></textarea>
            <label v-if="auth.isStaff" class="internal-toggle">
              <input v-model="isInternal" type="checkbox" />
              Internal note, hidden from the requester
            </label>
            <p v-if="commentError" class="field-error">{{ commentError }}</p>
            <div class="composer-actions">
              <button class="btn btn-ghost" type="button" @click="draft = ''">Cancel</button>
              <button class="btn btn-blue" type="submit" :disabled="posting">
                {{ posting ? 'Posting...' : 'Post Comment' }}
                <AppIcon v-if="!posting" name="arrow" />
              </button>
            </div>
          </form>
        </section>
      </div>
    </template>
  </div>
</template>

<style scoped>
.crumb {
  display: flex;
  gap: 8px;
  color: var(--muted);
  font-size: 14px;
  margin-bottom: 12px;
}

.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 16px;
}

.code-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.detail-head h1 {
  margin: 0;
  font-size: 32px;
  letter-spacing: -0.03em;
}

.detail-head p {
  margin: 0;
  color: var(--muted);
  font-size: 14px;
}

.detail-grid {
  display: grid;
  grid-template-columns: minmax(300px, 420px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.info,
.comments {
  padding: 22px;
}

.kicker {
  margin: 18px 0 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #9ca3af;
}

.info .kicker:first-child {
  margin-top: 0;
}

.info h2 {
  margin: 0;
  font-size: 20px;
  letter-spacing: -0.02em;
}

.description,
.comment p,
.value {
  margin: 0;
  white-space: pre-wrap;
  line-height: 1.6;
  color: #374151;
}

.status-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.inline-success {
  margin: 8px 0 0;
  color: var(--success);
  font-size: 13px;
  font-weight: 600;
}

.person-block {
  display: flex;
  gap: 12px;
  align-items: center;
}

.person-name {
  font-weight: 650;
}

.person-email {
  color: var(--muted);
  font-size: 13px;
  margin-top: 2px;
  overflow-wrap: anywhere;
}

.assign-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.assign-row select {
  flex: 1;
  min-width: 180px;
  height: 38px;
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 0 10px;
  background: #fff;
  font: inherit;
}

.comment.internal {
  background: #fffbeb;
  margin: 0 -8px;
  padding-left: 8px;
  padding-right: 8px;
  border-radius: 12px;
}

.internal-badge {
  font-size: 11px;
  font-weight: 700;
  color: #92400e;
  background: #fef3c7;
  border-radius: 999px;
  padding: 2px 6px;
}

.internal-toggle {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0 8px;
  color: #92400e;
  font-size: 13px;
  font-weight: 600;
}

.comments-head {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4b5563;
}

.comments-head h2 {
  margin: 0;
  font-size: 16px;
}

.count {
  min-width: 22px;
  height: 22px;
  padding: 0 6px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #4b5563;
  display: inline-grid;
  place-items: center;
  font-size: 12px;
  font-weight: 700;
}

.thread {
  list-style: none;
  margin: 8px 0 0;
  padding: 0;
  max-height: 560px;
  overflow: auto;
}

.comment {
  display: grid;
  grid-template-columns: 36px 1fr;
  gap: 12px;
  padding: 14px 0;
}

.comment + .comment {
  border-top: 1px solid #f3f4f6;
}

.comment-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: baseline;
}

.comment-meta time {
  color: var(--muted);
  font-size: 12px;
}

.comment p {
  margin-top: 8px;
}

.empty-comments {
  color: var(--muted);
  margin: 16px 0;
}

.composer {
  margin-top: 12px;
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 10px 12px;
}

.composer textarea {
  width: 100%;
  border: 0;
  resize: vertical;
  min-height: 72px;
  font: inherit;
  outline: none;
  background: transparent;
}

.composer textarea:focus {
  box-shadow: none;
}

.composer-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 1100px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .thread {
    max-height: none;
  }
}

@media (max-width: 640px) {
  .detail-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .detail-head h1 {
    font-size: 26px;
  }

  .composer-actions {
    flex-direction: column-reverse;
  }

  .composer-actions .btn {
    width: 100%;
  }
}
</style>
