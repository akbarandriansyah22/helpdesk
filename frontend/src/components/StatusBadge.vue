<script setup>
import { statusLabel } from '../utils/format'

defineProps({
  status: { type: String, required: true },
  dot: { type: Boolean, default: false },
})
</script>

<template>
  <span class="badge" :class="`status-${status}`">
    <span v-if="dot" class="badge-dot"></span>
    {{ statusLabel(status) }}
  </span>
</template>
