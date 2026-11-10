<script setup>
import { reactive, ref } from 'vue'
import { CATEGORIES, PRIORITIES } from '../constants'
import { createTicket, getErrorMessage } from '../services/api'
import { priorityLabel, ticketCode } from '../utils/format'

const form = reactive({
  title: '',
  description: '',
  category: '',
  priority: '',
})
const errors = reactive({
  title: '',
  description: '',
  category: '',
  priority: '',
})
const formError = ref('')
const success = ref(null)
const submitting = ref(false)

function validate() {
  const title = form.title.trim()
  const description = form.description.trim()
  errors.title = title.length >= 3 && title.length <= 200 ? '' : 'Enter a title between 3 and 200 characters.'
  errors.description = description && description.length <= 5000 ? '' : 'Enter a description (5000 characters max).'
  errors.category = form.category ? '' : 'Choose a category.'
  errors.priority = PRIORITIES.includes(form.priority) ? '' : 'Choose a priority.'
  return !errors.title && !errors.description && !errors.category && !errors.priority
}

async function submit() {
  formError.value = ''
  success.value = null
  if (!validate()) return

  submitting.value = true
  try {
    success.value = await createTicket({
      title: form.title.trim(),
      description: form.description.trim(),
      category: form.category,
      priority: form.priority,
    })
    form.title = ''
    form.description = ''
    form.category = ''
    form.priority = ''
  } catch (err) {
    formError.value = getErrorMessage(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="create">
    <header class="page-head">
      <div>
        <RouterLink class="back" to="/tickets">← Tickets</RouterLink>
        <h1>New ticket</h1>
        <p>Tell us what you need help with.</p>
      </div>
    </header>

    <form class="panel form-card" novalidate @submit.prevent="submit">
      <p v-if="success" class="alert success" role="status">
        Ticket {{ ticketCode(success) }} was created.
        <RouterLink :to="`/tickets/${success.id}`">Open ticket</RouterLink>
      </p>
      <p v-if="formError" class="alert error" role="alert">{{ formError }}</p>

      <label class="field" for="title">
        Title
        <input id="title" v-model="form.title" type="text" maxlength="200" placeholder="Short summary of the issue" />
        <span v-if="errors.title" class="field-error">{{ errors.title }}</span>
      </label>

      <label class="field" for="description">
        Description
        <textarea
          id="description"
          v-model="form.description"
          maxlength="5000"
          rows="7"
          placeholder="What happened, and what have you already tried?"
        ></textarea>
        <span v-if="errors.description" class="field-error">{{ errors.description }}</span>
      </label>

      <label class="field" for="category">
        Category
        <select id="category" v-model="form.category">
          <option value="" disabled>Choose a category</option>
          <option v-for="item in CATEGORIES" :key="item" :value="item">{{ item }}</option>
        </select>
        <span v-if="errors.category" class="field-error">{{ errors.category }}</span>
      </label>

      <fieldset class="field">
        <legend>Priority</legend>
        <div class="pills" role="radiogroup">
          <label v-for="item in PRIORITIES" :key="item" class="pill" :class="item">
            <input v-model="form.priority" type="radio" name="priority" :value="item" />
            {{ priorityLabel(item) }}
          </label>
        </div>
        <span v-if="errors.priority" class="field-error">{{ errors.priority }}</span>
      </fieldset>

      <div class="actions">
        <RouterLink class="btn btn-ghost" to="/tickets">Cancel</RouterLink>
        <button class="btn btn-purple" type="submit" :disabled="submitting">
          {{ submitting ? 'Creating...' : 'Create ticket' }}
        </button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.create {
  max-width: 760px;
}

.back {
  display: inline-block;
  margin-bottom: 8px;
  color: var(--muted);
  font-size: 14px;
  font-weight: 600;
}

.page-head h1 {
  margin: 0;
  font-size: 30px;
  letter-spacing: -0.03em;
}

.page-head p {
  margin: 4px 0 16px;
  color: var(--muted);
}

.form-card {
  padding: 22px;
}

fieldset.field {
  border: 0;
  padding: 0;
}

legend {
  font-size: 14px;
  font-weight: 650;
  margin-bottom: 8px;
}

.pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pill {
  position: relative;
  border: 1px solid var(--border);
  border-radius: 999px;
  padding: 8px 14px;
  font-size: 14px;
  font-weight: 650;
  color: #4b5563;
  cursor: pointer;
}

.pill input {
  position: absolute;
  opacity: 0;
  inset: 0;
  margin: 0;
}

.pill:has(input:checked).low {
  background: #e7f8ee;
  color: #15803d;
  border-color: transparent;
}

.pill:has(input:checked).medium {
  background: #ffedd5;
  color: #c2410c;
  border-color: transparent;
}

.pill:has(input:checked).high {
  background: #fde8ee;
  color: #e11d48;
  border-color: transparent;
}

.pill:has(input:checked).urgent {
  background: #fee2e2;
  color: #b91c1c;
  border-color: transparent;
}

.pill:has(input:focus-visible) {
  outline: 2px solid #7c3aed;
  outline-offset: 2px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

@media (max-width: 640px) {
  .actions {
    flex-direction: column-reverse;
  }

  .actions .btn {
    width: 100%;
  }
}
</style>
