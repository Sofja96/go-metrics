package exitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// Тестируем ExitAnalyzer
func TestExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ExitAnalyzer, "./...")
}
