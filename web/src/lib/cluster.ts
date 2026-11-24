import { ref } from "vue";
import { api, type UiConfig } from "./api";
import type { ClusterHealth } from "./types";

export const clusterHealth = ref<ClusterHealth>();
export const clusterVersion = ref<string>();
export const uiConfig = ref<UiConfig>({ s3Endpoint: "", s3Region: "garage" });

export async function refreshClusterSummary() {
  try {
    const [health, status] = await Promise.all([api.clusterHealth(), api.clusterStatus()]);
    clusterHealth.value = health;
    clusterVersion.value = status.nodes.find((n) => n.garageVersion)?.garageVersion ?? undefined;
  } catch {
    clusterHealth.value = undefined;
  }
}

export async function loadUiConfig() {
  try {
    uiConfig.value = await api.uiConfig();
  } catch {
    /* defaults stay in place; connection snippets fall back to placeholders */
  }
}
