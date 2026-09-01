import React from "react";
import { Layers, Activity, Clock, CheckCircle2, AlertTriangle, TrendingUp } from "lucide-react";
import type { JobStats } from "../../types";

interface ClusterStatsProps {
  stats: JobStats | null;
  loading?: boolean;
}

export const ClusterStats: React.FC<ClusterStatsProps> = ({ stats, loading }) => {
  const total = stats?.total || 0;
  const running = stats?.running || 0;
  const queued = stats?.queued || 0;
  const completed = stats?.completed || 0;
  const failed = stats?.failed || 0;

  const finishedCount = completed + failed;
  const successRate = finishedCount > 0 ? Math.round((completed / finishedCount) * 100) : 100;

  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4">
      {/* Total Jobs */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-lg hover:border-cyan-500/40 transition-all group">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-slate-400">Total Jobs</span>
          <div className="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            <Layers className="w-4 h-4" />
          </div>
        </div>
        <div className="mt-3 flex items-baseline justify-between">
          <span className="text-3xl font-bold font-mono text-white tracking-tight">
            {loading ? "..." : total.toLocaleString()}
          </span>
          <span className="text-xs text-slate-500 font-mono">All-time</span>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-cyan-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>

      {/* Running */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-lg hover:border-emerald-500/40 transition-all group">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-slate-400">Active Executing</span>
          <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <Activity className={`w-4 h-4 ${running > 0 ? "animate-pulse text-emerald-300" : ""}`} />
          </div>
        </div>
        <div className="mt-3 flex items-baseline justify-between">
          <span className="text-3xl font-bold font-mono text-emerald-400 tracking-tight">
            {loading ? "..." : running.toLocaleString()}
          </span>
          {running > 0 && (
            <span className="flex items-center text-xs text-emerald-400 font-mono">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 mr-1 animate-ping" />
              Live
            </span>
          )}
        </div>
        <div className="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-emerald-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>

      {/* Queued */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-lg hover:border-amber-500/40 transition-all group">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-slate-400">In Queue</span>
          <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Clock className="w-4 h-4" />
          </div>
        </div>
        <div className="mt-3 flex items-baseline justify-between">
          <span className="text-3xl font-bold font-mono text-amber-300 tracking-tight">
            {loading ? "..." : queued.toLocaleString()}
          </span>
          <span className="text-xs text-slate-500 font-mono">Buffered</span>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-amber-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>

      {/* Completed */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-lg hover:border-indigo-500/40 transition-all group">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-slate-400">Completed</span>
          <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <CheckCircle2 className="w-4 h-4" />
          </div>
        </div>
        <div className="mt-3 flex items-baseline justify-between">
          <span className="text-3xl font-bold font-mono text-indigo-300 tracking-tight">
            {loading ? "..." : completed.toLocaleString()}
          </span>
          <span className="flex items-center text-xs text-indigo-400 font-mono">
            <TrendingUp className="w-3 h-3 mr-0.5" />
            {successRate}% rate
          </span>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-indigo-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>

      {/* Failed */}
      <div className="col-span-2 sm:col-span-1 relative overflow-hidden rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-lg hover:border-rose-500/40 transition-all group">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium uppercase tracking-wider text-slate-400">Failed / Retried</span>
          <div className="p-2 rounded-xl bg-rose-500/10 text-rose-400 border border-rose-500/20">
            <AlertTriangle className="w-4 h-4" />
          </div>
        </div>
        <div className="mt-3 flex items-baseline justify-between">
          <span className="text-3xl font-bold font-mono text-rose-400 tracking-tight">
            {loading ? "..." : failed.toLocaleString()}
          </span>
          <span className="text-xs text-rose-400/70 font-mono">Exhausted</span>
        </div>
        <div className="absolute inset-x-0 bottom-0 h-1 bg-gradient-to-r from-transparent via-rose-500/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>
    </div>
  );
};
