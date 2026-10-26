package ui

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/manifoldco/promptui"

	"taller-cripto-go/internal/crypto"
	"taller-cripto-go/internal/domain"
	"taller-cripto-go/internal/generator"
	"taller-cripto-go/internal/storage"
)

// Estructura de consolidación de resultados para la capa de presentación
type ExecutionResult struct {
	TotalEvents       int
	RegistroCount     int
	ModificacionCount int
	Events            []domain.LogEvent
	OriginalLogPath   string
	OriginalHash      string
	EncryptedLogPath  string
	EncryptedHash     string
	KeystorePath      string
	DecryptedLogPath  string
	DecryptedHash     string
	IsIntegrityValid  bool
}

func RunMenu() error {
	menuSelect := promptui.Select{
		Label: "Seleccione una opción",
		Items: []string{
			"1. Ejecutar flujo completo de protección y auditoría",
			"2. Salir de la aplicación",
		},
	}

	idx, _, err := menuSelect.Run()
	if err != nil {
		return err
	}

	if idx == 1 {
		fmt.Println("Saliendo del programa...")
		return nil
	}

	// Paso 1: Definición del volumen de eventos (N)
	n, err := promptEventVolume()
	if err != nil {
		return err
	}

	// Definición de rutas de los artefactos
	originalLogPath := filepath.Join("artifacts", "logs.json")
	originalHashPath := filepath.Join("artifacts", "logs_original.hash")
	encryptedLogPath := filepath.Join("artifacts", "logs.enc")
	keystorePath := filepath.Join("artifacts", "keystore.json")
	encryptedHashPath := filepath.Join("artifacts", "logs_cifrado.hash")
	decryptedLogPath := filepath.Join("artifacts", "logs_descifrado.json")

	// Paso 2: Generación aleatoria de eventos 
	events, err := generator.GenerateEvents(n)
	if err != nil {
		return fmt.Errorf("error en Paso 2 (generator): %w", err)
	}

	// Paso 3: Hash individual por payload canónico
	eventsWithHash, err := crypto.ComputePayloadHashes(events)
	if err != nil {
		return fmt.Errorf("error en Paso 3 (crypto): %w", err)
	}

	// Paso 4: Consolidación del log estructurado en texto plano
	if err := storage.SaveLogsJSON(originalLogPath, eventsWithHash); err != nil {
		return fmt.Errorf("error en Paso 4 (storage): %w", err)
	}

	// Paso 5: Hash del log original en claro
	origHash, err := crypto.ComputeFileHash(originalLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 5 (crypto): %w", err)
	}
	if err := storage.SaveHash(originalHashPath, origHash); err != nil {
		return fmt.Errorf("error en Paso 5 (storage): %w", err)
	}

	// Paso 6: Cifrado en reposo con AES-256-GCM
	if err := crypto.EncryptLog(originalLogPath, encryptedLogPath, keystorePath); err != nil {
		return fmt.Errorf("error en Paso 6 (crypto): %w", err)
	}

	// Paso 7: Hash del log binario cifrado
	encHash, err := crypto.ComputeFileHash(encryptedLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 7 (crypto): %w", err)
	}
	if err := storage.SaveHash(encryptedHashPath, encHash); err != nil {
		return fmt.Errorf("error en Paso 7 (storage): %w", err)
	}

	// Paso 8: Descifrado del log en reposo (capa crypto)
	if err := crypto.DecryptLog(encryptedLogPath, keystorePath, decryptedLogPath); err != nil {
		return fmt.Errorf("error en Paso 8 (crypto): %w", err)
	}

	// Paso 9: Hash del log descifrado y validación de integridad bit a bit 
	decHash, err := crypto.ComputeFileHash(decryptedLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 9 (crypto): %w", err)
	}
	isMatch := crypto.CompareHashes(origHash, decHash)

	// Consolidación de métricas de eventos
	regCount, modCount := 0, 0
	for _, ev := range eventsWithHash {
		if ev.Tipo == "REGISTRO_NOTA" {
			regCount++
		} else {
			modCount++
		}
	}

	result := &ExecutionResult{
		TotalEvents:       n,
		RegistroCount:     regCount,
		ModificacionCount: modCount,
		Events:            eventsWithHash,
		OriginalLogPath:   originalLogPath,
		OriginalHash:      origHash,
		EncryptedLogPath:  encryptedLogPath,
		EncryptedHash:     encHash,
		KeystorePath:      keystorePath,
		DecryptedLogPath:  decryptedLogPath,
		DecryptedHash:     decHash,
		IsIntegrityValid:  isMatch,
	}

	// Paso 10: Presentación interactiva de resultados 
	presenter := NewPresenter()
	return presenter.ShowResults(result)
}

// Paso 1: Helper para solicitud y validación de volumen N
func promptEventVolume() (int, error) {
	prompt := promptui.Prompt{
		Label:   "Paso 1: Ingrese la cantidad de eventos a simular (N)",
		Default: "5",
	}

	input, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(input)
}
