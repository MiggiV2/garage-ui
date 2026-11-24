<script setup lang="ts">
import { ref } from "vue";
import Icon from "./Icon.vue";
import { notify } from "../lib/notify";

const props = defineProps<{ value: string; label?: string }>();
const copied = ref(false);

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value);
    copied.value = true;
    setTimeout(() => (copied.value = false), 1400);
  } catch {
    notify("error", "Clipboard not available (requires HTTPS or localhost)");
  }
}
</script>

<template>
  <button
    type="button"
    class="copy-btn"
    :class="{ copied }"
    :title="copied ? 'Copied' : `Copy ${label ?? ''}`.trim()"
    :aria-label="`Copy ${label ?? 'value'}`"
    @click.stop="copy"
  >
    <Icon :name="copied ? 'check' : 'copy'" :size="14" />
  </button>
</template>
