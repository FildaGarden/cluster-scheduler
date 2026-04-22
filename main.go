package main

import (
	"cluster-scheduler/commands"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		commands.PrintUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "master":
		commands.RunMaster(os.Args[2:])
	case "agent":
		commands.RunAgent(os.Args[2:])
	case "help", "--help", "-h":
		commands.PrintUsage()
		os.Exit(0)
	default:
		fmt.Printf("Neznámý příkaz %s\n", os.Args[1])
		commands.PrintUsage()
		os.Exit(2)
	}
}
