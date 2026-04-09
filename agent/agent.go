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

// Interval v sekundach
const HeartbeatInterval = 5

type Agent struct {
	ID        string
	MasterURL string
	Port      int
	jobs      map[string]*exec.Cmd
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(id, masterURL string, port int) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &Agent{
		ID:        id,
		MasterURL: masterURL,
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

	// Automatické zjištění IP adresy pro Mastera
	ip := metrics.GetLocalIP()

	node := proto.Node{
		ID:             a.ID,
		Address:        fmt.Sprintf("http://%s:%d", ip, a.Port),
		Status:         proto.NodeIdle,
		TotalCores:     hb.TotalCores,
		TotalMemoryMB:  hb.TotalMemoryMB,
		AvailableCores: hb.AvailableCores,
		FreeMemoryMB:   hb.FreeMemoryMB,
	}

	data, _ := json.Marshal(node)
	res, err := http.Post(a.MasterURL+"/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("Registrace selhala: %v", err)
	}
	defer res.Body.Close()
	log.Printf("Agent zaregistrovan na adrese: http://%s:%d", ip, a.Port)

}

// Heartbeat -> updatuje status vytizeni Masterovi
func (a *Agent) heartbeatLoop() {
	ticker := time.NewTicker(HeartbeatInterval * time.Second)
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
			res, err := http.Post(a.MasterURL+"/heartbeat", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Heartbeat failed: %v", err)
				continue
			}
			res.Body.Close()

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

	// Vytvoreni noveho contextu - zamezi vzniku zombie processes
	cmd := exec.CommandContext(a.ctx, "sh", "-c", job.Command)

	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start the job", http.StatusInternalServerError)
		return
	}

	a.mu.Lock()
	a.jobs[job.ID] = cmd
	a.mu.Unlock()

	log.Printf("Spouštím úlohu %s: %s", job.ID, job.Command)

	go func() {
		err := cmd.Wait()

		a.mu.Lock()
		defer a.mu.Unlock()
		delete(a.jobs, job.ID)

		job.FinishedAt = time.Now()
		if err != nil {
			log.Printf("Úloha %s selhala: %v", job.ID, err)
			job.Status = proto.JobFailed
		} else {
			log.Printf("Úloha %s dokončena", job.ID)
			job.Status = proto.JobDone
		}

		data, _ := json.Marshal(job)
		resp, err := http.Post(a.MasterURL+"/update_job", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Failed to notify master of job completion: %v", err)
			return
		}
		resp.Body.Close()
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
