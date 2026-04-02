# 📚 Studijní zdroje a literatura

Tento dokument slouží jako seznam klíčových zdrojů pro teoretickou část bakalářské práce i pro technickou implementaci systému.

---

## 🏗️ Architektura distribuovaných systémů
*Základy pro pochopení komunikace, konzistence a správy uzlů.*

*   **Designing Data-Intensive Applications (Martin Kleppmann)** – "Bible" moderních distribuovaných systémů. Klíčové kapitoly: *Replication*, *Partitioning*, *Distributed Content*.
*   **Distributed Systems: Principles and Paradigms (Andrew S. Tanenbaum)** – Akademický standard pro teorii RPC, message passing a modely konzistence.
*   **Google Borg Paper** ([Research PDF](https://research.google/pubs/pub43438/)) – Článek o systému, ze kterého vznikl Kubernetes. Zásadní pro pochopení správy clusterů v obrovském měřítku.

## 🧠 Algoritmy plánování (Scheduling)
*Logika rozhodování o tom, kde a kdy se úloha spustí.*

*   **Slurm Documentation: [Multifactor Priority](https://slurm.schedmd.com/priority_multifactor.html)** – Detailní popis prioritizace úloh (Fair-share, QOS, Age).
*   **Backfill Scheduling** – Strategie pro vyplňování "děr" v časovém plánu. Výborný zdroj je dokumentace Slurmu nebo články od *Drora G. Feitelsona*.
*   **Fair-share Scheduling** – Algoritmy pro spravedlivé rozdělení výkonu mezi uživatele (např. Hierarchical Fair Service Curve - HFSC).

## 🐹 Go (Golang) & Concurrency
*Efektivní a bezpečné programování v systémovém jazyce.*

*   **Concurrency in Go (Katherine Cox-Buday)** – Detailní rozbor patternů (Worker Pools, Fan-in/out) a bezpečná práce s kanály.
*   **Effective Go** ([go.dev/doc/effective_go](https://go.dev/doc/effective_go)) – Oficiální průvodce psaním idiomatického kódu.
*   **Go Blog: Share Memory By Communicating** – Filozofie Go: "Nepředávejte data sdílením paměti, sdílejte paměť předáváním dat."

## 📦 Kontejnerizace a izolace (Linux Internals)
*Zabezpečení a oddělení úloh na jednom uzlu.*

*   **Linux Cgroups (Control Groups)** – Dokumentace jádra a články na LWN.net o tom, jak omezit CPU/RAM na úrovni OS.
*   **Apptainer (Singularity) User Guide** ([apptainer.org](https://apptainer.org/)) – Standard pro kontejnery v HPC prostředí. Důležité pro srovnání bezpečnosti s Dockerem.
*   **Docker: Runtime resource constraints** – Oficiální dokumentace k omezování prostředků kontejnerů.

## 📊 Monitoring a telemetrie
*Sledování stavu clusteru v reálném čase.*

*   **Prometheus: Up & Running (Brian Brazil)** – Jak navrhovat metriky, rozdíl mezi Counter, Gauge a Histogram.
*   **The Four Golden Signals** (Google SRE Book) – Co sledovat: *Latency, Traffic, Errors, Saturation*.

---
> [!TIP]
> **Pro citace v BP:** Prioritizujte Kleppmanna, Tanenbauma a Google Borg Paper. Tyto zdroje mají nejvyšší akademickou váhu u obhajoby.
