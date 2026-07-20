# Concurrent Task Scheduler

## 🚀 Overview

This component implements a worker pool pattern for heterogeneous task processing, context cancellation, and queue metrics to understand Go's memory model and G-M-P scheduler internals.

## 📅 Schedule

- **Day 1: Worker Pool Foundation**
  - Define `Task` interface/struct (with CPU/GPU `Kind`) and `Pool`.
  - Fixed-size worker goroutines pulling from a buffered task channel.
  - Analyze escape analysis (`go build -gcflags="-m"`).
- **Day 2: Heterogeneous Scheduling & Cancellation**
  - Separate CPU-bound tasks (`runtime.NumCPU()`) from GPU-bound tasks.
  - Thread `context.Context` through `Submit` and worker loops.
  - Graceful shutdown via `Pool.Close()` using `sync.Once`.
- **Day 3: G-M-P Internals & Queue Metrics**
  - Queue depth tracking and execution time metrics.
  - Stress testing under burst load.

## 🛠️ Getting Started

```bash
cd task_scheduler
go run main.go
```
