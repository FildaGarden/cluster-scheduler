package master

import (
	"bytes"
	"cluster-scheduler/proto"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxControlPayload je limit pro heartbeat a registraci (4KB)
	MaxControlPayload = 4096
	// MaxJobPayload je limit pro odeslání jobu, kde může být dlouhý příkaz (16KB)
	MaxJobPayload = 16384
)

type Master struct {
	listenAddr string
	nodes      map[string]*proto.Node
	jobs       map[string]*proto.Job
	mu         sync.RWMutex
}

func New(listenAddr string) *Master {
	return &Master{
		listenAddr: listenAddr,
		nodes:      make(map[string]*proto.Node),
		jobs:       make(map[string]*proto.Job),
	}

}

// Spusteni http serveru a handling routes
func (m *Master) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", m.handleRegister)
	mux.HandleFunc("/heartbeat", m.handleHeartbeat)
	mux.HandleFunc("/submit", m.handleSubmit)
	mux.HandleFunc("/update_job", m.handleUpdateJob)

	log.Printf("Master poslouchá na %s", m.listenAddr)
	server := &http.Server{
		Addr:    m.listenAddr,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

// Status update
func (m *Master) handleUpdateJob(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var jobUpdate proto.Job
	if err := json.NewDecoder(req.Body).Decode(&jobUpdate); err != nil {
		http.Error(w, "Invalid job update", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if job, ok := m.jobs[jobUpdate.ID]; ok {
		job.Status = jobUpdate.Status
		job.FinishedAt = jobUpdate.FinishedAt
		log.Printf("Job %s updated to status: %s", job.ID, job.Status)
	}

	w.WriteHeader(http.StatusOK)
}

// Registrace noveho Node do Clusteru
func (m *Master) handleRegister(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var node proto.Node
	if err := json.NewDecoder(req.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid node registration", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	node.LastSeen = time.Now()
	m.nodes[node.ID] = &node

	log.Printf("Uzel zaregistrován: %s (%s)", node.ID, node.Address)
	w.WriteHeader(http.StatusOK)
}

// Node update
func (m *Master) handleHeartbeat(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var hb proto.Heartbeat
	if err := json.NewDecoder(req.Body).Decode(&hb); err != nil {
		http.Error(w, "Invalid heartbeat", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[hb.NodeID]
	if !ok {
		log.Printf("⚠️ Heartbeat od neznámého uzlu: %s", hb.NodeID)
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	node.CPUPercent = hb.CPUPercent
	node.MemoryPercent = hb.MemoryPercent
	node.AvailableCores = hb.AvailableCores
	node.TotalCores = hb.TotalCores
	node.FreeMemoryMB = hb.FreeMemoryMB
	node.TotalMemoryMB = hb.TotalMemoryMB
	node.LastSeen = time.Now()

	log.Printf("💓 Heartbeat [%s]: CPU %.1f%% | RAM %d/%d MB (%d Cores volno)",
		hb.NodeID, node.CPUPercent, node.FreeMemoryMB, node.TotalMemoryMB, node.AvailableCores)

	w.WriteHeader(http.StatusOK)
}

// Provizorni prirazeni Job -> Node (FIFO)
// TODO: Implementovat vice algoritmu
func (m *Master) handleSubmit(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxJobPayload)
	var job proto.Job
	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid job submission", http.StatusBadRequest)
		return
	}

	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%s", uuid.New().String())
	}
	job.Status = proto.JobPending
	job.CreatedAt = time.Now()

	m.mu.Lock()
	m.jobs[job.ID] = &job

	var selectedNode *proto.Node
	for _, node := range m.nodes {
		if node.AvailableCores >= job.CPUCores && node.FreeMemoryMB >= job.MemoryMB {
			selectedNode = node
			break
		}
	}

	if selectedNode == nil {
		m.mu.Unlock()
		http.Error(w, "No available nodes for this job", http.StatusServiceUnavailable)
		return
	}

	// Priprava na prirazeni pod zamkem
	job.NodeID = selectedNode.ID
	job.Status = proto.JobRunning
	job.StartedAt = time.Now()
	m.mu.Unlock()

	data, _ := json.Marshal(job)
	res, err := http.Post(selectedNode.Address+"/run", "application/json", bytes.NewBuffer(data))

	if err != nil {
		m.mu.Lock()
		job.Status = proto.JobFailed
		m.mu.Unlock()
		log.Printf("Dispatch job %s to node %s failed: %v", job.ID, selectedNode.ID, err)
		http.Error(w, "Failed to dispatch job to agent", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		m.mu.Lock()
		job.Status = proto.JobFailed
		m.mu.Unlock()
		log.Printf("Dispatch job %s to node %s failed with status: %d", job.ID, selectedNode.ID, res.StatusCode)
		http.Error(w, "Failed to dispatch job to agent", http.StatusInternalServerError)
		return
	}

	log.Printf("Job %s dispatched to node %s", job.ID, selectedNode.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
