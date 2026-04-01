package agent

import (
	"bytes"
	"cluster-scheduler/metrics"
	"cluster-scheduler/proto"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

type Agent struct {
	ID        string
	MasterURL string
	Port      int
	jobs      map[string]*exec.Cmd
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(id, masterUrl string, port int) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &Agent{
		ID:        id,
		MasterURL: masterUrl,
		Port:      port,
		jobs:      make(map[string]*exec.Cmd),
		ctx:       ctx,
		cancel:    cancel,
	}

}

// Zapnuti noveho Agenta
func (a *Agent) Start() {
	a.register()
	go a.heartbeatLoop()

	mux := http.NewServeMux()
	mux.HandleFunc("/run", a.handleRun)
	mux.HandleFunc("/status", a.handleStatus)

	log.Printf("Agent %s poslouchá na portu %d", a.ID, a.Port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.Port),
		Handler: mux,
	}
	log.Fatal(server.ListenAndServe())
}

// Zastaveni agenta
func (a *Agent) Stop() {
	a.cancel()
}

// Registrace agenta
func (a *Agent) register() {
	// Pocatecni metriky pro registraci
	hb, err := metrics.CollectStats(a.ID)
	if err != nil {
		log.Printf("Chyba při sběru metrik: %v", err)
		// Pokračujeme s nulovými metrikami, pokud sběr selže
		hb = &proto.Heartbeat{}
	}

	node := proto.Node{
		ID:             a.ID,
		Address:        fmt.Sprintf("http://localhost:%d", a.Port),
		Status:         proto.NodeIdle,
		TotalCores:     hb.TotalCores,
		TotalMemoryMB:  hb.TotalMemoryMB,
		AvailableCores: hb.AvailableCores,
		FreeMemoryMB:   hb.FreeMemoryMB,
	}

	data, _ := json.Marshal(node)
	_, err = http.Post(a.MasterURL+"/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Registrace selhala: %v", err)
	}
	log.Printf("Agent zaregistrovan: %v", a.ID)

}

// Heartbeat -> updatuje status vytizeni Masterovi
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			hb, err := metrics.CollectStats(a.ID)
			if err != nil {
				log.Printf("Chyba při sběru metrik: %v", err)
				continue
			}

			// Doplnime seznam bezicich uloh
			a.mu.Lock()
			running := make([]string, 0, len(a.jobs))
			for id := range a.jobs {
				running = append(running, id)
			}
			a.mu.Unlock()
			hb.RunningJobs = running

			data, _ := json.Marshal(hb)
			http.Post(a.MasterURL+"/heartbeat", "application/json", bytes.NewBuffer(data))
		}
	}

}

// Prijati Job od Mastera -> Execute
func (a *Agent) handleRun(w http.ResponseWriter, req *http.Request) {
	var job proto.Job
	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid job", http.StatusBadRequest)
		return
	}

	// Vytvoreni noveho contextu - predchazi vzniku zombie processes
	cmd := exec.CommandContext(a.ctx, "sh", "-c", job.Command)

	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start the job", http.StatusInternalServerError)
		return
	}

	a.mu.Lock()
	a.jobs[job.ID] = cmd
	a.mu.Unlock()

	log.Printf("Spouštím úlohu %s: %s", job.ID, job.Command)

	// Cekani na dokonceni goroutine
	go func() {
		err := cmd.Wait()

		a.mu.Lock()
		delete(a.jobs, job.ID)
		a.mu.Unlock()

		job.FinishedAt = time.Now()
		if err != nil {
			log.Printf("Úloha %s selhala: %v", job.ID, err)
			job.Status = proto.JobFailed
		} else {
			log.Printf("Úloha %s dokončena", job.ID)
			job.Status = proto.JobDone
		}

		// Final status update
		data, _ := json.Marshal(job)
		_, postErr := http.Post(a.MasterURL+"/update_job", "application/json", bytes.NewBuffer(data))
		if postErr != nil {
			log.Printf("Failed to notify master of job completion: %v", postErr)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}

// Vraci list probihajicich Jobs na konkretni Node
func (a *Agent) handleStatus(w http.ResponseWriter, req *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	running := make([]string, 0, len(a.jobs))
	for id := range a.jobs {
		running = append(running, id)
	}

	status := map[string]any{
		"node_id":      a.ID,
		"running_jobs": running,
		"job_count":    len(a.jobs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

}
