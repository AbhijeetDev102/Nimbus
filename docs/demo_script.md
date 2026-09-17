# Nimbus Distributed Execution & Resilience Demo Script 🎬

This guide provides a step-by-step procedure to showcase Nimbus's core distributed features: **multi-worker load distribution**, **TTL-based distributed lease fencing**, **crash recovery via Job Reaper**, and **live cluster observability**.

---

## 🛠 Prerequisites & Setup

Ensure Docker and Docker Compose are running.

```bash
# 1. Start Nimbus infrastructure with 3 worker replicas
docker compose up -d --scale worker=3

# 2. Register Debezium CDC Outbox Connector (one-time setup)
go run tools/registerConnector/main.go

# 3. Open the Nimbus Control Plane Dashboard
# Navigate to http://localhost:5173 (or http://localhost:3000)
```

---

## 🖥 Demo Tour: What to Highlight in the Dashboard

When you open the Nimbus Dashboard, highlight these components:

1. **Worker Fleet Roster (Top Right):**
   - Shows all active worker replicas (`worker-1`, `worker-2`, `worker-3`).
   - Pulsing green indicators show heartbeat health from Redis (`nimbus:workers`).
   - Real-time status shows `IDLE` when waiting, and `BUSY (processing #jobId)` when crunching.
2. **Cluster Activity Feed (Left Column):**
   - Live terminal-like timeline showing distributed cluster events:
     - `DISPATCHED`: Job accepted and committed to Postgres Outbox.
     - `CLAIMED`: Exclusive lease won by specific worker replica.
     - `COMPLETED`: Workload finished and results written.
     - `RETRY`: Distributed retry scheduled via Outbox.
3. **Hero Job Inspector Modal:**
   - Click any job to inspect its **Claimed Worker ID**, **Distributed Lease TTL**, and live WebSocket progress bar.

---

## ⚡ Demo Scenarios

### Scenario 1: Multi-Worker Concurrent Load Distribution

**Concept to Demonstrate:** Workload agnosticism, Kafka partitioning, and race-free lease claims across multiple workers.

1. Go to the **Workload Studio** tab.
2. Select **Matrix Compute / Stress Test** (or Video Transcoding).
3. Set payload to 1,000,000 operations.
4. Click **Dispatch Workload** 3–5 times in quick succession.
5. **Observe in Dashboard:**
   - In the **Worker Fleet** panel: Worker 1, Worker 2, and Worker 3 immediately turn `BUSY` simultaneously.
   - In the **Activity Feed**: Observe individual `CLAIMED by worker-XXXX` entries confirming that Kafka partitioned the jobs across all 3 nodes without race conditions.

---

### Scenario 2: Worker Crash, Lease Expiry & Reaper Recovery (The Hero Demo)

**Concept to Demonstrate:** Crash resilience, TTL lease recovery, and zero data loss.

1. Submit a long-running job:
   - Select **Video Transcode** or heavy **Compute**.
   - Click **Dispatch Workload**.
2. Identify which worker claimed the job in the **Worker Fleet** roster (e.g. `worker-1` / `worker-a3f2`).
3. Open a terminal and immediately kill that specific worker container:
   ```bash
   # Find container name/ID:
   docker compose ps | grep worker

   # Kill the worker actively processing the job:
   docker kill <worker_container_name>
   ```
4. **Observe in Dashboard:**
   - Within 15 seconds, the killed worker's status flips to `OFFLINE` (red badge) as its Redis heartbeats cease.
   - The job's lease in Postgres expires (`lease_expires_at < NOW()`).
   - The **Job Reaper daemon** detects the abandoned job, logs `[JobReaper] Re-queuing expired job`, and re-emits a `JobCreated` Outbox event.
   - A surviving worker (`worker-2` or `worker-3`) claims the job.
   - In the **Activity Feed**, watch the transition: `RETRY (attempt 2/3) -> CLAIMED by worker-b9c1 -> COMPLETED`.
   - **Result:** Zero manual intervention, zero job loss, 100% resilient.

---

### Scenario 3: Fencing Stale Workers (Split-Brain Prevention)

**Concept to Demonstrate:** Optimistic concurrency fencing (`WHERE worker_id = :id AND status = 'RUNNING'`).

1. If a stalled worker pauses (e.g., GC pause or network partition) and wakes back up *after* its lease was stolen:
   - It attempts to call `CompleteJob` or `Heartbeat`.
   - The conditional SQL update returns `RowsAffected = 0`.
   - The worker detects:
     ```
     [Heartbeat] Lease lost! Another worker or reaper claimed it.
     ```
   - The resurrected worker is strictly blocked from overwriting state, completely preventing split-brain corruption.

---

## 🎥 Video Recording Checklist

For recording a high-impact GIF or demo video:
- [ ] Browser window on the left displaying the Nimbus Dashboard.
- [ ] Terminal window on the right with `docker compose logs -f job-service worker`.
- [ ] Dispatch a job, show progress bar moving.
- [ ] Run `docker kill <worker>` in terminal.
- [ ] Show dashboard worker roster turning red, reaper log in terminal, and surviving worker finishing the job.
- [ ] Show final `COMPLETED` state in dashboard.
