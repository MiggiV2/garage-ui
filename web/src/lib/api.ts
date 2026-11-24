import type {
  BucketInfo,
  BucketSummary,
  ClusterHealth,
  ClusterStatus,
  KeyInfo,
  KeySummary,
  Permissions,
  Quotas,
} from "./types";

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type Query = Record<string, string | boolean | undefined>;

async function call<T>(method: "GET" | "POST", operation: string, query: Query = {}, body?: unknown): Promise<T> {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== false) params.set(key, String(value));
  }
  const qs = params.toString();
  const response = await fetch(`/api/v2/${operation}${qs ? `?${qs}` : ""}`, {
    method,
    headers: {
      "X-Requested-With": "garage-ui",
      ...(body === undefined ? {} : { "Content-Type": "application/json" }),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const text = await response.text();
  if (!response.ok) {
    throw new ApiError(garageMessage(text) ?? `HTTP ${response.status}`, response.status);
  }
  return (text ? JSON.parse(text) : undefined) as T;
}

function garageMessage(text: string): string | undefined {
  try {
    const parsed = JSON.parse(text);
    return typeof parsed?.message === "string" ? parsed.message : undefined;
  } catch {
    return undefined;
  }
}

export interface WebsiteAccess {
  enabled: boolean;
  indexDocument?: string;
  errorDocument?: string | null;
}

export interface UiConfig {
  s3Endpoint: string;
  s3Region: string;
}

export const api = {
  uiConfig: async (): Promise<UiConfig> => {
    const response = await fetch("/api/config");
    if (!response.ok) throw new ApiError(`HTTP ${response.status}`, response.status);
    return response.json();
  },
  clusterHealth: () => call<ClusterHealth>("GET", "GetClusterHealth"),
  clusterStatus: () => call<ClusterStatus>("GET", "GetClusterStatus"),

  listBuckets: () => call<BucketSummary[]>("GET", "ListBuckets"),
  getBucketInfo: (id: string) => call<BucketInfo>("GET", "GetBucketInfo", { id }),
  createBucket: (globalAlias: string) => call<BucketInfo>("POST", "CreateBucket", {}, { globalAlias }),
  updateBucket: (id: string, update: { quotas?: Quotas; websiteAccess?: WebsiteAccess }) =>
    call<BucketInfo>("POST", "UpdateBucket", { id }, update),
  deleteBucket: (id: string) => call<void>("POST", "DeleteBucket", { id }),

  listKeys: () => call<KeySummary[]>("GET", "ListKeys"),
  getKeyInfo: (id: string, showSecretKey = false) => call<KeyInfo>("GET", "GetKeyInfo", { id, showSecretKey }),
  createKey: (name: string) => call<KeyInfo>("POST", "CreateKey", {}, { name }),
  setCreateBucketPermission: (id: string, allowed: boolean) =>
    call<KeyInfo>(
      "POST",
      "UpdateKey",
      { id },
      allowed ? { allow: { createBucket: true } } : { deny: { createBucket: true } },
    ),
  deleteKey: (id: string) => call<void>("POST", "DeleteKey", { id }),

  allowBucketKey: (bucketId: string, accessKeyId: string, permissions: Permissions) =>
    call<BucketInfo>("POST", "AllowBucketKey", {}, { bucketId, accessKeyId, permissions }),
  denyBucketKey: (bucketId: string, accessKeyId: string, permissions: Permissions) =>
    call<BucketInfo>("POST", "DenyBucketKey", {}, { bucketId, accessKeyId, permissions }),
};
