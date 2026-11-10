<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { ROLES } from '../constants'
import { fetchUsers, getErrorMessage, updateUserRole } from '../services/api'
import { avatarColor, initials, roleLabel } from '../utils/format'

const auth = useAuthStore()
const users = ref([])
const drafts = reactive({})
const loading = ref(true)
const error = ref('')
const success = ref('')
const savingId = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    users.value = (await fetchUsers()) || []
    for (const user of users.value) drafts[user.id] = user.role
  } catch (err) {
    error.value = getErrorMessage(err)
  } finally {
    loading.value = false
  }
}

async function save(user) {
  savingId.value = user.id
  error.value = ''
  success.value = ''
  try {
    const updated = await updateUserRole(user.id, drafts[user.id])
    user.role = updated.role
    drafts[user.id] = updated.role
    success.value = `Role updated for ${updated.name}.`
  } catch (err) {
    drafts[user.id] = user.role
    error.value = getErrorMessage(err)
  } finally {
    savingId.value = ''
  }
}

onMounted(load)
</script>

<template>
  <div>
    <header class="page-head">
      <div>
        <h1>Users</h1>
        <p>Change who can request help, work tickets, or administer the desk.</p>
      </div>
    </header>

    <p v-if="success" class="alert success" role="status">{{ success }}</p>
    <p v-if="error && users.length" class="alert error" role="alert">{{ error }}</p>

    <section class="panel">
      <div v-if="loading" class="state" role="status">
        <span class="spinner"></span>
        Loading users...
      </div>

      <div v-else-if="error" class="state">
        <h2>Could not load users</h2>
        <p>{{ error }}</p>
        <button class="btn btn-ghost" type="button" @click="load">Try again</button>
      </div>

      <div v-else-if="users.length === 0" class="state">
        <h2>No users found</h2>
      </div>

      <template v-else>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Role</th>
                <th>Update</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>
                  <div class="person">
                    <span class="avatar" :style="{ background: avatarColor(user.name) }">
                      {{ initials(user.name) }}
                    </span>
                    <span class="name">
                      {{ user.name }}
                      <span v-if="user.id === auth.user?.id" class="you">You</span>
                    </span>
                  </div>
                </td>
                <td class="email">{{ user.email }}</td>
                <td><span class="role-badge" :class="user.role">{{ roleLabel(user.role) }}</span></td>
                <td>
                  <div class="role-edit">
                    <select
                      v-model="drafts[user.id]"
                      :aria-label="`Role for ${user.name}`"
                      :disabled="user.id === auth.user?.id"
                    >
                      <option v-for="role in ROLES" :key="role" :value="role">{{ roleLabel(role) }}</option>
                    </select>
                    <button
                      class="btn btn-small btn-purple"
                      type="button"
                      :disabled="user.id === auth.user?.id || drafts[user.id] === user.role || savingId === user.id"
                      @click="save(user)"
                    >
                      {{ savingId === user.id ? 'Saving...' : 'Save' }}
                    </button>
                  </div>
                  <p v-if="user.id === auth.user?.id" class="hint">You can't change your own role.</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="cards">
          <article v-for="user in users" :key="user.id" class="user-card">
            <div class="person">
              <span class="avatar" :style="{ background: avatarColor(user.name) }">{{ initials(user.name) }}</span>
              <div>
                <div class="name">
                  {{ user.name }}
                  <span v-if="user.id === auth.user?.id" class="you">You</span>
                </div>
                <div class="email">{{ user.email }}</div>
              </div>
            </div>
            <span class="role-badge" :class="user.role">{{ roleLabel(user.role) }}</span>
            <div class="role-edit">
              <select v-model="drafts[user.id]" :aria-label="`Role for ${user.name}`" :disabled="user.id === auth.user?.id">
                <option v-for="role in ROLES" :key="role" :value="role">{{ roleLabel(role) }}</option>
              </select>
              <button
                class="btn btn-small btn-purple"
                type="button"
                :disabled="user.id === auth.user?.id || drafts[user.id] === user.role || savingId === user.id"
                @click="save(user)"
              >
                {{ savingId === user.id ? 'Saving...' : 'Save' }}
              </button>
            </div>
            <p v-if="user.id === auth.user?.id" class="hint">You can't change your own role.</p>
          </article>
        </div>
      </template>
    </section>
  </div>
</template>

<style scoped>
.page-head h1 {
  margin: 0;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.page-head p {
  margin: 4px 0 16px;
  color: var(--muted);
}

.panel {
  overflow: hidden;
}

.table-wrap {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 720px;
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

.name {
  font-weight: 650;
}

.email {
  color: #4b5563;
  overflow-wrap: anywhere;
}

.you {
  margin-left: 6px;
  font-size: 11px;
  font-weight: 700;
  color: #7c3aed;
  background: #f3e8ff;
  border-radius: 999px;
  padding: 2px 6px;
  vertical-align: 1px;
}

.role-edit {
  display: flex;
  gap: 8px;
  align-items: center;
}

.role-edit select {
  height: 36px;
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 0 10px;
  background: #fff;
  font: inherit;
}

.hint {
  margin: 6px 0 0;
  color: var(--muted);
  font-size: 12px;
}

.cards {
  display: none;
}

@media (max-width: 800px) {
  .table-wrap {
    display: none;
  }

  .cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px;
  }

  .user-card {
    border: 1px solid var(--border);
    border-radius: 14px;
    padding: 14px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .role-edit {
    flex-wrap: wrap;
  }

  .role-edit select,
  .role-edit .btn {
    flex: 1;
  }
}
</style>
