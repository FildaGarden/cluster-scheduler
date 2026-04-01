package main

import (
	"cluster-scheduler/agent"
	"cluster-scheduler/config"
	"cluster-scheduler/master"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.Default()

	switch os.Args[1] {

	case "master":
		m := master.New(cfg.ListenAddr)
		m.Start()
		fmt.Println("Spoustim master na: " + cfg.ListenAddr)

	case "agent":
		if len(os.Args) < 5 {
			fmt.Println("Chyba: agent vyžaduje id, masterURL a port")
			fmt.Println("  scheduler agent <id> <masterURL> <port>")
			os.Exit(1)
		}
		port, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Port musi byt int")
			os.Exit(1)
		}

		a := agent.New(os.Args[2], os.Args[3], port)
		a.Start()

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Neznamy prikaz %s\n", os.Args[1])
		printUsage()
		os.Exit(1)

	}

}

func printUsage() {
	fmt.Println("Použití:")
	fmt.Println("  scheduler master                          - spustí master uzel")
	fmt.Println("  scheduler agent <id> <masterURL> <agentPort>  - spustí agenta")
	fmt.Println()
	fmt.Println("Příklady:")
	fmt.Println("  scheduler master")
	fmt.Println("  scheduler agent server01 http://localhost:8080 9001")
}
