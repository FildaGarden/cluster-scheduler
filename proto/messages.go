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
	ID         string    `json:"id"`
	Command    string    `json:"command"`
	CPUCores   int       `json:"cpu_cores"`
	MemoryMB   uint64    `json:"memory_mb"`
	Priority   int       `json:"priority"`
	UserID     string    `json:"user_id"`
	Status     JobStatus `json:"status"`
	NodeID     string    `json:"node_id"`
	CreatedAt  time.Time `json:"created_at"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

type Node struct {
	ID             string     `json:"id"`
	Address        string     `json:"address"`
	CPUPercent     float64    `json:"cpu_percent"`
	MemoryPercent  float64    `json:"memory_percent"`
	AvailableCores int        `json:"available_cores"`
	TotalCores     int        `json:"total_cores"`
	FreeMemoryMB   uint64     `json:"free_memory_mb"`
	TotalMemoryMB  uint64     `json:"total_memory_mb"`
	Status         NodeStatus `json:"status"`
	LastSeen       time.Time  `json:"last_seen"`
}

type Heartbeat struct {
	NodeID         string   `json:"node_id"`
	CPUPercent     float64  `json:"cpu_percent"`
	MemoryPercent  float64  `json:"memory_percent"`
	AvailableCores int      `json:"available_cores"`
	TotalCores     int      `json:"total_cores"`
	FreeMemoryMB   uint64   `json:"free_memory_mb"`
	TotalMemoryMB  uint64   `json:"total_memory_mb"`
	RunningJobs    []string `json:"running_jobs"`
}

type AlgorithmType string
