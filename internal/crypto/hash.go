package crypto

import (
	"taller-cripto-go/internal/domain"
)

// ComputePayloadHashes calcula el hash SHA-256 del payload de cada evento.
func ComputePayloadHashes(events []domain.LogEvent) ([]domain.LogEvent, error) {
	return events, nil
}

// ComputeFileHash calcula el hash SHA-256 de un archivo en disco.
func ComputeFileHash(path string) (string, error) {
	return "", nil
}

// CompareHashes compara dos cadenas hexadecimales de hash.
func CompareHashes(h1, h2 string) bool {
	return h1 == h2 && h1 != ""
}

