<script setup lang="ts">
import { computed } from "vue";
import { formatBytes } from "../lib/format";

const props = defineProps<{ label: string; used: number; total: number }>();
const percent = computed(() => (props.total > 0 ? Math.min(100, (props.used / props.total) * 100) : 0));
const level = computed(() => (percent.value >= 90 ? "crit" : percent.value >= 75 ? "warn" : "ok"));
</script>

<template>
  <div class="usage">
    <div class="usage-head">
      <span class="field-label">{{ label }}</span>
      <span class="mono-num">
        {{ formatBytes(used) }} <span class="muted">/ {{ formatBytes(total) }}</span>
        <span class="usage-pct" :class="level">{{ percent.toFixed(1) }}%</span>
      </span>
    </div>
    <div
      class="usage-track"
      role="meter"
      :aria-valuenow="Math.round(percent)"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-label="`${label} usage`"
    >
      <div class="usage-fill" :class="level" :style="{ width: `${percent}%` }"></div>
    </div>
  </div>
</template>
