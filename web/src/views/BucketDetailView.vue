<script setup lang="ts">
import { computed, ref, useTemplateRef, watch } from "vue";
import { useRouter } from "vue-router";
import { api } from "../lib/api";
import { formatBytes, parseSize, shortId } from "../lib/format";
import { errorText, notify } from "../lib/notify";
import type { BucketKey, Permissions } from "../lib/types";
import { useResource } from "../lib/useResource";
import ConfirmDelete from "../components/ConfirmDelete.vue";
import Icon from "../components/Icon.vue";
import IdText from "../components/IdText.vue";
import PageState from "../components/PageState.vue";
import UsageBar from "../components/UsageBar.vue";

const props = defineProps<{ id: string }>();
const router = useRouter();

const bucket = useResource(() => api.getBucketInfo(props.id));
const keys = useResource(api.listKeys);

const name = computed(() => bucket.data.value?.globalAliases[0] ?? shortId(props.id));
const busy = ref(false);

async function run(action: () => Promise<unknown>, success: string): Promise<boolean> {
  busy.value = true;
  try {
    await action();
    notify("success", success);
    await bucket.refresh();
    return true;
  } catch (e) {
    notify("error", errorText(e));
    await bucket.refresh();
    return false;
  } finally {
    busy.value = false;
  }
}

const maxSize = ref("");
const maxObjects = ref("");
function resetQuotaForm() {
  const quotas = bucket.data.value?.quotas;
  maxSize.value = quotas?.maxSize != null ? formatBytes(quotas.maxSize) : "";
  maxObjects.value = quotas?.maxObjects != null ? String(quotas.maxObjects) : "";
}
watch(() => bucket.data.value?.quotas, resetQuotaForm, { immediate: true });

function saveQuotas() {
  let size: number | null;
  try {
    size = parseSize(maxSize.value);
  } catch (e) {
    notify("error", errorText(e));
    return;
  }
  const rawObjects = maxObjects.value.replace(/[_,\s]/g, "");
  const objects = rawObjects === "" ? null : Number(rawObjects);
  if (objects !== null && (!Number.isInteger(objects) || objects < 0)) {
    notify("error", "Max objects must be a non-negative integer");
    return;
  }
  void run(() => api.updateBucket(props.id, { quotas: { maxSize: size, maxObjects: objects } }), "Quotas updated");
}

function toggleWebsite(enabled: boolean) {
  void run(
    () =>
      api.updateBucket(props.id, {
        websiteAccess: enabled ? { enabled, indexDocument: "index.html" } : { enabled },
      }),
    enabled ? "Website access enabled" : "Website access disabled",
  );
}

function setPermission(key: BucketKey, perm: keyof Permissions, value: boolean) {
  const change: Permissions = { read: false, write: false, owner: false, [perm]: true };
  void run(
    () => (value ? api.allowBucketKey : api.denyBucketKey)(props.id, key.accessKeyId, change),
    `${value ? "Granted" : "Revoked"} ${perm} for ${key.name || key.accessKeyId}`,
  );
}

function revoke(key: BucketKey) {
  void run(
    () => api.denyBucketKey(props.id, key.accessKeyId, { read: true, write: true, owner: true }),
    `Revoked all access for ${key.name || key.accessKeyId}`,
  );
}

const grantKeyId = ref("");
const grantPerms = ref<Permissions>({ read: true, write: true, owner: false });
const grantableKeys = computed(() => {
  const assigned = new Set(
    bucket.data.value?.keys
      .filter((k) => k.permissions.read || k.permissions.write || k.permissions.owner)
      .map((k) => k.accessKeyId),
  );
  return (keys.data.value ?? []).filter((k) => !assigned.has(k.id)).sort((a, b) => a.name.localeCompare(b.name));
});
const activeKeys = computed(() =>
  (bucket.data.value?.keys ?? []).filter((k) => k.permissions.read || k.permissions.write || k.permissions.owner),
);
async function grant() {
  const keyId = grantKeyId.value;
  if (!keyId || !Object.values(grantPerms.value).some(Boolean)) return;
  if (await run(() => api.allowBucketKey(props.id, keyId, grantPerms.value), "Access granted")) {
    grantKeyId.value = "";
  }
}

const confirmDelete = useTemplateRef<InstanceType<typeof ConfirmDelete>>("confirmDelete");
async function remove() {
  busy.value = true;
  try {
    await api.deleteBucket(props.id);
    notify("success", `Bucket ${name.value} deleted`);
    await router.push("/buckets");
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    busy.value = false;
  }
}

const permissionColumns: (keyof Permissions)[] = ["read", "write", "owner"];
</script>

