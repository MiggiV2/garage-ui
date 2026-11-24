export interface ClusterHealth {
  status: "healthy" | "degraded" | "unavailable";
  knownNodes: number;
  connectedNodes: number;
  storageNodes: number;
  storageNodesUp: number;
  partitions: number;
  partitionsQuorum: number;
  partitionsAllOk: number;
}

export interface FreeSpace {
  available: number;
  total: number;
}

export interface NodeStatus {
  id: string;
  garageVersion: string | null;
  addr: string | null;
  hostname: string | null;
  isUp: boolean;
  lastSeenSecsAgo: number | null;
  role: { zone: string; tags: string[]; capacity: number | null } | null;
  draining: boolean;
  dataPartition: FreeSpace | null;
  metadataPartition: FreeSpace | null;
  resyncQueueLen?: number | null;
}

export interface ClusterStatus {
  layoutVersion: number;
  nodes: NodeStatus[];
}

export interface LocalAlias {
  accessKeyId: string;
  alias: string;
}

export interface BucketSummary {
  id: string;
  created: string;
  globalAliases: string[];
  localAliases: LocalAlias[];
}

export interface Permissions {
  read: boolean;
  write: boolean;
  owner: boolean;
}

export interface Quotas {
  maxSize: number | null;
  maxObjects: number | null;
}

export interface BucketKey {
  accessKeyId: string;
  name: string;
  permissions: Permissions;
  bucketLocalAliases: string[];
}

export interface BucketInfo {
  id: string;
  created: string;
  globalAliases: string[];
  websiteAccess: boolean;
  websiteConfig?: { indexDocument: string; errorDocument: string | null } | null;
  keys: BucketKey[];
  objects: number;
  bytes: number;
  unfinishedUploads: number;
  unfinishedMultipartUploads: number;
  unfinishedMultipartUploadParts: number;
  unfinishedMultipartUploadBytes: number;
  quotas: Quotas;
}

export interface KeySummary {
  id: string;
  name: string;
  created: string | null;
  expiration: string | null;
  expired: boolean;
}

export interface KeyBucket {
  id: string;
  globalAliases: string[];
  localAliases: string[];
  permissions: Permissions;
}

export interface KeyInfo {
  accessKeyId: string;
  name: string;
  created: string | null;
  expiration: string | null;
  expired: boolean;
  secretAccessKey: string | null;
  permissions: { createBucket?: boolean };
  buckets: KeyBucket[];
}
