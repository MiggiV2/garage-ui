<script setup lang="ts">
defineProps<{ loading: boolean; error?: string }>();
defineEmits<{ retry: [] }>();
</script>

<template>
  <div v-if="error" class="state error-state" role="alert">
    <div>
      <strong>Request failed</strong>
      <p class="mono">{{ error }}</p>
    </div>
    <button class="btn" @click="$emit('retry')">Retry</button>
  </div>
  <div v-else-if="loading" class="state loading-state">
    <span class="spinner" aria-hidden="true"></span>
    Loading…
  </div>
</template>
