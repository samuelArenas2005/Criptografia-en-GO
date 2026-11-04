package domain

// Contratos e interfaces del dominio
type StorageRepository interface {
	SaveLogsJSON(path string, events []LogEvent) error
	SaveHash(path string, hash string) error
}

type CryptoService interface {
	ComputePayloadHashes(events []LogEvent) ([]LogEvent, error)
	ComputeFileHash(path string) (string, error)
	EncryptLog(plainPath, encPath, keystorePath string) error
	DecryptLog(encPath, keystorePath, decPath string) error
	CompareHashes(h1, h2 string) bool
}

type GeneratorService interface {
	GenerateEvents(n int) ([]LogEvent, error)
}

