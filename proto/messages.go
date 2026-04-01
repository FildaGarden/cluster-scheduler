package proto

import "time"

type JobStatus string
type NodeStatus string

const (
	JobPending JobStatus = "pending"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
)

const (
	NodeIdle    NodeStatus = "idle"
	NodeBusy    NodeStatus = "busy"
	NodePending NodeStatus = "pending"
)

const (
	AlgorithmFIFO      AlgorithmType = "fifo"
	AlgorithmPriority  AlgorithmType = "priority"
	AlgorithmFairShare AlgorithmType = "fair-share"
)

type Job struct {
	ID         string
	Command    string
	CPUCores   int
	MemoryMB   int
	Priority   int
	UserID     string
	Status     JobStatus
	NodeID     string
	CreatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time
}

type Node struct {
	ID             string
	Address        string
	CPUPercent     float64
	MemoryPercent  float64
	AvailableCores int
	TotalCores     int
	FreeMemoryMB   uint64
	TotalMemoryMB  uint64
	Status         NodeStatus
	LastSeen       time.Time
}

type Heartbeat struct {
	NodeID         string
	CPUPercent     float64
	MemoryPercent  float64
	AvailableCores int
	TotalCores     int
	FreeMemoryMB   uint64
	TotalMemoryMB  uint64
	RunningJobs    []string
}

type AlgorithmType string
