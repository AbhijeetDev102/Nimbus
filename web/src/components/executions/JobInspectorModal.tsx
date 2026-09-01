import React, { useEffect, useState } from "react";
import {
  X,
  Copy,
  Check,
  Film,
  Calculator,
  Download,
  AlertTriangle,
  Sparkles,
  Zap,
} from "lucide-react";
import type { JobRecord, ProgressUpdate } from "../../types";
import { connectJobProgressWS, getDownloadUrl, fetchJobById } from "../../services/api";

interface JobInspectorModalProps {
  job: JobRecord | null;
  onClose: () => void;
  onJobUpdated?: (updatedJob: JobRecord) => void;
}

export const JobInspectorModal: React.FC<JobInspectorModalProps> = ({
  job: initialJob,
  onClose,
  onJobUpdated,
}) => {
  const [job, setJob] = useState<JobRecord | null>(initialJob);
  const [liveProgress, setLiveProgress] = useState<ProgressUpdate | null>(null);
  const [downloadUrl, setDownloadUrl] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [loadingDownload, setLoadingDownload] = useState(false);

  useEffect(() => {
    setJob(initialJob);
    setLiveProgress(null);
    setDownloadUrl(null);
  }, [initialJob]);

  // 1. If job has an outputResourceId, fetch presigned download URL
  useEffect(() => {
    if (!job?.outputResourceId) return;

    let isMounted = true;
    setLoadingDownload(true);

    getDownloadUrl(job.outputResourceId)
      .then((res) => {
        if (isMounted) {
          setDownloadUrl(res.download_url);
          setLoadingDownload(false);
        }
      })
      .catch((err) => {
        console.error("Failed to fetch download url:", err);
        if (isMounted) setLoadingDownload(false);
      });

    return () => {
      isMounted = false;
    };
  }, [job?.outputResourceId]);

  // 2. If job is RUNNING or QUEUED, connect to WebSocket for live progress
  useEffect(() => {
    if (!job?.jobId || (job.status !== "RUNNING" && job.status !== "QUEUED")) return;

    const disconnectWS = connectJobProgressWS(
      job.jobId,
      (update) => {
        setLiveProgress(update);

        // If job completed over WS, refresh full record
        if (update.status === "COMPLETED" || update.progress >= 100) {
          fetchJobById(job.jobId).then((updated) => {
            setJob(updated);
            if (onJobUpdated) onJobUpdated(updated);
          });
        }
      },
      () => {
        // WS closed, do a final sync
        fetchJobById(job.jobId).then((updated) => {
          setJob(updated);
          if (onJobUpdated) onJobUpdated(updated);
        });
      }
    );

    return () => {
      disconnectWS();
    };
  }, [job?.jobId, job?.status]);

  if (!job) return null;

  const copyJobId = () => {
    navigator.clipboard.writeText(job.jobId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const progressValue =
    liveProgress?.progress !== undefined
      ? liveProgress.progress
      : job.status === "COMPLETED"
      ? 100
      : 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="relative w-full max-w-4xl max-h-[90vh] flex flex-col rounded-2xl bg-gradient-to-b from-slate-900 to-[#080d16] border border-white/10 shadow-2xl overflow-hidden">
        {/* Modal Header */}
        <div className="p-5 border-b border-white/10 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              {job.jobType.includes("VIDEO") ? (
                <Film className="w-5 h-5" />
              ) : (
                <Calculator className="w-5 h-5" />
              )}
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="text-base font-bold text-white font-mono">{job.jobType}</span>
                <span
                  className={`text-[10px] font-mono uppercase px-2 py-0.5 rounded-full ${
                    job.status === "COMPLETED"
                      ? "bg-indigo-500/20 text-indigo-300 border border-indigo-500/30"
                      : job.status === "RUNNING"
                      ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 animate-pulse"
                      : job.status === "FAILED"
                      ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                      : "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                  }`}
                >
                  {job.status}
                </span>
              </div>
              <div className="flex items-center space-x-2 mt-0.5 text-xs text-slate-400 font-mono">
                <span>ID: {job.jobId}</span>
                <button
                  onClick={copyJobId}
                  className="hover:text-cyan-400 transition-colors p-0.5 cursor-pointer"
                  title="Copy Job ID"
                >
                  {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 rounded-xl bg-slate-950 border border-white/10 text-slate-400 hover:text-white transition-colors cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Scrollable Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Live Progress Bar */}
          <div className="bg-slate-950/80 rounded-2xl p-5 border border-white/10">
            <div className="flex items-center justify-between text-xs font-mono mb-2">
              <span className="text-slate-300 flex items-center space-x-1.5">
                <Zap className="w-3.5 h-3.5 text-cyan-400" />
                <span>Execution Progress</span>
              </span>
              <span className="text-cyan-400 font-bold">{progressValue}%</span>
            </div>

            <div className="w-full bg-slate-800 rounded-full h-2.5 overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-cyan-500 via-indigo-500 to-emerald-400 transition-all duration-300 rounded-full"
                style={{ width: `${progressValue}%` }}
              />
            </div>

            {liveProgress && (
              <div className="mt-3 flex items-center justify-between text-[11px] font-mono text-slate-400 border-t border-white/5 pt-2">
                <span>Speed: {liveProgress.speed || "—"}</span>
                <span>FPS: {liveProgress.fps || "—"}</span>
                <span>Status: {liveProgress.status}</span>
              </div>
            )}
          </div>

          {/* Lifecycle & Resilience Summary */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
            <div className="bg-slate-950/60 p-3 rounded-xl border border-white/5">
              <span className="text-slate-500 block text-[10px] uppercase">Retries</span>
              <span className="text-white font-bold">
                {job.retryCount} / {job.maxRetries}
              </span>
            </div>

            <div className="bg-slate-950/60 p-3 rounded-xl border border-white/5">
              <span className="text-slate-500 block text-[10px] uppercase">Created</span>
              <span className="text-slate-300">{new Date(job.createdAt).toLocaleTimeString()}</span>
            </div>

            <div className="bg-slate-950/60 p-3 rounded-xl border border-white/5">
              <span className="text-slate-500 block text-[10px] uppercase">Started</span>
              <span className="text-slate-300">
                {job.startedAt ? new Date(job.startedAt).toLocaleTimeString() : "—"}
              </span>
            </div>

            <div className="bg-slate-950/60 p-3 rounded-xl border border-white/5">
              <span className="text-slate-500 block text-[10px] uppercase">Completed</span>
              <span className="text-slate-300">
                {job.completedAt ? new Date(job.completedAt).toLocaleTimeString() : "—"}
              </span>
            </div>
          </div>

          {/* Failure Alert (if any) */}
          {job.errorMessage && (
            <div className="bg-rose-500/10 border border-rose-500/20 rounded-xl p-4 flex items-start space-x-3 text-xs text-rose-400">
              <AlertTriangle className="w-5 h-5 shrink-0 text-rose-400 mt-0.5" />
              <div>
                <span className="font-bold block text-rose-300">Execution Error:</span>
                <span className="font-mono mt-0.5 block">{job.errorMessage}</span>
              </div>
            </div>
          )}

          {/* Dual Input / Output Panels */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Left: Input Parameters */}
            <div className="bg-slate-950/80 rounded-2xl p-5 border border-white/10 flex flex-col">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-300 mb-3 flex items-center space-x-1.5">
                <Sparkles className="w-3.5 h-3.5 text-cyan-400" />
                <span>Input Parameters</span>
              </span>

              <div className="flex-1 bg-slate-900/90 rounded-xl p-3 border border-white/5 overflow-x-auto">
                <pre className="text-xs text-cyan-300 font-mono">
                  {JSON.stringify(job.parameters || {}, null, 2)}
                </pre>
              </div>

              {job.resourceId && (
                <p className="text-[11px] text-slate-400 font-mono mt-3">
                  Input Resource ID: <span className="text-white">{job.resourceId}</span>
                </p>
              )}
            </div>

            {/* Right: Output Results */}
            <div className="bg-slate-950/80 rounded-2xl p-5 border border-white/10 flex flex-col">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-300 mb-3 flex items-center justify-between">
                <span className="flex items-center space-x-1.5">
                  <Film className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Output Result</span>
                </span>
                {downloadUrl && (
                  <a
                    href={downloadUrl}
                    target="_blank"
                    rel="noreferrer"
                    download
                    className="text-xs text-emerald-400 hover:text-emerald-300 font-mono flex items-center space-x-1"
                  >
                    <Download className="w-3 h-3" />
                    <span>Download</span>
                  </a>
                )}
              </span>

              {job.outputResourceId ? (
                /* Video Player Output */
                <div className="space-y-3">
                  {downloadUrl ? (
                    <div className="rounded-xl overflow-hidden border border-white/10 bg-black aspect-video flex items-center justify-center">
                      <video
                        src={downloadUrl}
                        controls
                        className="w-full h-full object-contain"
                      />
                    </div>
                  ) : loadingDownload ? (
                    <div className="rounded-xl border border-white/10 bg-slate-900 aspect-video flex items-center justify-center text-xs text-slate-500 font-mono">
                      Generating S3 Presigned URL...
                    </div>
                  ) : (
                    <div className="rounded-xl border border-white/10 bg-slate-900 aspect-video flex items-center justify-center text-xs text-slate-500 font-mono">
                      Output ready in MinIO ({job.outputResourceId.slice(0, 8)}...)
                    </div>
                  )}

                  {job.metadata && (
                    <div className="bg-slate-900/90 rounded-xl p-3 border border-white/5 overflow-x-auto">
                      <pre className="text-xs text-emerald-300 font-mono">
                        {JSON.stringify(job.metadata, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>
              ) : job.metadata ? (
                /* JSON Compute Metadata Output */
                <div className="flex-1 bg-slate-900/90 rounded-xl p-3 border border-white/5 overflow-x-auto">
                  <pre className="text-xs text-emerald-300 font-mono">
                    {JSON.stringify(job.metadata, null, 2)}
                  </pre>
                </div>
              ) : (
                <div className="flex-1 rounded-xl border border-dashed border-slate-800 flex items-center justify-center p-6 text-center text-xs text-slate-500">
                  {job.status === "RUNNING"
                    ? "Execution in progress..."
                    : job.status === "QUEUED"
                    ? "Waiting in queue for worker lease..."
                    : "No output payload generated."}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
