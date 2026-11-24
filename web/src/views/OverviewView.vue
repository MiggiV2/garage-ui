<script setup lang="ts">
import { computed } from "vue";
import { api } from "../lib/api";
import { formatBytes, shortId } from "../lib/format";
import { useResource } from "../lib/useResource";
import Icon from "../components/Icon.vue";
import IdText from "../components/IdText.vue";
import PageState from "../components/PageState.vue";
import UsageBar from "../components/UsageBar.vue";

const health = useResource(api.clusterHealth);
const status = useResource(api.clusterStatus);
const buckets = useResource(async () => {
  const list = await api.listBuckets();
  const infos = await Promise.all(list.map((b) => api.getBucketInfo(b.id)));
  return {
    count: list.length,
    bytes: infos.reduce((sum, b) => sum + b.bytes, 0),
    objects: infos.reduce((sum, b) => sum + b.objects, 0),
  };
});
const keys = useResource(api.listKeys);

const loading = computed(() => health.loading.value || status.loading.value);

function refresh() {
  void health.refresh();
  void status.refresh();
  void buckets.refresh();
  void keys.refresh();
}
</script>

<template>
  <section>
    <header class="page-head">
      <div>
        <h1>Overview</h1>
        <p class="subtitle">Cluster health, nodes and storage at a glance.</p>
      </div>
      <button class="btn" :disabled="loading" @click="refresh">
        <Icon name="refresh" :class="{ spin: loading }" /> Refresh
      </button>
    </header>

    <PageState :loading="health.loading.value && !health.data.value" :error="health.error.value" @retry="refresh" />

    <div v-if="health.data.value" class="stats">
      <div class="stat">
        <span class="field-label">Cluster status</span>
        <span class="status" :class="health.data.value.status">
          <span class="dot"></span>{{ health.data.value.status }}
        </span>
      </div>
      <div class="stat">
        <span class="field-label">Storage nodes up</span>
        <span class="stat-value">
          {{ health.data.value.storageNodesUp }}<span class="muted"> / {{ health.data.value.storageNodes }}</span>
        </span>
      </div>
      <div class="stat">
        <span class="field-label">Partitions healthy</span>
        <span class="stat-value">
          {{ health.data.value.partitionsAllOk }}<span class="muted"> / {{ health.data.value.partitions }}</span>
        </span>
      </div>
      <RouterLink to="/buckets" class="stat stat-link">
        <span class="field-label">Buckets</span>
        <span class="stat-value">
          {{ buckets.data.value?.count ?? "–" }}
          <span class="stat-sub" v-if="buckets.data.value">{{ formatBytes(buckets.data.value.bytes) }}</span>
        </span>
      </RouterLink>
      <RouterLink to="/keys" class="stat stat-link">
        <span class="field-label">Access keys</span>
        <span class="stat-value">{{ keys.data.value?.length ?? "–" }}</span>
      </RouterLink>
    </div>

    <h2 class="section-title">Nodes</h2>
    <PageState :loading="status.loading.value && !status.data.value" :error="status.error.value" @retry="refresh" />
    <div v-if="status.data.value" class="cards">
      <article v-for="node in status.data.value.nodes" :key="node.id" class="card">
        <header class="card-head">
          <div class="card-title">
            <h3>{{ node.hostname ?? shortId(node.id) }}</h3>
            <IdText :value="node.id" :display="shortId(node.id)" label="node ID" />
          </div>
          <span class="status" :class="node.isUp ? 'healthy' : 'unavailable'">
            <span class="dot"></span>{{ node.isUp ? "up" : "down" }}
          </span>
        </header>
        <dl class="kv">
          <dt>Version</dt>
          <dd class="mono">{{ node.garageVersion ?? "–" }}</dd>
          <dt>Address</dt>
          <dd class="mono">{{ node.addr ?? "–" }}</dd>
          <dt>Zone</dt>
          <dd>{{ node.role?.zone ?? "no role" }}</dd>
          <dt>Capacity</dt>
          <dd class="mono">{{ formatBytes(node.role?.capacity ?? null) }}</dd>
          <dt>Resync queue</dt>
          <dd class="mono">{{ node.resyncQueueLen ?? "–" }}</dd>
          <template v-if="node.role?.tags.length">
            <dt>Tags</dt>
            <dd>
              <span v-for="tag in node.role.tags" :key="tag" class="tag">{{ tag }}</span>
            </dd>
          </template>
        </dl>
        <UsageBar
          v-if="node.dataPartition"
          label="Data partition"
          :used="node.dataPartition.total - node.dataPartition.available"
          :total="node.dataPartition.total"
        />
        <UsageBar
          v-if="node.metadataPartition"
          label="Metadata partition"
          :used="node.metadataPartition.total - node.metadataPartition.available"
          :total="node.metadataPartition.total"
        />
      </article>
    </div>
  </section>
</template>
