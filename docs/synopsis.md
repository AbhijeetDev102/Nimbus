# Project Synopsis: Nimbus ⚡
## Cloud-Native Distributed Job Execution Engine & Control Plane

---

### Table of Contents
1. [Acknowledgment](#1-acknowledgment) *(Pending)*
2. [Certificate](#2-certificate) *(Pending)*
3. [Introduction](#3-introduction)
   - [3.1 Background and Motivation](#31-background-and-motivation)
   - [3.2 Problem Statement](#32-problem-statement)
   - [3.3 Overview of the Proposed System (Nimbus)](#33-overview-of-the-proposed-system-nimbus)
4. [Objectives](#4-objectives) *(Up Next)*
5. [Literature Review](#5-literature-review)
6. [Opportunities and Challenges](#6-opportunities-and-challenges)
7. [Future Scope](#7-future-scope)
8. [Hardware & Software Requirements](#8-hardware--software-requirements)
9. [Conclusion](#9-conclusion)
10. [References / Bibliography](#10-references--bibliography)

---

## 1. Acknowledgment
*[To be added in final draft]*

---

## 2. Certificate
*[To be added in final draft]*

---

## 3. Introduction

When you use modern web applications—whether uploading a video to YouTube, requesting a large PDF bank statement, or generating images with an AI model—these operations take seconds, minutes, or even hours to complete. A standard web server cannot handle these operations directly in a normal web request. If a browser waits longer than a few seconds, the connection times out, the browser freezes, and web servers quickly run out of memory.

To solve this, modern applications use **background job execution systems**: the web server accepts the user's request, saves it as a "job," returns an immediate confirmation ("Your task has started!"), and delegates the heavy computation to background worker machines.

However, building a background job system that runs reliably at production scale is notoriously difficult. Naive queues and simple background scripts frequently fail when servers crash mid-operation, files get huge, or multiple machines try to process the same job at once.

**Nimbus** is a cloud-native, distributed background job execution platform and control plane built in Go. It provides a rock-solid, crash-resilient engine where heavy workloads (such as video transcoding, AI inference, and data processing) run reliably, recover automatically from crashes, and stream real-time progress updates to users—all without losing data or overloading the system.

---

### 3.1 Background and Motivation

In any large-scale web platform, keeping the user interface fast and responsive requires offloading slow, compute-heavy tasks away from the main web servers into dedicated worker clusters.

Traditionally, development teams build these systems using simple message queues or in-memory brokers like Redis (e.g., Celery, BullMQ) or RabbitMQ. While these tools work well for small prototype projects, they introduce serious failure modes when deployed in real-world production environments:

1. **The "Lost Job" Problem (The Dual-Write Dilemma):**
   When a user clicks "Submit", the system needs to do two things: save the job in the database (for record-keeping) and send a message to the queue (so a worker knows to run it). If the server saves the job to the database, but crashes or loses its network connection right before notifying the queue, the job sits in the database forever and never runs. The user is left waiting indefinitely.
   
2. **Web Servers Crashing on Big Uploads (Network & Memory Bottlenecks):**
   If a user uploads a 2 GB video file, naive systems pass the entire file through the API server into memory before forwarding it to storage. When dozens of users upload at the same time, the API servers run out of RAM (Out of Memory crashes), bringing down the entire website for all users.

3. **Duplicate Work and Race Conditions:**
   In distributed cloud environments, multiple worker machines listen for jobs. If two workers grab the same job at the exact same fraction of a second, both start processing it. This wastes expensive compute power, generates duplicate outputs, and can corrupt user data.

4. **Stuck / Ghost Jobs when Workers Crash:**
   If a worker machine running a 30-minute transcoding job suddenly runs out of memory or gets restarted by cloud providers (e.g., spot instances), the job remains stuck in a "RUNNING" state forever. Traditional systems have no built-in way to detect that the worker died and reassign the job.

5. **Database Meltdowns from Progress Polling:**
   Users want to see a live progress bar ("45% completed..."). If thousands of users refresh their screen or poll the database every second to check their progress, the database gets flooded with read and write requests and slows down to a crawl.

These everyday production headaches motivated the creation of **Nimbus**. Nimbus was engineered from the ground up to solve these exact problems using battle-tested distributed systems patterns.

---

### 3.2 Problem Statement

The central problem addressed by this project is:

> **How can we design and implement a distributed background job execution engine that guarantees zero lost jobs, prevents duplicate executions, protects API servers from heavy file uploads, automatically recovers from worker crashes, and provides live sub-millisecond progress updates—without overwhelming core databases?**

Specifically, the project addresses four key technical hurdles:
- **Ensuring 100% Reliable Ingestion:** Eliminating the dual-write problem so that every submitted job is guaranteed to be picked up, even if servers crash midway.
- **Decoupling Compute from API Bandwidth:** Ensuring multi-gigabyte media uploads never pass through API server memory.
- **Fail-Safe Worker Coordination:** Providing distributed locking (leases), worker heartbeats, and panic recovery so dead workers never leave jobs permanently frozen.
- **Scalable Real-Time Telemetry:** Delivering smooth progress updates to browser dashboards without causing database bottlenecks or write amplification.

---

### 3.3 Overview of the Proposed System (Nimbus)

**Nimbus** solves these challenges by combining proven cloud-native technologies into a cohesive, high-performance platform:

```
[ Web Client / Dashboard ] 
       │                 │
       │ (1. Direct      │ (2. Submit Job)
       │    S3 Upload)   ▼
       │          [ API Gateway ]
       ▼                 │
[ MinIO Object Storage ] │ (gRPC)
                         ▼
                  [ Job Service ]
                         │
                         ▼ (ACID Transaction)
             [ PostgreSQL Database ] ──(WAL)──► [ Debezium CDC ]
                                                       │
                                                       ▼
                                            [ Apache Kafka Stream ]
