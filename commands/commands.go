package commands

import (
	"cluster-scheduler/agent"
	"cluster-scheduler/config"
	"cluster-scheduler/master"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// RunMaster zpracuje parametry pro master uzel
func RunMaster(args []string) {
	cfg := config.Default()
	masterCmd := flag.NewFlagSet("master", flag.ExitOnError)

	port := masterCmd.String("p", cfg.ListenAddr, "Port, na kterém master poslouchá (např. :8080)")

	masterCmd.Parse(args)

	fmt.Printf("🚀 Spouštím master na portu %s\n", *port)
	cfg.ListenAddr = *port
	m := master.New(cfg)
	m.Start()

	// Graceful shutdown - čekáme na CTRL+C nebo SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	m.Stop()
	fmt.Println("👋 Master ukončen.")
}

// RunAgent zpracuje parametry pro agent uzel
func RunAgent(args []string) {
	agentCmd := flag.NewFlagSet("agent", flag.ExitOnError)

	id := agentCmd.String("id", "agent01", "Unikátní ID agenta")
	masterURL := agentCmd.String("master", "http://localhost:8080", "URL adresa mastera")
	port := agentCmd.Int("p", 9001, "Port, na kterém agent poslouchá")

	agentCmd.Parse(args)

	fmt.Printf("📡 Spouštím agenta %s (Master: %s, Port: %d)\n", *id, *masterURL, *port)
	a := agent.New(*id, *masterURL, *port)
	a.Start()

	// Graceful shutdown i pro Agenta
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	a.Stop()
	fmt.Println("👋 Agent ukončen.")
}

// PrintUsage vypíše nápovědu
func PrintUsage() {
	fmt.Println("Použití: scheduler <příkaz> [parametry]")
	fmt.Println("\nPříkazy:")
	fmt.Println("  master  - Spustí řídicí uzel clusteru")
	fmt.Println("  agent   - Spustí výpočetní uzel (agent)")
	fmt.Println("\nParametry pro master:")
	fmt.Println("  -p      Port (výchozí :8080)")
	fmt.Println("\nParametry pro agent:")
	fmt.Println("  -id     Unikátní identifikátor (výchozí agent01)")
	fmt.Println("  -master URL mastera (např. http://192.168.0.2:8080)")
	fmt.Println("  -p      Lokální port agenta (výchozí 9001)")
}

