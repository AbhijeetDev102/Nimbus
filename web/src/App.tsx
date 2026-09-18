import { useState, useEffect, useCallback, useRef } from "react";
import { Navbar } from "./components/layout/Navbar";
import { ClusterStats } from "./components/overview/ClusterStats";
import { WorkloadStudio } from "./components/studio/WorkloadStudio";
import { ExecutionsTable } from "./components/executions/ExecutionsTable";
import { JobInspectorModal } from "./components/executions/JobInspectorModal";
import { WorkerRoster } from "./components/overview/WorkerRoster";
import { ActivityFeed, type ActivityEvent } from "./components/overview/ActivityFeed";
import type { JobRecord, JobStats, WorkerInfo } from "./types";
import { fetchJobStats, fetchJobs, fetchJobById, fetchWorkers } from "./services/api";
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

  const [workers, setWorkers] = useState<WorkerInfo[]>([]);
  const [activeWorkerCount, setActiveWorkerCount] = useState(0);
  const [workerFilter, setWorkerFilter] = useState<string | null>(null);
  const [activityEvents, setActivityEvents] = useState<ActivityEvent[]>([]);

  const [inspectedJob, setInspectedJob] = useState<JobRecord | null>(null);
  const [loadingStats, setLoadingStats] = useState(false);
  const [loadingJobs, setLoadingJobs] = useState(false);
  const [loadingWorkers, setLoadingWorkers] = useState(false);

  const prevJobsMapRef = useRef<Map<string, JobRecord>>(new Map());

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

  // 2. Fetch Workers
  const loadWorkers = useCallback(async () => {
    try {
      setLoadingWorkers(true);
      const data = await fetchWorkers();
      setWorkers(data.workers || []);
      setActiveWorkerCount(data.activeCount || 0);
    } catch (e) {
      console.error("Failed to load workers:", e);
    } finally {
      setLoadingWorkers(false);
    }
  }, []);

  // 3. Fetch Jobs & synthesize activity events
  const loadJobs = useCallback(async (silent = false) => {
    try {
      if (!silent) setLoadingJobs(true);
      const data = await fetchJobs({
        limit,
        offset,
        status: statusFilter,
        jobType: jobTypeFilter,
      });

      const newJobs = data.jobs || [];
      setJobs(newJobs);
      setTotalCount(data.totalCount || 0);

      // Event synthesis
      const prevMap = prevJobsMapRef.current;
      const detectedEvents: ActivityEvent[] = [];

      if (prevMap.size === 0 && newJobs.length > 0) {
        // Initial populate (newest 5)
        newJobs.slice(0, 5).forEach((j) => {
          detectedEvents.push({
            id: `init-${j.jobId}-${Date.now()}`,
            type:
              j.status === "COMPLETED"
                ? "COMPLETED"
                : j.status === "RUNNING"
                  ? "CLAIMED"
                  : j.status === "FAILED"
                    ? "FAILED"
                    : "DISPATCHED",
            jobId: j.jobId,
            jobType: j.jobType,
            workerId: j.workerId,
            timestamp: new Date(j.createdAt),
          });
        });
      } else {
        newJobs.forEach((j) => {
          const prev = prevMap.get(j.jobId);
          if (!prev) {
            // Brand new job
            detectedEvents.unshift({
              id: `disp-${j.jobId}-${Date.now()}`,
              type: "DISPATCHED",
              jobId: j.jobId,
              jobType: j.jobType,
              timestamp: new Date(),
            });
          } else {
            // Status transitions
            if (prev.status !== "RUNNING" && j.status === "RUNNING") {
              detectedEvents.unshift({
                id: `claim-${j.jobId}-${Date.now()}`,
                type: "CLAIMED",
                jobId: j.jobId,
                jobType: j.jobType,
                workerId: j.workerId,
                timestamp: new Date(),
              });
            } else if (prev.status !== "COMPLETED" && j.status === "COMPLETED") {
              detectedEvents.unshift({
                id: `comp-${j.jobId}-${Date.now()}`,
                type: "COMPLETED",
                jobId: j.jobId,
                jobType: j.jobType,
                workerId: j.workerId,
                timestamp: new Date(),
              });
            } else if (prev.status !== "FAILED" && j.status === "FAILED") {
              detectedEvents.unshift({
                id: `fail-${j.jobId}-${Date.now()}`,
                type: "FAILED",
                jobId: j.jobId,
                jobType: j.jobType,
                workerId: j.workerId,
                timestamp: new Date(),
              });
            } else if (j.retryCount > prev.retryCount) {
              detectedEvents.unshift({
                id: `retry-${j.jobId}-${Date.now()}`,
                type: "RETRY",
                jobId: j.jobId,
                jobType: j.jobType,
                workerId: j.workerId,
                timestamp: new Date(),
                details: `attempt ${j.retryCount + 1}/${j.maxRetries}`,
              });
            }
          }
        });
      }

      if (detectedEvents.length > 0) {
        setActivityEvents((prev) => [...detectedEvents, ...prev].slice(0, 35));
      }

      // Update map
      const nextMap = new Map<string, JobRecord>();
      newJobs.forEach((j) => nextMap.set(j.jobId, j));
      prevJobsMapRef.current = nextMap;
    } catch (e) {
      console.error("Failed to load jobs:", e);
    } finally {
      if (!silent) setLoadingJobs(false);
    }
  }, [limit, offset, statusFilter, jobTypeFilter]);

  // Initial Load & Auto-polling
  useEffect(() => {
    loadStats();
    loadJobs();
    loadWorkers();

    const interval = setInterval(() => {
      loadStats();
      loadWorkers();
      loadJobs(true);
    }, 3000);

    return () => clearInterval(interval);
  }, [loadStats, loadJobs, loadWorkers]);

  const handleJobDispatched = async (_jobId: string) => {
    loadStats();
    loadJobs();
    loadWorkers();
    // Do not automatically pop up the modal; user can click any job to inspect
  };

  const handleRefreshAll = () => {
    loadStats();
    loadJobs();
    loadWorkers();
  };

  const displayedJobs = workerFilter
    ? jobs.filter((j) => j.workerId === workerFilter)
    : jobs;

  return (
    <div className="min-h-screen bg-[#06090e] text-slate-100 flex flex-col selection:bg-cyan-500 selection:text-black">
      {/* Top Navigation */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onRefresh={handleRefreshAll}
        isRefreshing={loadingStats || loadingJobs || loadingWorkers}
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
              {/* Workload Studio + Activity Feed on Left */}
              <div className="lg:col-span-7 space-y-6">
                <WorkloadStudio onJobDispatched={handleJobDispatched} />
                <ActivityFeed
                  events={activityEvents}
                  onSelectJob={(j) => setInspectedJob(j)}
                  allJobs={jobs}
                />
              </div>

              {/* Live Cluster Workers + Architecture on Right */}
              <div className="lg:col-span-5 space-y-4">
                <WorkerRoster
                  workers={workers}
                  activeCount={activeWorkerCount}
                  loading={loadingWorkers}
                  activeWorkerFilter={workerFilter}
                  onFilterByWorker={(wId) => setWorkerFilter(wId || null)}
                  onSelectJob={async (jId) => {
                    const found = jobs.find((j) => j.jobId === jId);
                    if (found) {
                      setInspectedJob(found);
                    } else {
                      try {
                        const fetched = await fetchJobById(jId);
                        setInspectedJob(fetched);
                      } catch (e) {
                        console.error(e);
                      }
                    }
                  }}
                />

                <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-xl space-y-3">
                  <div className="flex items-center space-x-2 text-cyan-400 font-mono text-xs uppercase tracking-wider font-semibold">
                    <Activity className="w-4 h-4" />
                    <span>Resilience Architecture</span>
                  </div>
                  <p className="text-xs text-slate-400 leading-relaxed">
                    Transactional Outbox pattern with Debezium CDC and Kafka guarantees zero job loss. Distributed conditional leases prevent split-brain execution across worker crashes.
                  </p>

                  <div className="pt-2 border-t border-white/10 grid grid-cols-2 gap-2 text-xs font-mono">
                    <div className="bg-slate-950 p-2 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Storage</span>
                      <span className="text-cyan-300 font-semibold">MinIO / S3</span>
                    </div>
                    <div className="bg-slate-950 p-2 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Stream</span>
                      <span className="text-purple-300 font-semibold">Kafka + Franz-go</span>
                    </div>
                    <div className="bg-slate-950 p-2 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Registry</span>
                      <span className="text-emerald-300 font-semibold">Redis Heartbeats</span>
                    </div>
                    <div className="bg-slate-950 p-2 rounded-xl border border-white/5">
                      <span className="text-slate-500 block text-[10px]">Lease Recovery</span>
                      <span className="text-amber-300 font-semibold">Job Reaper Daemon</span>
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
                <div className="flex items-center space-x-3">
                  <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400 font-mono">
                    Recent Executions
                  </h3>
                  {workerFilter && (
                    <span className="text-xs font-mono px-2 py-0.5 rounded-md bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                      Filtered by worker-{workerFilter.slice(0, 8)}
                    </span>
                  )}
                </div>
                <button
                  onClick={() => setActiveTab("executions")}
                  className="text-xs text-cyan-400 hover:underline font-mono cursor-pointer"
                >
                  View full table →
                </button>
              </div>
              <ExecutionsTable
                jobs={displayedJobs.slice(0, 5)}
                totalCount={workerFilter ? displayedJobs.length : totalCount}
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
              {workerFilter && (
                <button
                  onClick={() => setWorkerFilter(null)}
                  className="text-xs font-mono px-3 py-1.5 rounded-xl bg-slate-800 text-slate-300 hover:text-white border border-white/10 transition-colors"
                >
                  Clear Worker Filter ✕
                </button>
              )}
            </div>

            <ExecutionsTable
              jobs={displayedJobs}
              totalCount={workerFilter ? displayedJobs.length : totalCount}
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
            setInspectedJob((prev) => (prev ? updated : null));
            loadStats();
            loadJobs();
            loadWorkers();
          }}
        />
      )}
    </div>
  );
}

export default App;
