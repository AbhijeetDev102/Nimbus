import type {
  JobRecord,
  JobStats,
  ListJobsResponse,
  ProgressUpdate,
  UploadUrlResponse,
  DownloadUrlResponse,
} from "../types";

const API_BASE = import.meta.env.VITE_API_URL || "http://localhost:8081";

export async function fetchJobStats(): Promise<JobStats> {
  const res = await fetch(`${API_BASE}/jobs/stats`);
  if (!res.ok) {
    throw new Error(`Failed to fetch stats: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchJobs(params: {
  limit?: number;
  offset?: number;
  status?: string;
  jobType?: string;
}): Promise<ListJobsResponse> {
  const query = new URLSearchParams();
  if (params.limit) query.set("limit", params.limit.toString());
  if (params.offset !== undefined) query.set("offset", params.offset.toString());
  if (params.status && params.status !== "ALL") query.set("status", params.status);
  if (params.jobType && params.jobType !== "ALL") query.set("jobType", params.jobType);

  const res = await fetch(`${API_BASE}/jobs?${query.toString()}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch jobs: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchJobById(jobId: string): Promise<JobRecord> {
  const res = await fetch(`${API_BASE}/jobs/${jobId}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch job ${jobId}: ${res.statusText}`);
  }
  return res.json();
}

export async function createJob(payload: {
  jobType: string;
  resourceID?: string;
  parameters: Record<string, any>;
}): Promise<{ jobID: string; status: string }> {
  const res = await fetch(`${API_BASE}/jobs/create`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const errText = await res.text();
    throw new Error(errText || "Failed to create job");
  }
  return res.json();
}

export async function getUploadUrl(
  file: File,
  resourceType = "VIDEO"
): Promise<UploadUrlResponse> {
  const res = await fetch(`${API_BASE}/resource/upload-url`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      fileName: file.name,
      contentType: file.type || "video/mp4",
      fileSize: file.size,
      resourceType,
    }),
  });
  if (!res.ok) {
    throw new Error("Failed to get presigned upload URL");
  }
  return res.json();
}

export function uploadFileToS3(
  uploadUrl: string,
  file: File,
  onProgress?: (percent: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", uploadUrl);
    xhr.setRequestHeader("Content-Type", file.type || "video/mp4");

    if (xhr.upload && onProgress) {
      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = Math.round((event.loaded / event.total) * 100);
          onProgress(percent);
        }
      };
    }

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        reject(new Error(`S3 upload failed with status ${xhr.status}`));
      }
    };

    xhr.onerror = () => reject(new Error("Network error during S3 upload"));
    xhr.send(file);
  });
}

export async function getDownloadUrl(resourceId: string): Promise<DownloadUrlResponse> {
  const res = await fetch(`${API_BASE}/resource/${resourceId}/download`);
  if (!res.ok) {
    throw new Error("Failed to get presigned download URL");
  }
  return res.json();
}

export function connectJobProgressWS(
  jobId: string,
  onUpdate: (update: ProgressUpdate) => void,
  onClose?: () => void
): () => void {
  const wsProto = API_BASE.startsWith("https") ? "wss:" : "ws:";
  const host = API_BASE.replace(/^https?:\/\//, "");
  const wsUrl = `${wsProto}//${host}/ws/jobs/${jobId}`;

  const ws = new WebSocket(wsUrl);

  ws.onmessage = (event) => {
    try {
      const data: ProgressUpdate = JSON.parse(event.data);
      onUpdate(data);
    } catch (e) {
      console.error("Failed to parse WS update:", e);
    }
  };

  ws.onerror = (err) => {
    console.error("WebSocket error:", err);
  };

  ws.onclose = () => {
    if (onClose) onClose();
  };

  return () => {
    if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
      ws.close();
    }
  };
}
