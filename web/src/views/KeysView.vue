<script setup lang="ts">
import { computed, ref, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { api } from "../lib/api";
import { errorText, notify } from "../lib/notify";
import type { KeyInfo } from "../lib/types";
import { useResource } from "../lib/useResource";
import CopyField from "../components/CopyField.vue";
import Icon from "../components/Icon.vue";
import IdText from "../components/IdText.vue";
import PageState from "../components/PageState.vue";

const router = useRouter();
const keys = useResource(async () => {
  const list = await api.listKeys();
  return Promise.all(list.map(async (summary) => ({ summary, info: await api.getKeyInfo(summary.id) })));
});

const filter = ref("");
const rows = computed(() => {
  const needle = filter.value.trim().toLowerCase();
  return (keys.data.value ?? [])
    .filter(
      ({ summary }) =>
        !needle || summary.name.toLowerCase().includes(needle) || summary.id.toLowerCase().includes(needle),
    )
    .sort((a, b) => a.summary.name.localeCompare(b.summary.name));
});

const newName = ref("");
const creating = ref(false);
const created = ref<KeyInfo>();
const dialog = useTemplateRef<HTMLDialogElement>("dialog");

async function create() {
  const name = newName.value.trim();
  if (!name) return;
  creating.value = true;
  try {
    created.value = await api.createKey(name);
    newName.value = "";
    dialog.value?.showModal();
    await keys.refresh();
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    creating.value = false;
  }
}

function closeDialog() {
  dialog.value?.close();
  if (created.value) void router.push(`/keys/${created.value.accessKeyId}`);
}

function bucketLabel(bucket: { id: string; globalAliases: string[]; localAliases: string[] }) {
  return bucket.globalAliases[0] ?? bucket.localAliases[0] ?? bucket.id.slice(0, 12);
}
</script>

<template>
  <section>
    <header class="page-head">
      <div>
        <h1>Access keys</h1>
        <p v-if="keys.data.value" class="subtitle">{{ keys.data.value.length }} keys</p>
      </div>
      <form class="inline-form" @submit.prevent="create">
        <input v-model="newName" placeholder="key-name" class="mono" aria-label="New key name" required />
        <button class="btn primary" :disabled="creating"><Icon name="plus" /> Create key</button>
      </form>
    </header>

    <PageState :loading="keys.loading.value && !keys.data.value" :error="keys.error.value" @retry="keys.refresh" />

    <template v-if="keys.data.value">
      <div class="toolbar">
        <label class="search">
          <Icon name="search" />
          <input v-model="filter" type="search" placeholder="Filter by name or key ID" aria-label="Filter keys" />
        </label>
        <button class="btn ghost" :disabled="keys.loading.value" @click="keys.refresh">
          <Icon name="refresh" :class="{ spin: keys.loading.value }" /> Refresh
        </button>
      </div>

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Access key ID</th>
              <th>Buckets</th>
              <th>Created</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="{ summary, info } in rows"
              :key="summary.id"
              class="clickable"
              @click="router.push(`/keys/${summary.id}`)"
            >
              <td>
                <RouterLink :to="`/keys/${summary.id}`" class="row-title" @click.stop>
                  {{ summary.name || "unnamed" }}
                </RouterLink>
              </td>
              <td><IdText :value="summary.id" label="access key ID" /></td>
              <td>
                <span v-for="bucket in info.buckets" :key="bucket.id" class="tag">{{ bucketLabel(bucket) }}</span>
                <span v-if="!info.buckets.length" class="muted">none</span>
              </td>
              <td class="mono-num nowrap">{{ summary.created?.slice(0, 10) ?? "—" }}</td>
              <td>
                <span class="status" :class="summary.expired ? 'unavailable' : 'healthy'">
                  <span class="dot"></span>{{ summary.expired ? "expired" : "active" }}
                </span>
              </td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="5" class="empty">
                {{ filter ? "No key matches this filter." : "No access keys yet. Create one above." }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>

    <dialog ref="dialog" class="modal" @cancel.prevent="closeDialog">
      <template v-if="created">
        <h2>Key {{ created.name }} created</h2>
        <p class="callout warn">
          <Icon name="alert" /> Copy the secret now. You can reveal it again later on the key's page.
        </p>
        <CopyField label="Access key ID" :value="created.accessKeyId" />
        <CopyField label="Secret access key" :value="created.secretAccessKey ?? ''" />
        <div class="actions">
          <button class="btn primary" @click="closeDialog">Done</button>
        </div>
      </template>
    </dialog>
  </section>
</template>
