package master

import (
	"bytes"
	"cluster-scheduler/config"
	"cluster-scheduler/proto"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const (
	// MaxControlPayload je limit pro heartbeat a registraci (4KB)
	MaxControlPayload = 4096
	// MaxJobPayload je limit pro odeslání jobu (1MB)
	MaxJobPayload = 1024 * 1024
)

type SchedulingStrategy interface {
	SelectNode(job *proto.Job, nodes map[string]*proto.Node) *proto.Node
}

type FirstAvailableScheduler struct{}

func (s *FirstAvailableScheduler) SelectNode(job *proto.Job, nodes map[string]*proto.Node) *proto.Node {
	for _, node := range nodes {
		if node.AvailableCores >= job.CPUCores && node.FreeMemoryMB >= job.MemoryMB {
			return node
		}
	}
	return nil
}

type LeastLoadedScheduler struct{}

func (s *LeastLoadedScheduler) SelectNode(job *proto.Job, nodes map[string]*proto.Node) *proto.Node {
	var bestNode *proto.Node
	maxFreeCores := -1

	for _, node := range nodes {
		if node.AvailableCores >= job.CPUCores && node.FreeMemoryMB >= job.MemoryMB {
			// Strategie: Vyber uzel s nejvíce volnými jádry
			if node.AvailableCores > maxFreeCores {
				maxFreeCores = node.AvailableCores
				bestNode = node
			}
		}
	}
	return bestNode
}

type Master struct {
	listenAddr string
	nodes      map[string]*proto.Node
	mu         sync.RWMutex
	db         *sql.DB
	scheduler  SchedulingStrategy
}

func New(cfg *config.Config) *Master {
	// Zajistit existenci složky pro DB
	dbPath := "db/scheduler.db"
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Chyba při vytváření složky pro DB: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		log.Fatalf("Chyba pri otvirani databaze %v", err)
	}

	var strategy SchedulingStrategy
	switch cfg.Algorithm {
	case proto.AlgorithmFirstAvailable:
		strategy = &FirstAvailableScheduler{}
	case proto.AlgorithmLeastLoaded:
		strategy = &LeastLoadedScheduler{}
	default:
		strategy = &LeastLoadedScheduler{}
	}
	m := &Master{
		listenAddr: cfg.ListenAddr,
		db:         db,
		nodes:      make(map[string]*proto.Node),
		scheduler:  strategy,
	}
	m.initDB()

	return m
}

// Start spustí HTTP server a rutiny na pozadí
func (m *Master) Start() {
	go m.runSchedulingLoop()
	go m.runHealthCheck()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", m.handleRegister)
	mux.HandleFunc("/heartbeat", m.handleHeartbeat)
	mux.HandleFunc("/submit", m.handleSubmit)
	mux.HandleFunc("/update_job", m.handleUpdateJob)
	mux.HandleFunc("/nodes", m.handleListNodes)
	mux.HandleFunc("/jobs", m.handleListJobs)

	log.Printf("[INFO] [MASTER] Master uzel spuštěn na %s", m.listenAddr)
	server := &http.Server{
		Addr:    m.listenAddr,
		Handler: mux,
	}

	// Spustíme server v gorutině, aby Start() nebyl blokující
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERR] [MASTER] Chyba serveru: %v", err)
		}
	}()
}

// Stop bezpečně zastaví Mastera a zavře databázi
func (m *Master) Stop() {
	log.Println("[INFO] [MASTER] Zastavování master uzlu a ukončování databázových spojení...")
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			log.Printf("[ERR] [DB] Chyba při zavírání databáze: %v", err)
		} else {
			log.Println("[INFO] [DB] Databáze byla úspěšně uzavřena.")
		}
	}
}

