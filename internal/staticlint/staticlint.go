package staticlint

import (
	"github.com/kisielk/errcheck/errcheck"
	"github.com/mdempsky/maligned/passes/maligned"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

/*
Package staticlint содержит пользовательский анализатор ExitAnalyzer.

ExitAnalyzer запрещает использование os.Exit в функции main пакета main.

Принцип работы:
1. Анализатор проверяет, является ли анализируемый файл частью пакета main.
2. Находит функцию main.
3. Проверяет, используется ли os.Exit внутри этой функции.
4. Если обнаружен вызов os.Exit, выдает предупреждение.

Пример использования:
1. Добавьте ExitAnalyzer в свой multichecker.
2. Запустите multichecker для анализа вашего кода.

Сообщение об ошибке:
`запрещено использовать os.Exit в функции main пакета main`
*/

func Run() {
	var analyzers []*analysis.Analyzer
	analyzers = append(
		analyzers,
		printf.Analyzer,       //pkg pass
		shadow.Analyzer,       //pkg pass
		structtag.Analyzer,    //pkg pass
		nilness.Analyzer,      //pkg pass
		unusedresult.Analyzer, //pkg pass
		errcheck.Analyzer,     //public analyzer
		maligned.Analyzer,     //public analyzer
		ExitAnalyzer,          //my analyzer
	)

	checks := map[string]bool{
		"SA":    true,
		"S1008": true,
		"S1016": true,
	}

	for _, v := range staticcheck.Analyzers {
		// добавляем в массив нужные проверки
		if checks[v.Analyzer.Name] {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	multichecker.Main(
		analyzers...,
	)
}
