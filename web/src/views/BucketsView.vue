<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../lib/api";
import { bucketName, formatBytes, shortId } from "../lib/format";
import { errorText, notify } from "../lib/notify";
import { useResource } from "../lib/useResource";
import Icon from "../components/Icon.vue";
import IdText from "../components/IdText.vue";
import PageState from "../components/PageState.vue";

const router = useRouter();

const buckets = useResource(async () => {
  const list = await api.listBuckets();
  return Promise.all(list.map(async (summary) => ({ summary, info: await api.getBucketInfo(summary.id) })));
});

const filter = ref("");
const rows = computed(() => {
  const needle = filter.value.trim().toLowerCase();
  return (buckets.data.value ?? [])
    .filter(
      ({ summary }) => !needle || bucketName(summary).toLowerCase().includes(needle) || summary.id.startsWith(needle),
    )
    .sort((a, b) => bucketName(a.summary).localeCompare(bucketName(b.summary)));
});
const totals = computed(() =>
  (buckets.data.value ?? []).reduce(
    (acc, { info }) => ({ bytes: acc.bytes + info.bytes, objects: acc.objects + info.objects }),
    { bytes: 0, objects: 0 },
  ),
);

const newName = ref("");
const creating = ref(false);
async function create() {
  const name = newName.value.trim();
  if (!name) return;
  creating.value = true;
  try {
    const bucket = await api.createBucket(name);
    notify("success", `Bucket ${name} created`);
    newName.value = "";
    await router.push(`/buckets/${bucket.id}`);
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    creating.value = false;
  }
}
</script>

<template>
  <section>
    <header class="page-head">
      <div>
        <h1>Buckets</h1>
        <p v-if="buckets.data.value" class="subtitle">
          {{ buckets.data.value.length }} buckets · {{ formatBytes(totals.bytes) }} ·
          {{ totals.objects.toLocaleString("en-US") }} objects
        </p>
      </div>
      <form class="inline-form" @submit.prevent="create">
        <input
          v-model="newName"
          class="mono"
          placeholder="new-bucket-name"
          aria-label="New bucket name"
          pattern="[a-z0-9][a-z0-9.\-]{1,61}[a-z0-9]"
          title="3–63 characters: lowercase letters, digits, dots and hyphens"
          required
        />
        <button class="btn primary" :disabled="creating"><Icon name="plus" /> Create bucket</button>
      </form>
    </header>

    <PageState
      :loading="buckets.loading.value && !buckets.data.value"
      :error="buckets.error.value"
      @retry="buckets.refresh"
    />

    <template v-if="buckets.data.value">
      <div class="toolbar">
        <label class="search">
          <Icon name="search" />
          <input v-model="filter" type="search" placeholder="Filter by name or ID" aria-label="Filter buckets" />
        </label>
        <button class="btn ghost" :disabled="buckets.loading.value" @click="buckets.refresh">
          <Icon name="refresh" :class="{ spin: buckets.loading.value }" /> Refresh
        </button>
      </div>

      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Bucket</th>
              <th class="num">Size</th>
              <th class="num">Objects</th>
              <th class="num">Quota</th>
              <th>Access</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="{ summary, info } in rows"
              :key="summary.id"
              class="clickable"
              @click="router.push(`/buckets/${summary.id}`)"
            >
              <td>
                <div class="name-cell">
                  <RouterLink :to="`/buckets/${summary.id}`" class="row-title" @click.stop>
                    {{ bucketName(summary) }}
                  </RouterLink>
                  <span v-if="info.websiteAccess" class="tag accent" title="Website access enabled">
                    <Icon name="globe" :size="12" /> web
                  </span>
                </div>
                <IdText :value="summary.id" :display="shortId(summary.id)" label="bucket ID" class="sub" />
              </td>
              <td class="num">{{ formatBytes(info.bytes) }}</td>
              <td class="num">{{ info.objects.toLocaleString("en-US") }}</td>
              <td class="num">
                <span v-if="info.quotas.maxSize !== null">{{ formatBytes(info.quotas.maxSize) }}</span>
                <span v-else class="muted">—</span>
              </td>
              <td>
                <span v-for="key in info.keys" :key="key.accessKeyId" class="tag">{{
                  key.name || key.accessKeyId
                }}</span>
                <span v-if="!info.keys.length" class="muted">no keys</span>
              </td>
              <td class="mono-num nowrap">{{ summary.created.slice(0, 10) }}</td>
            </tr>
            <tr v-if="!rows.length">
              <td colspan="6" class="empty">
                {{ filter ? "No bucket matches this filter." : "No buckets yet. Create one above." }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </section>
</template>
