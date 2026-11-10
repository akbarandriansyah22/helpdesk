<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppIcon from '../components/AppIcon.vue'
import PriorityBadge from '../components/PriorityBadge.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { useDirectory } from '../composables/useDirectory'
import { PAGE_SIZE, PRIORITIES, STATUSES } from '../constants'
import { fetchTickets, getErrorMessage } from '../services/api'
import { avatarColor, formatDate, formatTime, initials, priorityLabel, statusLabel, ticketCode } from '../utils/format'
import { directoryPerson } from '../utils/people'

const SORTS = [
  { value: 'created_at:desc', label: 'Newest' },
  { value: 'created_at:asc', label: 'Oldest' },
  { value: 'updated_at:desc', label: 'Recently updated' },
  { value: 'title:asc', label: 'Title A–Z' },
  { value: 'status:asc', label: 'Status' },
  { value: 'priority:asc', label: 'Priority' },
]

const router = useRouter()
const { people, staff, directoryError, loadDirectory } = useDirectory()
const searchRef = ref(null)
const tickets = ref([])
const meta = ref({ page: 1, limit: PAGE_SIZE, total: 0, total_pages: 0 })
const loading = ref(true)
const loadedOnce = ref(false)
const error = ref('')
const search = ref('')
const appliedSearch = ref('')
const status = ref('')
const priority = ref('')
const assignee = ref('')
const sort = ref('created_at:desc')
const page = ref(1)
let searchTimer = 0
let requestId = 0

const shortcut = /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘K' : 'Ctrl K'

const filtersActive = computed(() => Boolean(status.value || priority.value || assignee.value))
const queryActive = computed(() => filtersActive.value || Boolean(appliedSearch.value))
const pageCount = computed(() => meta.value.total_pages || 0)
const rangeStart = computed(() => (meta.value.total === 0 ? 0 : (page.value - 1) * meta.value.limit + 1))
const rangeEnd = computed(() => Math.min(page.value * meta.value.limit, meta.value.total))
const pageNumbers = computed(() => {
  const total = pageCount.value
  if (total <= 1) return []
  const end = Math.min(total, Math.max(page.value + 2, 5))
  const start = Math.max(1, end - 4)
  return Array.from({ length: end - start + 1 }, (_, index) => start + index)
})
const rows = computed(() =>
  tickets.value.map((ticket) => ({
    ...ticket,
    requester: ticket.requester || directoryPerson(people.value, ticket.requester_id),
    assignee: ticket.assignee_id
      ? ticket.assignee || directoryPerson(people.value, ticket.assignee_id, 'Assigned')
      : null,
  })),
)

watch(search, (value) => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    appliedSearch.value = value.trim()
  }, 300)
})

watch([appliedSearch, status, priority, assignee, sort], () => {
  page.value = 1
  load()
})

function setPage(next) {
  if (next === page.value || next < 1 || (pageCount.value && next > pageCount.value)) return
  page.value = next
  load()
}

async function load() {
  const current = ++requestId
  loading.value = true
  error.value = ''
  const [sortBy, order] = sort.value.split(':')
  try {
    const result = await fetchTickets({
      page: page.value,
      limit: PAGE_SIZE,
      status: status.value,
      priority: priority.value,
      assignee_id: assignee.value,
      search: appliedSearch.value,
      sort_by: sortBy,
      order,
    })
    if (current !== requestId) return
    const nextMeta = result.meta || {
      page: page.value,
      limit: PAGE_SIZE,
      total: 0,
      total_pages: 0,
    }
    // The filters can shrink the result while the user is past the last page.
    if (nextMeta.total_pages > 0 && page.value > nextMeta.total_pages) {
      page.value = nextMeta.total_pages
      return load()
    }
    if (nextMeta.total_pages === 0 && page.value !== 1) {
      page.value = 1
      return load()
    }
    tickets.value = result.tickets || []
    meta.value = nextMeta
    loadedOnce.value = true
  } catch (err) {
    if (current !== requestId) return
    error.value = getErrorMessage(err)
    if (!loadedOnce.value) tickets.value = []
  } finally {
    if (current === requestId) loading.value = false
  }
}

function clearFilters() {
  status.value = ''
  priority.value = ''
  assignee.value = ''
}

function clearAll() {
  search.value = ''
  appliedSearch.value = ''
  clearFilters()
}

function openTicket(id) {
  router.push(`/tickets/${id}`)
}

function onKeydown(event) {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault()
    searchRef.value?.focus()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  loadDirectory()
  load()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
  clearTimeout(searchTimer)
})
</script>

