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
	id        string
	masterURL string
	port      int
	jobs      map[string]*exec.Cmd
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func New(id, masterURL string, port int) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &Agent{
		id:        id,
		masterURL: masterURL,
		port:      port,
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

	log.Printf("[INFO] [AGENT] Agent %s naslouchá na portu %d", a.id, a.port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
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
	hb, err := metrics.CollectStats(a.id)
	if err != nil {
		log.Printf("[ERR] [AGENT] Chyba při sběru systémových metrik: %v", err)
		// Pokračujeme s nulovými metrikami, pokud sběr selže
		hb = &proto.Heartbeat{}
	}

	// Automatické zjištění IP adresy pro Mastera
	ip := metrics.GetLocalIP()

	node := proto.Node{
		ID:             a.id,
		Address:        fmt.Sprintf("http://%s:%d", ip, a.port),
		Status:         proto.NodeIdle,
		TotalCores:     hb.TotalCores,
		TotalMemoryMB:  hb.TotalMemoryMB,
		AvailableCores: hb.AvailableCores,
		FreeMemoryMB:   hb.FreeMemoryMB,
	}

	data, _ := json.Marshal(node)
	res, err := http.Post(a.masterURL+"/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatalf("[ERR] [AGENT] Registrace u master uzlu selhala: %v", err)
	}
	defer res.Body.Close()
	log.Printf("[INFO] [AGENT] Agent úspěšně zaregistrován na adrese: http://%s:%d", ip, a.port)

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
			hb, err := metrics.CollectStats(a.id)
			if err != nil {
				log.Printf("[ERR] [AGENT] Chyba při sběru systémových metrik: %v", err)
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
			res, err := http.Post(a.masterURL+"/heartbeat", "application/json", bytes.NewBuffer(data))
			if err != nil {
				log.Printf("[ERR] [AGENT] Heartbeat selhal: %v", err)
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
		http.Error(w, "Neplatné zadání úlohy", http.StatusBadRequest)
		return
	}

	// Vytvoreni noveho contextu - zamezi vzniku zombie processes
	cmd := exec.CommandContext(a.ctx, "sh", "-c", job.Command)

	if err := cmd.Start(); err != nil {
		http.Error(w, "Chyba: Nepodařilo se spustit úlohu", http.StatusInternalServerError)
		return
	}

	a.mu.Lock()
	a.jobs[job.ID] = cmd
	a.mu.Unlock()

	log.Printf("[INFO] [AGENT] Spouštění úlohy %s: %s", job.ID, job.Command)

	go func() {
		err := cmd.Wait()

		a.mu.Lock()
		defer a.mu.Unlock()
		delete(a.jobs, job.ID)

		job.FinishedAt = time.Now()
		if err != nil {
			log.Printf("[ERR] [AGENT] Úloha %s selhala: %v", job.ID, err)
			job.Status = proto.JobFailed
		} else {
			log.Printf("[INFO] [AGENT] Úloha %s byla úspěšně dokončena", job.ID)
			job.Status = proto.JobDone
		}

		data, _ := json.Marshal(job)
		resp, err := http.Post(a.masterURL+"/update_job", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("[ERR] [AGENT] Nepodařilo se odeslat stav úlohy master uzlu: %v", err)
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
		"node_id":      a.id,
		"running_jobs": running,
		"job_count":    len(a.jobs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

}
