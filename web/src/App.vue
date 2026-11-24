<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import Icon from "./components/Icon.vue";
import ToastStack from "./components/ToastStack.vue";
import { clusterHealth, clusterVersion, loadUiConfig, refreshClusterSummary } from "./lib/cluster";
import { cycleTheme, theme } from "./lib/theme";

const route = useRoute();

const links = [
  { to: "/", label: "Overview", icon: "overview" },
  { to: "/buckets", label: "Buckets", icon: "bucket" },
  { to: "/keys", label: "Access keys", icon: "key" },
] as const;

function isActive(to: string) {
  return to === "/" ? route.path === "/" : route.path.startsWith(to);
}

const healthLabel = computed(() => clusterHealth.value?.status ?? "unreachable");
const themeIcons = { system: "monitor", light: "sun", dark: "moon" } as const;
const themeIcon = computed(() => themeIcons[theme.value]);

let timer: ReturnType<typeof setInterval>;
onMounted(() => {
  void loadUiConfig();
  void refreshClusterSummary();
  timer = setInterval(refreshClusterSummary, 30_000);
});
onUnmounted(() => clearInterval(timer));
</script>

<template>
  <div class="shell">
    <aside class="sidebar">
      <RouterLink to="/" class="brand">
        <img src="/favicon.svg" alt="" width="24" height="24" />
        <span>Garage</span>
        <span class="brand-sub">console</span>
      </RouterLink>

      <nav aria-label="Main">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="nav-link"
          :class="{ active: isActive(link.to) }"
          :aria-current="isActive(link.to) ? 'page' : undefined"
          :title="link.label"
        >
          <Icon :name="link.icon" />
          <span>{{ link.label }}</span>
        </RouterLink>
      </nav>

      <div class="sidebar-foot">
        <div class="cluster-pill" :class="healthLabel" :title="`Cluster ${healthLabel}`">
          <span class="dot"></span>
          <span class="cluster-pill-label">{{ healthLabel }}</span>
          <span v-if="clusterVersion" class="mono muted">{{ clusterVersion }}</span>
        </div>
        <button
          class="icon-btn theme-btn"
          :title="`Theme: ${theme}`"
          :aria-label="`Theme: ${theme}`"
          @click="cycleTheme"
        >
          <Icon :name="themeIcon" />
        </button>
      </div>
    </aside>

    <main class="content">
      <RouterView :key="route.fullPath" />
    </main>
    <ToastStack />
  </div>
</template>
