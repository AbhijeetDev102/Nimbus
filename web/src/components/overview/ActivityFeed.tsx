import React from "react";
import { Zap, CheckCircle2, XCircle, RefreshCw, Play, Clock, Sparkles } from "lucide-react";
import type { JobRecord } from "../../types";

export interface ActivityEvent {
  id: string;
  type: "DISPATCHED" | "CLAIMED" | "COMPLETED" | "FAILED" | "RETRY";
  jobId: string;
  jobType: string;
  workerId?: string;
  timestamp: Date;
  details?: string;
}

interface ActivityFeedProps {
  events: ActivityEvent[];
  onSelectJob?: (job: JobRecord) => void;
  allJobs: JobRecord[];
}

export const ActivityFeed: React.FC<ActivityFeedProps> = ({
  events,
  onSelectJob,
  allJobs,
}) => {
  const getEventIcon = (type: ActivityEvent["type"]) => {
    switch (type) {
      case "DISPATCHED":
        return <Zap className="w-3.5 h-3.5 text-cyan-400" />;
      case "CLAIMED":
        return <Play className="w-3.5 h-3.5 text-emerald-400" />;
      case "COMPLETED":
        return <CheckCircle2 className="w-3.5 h-3.5 text-indigo-400" />;
      case "FAILED":
        return <XCircle className="w-3.5 h-3.5 text-rose-400" />;
      case "RETRY":
        return <RefreshCw className="w-3.5 h-3.5 text-amber-400 animate-spin" />;
    }
  };

  const getEventBadge = (type: ActivityEvent["type"]) => {
    switch (type) {
      case "DISPATCHED":
        return "bg-cyan-500/10 text-cyan-400 border-cyan-500/30";
      case "CLAIMED":
        return "bg-emerald-500/10 text-emerald-400 border-emerald-500/30";
      case "COMPLETED":
        return "bg-indigo-500/10 text-indigo-400 border-indigo-500/30";
      case "FAILED":
        return "bg-rose-500/10 text-rose-400 border-rose-500/30";
      case "RETRY":
        return "bg-amber-500/10 text-amber-400 border-amber-500/30";
    }
  };

  const handleJobClick = (jobId: string) => {
    if (!onSelectJob) return;
    const found = allJobs.find((j) => j.jobId === jobId);
    if (found) {
      onSelectJob(found);
    }
  };

  return (
    <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-5 border border-white/10 shadow-xl space-y-3">
      <div className="flex items-center justify-between border-b border-white/10 pb-3">
        <div className="flex items-center space-x-2">
          <div className="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <Sparkles className="w-3.5 h-3.5" />
          </div>
          <h3 className="text-xs font-bold text-white uppercase tracking-wider font-mono">
            Cluster Activity Feed
          </h3>
        </div>
        <div className="flex items-center space-x-1.5">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          <span className="text-[10px] font-mono text-emerald-400 uppercase font-semibold">
            Live Stream
          </span>
        </div>
      </div>

      <div className="space-y-2 max-h-[300px] overflow-y-auto pr-1">
        {events.length === 0 ? (
          <div className="py-8 text-center text-xs text-slate-500 font-mono">
            Waiting for cluster events... Dispatch a job above to watch distributed execution.
          </div>
        ) : (
          events.slice(0, 25).map((evt) => {
            const shortJob = evt.jobId.slice(0, 8);
            const shortWorker = evt.workerId ? evt.workerId.slice(0, 8) : null;

            return (
              <div
                key={evt.id}
                onClick={() => handleJobClick(evt.jobId)}
                className="p-2.5 rounded-xl bg-slate-950/70 hover:bg-slate-900/80 border border-white/5 transition-all text-xs flex items-center justify-between gap-2 cursor-pointer group"
              >
                <div className="flex items-center space-x-2.5 min-w-0">
                  <div className="shrink-0">{getEventIcon(evt.type)}</div>

                  <div className="truncate font-mono">
                    <span
                      className={`inline-block px-1.5 py-0.2 rounded text-[9px] uppercase font-bold border mr-1.5 ${getEventBadge(
                        evt.type
                      )}`}
                    >
                      {evt.type}
                    </span>

                    <span className="text-cyan-300 group-hover:underline">
                      #{shortJob}
                    </span>

                    <span className="text-slate-500 mx-1.5">•</span>
                    <span className="text-slate-400 text-[11px]">{evt.jobType}</span>

                    {shortWorker && (
                      <span className="text-emerald-400/90 text-[10px] ml-1.5">
                        by worker-{shortWorker}
                      </span>
                    )}

                    {evt.details && (
                      <span className="text-slate-500 text-[10px] ml-1.5">
                        ({evt.details})
                      </span>
                    )}
                  </div>
                </div>

                <div className="shrink-0 text-[10px] font-mono text-slate-500 flex items-center space-x-1">
                  <Clock className="w-2.5 h-2.5" />
                  <span>
                    {evt.timestamp.toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                      second: "2-digit",
                    })}
                  </span>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
