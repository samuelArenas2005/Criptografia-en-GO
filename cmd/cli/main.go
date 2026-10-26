package main

import (
	"fmt"
	"os"

	"taller-cripto-go/internal/ui"
)

func main() {
	// Inicialización y ejecución del menú interactivo CLI
	if err := ui.RunMenu(); err != nil {
		fmt.Fprintf(os.Stderr, "Error durante la ejecución: %v\n", err)
		os.Exit(1)
	}
}
