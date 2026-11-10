<script setup>
import { priorityLabel } from '../utils/format'

defineProps({
  priority: { type: String, required: true },
})
</script>

<template>
  <span class="badge" :class="`priority-${priority}`">
    {{ priorityLabel(priority) }}
  </span>
</template>