<template>
  <div class="page">
    <label class="search">
      <AppIcon name="search" />
      <input ref="searchRef" v-model="search" type="search" placeholder="Search tickets..." />
      <kbd>{{ shortcut }}</kbd>
    </label>

    <header class="page-head">
      <div>
        <h1>Tickets</h1>
        <p>View and manage all support tickets</p>
      </div>
      <RouterLink class="btn btn-purple" to="/tickets/create">
        <AppIcon name="plus" />
        New Ticket
      </RouterLink>
    </header>

    <section class="panel">
      <div class="filters">
        <select v-model="status" aria-label="Status">
          <option value="">All Status</option>
          <option v-for="item in STATUSES" :key="item" :value="item">{{ statusLabel(item) }}</option>
        </select>
        <select v-model="priority" aria-label="Priority">
          <option value="">All Priority</option>
          <option v-for="item in PRIORITIES" :key="item" :value="item">{{ priorityLabel(item) }}</option>
        </select>
        <select v-model="assignee" aria-label="Assignee">
          <option value="">All assignees</option>
          <option value="unassigned">Unassigned</option>
          <option v-for="user in staff" :key="user.id" :value="user.id">{{ user.name }}</option>
        </select>
        <select v-model="sort" aria-label="Sort">
          <option v-for="item in SORTS" :key="item.value" :value="item.value">{{ item.label }}</option>
        </select>
        <button class="text-button" type="button" :disabled="!filtersActive" @click="clearFilters">
          × Clear filters
        </button>
        <span v-if="loadedOnce && !error" class="result-count">{{ meta.total }} tickets</span>
      </div>

      <p v-if="directoryError" class="alert error list-alert" role="alert">{{ directoryError }}</p>
      <p v-if="error && tickets.length" class="alert error list-alert" role="alert">{{ error }}</p>

      <div v-if="loading && !tickets.length" class="state" role="status">
        <span class="spinner"></span>
        Loading tickets...
      </div>

      <div v-else-if="error && !tickets.length" class="state">
        <h2>Could not load tickets</h2>
        <p>{{ error }}</p>
        <button class="btn btn-ghost" type="button" @click="load">Try again</button>
      </div>

      <div v-else-if="tickets.length === 0" class="state">
        <h2>{{ queryActive ? 'No tickets found' : 'No tickets yet' }}</h2>
        <p v-if="queryActive">Try a different search or clear the filters.</p>
        <p v-else>Create a ticket and it will show up here.</p>
        <RouterLink v-if="!queryActive" class="btn btn-purple" to="/tickets/create">New Ticket</RouterLink>
        <button v-else class="btn btn-ghost" type="button" @click="clearAll">Clear filters</button>
      </div>

      <template v-else>
        <div class="table-wrap" :aria-busy="loading" :class="{ busy: loading }">
          <table>
            <thead>
              <tr>
                <th>Title</th>
                <th>Status</th>
                <th>Priority</th>
                <th>Assignee</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="ticket in rows"
                :key="ticket.id"
                @click="openTicket(ticket.id)"
              >
                <td>
                  <RouterLink class="title-link" :to="`/tickets/${ticket.id}`" @click.stop>
                    {{ ticket.title }}
                  </RouterLink>
                  <div class="meta">#{{ ticketCode(ticket) }} · {{ ticket.category }}</div>
                </td>
                <td><StatusBadge :status="ticket.status" /></td>
                <td><PriorityBadge :priority="ticket.priority" /></td>
                <td>
                  <div class="person">
                    <span
                      class="avatar"
                      :class="{ empty: !ticket.assignee }"
                      :style="ticket.assignee ? { background: avatarColor(ticket.assignee.name) } : null"
                    >
                      {{ ticket.assignee ? initials(ticket.assignee.name) : '—' }}
                    </span>
                    <span :class="{ muted: !ticket.assignee }">
                      {{ ticket.assignee?.name || 'Unassigned' }}
                    </span>
                  </div>
                </td>
                <td>
                  <div>{{ formatDate(ticket.created_at) }}</div>
                  <div class="time">{{ formatTime(ticket.created_at) }}</div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="cards" :aria-busy="loading" :class="{ busy: loading }">
          <RouterLink v-for="ticket in rows" :key="ticket.id" class="ticket-card" :to="`/tickets/${ticket.id}`">
            <div>
              <h3>{{ ticket.title }}</h3>
              <p class="meta">#{{ ticketCode(ticket) }} · {{ ticket.category }}</p>
            </div>
            <div class="badge-row">
              <StatusBadge :status="ticket.status" />
              <PriorityBadge :priority="ticket.priority" />
            </div>
            <div class="card-foot">
              <div class="person">
                <span
                  class="avatar"
                  :class="{ empty: !ticket.assignee }"
                  :style="ticket.assignee ? { background: avatarColor(ticket.assignee.name) } : null"
                >
                  {{ ticket.assignee ? initials(ticket.assignee.name) : '—' }}
                </span>
                <span :class="{ muted: !ticket.assignee }">{{ ticket.assignee?.name || 'Unassigned' }}</span>
              </div>
              <div class="time">
                {{ formatDate(ticket.created_at) }}
                {{ formatTime(ticket.created_at) }}
              </div>
            </div>
          </RouterLink>
        </div>

        <footer class="pager">
          <span>Showing {{ rangeStart }} to {{ rangeEnd }} of {{ meta.total }} tickets</span>
          <nav v-if="pageCount > 1" aria-label="Pagination">
            <button type="button" aria-label="Previous page" :disabled="page === 1 || loading" @click="setPage(page - 1)">‹</button>
            <button
              v-for="number in pageNumbers"
              :key="number"
              type="button"
              :aria-current="number === page ? 'page' : undefined"
              :disabled="loading"
              @click="setPage(number)"
            >
              {{ number }}
            </button>
            <button type="button" aria-label="Next page" :disabled="page === pageCount || loading" @click="setPage(page + 1)">›</button>
          </nav>
        </footer>
      </template>
    </section>
  </div>
