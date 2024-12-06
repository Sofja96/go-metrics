package staticlint

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

// Тестируем ExitAnalyzer
func TestExitAnalyzer(t *testing.T) {

	// Запускаем анализатор на тестовых данных
	analysistest.Run(t, analysistest.TestData(), ExitAnalyzer, "./...")
}
