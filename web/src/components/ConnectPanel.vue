<script setup lang="ts">
import { computed, ref } from "vue";
import { connectionSnippets, type Snippets } from "../lib/snippets";
import CopyButton from "./CopyButton.vue";

const props = defineProps<{
  endpoint: string;
  region: string;
  accessKeyId: string;
  secretAccessKey?: string;
  keyName: string;
  buckets: string[];
}>();

const tabs: { id: keyof Snippets; label: string }[] = [
  { id: "env", label: "Environment" },
  { id: "awsCli", label: "AWS CLI" },
  { id: "rclone", label: "rclone" },
];
const active = ref<keyof Snippets>("env");
const bucket = ref(props.buckets[0] ?? "");

const snippet = computed(
  () =>
    connectionSnippets({
      endpoint: props.endpoint,
      region: props.region,
      accessKeyId: props.accessKeyId,
      secretAccessKey: props.secretAccessKey,
      bucket: bucket.value || undefined,
      profile: props.keyName,
    })[active.value],
);
</script>

<template>
  <div class="connect">
    <div class="connect-bar">
      <div class="tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          role="tab"
          class="tab"
          :class="{ active: active === tab.id }"
          :aria-selected="active === tab.id"
          @click="active = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>
      <select v-if="active === 'awsCli' && buckets.length" v-model="bucket" class="compact" aria-label="Bucket">
        <option v-for="name in buckets" :key="name" :value="name">{{ name }}</option>
      </select>
    </div>
    <div class="code-block">
      <pre><code>{{ snippet }}</code></pre>
      <CopyButton :value="snippet" label="snippet" />
    </div>
    <p v-if="!endpoint" class="hint">
      Set <code>S3_ENDPOINT</code> on the garage-ui container to fill in the endpoint.
    </p>
    <p v-if="!secretAccessKey" class="hint">Reveal the secret above to include it in the snippet.</p>
  </div>
</template>
