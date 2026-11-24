<script setup lang="ts">
import { computed, ref, useTemplateRef } from "vue";
import { useRouter } from "vue-router";
import { api } from "../lib/api";
import { uiConfig } from "../lib/cluster";
import { shortId } from "../lib/format";
import { errorText, notify } from "../lib/notify";
import type { KeyBucket, Permissions } from "../lib/types";
import { useResource } from "../lib/useResource";
import ConfirmDelete from "../components/ConfirmDelete.vue";
import ConnectPanel from "../components/ConnectPanel.vue";
import CopyField from "../components/CopyField.vue";
import Icon from "../components/Icon.vue";
import PageState from "../components/PageState.vue";

const props = defineProps<{ id: string }>();
const router = useRouter();

const key = useResource(() => api.getKeyInfo(props.id));
const secret = ref<string>();
const revealing = ref(false);
const busy = ref(false);

const displayName = computed(() => key.data.value?.name || shortId(props.id));

async function revealSecret() {
  revealing.value = true;
  try {
    secret.value = (await api.getKeyInfo(props.id, true)).secretAccessKey ?? undefined;
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    revealing.value = false;
  }
}

async function toggleCreateBucket(allowed: boolean) {
  busy.value = true;
  try {
    await api.setCreateBucketPermission(props.id, allowed);
    notify("success", allowed ? "Key may now create buckets" : "Key may no longer create buckets");
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    await key.refresh();
    busy.value = false;
  }
}

const confirmDelete = useTemplateRef<InstanceType<typeof ConfirmDelete>>("confirmDelete");
async function remove() {
  busy.value = true;
  try {
    await api.deleteKey(props.id);
    notify("success", `Key ${displayName.value} deleted`);
    await router.push("/keys");
  } catch (e) {
    notify("error", errorText(e));
  } finally {
    busy.value = false;
  }
}

function bucketLabel(bucket: KeyBucket) {
  return bucket.globalAliases[0] ?? bucket.localAliases[0] ?? shortId(bucket.id);
}

const bucketNames = computed(() => (key.data.value?.buckets ?? []).map(bucketLabel));
const permissionColumns: (keyof Permissions)[] = ["read", "write", "owner"];
</script>

<template>
  <section>
    <nav class="breadcrumbs" aria-label="Breadcrumb">
      <RouterLink to="/keys">Access keys</RouterLink>
      <Icon name="chevron" :size="14" />
      <span>{{ displayName }}</span>
    </nav>

    <PageState :loading="key.loading.value && !key.data.value" :error="key.error.value" @retry="key.refresh" />

    <template v-if="key.data.value">
      <header class="page-head">
        <div>
          <h1 class="mono-title">{{ displayName }}</h1>
          <span class="status" :class="key.data.value.expired ? 'unavailable' : 'healthy'">
            <span class="dot"></span>{{ key.data.value.expired ? "expired" : "active" }}
          </span>
        </div>
        <button class="btn danger" :disabled="busy" @click="confirmDelete?.open()">
          <Icon name="trash" /> Delete key
        </button>
      </header>

      <div class="grid-2">
        <article class="card">
          <header class="card-head"><h2>Credentials</h2></header>
          <CopyField label="Access key ID" :value="key.data.value.accessKeyId" />
          <CopyField v-if="secret" label="Secret access key" :value="secret" secret />
          <div v-else class="field">
            <span class="field-label">Secret access key</span>
            <button class="btn" :disabled="revealing" @click="revealSecret"><Icon name="eye" /> Reveal secret</button>
          </div>
        </article>

        <article class="card">
          <header class="card-head"><h2>Key settings</h2></header>
          <label class="switch-row">
            <span>
              <strong>Allow bucket creation</strong>
              <span class="muted block">Lets this key create new buckets via the S3 API.</span>
            </span>
            <span class="switch">
              <input
                type="checkbox"
                :checked="key.data.value.permissions.createBucket === true"
                :disabled="busy"
                @change="toggleCreateBucket(($event.target as HTMLInputElement).checked)"
              />
            </span>
          </label>
          <dl class="kv">
            <dt>Created</dt>
            <dd class="mono-num">
              {{ key.data.value.created ? key.data.value.created.replace("T", " ").slice(0, 19) + " UTC" : "—" }}
            </dd>
            <dt>Expires</dt>
            <dd class="mono-num">{{ key.data.value.expiration?.replace("T", " ").slice(0, 19) ?? "never" }}</dd>
          </dl>
        </article>
      </div>

      <article class="card">
        <header class="card-head">
          <h2>Bucket permissions</h2>
          <span class="muted">Manage on the bucket page</span>
        </header>
        <div class="table-wrap flush">
          <table>
            <thead>
              <tr>
                <th>Bucket</th>
                <th v-for="perm in permissionColumns" :key="perm" class="center">{{ perm }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="bucket in key.data.value.buckets"
                :key="bucket.id"
                class="clickable"
                @click="router.push(`/buckets/${bucket.id}`)"
              >
                <td>
                  <RouterLink :to="`/buckets/${bucket.id}`" class="row-title" @click.stop>
                    {{ bucketLabel(bucket) }}
                  </RouterLink>
                </td>
                <td v-for="perm in permissionColumns" :key="perm" class="center">
                  <Icon v-if="bucket.permissions[perm]" name="check" class="perm-yes" :aria-label="perm" />
                  <span v-else class="muted">—</span>
                </td>
              </tr>
              <tr v-if="!key.data.value.buckets.length">
                <td colspan="4" class="empty">This key has no bucket access.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="card">
        <header class="card-head">
          <h2><Icon name="terminal" /> Connect</h2>
          <span class="muted mono">region {{ uiConfig.s3Region }}</span>
        </header>
        <ConnectPanel
          :endpoint="uiConfig.s3Endpoint"
          :region="uiConfig.s3Region"
          :access-key-id="key.data.value.accessKeyId"
          :secret-access-key="secret"
          :key-name="key.data.value.name"
          :buckets="bucketNames"
        />
      </article>

      <ConfirmDelete
        ref="confirmDelete"
        :title="`Delete key ${displayName}?`"
        :confirm-text="key.data.value.name || key.data.value.accessKeyId"
        description="Applications using this key lose access immediately. This cannot be undone."
        action="Delete key"
        @confirm="remove"
      />
    </template>
  </section>
</template>
