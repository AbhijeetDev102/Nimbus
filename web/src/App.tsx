import { useState, useEffect, useCallback } from "react";
import { Navbar } from "./components/layout/Navbar";
import { ClusterStats } from "./components/overview/ClusterStats";
import { WorkloadStudio } from "./components/studio/WorkloadStudio";
import { ExecutionsTable } from "./components/executions/ExecutionsTable";
import { JobInspectorModal } from "./components/executions/JobInspectorModal";
import type { JobRecord, JobStats } from "./types";
import { fetchJobStats, fetchJobs, fetchJobById } from "./services/api";
import { Activity, Terminal, ArrowRight } from "lucide-react";

export function App() {
  const [activeTab, setActiveTab] = useState<"overview" | "studio" | "executions">("overview");
  const [stats, setStats] = useState<JobStats | null>(null);
  const [jobs, setJobs] = useState<JobRecord[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [limit] = useState(20);
  const [offset, setOffset] = useState(0);
  const [statusFilter, setStatusFilter] = useState("ALL");
  const [jobTypeFilter, setJobTypeFilter] = useState("ALL");
  const [searchQuery, setSearchQuery] = useState("");

  const [inspectedJob, setInspectedJob] = useState<JobRecord | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);
  const [loadingJobs, setLoadingJobs] = useState(false);

  // 1. Fetch Stats
  const loadStats = useCallback(async () => {
    try {
      setLoadingStats(true);
      const data = await fetchJobStats();
      setStats(data);
    } catch (e) {
      console.error("Failed to load stats:", e);
    } finally {
      setLoadingStats(false);
    }
  }, []);

  // 2. Fetch Jobs
  const loadJobs = useCallback(async () => {
    try {
      setLoadingJobs(true);
      const data = await fetchJobs({
        limit,
        offset,
        status: statusFilter,
        jobType: jobTypeFilter,
      });
      setJobs(data.jobs || []);
      setTotalCount(data.totalCount || 0);
    } catch (e) {
      console.error("Failed to load jobs:", e);
    } finally {
      setLoadingJobs(false);
    }
  }, [limit, offset, statusFilter, jobTypeFilter]);

  // Initial Load & Auto-polling
  useEffect(() => {
    loadStats();
    loadJobs();

    const interval = setInterval(() => {
      loadStats();
    }, 4000);

    return () => clearInterval(interval);
  }, [loadStats, loadJobs]);

  const handleJobDispatched = async (jobId: string) => {
    loadStats();
    loadJobs();
    try {
      const newJob = await fetchJobById(jobId);
      setInspectedJob(newJob);
    } catch (e) {
      console.error("Failed to open dispatched job:", e);
    }
  };

  const handleRefreshAll = () => {
    loadStats();
    loadJobs();
  };

  return (
    <div className="min-h-screen bg-[#06090e] text-slate-100 flex flex-col selection:bg-cyan-500 selection:text-black">
      {/* Top Navigation */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onRefresh={handleRefreshAll}
        isRefreshing={loadingStats || loadingJobs}
      />

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Top Metric Cards (Always visible on Overview & Executions) */}
        <ClusterStats stats={stats} loading={loadingStats} />

        {/* Tab Views */}
        {activeTab === "overview" && (
          <div className="space-y-8">
            {/* Quick Dispatch Banner & Studio Shortcut */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
              {/* Workload Studio on Left */}
              <div className="lg:col-span-7">
                <WorkloadStudio onJobDispatched={handleJobDispatched} />
              </div>

              {/* Live Cluster Health / Info on Right */}
              <div className="lg:col-span-5 space-y-4">
                <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-6 border border-white/10 shadow-xl space-y-4">
                  <div className="flex items-center space-x-2 text-cyan-400 font-mono text-xs uppercase tracking-wider font-semibold">
                    <Activity className="w-4 h-4" />
                    <span>Engine Architecture</span>
                  </div>
                  <h3 className="text-base font-bold text-white tracking-tight">
                    Cloud-Native Job Execution
                  </h3>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    Nimbus decouples ingestion from compute via Transactional Outbox, Debezium CDC,
                    Kafka partitioning, and distributed atomic conditional leases.
                  </p>

                  <div className="pt-2 border-t border-white/10 grid grid-cols-2 gap-3 text-xs font-mono">
                    <div className="bg-slate-950 p-2.5 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Storage Provider</span>
                      <span className="text-cyan-300 font-semibold">MinIO / S3</span>
                    </div>
                    <div className="bg-slate-950 p-2.5 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Message Bus</span>
                      <span className="text-purple-300 font-semibold">Kafka + Franz-go</span>
                    </div>
                    <div className="bg-slate-950 p-2.5 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Telemetry</span>
                      <span className="text-emerald-300 font-semibold">Redis Pub/Sub</span>
                    </div>
                    <div className="bg-slate-950 p-2.5 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Resilience</span>
                      <span className="text-amber-300 font-semibold">Outbox Retries</span>
                    </div>
                  </div>
                </div>

                {/* Quick link to Executions */}
                <button
                  onClick={() => setActiveTab("executions")}
                  className="w-full p-4 rounded-2xl bg-slate-900/60 hover:bg-slate-900 border border-white/10 hover:border-cyan-500/40 text-left transition-all flex items-center justify-between group cursor-pointer"
                >
                  <div className="flex items-center space-x-3">
                    <div className="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                      <Terminal className="w-4 h-4" />
                    </div>
                    <div>
                      <h4 className="text-xs font-bold text-white font-mono">View All Executions</h4>
                      <p className="text-[11px] text-slate-400 font-mono">
                        Inspect {totalCount} jobs across cluster
                      </p>
                    </div>
                  </div>
                  <ArrowRight className="w-4 h-4 text-slate-500 group-hover:text-cyan-400 group-hover:translate-x-1 transition-all" />
                </button>
              </div>
            </div>

            {/* Recent Executions Preview */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400 font-mono">
                  Recent Executions
                </h3>
                <button
                  onClick={() => setActiveTab("executions")}
                  className="text-xs text-cyan-400 hover:underline font-mono cursor-pointer"
                >
                  View full table →
                </button>
              </div>
              <ExecutionsTable
                jobs={jobs.slice(0, 5)}
                totalCount={totalCount}
                limit={limit}
                offset={offset}
                statusFilter={statusFilter}
                setStatusFilter={setStatusFilter}
                jobTypeFilter={jobTypeFilter}
                setJobTypeFilter={setJobTypeFilter}
                searchQuery={searchQuery}
                setSearchQuery={setSearchQuery}
                onPageChange={setOffset}
                onSelectJob={(job) => setInspectedJob(job)}
                onRefresh={loadJobs}
                loading={loadingJobs}
              />
            </div>
          </div>
        )}

        {activeTab === "studio" && (
          <div className="max-w-3xl mx-auto">
            <WorkloadStudio onJobDispatched={handleJobDispatched} />
          </div>
        )}

        {activeTab === "executions" && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-bold text-white tracking-tight">Executions Explorer</h2>
                <p className="text-xs text-slate-400">
                  Real-time audit log of all jobs processed by worker replicas
                </p>
              </div>
            </div>

            <ExecutionsTable
              jobs={jobs}
              totalCount={totalCount}
              limit={limit}
              offset={offset}
              statusFilter={statusFilter}
              setStatusFilter={setStatusFilter}
              jobTypeFilter={jobTypeFilter}
              setJobTypeFilter={setJobTypeFilter}
              searchQuery={searchQuery}
              setSearchQuery={setSearchQuery}
              onPageChange={setOffset}
              onSelectJob={(job) => setInspectedJob(job)}
              onRefresh={loadJobs}
              loading={loadingJobs}
            />
          </div>
        )}
      </main>

      {/* Hero Job Inspector Modal */}
      {inspectedJob && (
        <JobInspectorModal
          job={inspectedJob}
          onClose={() => setInspectedJob(null)}
          onJobUpdated={(updated) => {
            setInspectedJob(updated);
            loadStats();
            loadJobs();
          }}
        />
      )}
    </div>
  );
}

export default App;
