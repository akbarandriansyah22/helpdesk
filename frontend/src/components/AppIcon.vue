<script setup>
defineProps({
  name: { type: String, required: true },
})
</script>

<template>
  <svg v-if="name === 'headset'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M4 13a8 8 0 0 1 16 0" />
    <rect x="3" y="13" width="4.2" height="7" rx="1.6" />
    <rect x="16.8" y="13" width="4.2" height="7" rx="1.6" />
  </svg>
  <svg v-else-if="name === 'ticket'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M4 8.5A2.5 2.5 0 0 1 6.5 6h11A2.5 2.5 0 0 1 20 8.5V10a2 2 0 0 0 0 4v1.5a2.5 2.5 0 0 1-2.5 2.5h-11A2.5 2.5 0 0 1 4 15.5V14a2 2 0 0 0 0-4V8.5Z" />
    <path d="M12 7.5v9" />
  </svg>
  <svg v-else-if="name === 'users'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M16 20v-1.5A3.5 3.5 0 0 0 12.5 15h-5A3.5 3.5 0 0 0 4 18.5V20" />
    <circle cx="10" cy="8" r="3" />
    <path d="M20 20v-1.5A3.5 3.5 0 0 0 17 15.1" />
    <path d="M16 5.1a3 3 0 0 1 0 5.8" />
  </svg>
  <svg v-else-if="name === 'search'" viewBox="0 0 24 24" aria-hidden="true">
    <circle cx="11" cy="11" r="6.5" />
    <path d="m16 16 4 4" />
  </svg>
  <svg v-else-if="name === 'plus'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M12 5v14M5 12h14" />
  </svg>
  <svg v-else-if="name === 'mail'" viewBox="0 0 24 24" aria-hidden="true">
    <rect x="3.5" y="5.5" width="17" height="13" rx="2" />
    <path d="m4 7 8 6 8-6" />
  </svg>
  <svg v-else-if="name === 'lock'" viewBox="0 0 24 24" aria-hidden="true">
    <rect x="5" y="11" width="14" height="9" rx="2" />
    <path d="M8 11V8a4 4 0 0 1 8 0v3" />
  </svg>
  <svg v-else-if="name === 'eye'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M2.5 12S6 6.5 12 6.5 21.5 12 21.5 12 18 17.5 12 17.5 2.5 12 2.5 12Z" />
    <circle cx="12" cy="12" r="2.5" />
  </svg>
  <svg v-else-if="name === 'eye-off'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M4 5.5 19 19" />
    <path d="M9.5 6.8A9.8 9.8 0 0 1 12 6.5c6 0 9.5 5.5 9.5 5.5a16 16 0 0 1-3.2 3.6" />
    <path d="M6.2 8.2C3.8 10 2.5 12 2.5 12S6 17.5 12 17.5a9 9 0 0 0 3.2-.6" />
  </svg>
  <svg v-else-if="name === 'menu'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M4 7h16M4 12h16M4 17h16" />
  </svg>
  <svg v-else-if="name === 'logout'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M10 7V5.5A1.5 1.5 0 0 1 11.5 4h7A1.5 1.5 0 0 1 20 5.5v13a1.5 1.5 0 0 1-1.5 1.5h-7A1.5 1.5 0 0 1 10 18.5V17" />
    <path d="M4 12h10" />
    <path d="m11 8 4 4-4 4" />
  </svg>
  <svg v-else-if="name === 'chat'" viewBox="0 0 24 24" aria-hidden="true">
    <path d="M6 17.5 3.5 20V6.5A2 2 0 0 1 5.5 4.5h13a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H6Z" />
  </svg>
  <svg v-else viewBox="0 0 24 24" aria-hidden="true">
    <path d="M5 12h14" />
    <path d="m13 6 6 6-6 6" />
  </svg>
</template>

<style scoped>
svg {
  width: 1.15em;
  height: 1.15em;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  display: block;
}
</style>