// handleUpdateJob zpracuje update statusu od Agenta a uvolní zdroje
func (m *Master) handleUpdateJob(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var jobUpdate proto.Job
	if err := json.NewDecoder(req.Body).Decode(&jobUpdate); err != nil {
		http.Error(w, "Invalid job update", http.StatusBadRequest)
		return
	}

	// 1. Zjistíme detaily o původní úloze z DB
	var cpuCores int
	var memoryMB int
	var nodeID string
	var currentStatus string
	err := m.db.QueryRow("SELECT cpu_cores, memory_mb, node_id, status FROM jobs WHERE id = ?", jobUpdate.ID).
		Scan(&cpuCores, &memoryMB, &nodeID, &currentStatus)

	if err != nil {
		log.Printf("[WARN] [JOB] Přijat update pro neznámou úlohu: %s", jobUpdate.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Pokud úloha končí a dříve běžela, uvolníme zdroje v RAM
	if (jobUpdate.Status == proto.JobDone || jobUpdate.Status == proto.JobFailed) && currentStatus == string(proto.JobRunning) {
		m.mu.Lock()
		if node, ok := m.nodes[strings.ToLower(nodeID)]; ok {
			node.AvailableCores += cpuCores
			node.FreeMemoryMB += memoryMB
			log.Printf("[INFO] [NODE] Uvolněny zdroje na uzlu %s: %d jader, %d MB RAM (Úloha %s)", nodeID, cpuCores, memoryMB, jobUpdate.ID)
		}
		m.mu.Unlock()
	}

	// 3. Update statusu v DB
	_, err = m.db.Exec("UPDATE jobs SET status = ?, finished_at = ? WHERE id = ?",
		jobUpdate.Status, time.Now(), jobUpdate.ID)

	if err != nil {
		log.Printf("[ERR] [DB] Chyba při aktualizaci databáze pro úlohu %s: %v", jobUpdate.ID, err)
	}

	log.Printf("[INFO] [JOB] Úloha %s aktualizována na stav: %s", jobUpdate.ID, jobUpdate.Status)
	w.WriteHeader(http.StatusOK)
}

// handleRegister zaregistruje nový výpočetní uzel
func (m *Master) handleRegister(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var node proto.Node
	if err := json.NewDecoder(req.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid node registration", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	node.ID = strings.ToLower(node.ID)
	node.LastSeen = time.Now()
	m.nodes[node.ID] = &node

	log.Printf("[INFO] [NODE] Uzel zaregistrován: %s (%s)", node.ID, node.Address)
	w.WriteHeader(http.StatusOK)
}

// handleHeartbeat aktualizuje metriky uzlu (bez přepisování rezervací)
func (m *Master) handleHeartbeat(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxControlPayload)
	var hb proto.Heartbeat
	if err := json.NewDecoder(req.Body).Decode(&hb); err != nil {
		http.Error(w, "Neplatný heartbeat", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	nodeID := strings.ToLower(hb.NodeID)
	node, ok := m.nodes[nodeID]
	if !ok {
		http.Error(w, "Uzel nenalezen", http.StatusNotFound)
		return
	}

	// Aktualizujeme pouze "tvrdá" data z OS
	node.CPUPercent = hb.CPUPercent
	node.MemoryPercent = hb.MemoryPercent
	node.TotalCores = hb.TotalCores
	node.TotalMemoryMB = hb.TotalMemoryMB
	node.LastSeen = time.Now()

	w.WriteHeader(http.StatusOK)
}

// handleSubmit přijme novou úlohu a uloží ji do SQLite
func (m *Master) handleSubmit(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, MaxJobPayload)
	var job proto.Job
	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		http.Error(w, "Neplatné zadání úlohy", http.StatusBadRequest)
		return
	}

	if job.ID == "" {
		job.ID = fmt.Sprintf("job-%s", uuid.New().String())
	}

	// Základní validace a defaulty
	if job.CPUCores <= 0 {
		job.CPUCores = 1
	}
	if job.MemoryMB <= 0 {
		job.MemoryMB = 128
	}

	_, err := m.db.Exec(`
		INSERT INTO jobs(id, command, cpu_cores, memory_mb, priority, status) 
		VALUES (?, ?, ?, ?, ?, 'pending')`,
		job.ID, job.Command, job.CPUCores, job.MemoryMB, job.Priority,
	)

	if err != nil {
		log.Printf("[ERR] [DB] Chyba při ukládání úlohy: %v", err)
		http.Error(w, "Chyba databáze", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] [JOB] Úloha přijata a uložena do fronty: %s", job.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"id": job.ID, "status": "pending"})
}

func (m *Master) initDB() {
	query := `CREATE TABLE IF NOT EXISTS jobs (
    id          TEXT     PRIMARY KEY,
    command     TEXT     NOT NULL,
    cpu_cores   INTEGER  NOT NULL DEFAULT 1,
    memory_mb   INTEGER  NOT NULL DEFAULT 128,
    priority    INTEGER  NOT NULL DEFAULT 0,
    status      TEXT     NOT NULL DEFAULT 'pending',
    node_id     TEXT     DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at  DATETIME,
    finished_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_jobs_pending ON jobs (status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_jobs_node ON jobs (node_id) WHERE status = 'running';`

	if _, err := m.db.Exec(query); err != nil {
		log.Fatalf("[ERR] [DB] Chyba při vytváření databáze: %v", err)
	}
}

func (m *Master) getNextPendingJob() (*proto.Job, error) {
	query := `SELECT id, command, COALESCE(cpu_cores, 1), COALESCE(memory_mb, 128), COALESCE(priority, 0) FROM jobs 
              WHERE status = 'pending' ORDER BY priority DESC, created_at ASC LIMIT 1`

	var job proto.Job
	err := m.db.QueryRow(query).Scan(&job.ID, &job.Command, &job.CPUCores, &job.MemoryMB, &job.Priority)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &job, err
}

func (m *Master) runSchedulingLoop() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		job, err := m.getNextPendingJob()
		if err != nil {
			log.Printf("[ERR] [DB] Chyba při čtení úlohy z databáze: %v", err)
			continue
		}
		if job == nil {
			continue
		}

		m.mu.Lock()
		node := m.scheduler.SelectNode(job, m.nodes)
		if node == nil {
			log.Printf("[INFO] [SCHED] Nedostatek zdrojů pro úlohu %s (%d CPU, %d MB RAM)", job.ID, job.CPUCores, job.MemoryMB)
			m.mu.Unlock()
			continue
		}

		// Optimistická rezervace
		node.AvailableCores -= job.CPUCores
		node.FreeMemoryMB -= job.MemoryMB
		m.mu.Unlock()

		_, err = m.db.Exec("UPDATE jobs SET status = 'running', node_id = ?, started_at = ? WHERE id = ?",
			node.ID, time.Now(), job.ID)
		if err != nil {
			log.Printf("[ERR] [DB] Chyba databáze při aktualizaci úlohy %s: %v", job.ID, err)
			// Vrátit zdroje, pokud update selže
			m.mu.Lock()
			if n, ok := m.nodes[strings.ToLower(node.ID)]; ok {
				n.AvailableCores += job.CPUCores
				n.FreeMemoryMB += job.MemoryMB
			}
			m.mu.Unlock()
			continue
		}

		go func(n *proto.Node, j *proto.Job) {
			data, _ := json.Marshal(j)
			client := http.Client{Timeout: 5 * time.Second}
			res, err := client.Post(n.Address+"/run", "application/json", bytes.NewBuffer(data))

			if err != nil {
				log.Printf("[ERR] [SCHED] Selhalo odeslání úlohy %s na uzel %s: %v", j.ID, n.ID, err)
				m.requeueJob(n, j)
				return
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusAccepted {
				log.Printf("[ERR] [SCHED] Uzel %s odmítl úlohu %s se stavem %d", n.ID, j.ID, res.StatusCode)
				m.requeueJob(n, j)
				return
			}

			log.Printf("[INFO] [SCHED] Úloha %s úspěšně odeslána na uzel %s", j.ID, n.ID)
		}(node, job)

	}
}

func (m *Master) runHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, node := range m.nodes {
			if now.Sub(node.LastSeen) > 30*time.Second {
				log.Printf("[WARN] [NODE] Uzel %s je offline, odstraňuji jej.", id)
				delete(m.nodes, id)
				// Restartování úloh z tohoto uzlu
				m.db.Exec("UPDATE jobs SET status = 'pending', node_id = '' WHERE node_id = ? AND status = 'running'", id)
			}
		}
		m.mu.Unlock()
	}
}

func (m *Master) requeueJob(n *proto.Node, j *proto.Job) {
	m.mu.Lock()
	if node, ok := m.nodes[strings.ToLower(n.ID)]; ok {
		node.AvailableCores += j.CPUCores
		node.FreeMemoryMB += j.MemoryMB
	}
	m.mu.Unlock()
	m.db.Exec("UPDATE jobs SET status = 'pending', node_id = '' WHERE id = ?", j.ID)
}
