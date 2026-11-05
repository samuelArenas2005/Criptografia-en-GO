package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"taller-cripto-go/internal/domain"
)

// ensureDir verifica y crea de forma idempotente el directorio padre si no existe.
func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error al crear el directorio %s: %w", dir, err)
	}
	return nil
}

// SaveLogsJSON persiste la lista de eventos en formato JSON estructurado con indentación.
func SaveLogsJSON(path string, events []domain.LogEvent) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar eventos a JSON: %w", err)
	}

	// Salto de línea final por convención de archivos de texto
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo %s: %w", path, err)
	}

	return nil
}

// SaveHash guarda la cadena de hash en un archivo de texto plano en disco.
func SaveHash(path string, hash string) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	content := []byte(hash + "\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("error al guardar el hash en %s: %w", path, err)
	}

	return nil
}
