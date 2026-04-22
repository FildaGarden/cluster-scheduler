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
	NodeOffline NodeStatus = "offline"
)

type JobAlgorithm string

const (
	JobAlgoFIFO      JobAlgorithm = "fifo"
	JobAlgoPriority  JobAlgorithm = "priority"
	JobAlgoFairShare JobAlgorithm = "fair-share"
)

type NodeAlgorithm string

const (
	NodeAlgoFirstAvailable NodeAlgorithm = "first-available"
	NodeAlgoLeastLoaded    NodeAlgorithm = "least-loaded"
)

type Job struct {
	ID         string    `json:"id"`
	Command    string    `json:"command"`
	CPUCores   int       `json:"cpu_cores"`
	MemoryMB   int       `json:"memory_mb"`
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
	FreeMemoryMB   int        `json:"free_memory_mb"`
	TotalMemoryMB  int        `json:"total_memory_mb"`
	Status         NodeStatus `json:"status"`
	LastSeen       time.Time  `json:"last_seen"`
}

type Heartbeat struct {
	NodeID         string   `json:"node_id"`
	CPUPercent     float64  `json:"cpu_percent"`
	MemoryPercent  float64  `json:"memory_percent"`
	AvailableCores int      `json:"available_cores"`
	TotalCores     int      `json:"total_cores"`
	FreeMemoryMB   int      `json:"free_memory_mb"`
	TotalMemoryMB  int      `json:"total_memory_mb"`
	RunningJobs    []string `json:"running_jobs"`
}
