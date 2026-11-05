package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"

	"taller-cripto-go/internal/domain"
)

// ComputePayloadHashes calcula el SHA-256 del payload serializado (evita circularidad).
func ComputePayloadHashes(events []domain.LogEvent) ([]domain.LogEvent, error) {
	for i := range events {
		bytes, _ := json.Marshal(events[i].Payload)
		hash := sha256.Sum256(bytes)
		events[i].PayloadHash = hex.EncodeToString(hash[:])
	}
	return events, nil
}

// ComputeFileHash calcula el hash SHA-256 de un archivo en disco.
func ComputeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// CompareHashes compara dos cadenas hexadecimales de hash.
func CompareHashes(h1, h2 string) bool {
	return h1 != "" && h1 == h2
}
