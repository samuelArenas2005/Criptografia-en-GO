package storage

import (
	"taller-cripto-go/internal/domain"
)

// SaveLogsJSON persiste la lista de eventos en formato JSON estructurado.
func SaveLogsJSON(path string, events []domain.LogEvent) error {
	return nil
}

// SaveHash guarda el hash en un archivo de texto en disco.
func SaveHash(path string, hash string) error {
	return nil
}

