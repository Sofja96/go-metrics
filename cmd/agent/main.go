package main

import (
	"fmt"

	"github.com/Sofja96/go-metrics.git/internal/agent"
)

const (
	NA string = "N/A"
)

var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
)

func PrintBuildInfo() {
	if len(BuildVersion) == 0 {
		BuildVersion = NA
	}

	if len(BuildCommit) == 0 {
		BuildCommit = NA
	}

	if len(BuildDate) == 0 {
		BuildDate = NA
	}
	fmt.Printf(" Build version: %s\n Build date: %s\n Build commit: %s\n",
		BuildVersion, BuildDate, BuildCommit)
}

func main() {
	PrintBuildInfo()
	err := agent.Run()
	if err != nil {
		panic(fmt.Errorf("error running agent: %w", err))
	}
}
