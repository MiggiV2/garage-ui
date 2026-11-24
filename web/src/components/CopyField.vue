<script setup lang="ts">
import { ref } from "vue";
import CopyButton from "./CopyButton.vue";
import Icon from "./Icon.vue";

const props = defineProps<{ label: string; value: string; secret?: boolean }>();
const revealed = ref(!props.secret);
</script>

<template>
  <div class="copy-field">
    <span class="field-label">{{ label }}</span>
    <div class="copy-row">
      <code class="copy-value">{{ revealed ? value : "•".repeat(32) }}</code>
      <button
        v-if="secret"
        type="button"
        class="copy-btn"
        :aria-label="revealed ? 'Hide secret' : 'Reveal secret'"
        :title="revealed ? 'Hide' : 'Reveal'"
        @click="revealed = !revealed"
      >
        <Icon :name="revealed ? 'eyeOff' : 'eye'" :size="14" />
      </button>
      <CopyButton :value="value" :label="label" />
    </div>
  </div>
</template>
