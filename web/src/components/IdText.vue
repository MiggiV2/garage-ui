<script setup lang="ts">
import CopyButton from "./CopyButton.vue";

defineProps<{ value: string; display?: string; label?: string }>();
</script>

<template>
  <span class="id-text">
    <code :title="value">{{ display ?? value }}</code>
    <CopyButton :value="value" :label="label" />
  </span>
</template>