<template>
  <section>
    <nav class="breadcrumbs" aria-label="Breadcrumb">
      <RouterLink to="/buckets">Buckets</RouterLink>
      <Icon name="chevron" :size="14" />
      <span>{{ name }}</span>
    </nav>

    <PageState
      :loading="bucket.loading.value && !bucket.data.value"
      :error="bucket.error.value"
      @retry="bucket.refresh"
    />

    <template v-if="bucket.data.value">
      <header class="page-head">
        <div>
          <h1 class="mono-title">{{ name }}</h1>
          <IdText :value="bucket.data.value.id" label="bucket ID" class="sub" />
        </div>
        <button class="btn danger" :disabled="busy" @click="confirmDelete?.open()">
          <Icon name="trash" /> Delete bucket
        </button>
      </header>

      <div class="stats">
        <div class="stat">
          <span class="field-label">Size</span>
          <span class="stat-value">{{ formatBytes(bucket.data.value.bytes) }}</span>
        </div>
        <div class="stat">
          <span class="field-label">Objects</span>
          <span class="stat-value">{{ bucket.data.value.objects.toLocaleString("en-US") }}</span>
        </div>
        <div class="stat">
          <span class="field-label">Unfinished uploads</span>
          <span class="stat-value">
            {{ bucket.data.value.unfinishedUploads }}
            <span v-if="bucket.data.value.unfinishedMultipartUploadBytes" class="stat-sub">
              {{ formatBytes(bucket.data.value.unfinishedMultipartUploadBytes) }}
            </span>
          </span>
        </div>
        <div class="stat">
          <span class="field-label">Created</span>
          <span class="stat-value small-value mono-num">
            {{ new Date(bucket.data.value.created).toISOString().replace("T", " ").slice(0, 19) }} UTC
          </span>
        </div>
      </div>

      <div class="grid-2">
        <article class="card">
          <header class="card-head">
            <h2>Quotas</h2>
          </header>
          <UsageBar
            v-if="bucket.data.value.quotas.maxSize"
            label="Size quota"
            :used="bucket.data.value.bytes"
            :total="bucket.data.value.quotas.maxSize"
          />
          <form class="form-grid" @submit.prevent="saveQuotas">
            <label class="field">
              <span class="field-label">Max size</span>
              <input v-model="maxSize" class="mono" placeholder="unlimited — e.g. 500 GiB" />
            </label>
            <label class="field">
              <span class="field-label">Max objects</span>
              <input v-model="maxObjects" class="mono" inputmode="numeric" placeholder="unlimited" />
            </label>
            <div class="actions">
              <button type="button" class="btn ghost" :disabled="busy" @click="resetQuotaForm">Reset</button>
              <button class="btn primary" :disabled="busy">Save quotas</button>
            </div>
          </form>
        </article>

        <article class="card">
          <header class="card-head">
            <h2>Website access</h2>
            <label class="switch">
              <input
                type="checkbox"
                :checked="bucket.data.value.websiteAccess"
                :disabled="busy"
                aria-label="Website access"
                @change="toggleWebsite(($event.target as HTMLInputElement).checked)"
              />
            </label>
          </header>
          <p class="muted">Serve this bucket as a static website through Garage's web endpoint.</p>
          <dl v-if="bucket.data.value.websiteAccess && bucket.data.value.websiteConfig" class="kv">
            <dt>Index document</dt>
            <dd class="mono">{{ bucket.data.value.websiteConfig.indexDocument }}</dd>
            <dt>Error document</dt>
            <dd class="mono">{{ bucket.data.value.websiteConfig.errorDocument ?? "—" }}</dd>
          </dl>
          <dl class="kv">
            <dt>Aliases</dt>
            <dd>
              <span v-for="alias in bucket.data.value.globalAliases" :key="alias" class="tag mono">{{ alias }}</span>
              <span v-if="!bucket.data.value.globalAliases.length" class="muted">none</span>
            </dd>
          </dl>
        </article>
      </div>

      <article class="card">
        <header class="card-head">
          <h2>Access keys</h2>
          <span class="muted">{{ activeKeys.length }} with access</span>
        </header>
        <div class="table-wrap flush">
          <table>
            <thead>
              <tr>
                <th>Key</th>
                <th v-for="perm in permissionColumns" :key="perm" class="center">{{ perm }}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="key in activeKeys" :key="key.accessKeyId">
                <td>
                  <RouterLink :to="`/keys/${key.accessKeyId}`" class="row-title">{{
                    key.name || "unnamed"
                  }}</RouterLink>
                  <IdText :value="key.accessKeyId" label="access key ID" class="sub" />
                </td>
                <td v-for="perm in permissionColumns" :key="perm" class="center">
                  <input
                    type="checkbox"
                    :aria-label="`${perm} for ${key.name}`"
                    :checked="key.permissions[perm]"
                    :disabled="busy"
                    @change="setPermission(key, perm, ($event.target as HTMLInputElement).checked)"
                  />
                </td>
                <td class="right">
                  <button class="btn small ghost" :disabled="busy" @click="revoke(key)">Revoke</button>
                </td>
              </tr>
              <tr v-if="!activeKeys.length">
                <td colspan="5" class="empty">No key has access to this bucket yet.</td>
              </tr>
            </tbody>
          </table>
        </div>

        <form class="grant" @submit.prevent="grant">
          <select v-model="grantKeyId" aria-label="Select key" required>
            <option value="" disabled>
              {{ grantableKeys.length ? "Select a key…" : "All keys already have access" }}
            </option>
            <option v-for="key in grantableKeys" :key="key.id" :value="key.id">
              {{ key.name || key.id }} ({{ key.id }})
            </option>
          </select>
          <div class="check-group">
            <label v-for="perm in permissionColumns" :key="perm" class="check">
              <input v-model="grantPerms[perm]" type="checkbox" /> {{ perm }}
            </label>
          </div>
          <button class="btn primary" :disabled="busy || !grantKeyId"><Icon name="plus" /> Grant access</button>
        </form>
      </article>

      <ConfirmDelete
        ref="confirmDelete"
        :title="`Delete bucket ${name}?`"
        :confirm-text="name"
        description="Garage only deletes empty buckets. This cannot be undone."
        action="Delete bucket"
        @confirm="remove"
      />
    </template>
  </section>
</template>