</template>

<style scoped>
.page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
  margin: 22px 0 16px;
}

.page-head h1 {
  margin: 0;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.page-head p {
  margin: 4px 0 0;
  color: var(--muted);
}

.search {
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: 640px;
  height: 44px;
  padding: 0 12px;
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 12px;
  color: #9ca3af;
}

.search input {
  flex: 1;
  min-width: 0;
  border: 0;
  outline: none;
  background: transparent;
  color: var(--text);
  font: inherit;
  height: 100%;
}

.search input:focus {
  box-shadow: none;
}

.search kbd {
  font-family: inherit;
  font-size: 12px;
  color: #9ca3af;
  border: 1px solid var(--border);
  border-bottom-width: 2px;
  border-radius: 6px;
  padding: 2px 6px;
  background: #fafafa;
}

.panel {
  background: #fff;
  border: 1px solid var(--border);
  border-radius: 16px;
}

.filters {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
}

.filters select {
  appearance: none;
  height: 38px;
  padding: 0 32px 0 12px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' fill='none' stroke='%236b7280' stroke-width='1.8' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m4 6 4 4 4-4'/%3E%3C/svg%3E") no-repeat right 10px center;
  color: var(--text);
  font: inherit;
  font-size: 14px;
}

.text-button {
  border: 0;
  background: transparent;
  color: #6b7280;
  font-weight: 600;
  font-size: 14px;
  padding: 8px 4px;
}

.text-button:hover:not(:disabled) {
  color: var(--text);
}

.result-count {
  margin-left: auto;
  color: var(--muted);
  font-size: 14px;
}

.list-alert {
  margin: 12px 16px 0;
}

.table-wrap {
  overflow-x: auto;
}

.busy {
  opacity: 0.55;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}

th {
  text-align: left;
  font-size: 13px;
  font-weight: 550;
  color: #6b7280;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

td {
  padding: 14px 16px;
  border-bottom: 1px solid #f3f4f6;
  vertical-align: middle;
  font-size: 14px;
}

tbody tr {
  cursor: pointer;
}

tbody tr:hover {
  background: #fafafa;
}

.title-link {
  color: var(--text);
  font-weight: 650;
}

.title-link:hover {
  color: #6d28d9;
}

.meta {
  margin-top: 3px;
  color: #9ca3af;
  font-size: 12px;
}

.time {
  color: var(--muted);
  font-size: 12px;
  margin-top: 2px;
}

.cards {
  display: none;
}

.pager {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  padding: 14px 16px;
  color: var(--muted);
  font-size: 14px;
}

.pager nav {
  display: flex;
  gap: 4px;
}

.pager button {
  min-width: 36px;
  height: 36px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #4b5563;
  font-weight: 600;
}

.pager button[aria-current='page'] {
  background: #f3e8ff;
  color: #7c3aed;
}

.pager button:hover:not(:disabled) {
  background: #f3f4f6;
}

@media (max-width: 800px) {
  .page-head {
    flex-direction: column;
    align-items: stretch;
  }

  .page-head .btn {
    justify-content: center;
  }

  .search kbd,
  .table-wrap {
    display: none;
  }

  .cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px;
  }

  .ticket-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px;
    border: 1px solid var(--border);
    border-radius: 14px;
    background: #fff;
    color: inherit;
  }

  .ticket-card h3 {
    margin: 0;
    font-size: 16px;
    letter-spacing: -0.02em;
  }

  .badge-row,
  .card-foot,
  .person {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-foot {
    justify-content: space-between;
    gap: 12px;
  }

  .result-count {
    width: 100%;
    margin-left: 0;
  }

  .filters select {
    flex: 1 1 140px;
  }
}
</style>
