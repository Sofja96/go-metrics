package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit call in main"
}

func someOtherFunction() {
	// Здесь вызов os.Exit допустим, так как это не функция main
	os.Exit(1)
}
