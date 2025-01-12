package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/Sofja96/go-metrics.git/internal/server/handlers"
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	PrintBuildInfo()
	s := handlers.New(ctx)

	if err := s.Start(ctx); err != nil {
		log.Fatal(err)
	}
}
