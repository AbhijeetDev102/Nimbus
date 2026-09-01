import React, { useState, useRef } from "react";
import {
  UploadCloud,
  CheckCircle,
  Play,
  AlertCircle,
  FileVideo,
  Paperclip,
  Check,
  Code2,
} from "lucide-react";
import { getUploadUrl, uploadFileToS3, createJob } from "../../services/api";

interface WorkloadStudioProps {
  onJobDispatched: (jobId: string) => void;
}

const PRESETS = {
  VIDEO_TRANSCODE: {
    jobType: "VIDEO_TRANSCODE",
    attachResource: true,
    params: JSON.stringify(
      {
        resolution: "1280x720",
        codec: "libx264",
        bitrate: "2000k",
      },
      null,
      2
    ),
  },
  calculator: {
    jobType: "calculator",
    attachResource: false,
    params: JSON.stringify(
      {
        num1: 15,
        num2: 27,
      },
      null,
      2
    ),
  },
  CUSTOM: {
    jobType: "custom.compute",
    attachResource: false,
    params: JSON.stringify(
      {
        task: "data_transform",
        batch_size: 100,
        tags: ["production", "v1"],
      },
      null,
      2
    ),
  },
};

export const WorkloadStudio: React.FC<WorkloadStudioProps> = ({ onJobDispatched }) => {
  const [jobType, setJobType] = useState("VIDEO_TRANSCODE");
  const [paramsJSON, setParamsJSON] = useState(PRESETS.VIDEO_TRANSCODE.params);
  const [attachResource, setAttachResource] = useState(true);

  // File Upload State
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadPercent, setUploadPercent] = useState(0);
  const [uploadedResourceId, setUploadedResourceId] = useState<string | null>(null);

  // Form State
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jsonError, setJsonError] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement | null>(null);

  // Validate JSON on the fly
  const handleParamsChange = (val: string) => {
    setParamsJSON(val);
    try {
      JSON.parse(val);
      setJsonError(null);
    } catch (e: any) {
      setJsonError(e.message);
    }
  };

  const applyPreset = (key: keyof typeof PRESETS) => {
    const preset = PRESETS[key];
    setJobType(preset.jobType);
    setAttachResource(preset.attachResource);
    setParamsJSON(preset.params);
    setJsonError(null);
    setError(null);
  };

  const handleFileChange = async (file: File) => {
    setSelectedFile(file);
    setIsUploading(true);
    setUploadPercent(0);
    setError(null);

    try {
      // 1. Get Presigned S3 PUT URL
      const { uploadUrl, resourceID } = await getUploadUrl(file, "VIDEO");
      setUploadedResourceId(resourceID);

      // 2. Direct-to-storage upload
      await uploadFileToS3(uploadUrl, file, (percent) => {
        setUploadPercent(percent);
      });

      setIsUploading(false);
    } catch (err: any) {
      setError(err.message || "Failed to upload file to MinIO storage");
      setIsUploading(false);
    }
  };

  const handleDispatch = async () => {
    setError(null);

    // 1. Validate JobType
    if (!jobType.trim()) {
      setError("Please specify a Job Type");
      return;
    }

    // 2. Validate JSON
    let parsedParams: Record<string, any> = {};
    try {
      parsedParams = JSON.parse(paramsJSON);
    } catch (e: any) {
      setError(`Invalid JSON parameters: ${e.message}`);
      return;
    }

    // 3. Validate Resource if enabled
    if (attachResource && !uploadedResourceId) {
      setError("Please attach a file asset or disable 'Attach Resource'");
      return;
    }

    setIsSubmitting(true);

    try {
      const res = await createJob({
        jobType: jobType.trim(),
        resourceID: attachResource && uploadedResourceId ? uploadedResourceId : undefined,
        parameters: parsedParams,
      });

      setIsSubmitting(false);
      onJobDispatched(res.jobID);
    } catch (err: any) {
      setError(err.message || "Failed to dispatch job");
      setIsSubmitting(false);
    }
  };

  return (
    <div className="rounded-2xl bg-gradient-to-b from-slate-900/90 to-[#080d16]/90 p-6 border border-white/10 shadow-2xl space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-5 border-b border-white/10 gap-3">
        <div>
          <h2 className="text-lg font-bold text-white tracking-tight flex items-center space-x-2">
            <span>Workload Studio</span>
            <span className="text-[10px] px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 border border-cyan-500/30">
              Generic Dispatcher
            </span>
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Dispatch any registered worker task with custom parameters and optional assets
          </p>
        </div>

        {/* Quick Presets */}
        <div className="flex items-center space-x-1.5 text-xs">
          <span className="text-slate-500 text-[11px] font-mono mr-1 hidden sm:inline">Presets:</span>
          <button
            onClick={() => applyPreset("VIDEO_TRANSCODE")}
            className="px-2.5 py-1 rounded-lg bg-slate-950 border border-white/10 hover:border-cyan-500/40 text-slate-300 hover:text-cyan-300 font-mono transition-all text-xs cursor-pointer"
          >
            Video
          </button>
          <button
            onClick={() => applyPreset("calculator")}
            className="px-2.5 py-1 rounded-lg bg-slate-950 border border-white/10 hover:border-emerald-500/40 text-slate-300 hover:text-emerald-300 font-mono transition-all text-xs cursor-pointer"
          >
            Calculator
          </button>
          <button
            onClick={() => applyPreset("CUSTOM")}
            className="px-2.5 py-1 rounded-lg bg-slate-950 border border-white/10 hover:border-purple-500/40 text-slate-300 hover:text-purple-300 font-mono transition-all text-xs cursor-pointer"
          >
            Custom
          </button>
        </div>
      </div>

      {/* Main Form Fields */}
      <div className="space-y-5">
        {/* Job Type Input */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-300 mb-2">
            1. Job Type / Workload Name
          </label>
          <div className="relative">
            <input
              type="text"
              placeholder="e.g. VIDEO_TRANSCODE, calculator, ai.whisper, email.send"
              value={jobType}
              onChange={(e) => setJobType(e.target.value)}
              className="w-full bg-slate-950 border border-white/10 rounded-xl px-4 py-2.5 text-sm text-cyan-300 font-mono placeholder-slate-600 focus:outline-none focus:border-cyan-500"
            />
          </div>
          <p className="text-[11px] text-slate-500 font-mono mt-1">
            Matches the string registered in your Worker Dispatcher (<code className="text-slate-400">worker.Register(jobType, handler)</code>)
          </p>
        </div>

        {/* Resource Attachment Toggle */}
        <div className="rounded-xl bg-slate-950/60 p-4 border border-white/10 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Paperclip className="w-4 h-4 text-cyan-400" />
              <div>
                <span className="text-xs font-semibold text-white block">
                  Attach Resource Asset (Direct S3 / MinIO)
                </span>
                <span className="text-[11px] text-slate-400">
                  Enable for media/files; disable for compute/data workloads
                </span>
              </div>
            </div>

            {/* Toggle Switch */}
            <button
              type="button"
              onClick={() => {
                setAttachResource(!attachResource);
                if (attachResource) {
                  setSelectedFile(null);
                  setUploadedResourceId(null);
                }
              }}
              className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                attachResource ? "bg-cyan-500" : "bg-slate-800"
              }`}
            >
              <span
                className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-lg ring-0 transition duration-200 ease-in-out ${
                  attachResource ? "translate-x-5" : "translate-x-0"
                }`}
              />
            </button>
          </div>

          {/* Expandable Drag & Drop Zone */}
          {attachResource && (
            <div className="pt-2 animate-in fade-in duration-200">
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                onChange={(e) => {
                  if (e.target.files && e.target.files[0]) {
                    handleFileChange(e.target.files[0]);
                  }
                }}
              />

              {!selectedFile ? (
                <div
                  onClick={() => fileInputRef.current?.click()}
                  className="border-2 border-dashed border-slate-700 hover:border-cyan-500/60 rounded-xl p-5 text-center cursor-pointer transition-all bg-slate-950/40 hover:bg-slate-900/40 group"
                >
                  <UploadCloud className="w-7 h-7 mx-auto text-slate-500 group-hover:text-cyan-400 transition-colors" />
                  <p className="mt-1.5 text-xs font-medium text-slate-300">
                    Click or drag & drop file to upload directly to MinIO
                  </p>
                  <p className="text-[10px] text-slate-500 mt-0.5 font-mono">
                    Direct-to-storage Presigned PUT
                  </p>
                </div>
              ) : (
                <div className="bg-slate-900/80 rounded-xl p-3.5 border border-white/10">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                        <FileVideo className="w-4 h-4" />
                      </div>
                      <div>
                        <p className="text-xs font-medium text-white">{selectedFile.name}</p>
                        <p className="text-[11px] text-slate-400 font-mono">
                          {(selectedFile.size / (1024 * 1024)).toFixed(2)} MB
                          {uploadedResourceId && ` • Resource ID: ${uploadedResourceId.slice(0, 8)}...`}
                        </p>
                      </div>
                    </div>

                    <button
                      onClick={() => {
                        setSelectedFile(null);
                        setUploadedResourceId(null);
                        setUploadPercent(0);
                      }}
                      className="text-xs text-slate-400 hover:text-rose-400 transition-colors cursor-pointer"
                    >
                      Remove
                    </button>
                  </div>

                  {isUploading && (
                    <div className="mt-2.5">
                      <div className="flex justify-between text-[11px] text-cyan-400 font-mono mb-1">
                        <span>Uploading to storage...</span>
                        <span>{uploadPercent}%</span>
                      </div>
                      <div className="w-full bg-slate-800 rounded-full h-1.5 overflow-hidden">
                        <div
                          className="bg-cyan-400 h-full transition-all duration-200"
                          style={{ width: `${uploadPercent}%` }}
                        />
                      </div>
                    </div>
                  )}

                  {!isUploading && uploadedResourceId && (
                    <div className="mt-2 flex items-center space-x-1.5 text-xs text-emerald-400 font-mono">
                      <CheckCircle className="w-3.5 h-3.5" />
                      <span>Asset uploaded & attached</span>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Dynamic JSON Parameters Editor */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-xs font-semibold uppercase tracking-wider text-slate-300 flex items-center space-x-1.5">
              <Code2 className="w-3.5 h-3.5 text-cyan-400" />
              <span>2. Execution Parameters (Raw JSON)</span>
            </label>

            {jsonError ? (
              <span className="text-[11px] text-rose-400 font-mono flex items-center space-x-1">
                <AlertCircle className="w-3 h-3" />
                <span>Invalid JSON</span>
              </span>
            ) : (
              <span className="text-[11px] text-emerald-400 font-mono flex items-center space-x-1">
                <Check className="w-3 h-3" />
                <span>Valid JSON</span>
              </span>
            )}
          </div>

          <div className="relative rounded-xl overflow-hidden border border-white/10 focus-within:border-cyan-500 bg-slate-950">
            <textarea
              rows={6}
              value={paramsJSON}
              onChange={(e) => handleParamsChange(e.target.value)}
              className="w-full bg-slate-950 p-3.5 text-xs text-cyan-300 font-mono focus:outline-none resize-y selection:bg-cyan-500 selection:text-black"
              placeholder='{\n  "key": "value"\n}'
            />
          </div>
        </div>

        {/* Error Alert */}
        {error && (
          <div className="flex items-center space-x-2 text-xs text-rose-400 bg-rose-500/10 border border-rose-500/20 p-3 rounded-xl">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{error}</span>
          </div>
        )}

        {/* Dispatch Action Button */}
        <button
          onClick={handleDispatch}
          disabled={isSubmitting || !!jsonError || (attachResource && (!uploadedResourceId || isUploading))}
          className="w-full py-3.5 px-4 rounded-xl bg-gradient-to-r from-cyan-500 via-indigo-500 to-purple-600 hover:from-cyan-400 hover:to-purple-500 text-white font-semibold text-sm shadow-lg shadow-cyan-500/25 flex items-center justify-center space-x-2 transition-all disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
        >
          <Play className="w-4 h-4 fill-white" />
          <span>{isSubmitting ? "Dispatching to Outbox..." : `Dispatch ${jobType || "Job"}`}</span>
        </button>
      </div>
    </div>
  );
};
