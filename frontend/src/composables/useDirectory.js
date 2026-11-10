import { computed, ref } from 'vue'
import { fetchUsers, getErrorMessage } from '../services/api'
import { useAuthStore } from '../stores/auth'

// Staff can load the user directory. Requesters only know themselves,
// because GET /users is limited to agents and admins.
export function useDirectory() {
  const auth = useAuthStore()
  const users = ref([])
  const directoryError = ref('')

  const people = computed(() => {
    const map = new Map()
    if (auth.user?.id) map.set(auth.user.id, auth.user)
    for (const user of users.value) map.set(user.id, user)
    return map
  })

  const staff = computed(() =>
    users.value
      .filter((user) => user.role === 'agent' || user.role === 'admin')
      .slice()
      .sort((a, b) => a.name.localeCompare(b.name)),
  )

  async function loadDirectory() {
    if (!auth.isStaff) return
    directoryError.value = ''
    try {
      users.value = (await fetchUsers()) || []
    } catch (err) {
      directoryError.value = getErrorMessage(err)
    }
  }

  return { users, people, staff, directoryError, loadDirectory }
}
