# HPC Cluster Scheduler

Jednoduchý distribuovaný plánovač úloh pro HPC (High Performance Computing) clustery napsaný v Go. Systém umožňuje správu výpočetních uzlů, řazení úloh do fronty s prioritami a jejich následnou distribuci na dostupné agenty.

## Architektura

Systém se skládá ze dvou hlavních komponent:

1.  **Master**: Centrální uzel, který přijímá úlohy přes REST API, ukládá je do SQLite databáze a plánuje jejich spuštění na dostupných agentech.
2.  **Agent**: Výpočetní uzel, který se registruje u mastera, pravidelně zasílá heartbeat se stavem systémových prostředků (CPU, RAM) a vykonává přidělené úlohy v izolovaných procesech.

## Požadavky

*   Go 1.21 nebo novější
*   SQLite3

## Instalace a sestavení

Klonování repozitáře:
```bash
git clone https://github.com/vasi-jmeno/cluster-scheduler
cd cluster-scheduler
```

Stažení závislostí:
```bash
go mod download
```

Sestavení binárky:
```bash
go build -o scheduler main.go
```

## Použití

### Spuštění Master uzlu
Master uzel naslouchá na portu 8080 a spravuje SQLite databázi ve složce `db/`.
```bash
./scheduler master -p :8080
```

### Spuštění Agenta
Agent vyžaduje unikátní ID a URL adresu běžícího mastera.
```bash
./scheduler agent -id node-01 -master http://localhost:8080 -p 9001
```

### Odeslání úlohy
Úlohy lze odesílat pomocí libovolného HTTP klienta.
```bash
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "command": "sleep 30 && echo Success",
    "cpu_cores": 2,
    "memory_mb": 512,
    "priority": 10
  }'
```

## API Endpoints

*   `POST /submit` - Přidání nové úlohy do fronty.
*   `POST /register` - Registrace nového agenta.
*   `POST /heartbeat` - Aktualizace stavu agenta.
*   `POST /update_job` - Hlášení o dokončení úlohy.

## Funkce a vlastnosti

*   **Persistence**: Úlohy jsou trvale uloženy v SQLite, což umožňuje obnovu fronty po restartu mastera.
*   **Plánování**: Podpora strategií FIFO a Least Loaded (rozprostření zátěže).
*   **Monitorování**: Sběr reálných systémových metrik pomocí gopsutil.
*   **Health Check**: Automatické odpojování nedostupných uzlů a opětovné zařazení přerušených úloh do fronty.
