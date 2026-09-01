import React from "react";
import {
  Search,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  Film,
  Calculator,
  RefreshCw,
} from "lucide-react";
import type { JobRecord, JobStatus } from "../../types";

interface ExecutionsTableProps {
  jobs: JobRecord[];
  totalCount: number;
  limit: number;
  offset: number;
  statusFilter: string;
  setStatusFilter: (status: string) => void;
  jobTypeFilter: string;
  setJobTypeFilter: (type: string) => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  onPageChange: (newOffset: number) => void;
  onSelectJob: (job: JobRecord) => void;
  onRefresh: () => void;
  loading?: boolean;
}

export const ExecutionsTable: React.FC<ExecutionsTableProps> = ({
  jobs,
  totalCount,
  limit,
  offset,
  statusFilter,
  setStatusFilter,
  searchQuery,
  setSearchQuery,
  onPageChange,
  onSelectJob,
  onRefresh,
  loading,
}) => {
  const currentPage = Math.floor(offset / limit) + 1;
  const totalPages = Math.ceil(totalCount / limit) || 1;

  const filteredJobs = jobs.filter((j) => {
    if (!searchQuery) return true;
    return (
      j.jobId.toLowerCase().includes(searchQuery.toLowerCase()) ||
      (j.resourceId && j.resourceId.toLowerCase().includes(searchQuery.toLowerCase()))
    );
  });

  const getStatusBadge = (status: JobStatus, retryCount: number, maxRetries: number) => {
    switch (status) {
      case "RUNNING":
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
            <span>RUNNING</span>
          </span>
        );
      case "QUEUED":
        return (
          <span className="inline-flex items-center space-x-1.5 px-2.5 py-1 rounded-full text-xs font-medium font-mono bg-amber-500/10 text-amber-300 border border-amber-500/30">
            <span className="w-1.5 h-1.5 rounded-full bg-amber-400" />
            <span>{retryCount > 0 ? `RETRYING (${retryCount}/${maxRetries})` : "QUEUED"}</span>
          </span>
        );
      case "COMPLETED":
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-medium font-mono bg-indigo-500/10 text-indigo-300 border border-indigo-500/30">
            <span>✓ COMPLETED</span>
          </span>
        );
      case "FAILED":
        return (
          <span className="inline-flex items-center space-x-1 px-2.5 py-1 rounded-full text-xs font-medium font-mono bg-rose-500/10 text-rose-400 border border-rose-500/30">
            <span>✕ FAILED</span>
          </span>
        );
      default:
        return <span className="text-xs text-slate-400 font-mono">{status}</span>;
    }
  };

  const calculateDuration = (job: JobRecord) => {
    if (!job.startedAt) return "—";
    const start = new Date(job.startedAt).getTime();
    const end = job.completedAt ? new Date(job.completedAt).getTime() : Date.now();
    const diffMs = end - start;

    if (diffMs < 1000) return `${diffMs}ms`;
    if (diffMs < 60000) return `${(diffMs / 1000).toFixed(1)}s`;
    return `${Math.floor(diffMs / 60000)}m ${Math.floor((diffMs % 60000) / 1000)}s`;
  };

  return (
    <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 border border-white/10 shadow-2xl overflow-hidden">
      {/* Control / Filter Bar */}
      <div className="p-4 sm:p-5 border-b border-white/10 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        {/* Search */}
        <div className="relative flex-1 max-w-md">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search by Job ID..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-950 border border-white/10 rounded-xl pl-9 pr-4 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 font-mono"
          />
        </div>

        {/* Status Filters */}
        <div className="flex items-center space-x-2 overflow-x-auto pb-1 sm:pb-0">
          <div className="flex bg-slate-950 p-1 rounded-xl border border-white/10 text-xs">
            {["ALL", "RUNNING", "QUEUED", "COMPLETED", "FAILED"].map((s) => (
              <button
                key={s}
                onClick={() => setStatusFilter(s)}
                className={`px-2.5 py-1 rounded-lg transition-all ${
                  statusFilter === s
                    ? "bg-cyan-500/20 text-cyan-300 font-semibold border border-cyan-500/30"
                    : "text-slate-400 hover:text-white"
                }`}
              >
                {s}
              </button>
            ))}
          </div>

          <button
            onClick={onRefresh}
            className="p-2 rounded-xl bg-slate-950 border border-white/10 text-slate-400 hover:text-white transition-colors cursor-pointer"
            title="Refresh Table"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin text-cyan-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-white/10 bg-slate-950/60 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
              <th className="py-3 px-4">Job ID</th>
              <th className="py-3 px-4">Workload</th>
              <th className="py-3 px-4">Status</th>
              <th className="py-3 px-4">Attempts</th>
              <th className="py-3 px-4">Duration</th>
              <th className="py-3 px-4">Created</th>
              <th className="py-3 px-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5 text-xs text-slate-300">
            {loading && jobs.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-12 text-center text-slate-500">
                  <RefreshCw className="w-6 h-6 animate-spin mx-auto text-cyan-400 mb-2" />
                  <span>Loading cluster executions...</span>
                </td>
              </tr>
            ) : filteredJobs.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-12 text-center text-slate-500">
                  No jobs found matching the selected filters.
                </td>
              </tr>
            ) : (
              filteredJobs.map((job) => (
                <tr
                  key={job.jobId}
                  onClick={() => onSelectJob(job)}
                  className="hover:bg-slate-800/40 transition-colors cursor-pointer group"
                >
                  <td className="py-3.5 px-4 font-mono text-cyan-300 group-hover:text-cyan-200">
                    {job.jobId.slice(0, 8)}...{job.jobId.slice(-4)}
                  </td>
                  <td className="py-3.5 px-4">
                    <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-slate-800 border border-white/10 text-[11px] font-mono">
                      {job.jobType.includes("VIDEO") ? (
                        <Film className="w-3 h-3 text-cyan-400" />
                      ) : (
                        <Calculator className="w-3 h-3 text-emerald-400" />
                      )}
                      <span>{job.jobType}</span>
                    </span>
                  </td>
                  <td className="py-3.5 px-4">{getStatusBadge(job.status, job.retryCount, job.maxRetries)}</td>
                  <td className="py-3.5 px-4 font-mono text-slate-400">
                    {job.retryCount + 1} / {job.maxRetries}
                  </td>
                  <td className="py-3.5 px-4 font-mono text-slate-400">{calculateDuration(job)}</td>
                  <td className="py-3.5 px-4 font-mono text-slate-400">
                    {new Date(job.createdAt).toLocaleTimeString()}
                  </td>
                  <td className="py-3.5 px-4 text-right">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        onSelectJob(job);
                      }}
                      className="p-1.5 rounded-lg bg-slate-900 border border-white/10 text-slate-400 group-hover:text-cyan-300 group-hover:border-cyan-500/30 transition-colors inline-flex items-center space-x-1"
                    >
                      <ExternalLink className="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      <div className="p-4 border-t border-white/10 flex items-center justify-between text-xs text-slate-400">
        <div>
          Showing <span className="font-mono text-white">{offset + 1}</span> to{" "}
          <span className="font-mono text-white">{Math.min(offset + limit, totalCount)}</span> of{" "}
          <span className="font-mono text-white">{totalCount}</span> jobs
        </div>

        <div className="flex items-center space-x-2">
          <button
            onClick={() => onPageChange(Math.max(0, offset - limit))}
            disabled={offset === 0}
            className="p-1.5 rounded-lg bg-slate-950 border border-white/10 disabled:opacity-30 disabled:cursor-not-allowed hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="font-mono px-2">
            {currentPage} / {totalPages}
          </span>
          <button
            onClick={() => onPageChange(offset + limit)}
            disabled={offset + limit >= totalCount}
            className="p-1.5 rounded-lg bg-slate-950 border border-white/10 disabled:opacity-30 disabled:cursor-not-allowed hover:bg-slate-800 transition-colors cursor-pointer"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
};
