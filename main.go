package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "master":
		runMaster(os.Args[2:])
	case "agent":
		runAgent(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Neznámý příkaz %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
