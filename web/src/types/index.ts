export type JobStatus = "QUEUED" | "RUNNING" | "COMPLETED" | "FAILED";

export interface JobRecord {
  jobId: string;
  resourceId?: string;
  jobType: string;
  status: JobStatus;
  retryCount: number;
  maxRetries: number;
  errorMessage?: string;
  outputResourceId?: string;
  parameters?: Record<string, any>;
  metadata?: Record<string, any>;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
}

export interface JobStats {
  total: number;
  queued: number;
  running: number;
  completed: number;
  failed: number;
}

export interface ListJobsResponse {
  jobs: JobRecord[];
  totalCount: number;
  limit: number;
  offset: number;
}

export interface ProgressUpdate {
  job_id: string;
  progress: number;
  speed?: string;
  fps?: number;
  message?: string;
  metadata?: Record<string, any>;
  status: string;
}

export interface UploadUrlResponse {
  uploadUrl: string;
  objectKey: string;
  expiresIn: number;
  resourceID: string;
}

export interface DownloadUrlResponse {
  download_url: string;
  resourceId?: string;
}
