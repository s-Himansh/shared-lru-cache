"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { api, CacheState, Metrics, CacheEntry } from "../lib/api";

export default function Dashboard() {
  const [state, setState] = useState<CacheState | null>(null);
  const [connected, setConnected] = useState(false);
  const [key, setKey] = useState("");
  const [value, setValue] = useState("");
  const [lookupKey, setLookupKey] = useState("");
  const [lookupResult, setLookupResult] = useState<{ found: boolean; key: string; value?: string } | null>(null);
  const [metricsHistory, setMetricsHistory] = useState<Metrics[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<() => void>(() => {});

  const connect = useCallback(() => {
    const ws = new WebSocket(api.wsUrl());
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => {
      setConnected(false);
      setTimeout(reconnectRef.current, 2000);
    };
    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data);
        if (event.type === "state" && event.data) {
          const newState = typeof event.data === "string" ? JSON.parse(event.data) : event.data;
          setState(newState);
          setMetricsHistory((prev) => [...prev.slice(-59), event.metrics || newState.metrics]);
        }
      } catch {}
    };
  }, []);

  useEffect(() => {
    reconnectRef.current = connect;
  }, [connect]);

  useEffect(() => {
    connect();
    return () => wsRef.current?.close();
  }, [connect]);

  useEffect(() => {
    if (!connected) {
      api.state().then(setState).catch(() => {});
    }
  }, [connected]);

  const handlePut = async () => {
    if (!key) return;
    await api.put(key, value);
    setKey("");
    setValue("");
  };

  const handleGet = async () => {
    if (!lookupKey) return;
    const result = await api.get(lookupKey);
    setLookupResult(result);
  };

  const handleDelete = async (k: string) => {
    await api.delete(k);
  };

  const handleClear = async () => {
    await api.clear();
    setMetricsHistory([]);
  };

  const metrics = state?.metrics || { hits: 0, misses: 0, evictions: 0, puts: 0, gets: 0, deletes: 0, hit_rate: 0 };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-100 via-slate-50 to-blue-50/50 text-slate-800">
      {/* Header */}
      <header className="border-b border-slate-200/60 bg-slate-100/80 backdrop-blur-xl sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center shadow-lg shadow-cyan-500/25">
              <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
              </svg>
            </div>
            <div>
              <h1 className="text-xl font-bold">Sharded LRU Cache</h1>
              <p className="text-xs text-slate-500">High-performance concurrent cache dashboard</p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-slate-500">
            <div className={`w-2 h-2 rounded-full ${connected ? "bg-emerald-500 animate-pulse" : "bg-red-400"}`}></div>
            {connected ? "Live" : "Disconnected"}
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Left: Controls */}
          <div className="space-y-6">
            {/* Put */}
            <div className="bg-slate-50/80 border border-slate-200/50 rounded-2xl p-6 shadow-sm">
              <h2 className="text-sm font-bold text-slate-500 uppercase tracking-wider mb-4">Add / Update</h2>
              <div className="space-y-3">
                <input
                  type="text"
                  value={key}
                  onChange={(e) => setKey(e.target.value)}
                  placeholder="Key"
                  className="w-full bg-slate-100/60 border border-slate-200/50 rounded-xl px-4 py-3 text-slate-800 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30"
                  onKeyDown={(e) => e.key === "Enter" && handlePut()}
                />
                <input
                  type="text"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="Value"
                  className="w-full bg-slate-100/60 border border-slate-200/50 rounded-xl px-4 py-3 text-slate-800 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30"
                  onKeyDown={(e) => e.key === "Enter" && handlePut()}
                />
                <button onClick={handlePut} className="w-full py-3 rounded-xl bg-gradient-to-r from-cyan-500 to-blue-600 text-white font-bold shadow-lg shadow-cyan-500/25 hover:shadow-cyan-500/40 transition-all">
                  PUT
                </button>
              </div>
            </div>

            {/* Get */}
            <div className="bg-slate-50/80 border border-slate-200/50 rounded-2xl p-6 shadow-sm">
              <h2 className="text-sm font-bold text-slate-500 uppercase tracking-wider mb-4">Lookup</h2>
              <div className="space-y-3">
                <input
                  type="text"
                  value={lookupKey}
                  onChange={(e) => setLookupKey(e.target.value)}
                  placeholder="Key to lookup"
                  className="w-full bg-slate-100/60 border border-slate-200/50 rounded-xl px-4 py-3 text-slate-800 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500/30"
                  onKeyDown={(e) => e.key === "Enter" && handleGet()}
                />
                <button onClick={handleGet} className="w-full py-3 rounded-xl bg-slate-200/60 border border-slate-200/50 text-slate-600 font-bold hover:bg-slate-200 transition-all">
                  GET
                </button>
                {lookupResult && (
                  <div className={`rounded-xl p-3 text-sm font-mono ${lookupResult.found ? "bg-emerald-50 border border-emerald-200 text-emerald-700" : "bg-red-50 border border-red-200 text-red-600"}`}>
                    {lookupResult.found ? `${lookupResult.key} → ${lookupResult.value}` : `Key "${lookupResult.key}" not found`}
                  </div>
                )}
              </div>
            </div>

            {/* Clear */}
            <button onClick={handleClear} className="w-full py-3 rounded-xl bg-red-50 border border-red-200 text-red-600 font-bold hover:bg-red-100 transition-all">
              Clear Cache
            </button>
          </div>

          {/* Right: Dashboard */}
          <div className="lg:col-span-2 space-y-6">
            {/* Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard label="Size" value={`${state?.size || 0} / ${state?.capacity || 0}`} color="text-cyan-600" />
              <StatCard label="Hit Rate" value={`${metrics.hit_rate.toFixed(1)}%`} color="text-emerald-600" />
              <StatCard label="Hits" value={metrics.hits.toLocaleString()} color="text-emerald-600" />
              <StatCard label="Misses" value={metrics.misses.toLocaleString()} color="text-red-500" />
            </div>

            {/* Shard Distribution */}
            <div className="bg-slate-50/80 border border-slate-200/50 rounded-2xl p-6 shadow-sm">
              <h3 className="text-sm font-bold text-slate-500 uppercase tracking-wider mb-4">Shard Distribution</h3>
              <div className="flex gap-1 items-end h-24">
                {(state?.shards || []).map((size, i) => {
                  const max = Math.max(...(state?.shards || [1]), 1);
                  return (
                    <div key={i} className="flex-1 flex flex-col items-center gap-1">
                      <div
                        className="w-full rounded-t bg-gradient-to-t from-cyan-600 to-cyan-400 transition-all duration-300"
                        style={{ height: `${(size / max) * 100}%`, minHeight: size > 0 ? "4px" : "0" }}
                      />
                      <span className="text-[10px] text-slate-400">{size}</span>
                    </div>
                  );
                })}
              </div>
              <div className="flex justify-between mt-2 text-xs text-slate-400">
                <span>Shard 0</span>
                <span>Shard {(state?.num_shards || 16) - 1}</span>
              </div>
            </div>

            {/* Hit Rate Chart */}
            <div className="bg-slate-50/80 border border-slate-200/50 rounded-2xl p-6 shadow-sm">
              <h3 className="text-sm font-bold text-slate-500 uppercase tracking-wider mb-4">Hit Rate (last 60s)</h3>
              <div className="flex gap-px items-end h-16">
                {metricsHistory.map((m, i) => (
                  <div
                    key={i}
                    className="flex-1 rounded-t bg-gradient-to-t from-emerald-600 to-emerald-400 transition-all duration-200"
                    style={{ height: `${m.hit_rate}%`, minHeight: m.hit_rate > 0 ? "2px" : "0" }}
                  />
                ))}
                {metricsHistory.length === 0 && (
                  <div className="w-full text-center text-slate-400 text-sm py-4">Waiting for data...</div>
                )}
              </div>
            </div>

            {/* Cache Entries */}
            <div className="bg-slate-50/80 border border-slate-200/50 rounded-2xl p-6 shadow-sm">
              <h3 className="text-sm font-bold text-slate-500 uppercase tracking-wider mb-4">
                Cache Entries ({state?.entries?.length || 0})
              </h3>
              <div className="max-h-64 overflow-y-auto space-y-1">
                {(state?.entries || []).length === 0 && (
                  <p className="text-slate-400 text-sm text-center py-4">Cache is empty</p>
                )}
                {(state?.entries || []).map((entry) => (
                  <EntryRow key={entry.key} entry={entry} onDelete={handleDelete} />
                ))}
              </div>
            </div>

            {/* Operations */}
            <div className="grid grid-cols-3 gap-4">
              <StatCard label="PUTs" value={metrics.puts.toLocaleString()} color="text-blue-600" />
              <StatCard label="GETs" value={metrics.gets.toLocaleString()} color="text-purple-600" />
              <StatCard label="DELETEs" value={metrics.deletes.toLocaleString()} color="text-orange-500" />
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string; color: string }) {
  return (
    <div className="bg-slate-50/80 border border-slate-200/50 rounded-xl p-4 shadow-sm">
      <p className="text-xs text-slate-500 mb-1">{label}</p>
      <p className={`text-lg font-bold font-mono ${color}`}>{value}</p>
    </div>
  );
}

function EntryRow({ entry, onDelete }: { entry: CacheEntry; onDelete: (key: string) => void }) {
  return (
    <div className="flex items-center justify-between bg-slate-100/60 border border-slate-200/40 rounded-lg px-3 py-2 group">
      <div className="flex items-center gap-3 min-w-0">
        <span className="text-xs text-slate-400 font-mono w-8">S{entry.shard}</span>
        <span className="text-sm font-mono text-cyan-600 truncate">{entry.key}</span>
        <span className="text-slate-300">→</span>
        <span className="text-sm font-mono text-slate-600 truncate">{entry.value}</span>
      </div>
      <button
        onClick={() => onDelete(entry.key)}
        className="text-slate-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-all ml-2"
      >
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  );
}
