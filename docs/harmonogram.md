# 📅 Harmonogram Vývoje: HPC Cluster Scheduler

Tento dokument slouží jako roadmapa pro praktickou část bakalářské práce: *„Návrh a implementace systému pro správu a plánování výpočetních úloh na HPC clusteru s real-time monitoringem“*.

---

## 🏗️ Fáze 1: Jádro Systému & Reálná Data (HOTOVO ✅)
**Cíl:** Rozchodit stabilní komunikaci a sběr skutečných dat z hardwaru.

- [x] **Reálný Monitoring (Agent)**
  - [x] Implementace `metrics/metrics.go` pomocí knihovny `gopsutil`.
  - [x] Agent posílá v heartbeatu skutečné vytížení CPU a volnou RAM.
- [x] **Automatická Síťová Identifikace**
  - [x] Implementace IP resolveru (UDP dial trick).
  - [x] Úspěšné testování na více fyzických zařízeních v LAN.
- [x] **Správa Úloh (Master)**
  - [x] Rozšíření logování pro vizuální kontrolu clusteru v reálném čase.
- [x] **CLI Pomůcky**
  - [x] Vytvoření `scripts/submit.sh` pro odesílání úloh.

---

## 🧠 Fáze 2: Plánování (Scheduler) & Fronta
**Cíl:** Přeměna systému z "okamžitého spouštění" na inteligentní plánovač.

- [ ] **Job Queue (Fronta)**
  - [ ] Úlohy se po přijetí ukládají do DB se stavem `Pending`.
- [ ] **Scheduler Loop**
  - [ ] Background gorutina v Masterovi, která cyklicky prochází frontu.
  - [ ] Implementace algoritmů:
    - [ ] Výběr úloh (**FIFO**, **Priority**).
    - [ ] Výběr uzlů (**FirstAvailable**, **LeastLoaded**).
  - [ ] **Optimistická rezervace:** Zamezení race condition při plánování více úloh najednou.
  - [ ] **Alokace jader:** Přechod z "Real-time CPU" na "Allocated Cores" model.
- [ ] **Multi-node Test**
  - [ ] Testování na fakultním clusteru (3–5 reálných uzlů).

---

## 🐳 Fáze 3: Monitorovací Infrastruktura (Docker)
**Cíl:** Vizualizace nasbíraných dat pro potřeby analýzy v BP.

- [ ] **Prometheus Integration**
  - [ ] Master vystavuje `/metrics` endpoint pro Prometheus.
- [ ] **Docker Deployment**
  - [ ] `docker-compose.yml` pro spuštění **Promethea** a **Grafany**.
- [ ] **Dashboarding**
  - [ ] Tvorba dashboardů v Grafaně (grafy vytížení clusteru, stav fronty).

---

## 🛡️ Fáze 4: Robustnost & Pokročilé Funkce
**Cíl:** Zajištění stability a přidaná hodnota pro výzkumný tým.

- [ ] **Fault Tolerance**
  - [ ] Detekce výpadku agenta a automatické přeplánování úloh.
- [ ] **Log Management**
  - [ ] Zachytávání a vzdálené prohlížení `stdout/stderr` úloh.
- [ ] **Fair-Share Algoritmus** (Volitelné)
  - [ ] Spravedlivé dělení zdrojů mezi uživatele.

---

## 📊 Fáze 5: Analýza & Text BP
**Cíl:** Vyhodnocení a dokumentace.

- [ ] **Zátěžové testy** (Simulace 100+ uzlů v Dockeru).
- [ ] **Sběr grafů** pro praktickou část textu práce.

---
*Poslední aktualizace: 1. dubna 2026*
