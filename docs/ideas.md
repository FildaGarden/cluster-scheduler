# 💡 Ideas & Future Work: HPC Cluster Scheduler

Tento dokument slouží jako zásobník nápadů pro rozšíření systému. Nejsou povinné pro základní funkčnost, ale mohou výrazně zvýšit technickou úroveň bakalářské práce.

---

## 🏗️ Architektura & Perzistence

### 🗄️ Perzistentní úložiště (SQLite)
- **Problém:** Při pádu nebo restartu Mastera se ztratí fronta úloh v paměti.
- **Idea:** Použít lehkou SQL databázi (SQLite) pro ukládání stavu úloh a uzlů.
- **Status:** *Later phase / Optional*

### 🔐 Bezpečná komunikace (mTLS)
- **Problém:** Komunikace mezi Masterem a Agenty probíhá v plain HTTP (nešifrovaně).
- **Idea:** Implementovat vzájemné ověřování pomocí certifikátů (Mutual TLS).
- **Status:** *Low priority*

---

## 🧠 Plánování (Scheduling)

### ⚖️ Fair-Share Algoritmus
- **Idea:** Zamezit tomu, aby jeden uživatel zahltil celý cluster. Scheduler bude upřednostňovat ty uživatele, kteří v poslední době spotřebovali nejméně zdrojů.

### ⏳ Backfilling
- **Idea:** Vyplňování "děr" v časovém plánu clusteru menšími úlohami, které stihnou doběhnout před začátkem velké rezervované úlohy.

---

## 🛡️ Izolace & Bezpečnost

### 📦 Kontejnerizace (Singularity/Docker)
- **Idea:** Místo spouštění `sh -c` příkazů přímo na OS spouštět úlohy v izolovaných kontejnerech. V HPC prostředí preferovat **Singularity (Apptainer)** kvůli bezpečnosti.

### 🛡️ Resource Limits (Cgroups)
- **Idea:** Využít Linuxové `cgroups` k tomu, aby Agent mohl natvrdo omezit CPU a RAM pro konkrétní úlohu a zabránit tak shození celého uzlu.

---

## 📊 Monitoring & UI

### 🐚 Pokročilé CLI (TUI)
- **Idea:** Vytvořit interaktivní terminálové rozhraní (např. pomocí knihovny `bubbletea`), které by v reálném čase ukazovalo stav clusteru bez nutnosti otevírat Grafanu.

### 📁 Log Streaming
- **Idea:** Implementovat "live" sledování výstupu úlohy (`tail -f` styl), kdy Agent posílá `stdout` Masterovi v reálném čase přes WebSockets nebo gRPC stream.

---
> [!TIP]
> **Doporučení:** Tyto nápady jsou skvělým materiálem pro kapitolu „Budoucí rozvoj“ v textové části BP, i kdyby nebyly všechny implementovány.






























### PROMPT

Pokračujeme v práci na mé bakalářské práci: 'HPC Cluster Scheduler v Go'. Podívej se do složky docs/, konkrétně na harmonogram.md, poznatky.md a ideas.md, abys pochopil aktuální stav
  projektu. Máme hotovou Fázi 1 (reálný monitoring přes gopsutil, heartbeat, automatickou detekci IP a testování v síti). Dnes se chceme zaměřit na Fázi 2: Implementaci fronty úloh (Queue) v
  Masterovi a základní Scheduler loop (FIFO).
