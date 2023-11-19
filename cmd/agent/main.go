package main

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/agent"
)

func main() {
	err := agent.Run()
	if err != nil {
		panic(fmt.Errorf("error running agent: %w", err))
	}
}
