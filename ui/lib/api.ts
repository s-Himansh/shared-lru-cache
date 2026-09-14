const API_BASE = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080").replace(/\/$/, "");

export interface CacheEntry {
  key: string;
  value: string;
  shard: number;
}

export interface CacheState {
  entries: CacheEntry[];
  shards: number[];
  capacity: number;
  num_shards: number;
  size: number;
  metrics: Metrics;
}

export interface Metrics {
  hits: number;
  misses: number;
  evictions: number;
  puts: number;
  gets: number;
  deletes: number;
  hit_rate: number;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
  });
  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  health: () => request<{ status: string }>("/api/health"),
  state: () => request<CacheState>("/api/state"),
  keys: () => request<{ keys: string[]; count: number }>("/api/keys"),
  put: (key: string, value: string) =>
    request<{ status: string }>("/api/put", { method: "POST", body: JSON.stringify({ key, value }) }),
  get: (key: string) =>
    request<{ found: boolean; key: string; value?: string }>("/api/get", { method: "POST", body: JSON.stringify({ key }) }),
  delete: (key: string) =>
    request<{ deleted: boolean }>("/api/delete", { method: "POST", body: JSON.stringify({ key }) }),
  clear: () => request<{ status: string }>("/api/clear", { method: "POST" }),
  metrics: () => request<Metrics>("/api/metrics"),
  wsUrl: () => API_BASE.replace(/^http/, "ws") + "/ws",
};
