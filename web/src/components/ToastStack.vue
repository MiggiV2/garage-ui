<script setup lang="ts">
import { dismiss, toasts } from "../lib/notify";
import Icon from "./Icon.vue";
</script>

<template>
  <div class="toasts" aria-live="polite">
    <TransitionGroup name="toast">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="toast.kind"
        :role="toast.kind === 'error' ? 'alert' : 'status'"
      >
        <Icon :name="toast.kind === 'error' ? 'alert' : 'check'" class="toast-icon" />
        <span class="toast-text">{{ toast.text }}</span>
        <button class="icon-btn" aria-label="Dismiss" @click="dismiss(toast.id)">×</button>
      </div>
    </TransitionGroup>
  </div>
</template>
