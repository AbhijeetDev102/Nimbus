import React from "react";
import { Server, Cpu, Activity, Clock, ExternalLink, Filter } from "lucide-react";
import type { WorkerInfo } from "../../types";

interface WorkerRosterProps {
  workers: WorkerInfo[];
  activeCount: number;
  loading?: boolean;
  onSelectJob?: (jobId: string) => void;
  onFilterByWorker?: (workerId: string) => void;
  activeWorkerFilter?: string | null;
}

export const WorkerRoster: React.FC<WorkerRosterProps> = ({
  workers,
  activeCount,
  loading,
  onSelectJob,
  onFilterByWorker,
  activeWorkerFilter,
}) => {
  const formatTimeSince = (isoString: string) => {
    try {
      const ms = Date.now() - new Date(isoString).getTime();
      if (ms < 3000) return "just now";
      if (ms < 60000) return `${Math.floor(ms / 1000)}s ago`;
      return `${Math.floor(ms / 60000)}m ago`;
    } catch {
      return "—";
    }
  };

  return (
    <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-xl space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div className="flex items-center space-x-2.5">
          <div className="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            <Server className="w-4 h-4" />
          </div>
          <div>
            <h3 className="text-sm font-bold text-white tracking-tight flex items-center space-x-2">
              <span>Worker Fleet</span>
              <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                {activeCount} {activeCount === 1 ? "node" : "nodes"} active
              </span>
            </h3>
            <p className="text-[11px] text-slate-400">
              Distributed execution engines with TTL-based leases
            </p>
          </div>
        </div>

        {activeWorkerFilter && onFilterByWorker && (
          <button
            onClick={() => onFilterByWorker("")}
            className="text-[11px] font-mono px-2 py-1 rounded-lg bg-slate-800 text-slate-300 hover:text-white border border-white/10 transition-colors"
          >
            Clear Filter ✕
          </button>
        )}
      </div>

      {/* Workers List */}
      <div className="space-y-2.5">
        {loading && workers.length === 0 ? (
          <div className="py-6 text-center text-xs text-slate-500 font-mono">
            Scanning cluster for worker heartbeats...
          </div>
        ) : workers.length === 0 ? (
          <div className="py-6 px-4 rounded-xl bg-slate-950/60 border border-dashed border-white/10 text-center space-y-2">
            <Cpu className="w-6 h-6 text-slate-600 mx-auto" />
            <p className="text-xs text-slate-400">No worker replicas detected</p>
            <p className="text-[11px] text-slate-500 font-mono">
              Start one with: <code className="text-cyan-400">docker compose up --scale worker=2</code>
            </p>
          </div>
        ) : (
          workers.map((worker) => {
            const isFilterActive = activeWorkerFilter === worker.workerId;
            const shortId = worker.workerId.slice(0, 8);
            const isBusy = worker.status === "BUSY";
            const isOffline = worker.status === "OFFLINE";

            return (
              <div
                key={worker.workerId}
                className={`p-3 rounded-xl border transition-all ${
                  isFilterActive
                    ? "bg-cyan-950/40 border-cyan-500/50"
                    : isOffline
                    ? "bg-slate-950/40 border-rose-500/20 opacity-70"
                    : "bg-slate-950/80 hover:bg-slate-900/90 border-white/5 hover:border-white/15"
                }`}
              >
                <div className="flex items-center justify-between gap-2">
                  {/* Left: Worker Ident */}
                  <div className="flex items-center space-x-2.5 min-w-0">
                    <span className="relative flex h-2.5 w-2.5 shrink-0">
                      {isBusy ? (
                        <>
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                          <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500" />
                        </>
                      ) : isOffline ? (
                        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-rose-500" />
                      ) : (
                        <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-cyan-400" />
                      )}
                    </span>

                    <div className="truncate">
                      <div className="flex items-center space-x-2">
                        <span className="text-xs font-mono font-semibold text-slate-200">
                          {worker.hostname || "worker"}
                        </span>
                        <span className="text-[10px] font-mono text-slate-500">
                          ({shortId})
                        </span>
                      </div>

                      <div className="flex items-center space-x-2 text-[10px] font-mono text-slate-400">
                        <span className="flex items-center space-x-1">
                          <Clock className="w-2.5 h-2.5 text-slate-500" />
                          <span>{formatTimeSince(worker.lastHeartbeat)}</span>
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Right: Status / Actions */}
                  <div className="flex items-center space-x-2 shrink-0">
                    {isBusy ? (
                      <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                        <Activity className="w-2.5 h-2.5 animate-pulse" />
                        <span>BUSY</span>
                      </span>
                    ) : isOffline ? (
                      <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-rose-500/10 text-rose-400 border border-rose-500/30">
                        <span>OFFLINE</span>
                      </span>
                    ) : (
                      <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                        <span>IDLE</span>
                      </span>
                    )}

                    {onFilterByWorker && (
                      <button
                        onClick={() =>
                          onFilterByWorker(isFilterActive ? "" : worker.workerId)
                        }
                        title={isFilterActive ? "Remove filter" : "Filter jobs by this worker"}
                        className={`p-1.5 rounded-lg border transition-colors ${
                          isFilterActive
                            ? "bg-cyan-500/20 text-cyan-300 border-cyan-500/40"
                            : "bg-slate-900 border-white/5 text-slate-400 hover:text-white"
                        }`}
                      >
                        <Filter className="w-3 h-3" />
                      </button>
                    )}
                  </div>
                </div>

                {/* Current Job processing banner if BUSY */}
                {isBusy && worker.currentJobId && (
                  <div className="mt-2 pt-2 border-t border-white/5 flex items-center justify-between text-[11px] font-mono">
                    <span className="text-slate-400 text-[10px]">Processing:</span>
                    {onSelectJob ? (
                      <button
                        onClick={() => onSelectJob(worker.currentJobId!)}
                        className="inline-flex items-center space-x-1 text-cyan-400 hover:text-cyan-300 hover:underline"
                      >
                        <span>{worker.currentJobId.slice(0, 8)}...</span>
                        <ExternalLink className="w-2.5 h-2.5" />
                      </button>
                    ) : (
                      <span className="text-cyan-400">{worker.currentJobId.slice(0, 8)}...</span>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
