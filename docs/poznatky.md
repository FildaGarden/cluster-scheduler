# Zjisteni IP Agenta

```go
func getLocalIP() string {
       // Připoj se na 8.8.8.8 — Google DNS
       // Neposílá žádná data, jen zjistí jaké rozhraní by použil
       conn, err := net.Dial("udp", "8.8.8.8:80")
       if err != nil {
           return "localhost"
       }
       defer conn.Close()

       // Zjisti lokální adresu tohoto spojení
       localAddr := conn.LocalAddr().(*net.UDPAddr)
       return localAddr.IP.String()
   }}
```

   1. Žádný provoz: UDP spojení je "bezstavové". Nevysíláš žádná data na internet, jen se ptáš operačního systému: "Kdybych chtěl poslat paket ven, kterou IP
      adresu bys mi na tu obálku napsal jako zpáteční?"
   2. Spolehlivost: Vyhneš se těm 127.0.0.1 adresám, které by ti vrátilo prosté čtení seznamu rozhraní.
   3. Automatizace: Uživatel nemusí při startu agenta ručně zadávat IP adresu, což je pro "real-time monitoring systém" (tvé téma) mnohem elegantnější.

  Malé doporučení pro tvoji práci:
  Pokud to budeš psát do bakalářky, můžeš zmínit, že:
   * Tato metoda vyžaduje, aby stroj měl cestu k internetu (nebo aspoň k té IP 8.8.8.8).
   * V izolované síti (bez internetu) by tato funkce vrátila chybu a spadla by zpět na localhost.
       * Vylepšení: Místo 8.8.8.8 (Google) bys tam teoreticky mohl dát IP adresu tvého Mastera, protože to je ten cíl, se kterým ten agent stoprocentně komunikovat chce.

---

# Monitoring: PUSH vs. PULL Model

V projektu kombinujeme oba přístupy pro dosažení maximální efektivity a stability.

1. **PUSH Model (Agent -> Master):**
   - **Využití:** Heartbeat a registrace uzlů.
   - **Výhoda:** Master se o novém agentovi dozví okamžitě bez nutnosti skenování sítě.
   - **Nevýhoda:** Při velkém počtu agentů hrozí zahlcení Mastera (řešeno rozumným intervalem heartbeatu).

2. **PULL Model (Prometheus -> Master):**
   - **Využití:** Sběr metrik pro Grafanu.
   - **Výhoda:** Prometheus si sám řídí zátěž (stahuje data jen když potřebuje) aniž by ohrozil stabilitu Mastera.
   - **Nevýhoda:** Vyžaduje endpoint (`/metrics`), který musí Master neustále udržovat aktuální.

**Závěr pro BP:**
Zvolený **hybridní model** využívá Push pro rychlou orchestraci a registraci uzlů (Agent -> Master) a Pull pro robustní dlouhodobý monitoring. To odpovídá moderním standardům v distribuovaných systémech.

---

# Volba technologie: Proč Go (Golang)?

Argumentace pro obhajobu volby jazyka Go proti tradičním akademickým standardům (C++, Java, Python).

### 1. Srovnání s C/C++ (Bezpečnost & Paralelismus)
*   **Bezpečnost:** Go eliminuje časté chyby v C++ (např. buffer overflow, memory leaks) díky automatické správě paměti a bezpečné práci s ukazateli.
*   **Paralelismus:** Go používá **Gorutiny** (lehká "vlákna", startovní stack jen 2KB). Oproti tisícům vláken v C++, které by zahltily RAM a OS scheduler, Go zvládá miliony gorutin na jediném stroji.

### 2. Srovnání s Javou (Deployment & Režie)
*   **Nulové závislosti:** Go produkuje **statickou binárku**. Agent v Go nevyžaduje instalaci žádného runtime (jako JRE), což zjednodušuje nasazení na stovky uzlů (copy-and-run).
*   **Minimální footprint:** Java vyžaduje stovky MB RAM jen pro JVM. Go agent spotřebovává jednotky MB RAM, což je v HPC kritické – výkon uzlu má patřit výpočtu, nikoliv monitoringu.

### 3. Srovnání s Pythonem (Výkon & Multicore)
*   **Nativní výkon:** Go je kompilovaný jazyk, jehož výkon je srovnatelný s C++. Python je interpretovaný a pro masivní plánování tisíců úloh za sekundu příliš pomalý.
*   **Paralelní běh:** Na rozdíl od Pythonu (limitovaného GIL - Global Interpreter Lock), Go nativně využívá všechna jádra procesoru.

---

# Tok dat a architektura nasazení

Vizualizace cesty úlohy od odeslání až po zobrazení v monitoringu.

```text
Uživatel (CLI) 
      |
      | (1) HTTP POST /submit (JSON)
      v
Master (Go proces) <------- (3) HTTP GET /metrics (PULL) ------ Prometheus (Docker)
      |                                                            ^
      | (2) HTTP POST /run                                         |
      v                                                      (4) Data Source
Agent (Go proces)                                                  |
      |                                                         Grafana (Docker)
      |---> (Provádí výpočet)                                      |
                                                                   v
                                                            Uživatel (Dashboard)
```

---

# Bezpečnost a izolace Master uzlu 

Klíčové principy pro ochranu řídicí vrstvy:
1. **Separace:** Proces Mastera běží pod dedikovaným systémovým účtem (izolace od uživatelů).
2. **Rozhraní:** Uživatel přistupuje k Masteru výhradně přes CLI (HTTP API), nikoliv přímou manipulací s procesy či daty.
3. **Resource Control:** Omezení systémových zdrojů (cgroups) pro interaktivní SSH uživatele, aby jejich práce neovlivnila stabilitu scheduleru.
