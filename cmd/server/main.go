package main

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/server/handlers"
	"log"
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
	s := handlers.New()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
