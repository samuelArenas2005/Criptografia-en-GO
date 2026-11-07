package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"

	"taller-cripto-go/internal/crypto"
	"taller-cripto-go/internal/domain"
	"taller-cripto-go/internal/generator"
	"taller-cripto-go/internal/storage"
)

// ExecutionResult consolida los resultados para la capa de presentación.
type ExecutionResult struct {
	TotalEvents       int
	RegistroCount     int
	ModificacionCount int
	Events            []domain.LogEvent
	OriginalLogPath   string
	OriginalHashPath  string
	OriginalHash      string
	EncryptedLogPath  string
	EncryptedHashPath string
	EncryptedHash     string
	KeystorePath      string
	DecryptedLogPath  string
	DecryptedHash     string
	IsIntegrityValid  bool
}

// RunMenu despliega el menú principal interactivo de la CLI.
func RunMenu() error {
	for {
		menuSelect := promptui.Select{
			Label: "Seleccione una opción",
			Items: []string{
				"1. Ejecutar flujo completo de protección y auditoría",
				"2. Ejecutar el flujo paso a paso",
				"3. Salir de la aplicación",
			},
		}

		idx, _, err := menuSelect.Run()
		if err != nil {
			return err
		}

		switch idx {
		case 0:
			if err := runFullFlow(); err != nil {
				fmt.Printf("\n[ERROR]: Falló el flujo completo: %v\n\n", err)
			}
		case 1:
			if err := runStepByStepFlow(); err != nil {
				fmt.Printf("\n[ERROR]: Falló el flujo paso a paso: %v\n\n", err)
			}
		case 2:
			fmt.Println("Saliendo de la aplicación...")
			return nil
		}
	}
}

// runFullFlow ejecuta los 10 pasos de manera continua y muestra el reporte consolidado.
func runFullFlow() error {
	divider := strings.Repeat("=", 70)
	fmt.Println()
	fmt.Println(divider)
	fmt.Println("       MODO 1: FLUJO COMPLETO DE PROTECCIÓN Y AUDITORÍA EN REPOSO     ")
	fmt.Println(divider)

	n, err := promptEventVolume()
	if err != nil {
		return err
	}

	originalLogPath := filepath.Join("artifacts", "logs.json")
	originalHashPath := filepath.Join("artifacts", "logs_original.hash")
	encryptedLogPath := filepath.Join("artifacts", "logs.enc")
	keystorePath := filepath.Join("artifacts", "keystore.json")
	encryptedHashPath := filepath.Join("artifacts", "logs_cifrado.hash")
	decryptedLogPath := filepath.Join("artifacts", "logs_descifrado.json")

	// Paso 2: Generación aleatoria
	events, err := generator.GenerateEvents(n)
	if err != nil {
		return fmt.Errorf("error en Paso 2 (generator): %w", err)
	}

	// Paso 3: Hash individual de payload
	eventsWithHash, err := crypto.ComputePayloadHashes(events)
	if err != nil {
		return fmt.Errorf("error en Paso 3 (crypto): %w", err)
	}

	// Paso 4: Consolidación a JSON en disco
	if err := storage.SaveLogsJSON(originalLogPath, eventsWithHash); err != nil {
		return fmt.Errorf("error en Paso 4 (storage): %w", err)
	}

	// Paso 5: Hash del log en claro
	origHash, err := crypto.ComputeFileHash(originalLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 5 (crypto): %w", err)
	}
	if err := storage.SaveHash(originalHashPath, origHash); err != nil {
		return fmt.Errorf("error en Paso 5 (storage): %w", err)
	}

	// Paso 6: Cifrado en reposo (AES-256-GCM)
	if err := crypto.EncryptLog(originalLogPath, encryptedLogPath, keystorePath); err != nil {
		return fmt.Errorf("error en Paso 6 (crypto): %w", err)
	}

	// Paso 7: Hash del archivo cifrado
	encHash, err := crypto.ComputeFileHash(encryptedLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 7 (crypto): %w", err)
	}
	if err := storage.SaveHash(encryptedHashPath, encHash); err != nil {
		return fmt.Errorf("error en Paso 7 (storage): %w", err)
	}

	// Paso 8: Descifrado en reposo
	if err := crypto.DecryptLog(encryptedLogPath, keystorePath, decryptedLogPath); err != nil {
		return fmt.Errorf("error en Paso 8 (crypto): %w", err)
	}

	// Paso 9: Comparación bit a bit
	decHash, err := crypto.ComputeFileHash(decryptedLogPath)
	if err != nil {
		return fmt.Errorf("error en Paso 9 (crypto): %w", err)
	}
	isMatch := crypto.CompareHashes(origHash, decHash)

	regCount, modCount := 0, 0
	for _, ev := range eventsWithHash {
		if ev.Tipo == domain.TipoRegistroNota {
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
		OriginalHashPath:  originalHashPath,
		OriginalHash:      origHash,
		EncryptedLogPath:  encryptedLogPath,
		EncryptedHashPath: encryptedHashPath,
		EncryptedHash:     encHash,
		KeystorePath:      keystorePath,
		DecryptedLogPath:  decryptedLogPath,
		DecryptedHash:     decHash,
		IsIntegrityValid:  isMatch,
	}

	presenter := NewPresenter()
	return presenter.ShowResults(result)
}

// runStepByStepFlow guía al usuario paso a paso con navegación (Continuar, Atrás, Salir).
func runStepByStepFlow() error {
	divider := strings.Repeat("=", 70)
	fmt.Println()
	fmt.Println(divider)
	fmt.Println("             MODO 2: EJECUCIÓN DEL FLUJO PASO A PASO                  ")
	fmt.Println(divider)

	originalLogPath := filepath.Join("artifacts", "logs.json")
	originalHashPath := filepath.Join("artifacts", "logs_original.hash")
	encryptedLogPath := filepath.Join("artifacts", "logs.enc")
	keystorePath := filepath.Join("artifacts", "keystore.json")
	encryptedHashPath := filepath.Join("artifacts", "logs_cifrado.hash")
	decryptedLogPath := filepath.Join("artifacts", "logs_descifrado.json")

	// Paso 1 inicial
	printStepHeader(1, "DEFINICIÓN DEL VOLUMEN DE EVENTOS")
	n, err := promptEventVolume()
	if err != nil {
		return nil
	}

	var events []domain.LogEvent
	var eventsWithHash []domain.LogEvent
	var origHash string
	var encHash string
	var decHash string
	var isMatch bool
	regCount, modCount := 0, 0

	currentStep := 2

	for currentStep >= 2 && currentStep <= 9 {
		switch currentStep {
		case 2:
			printStepHeader(2, "GENERACIÓN ALEATORIA DE EVENTOS")
			if len(events) != n {
				evs, err := generator.GenerateEvents(n)
				if err != nil {
					return fmt.Errorf("error en Paso 2: %w", err)
				}
				events = evs
				eventsWithHash = nil // forzar recálculo
			}

			regCount, modCount = 0, 0
			for _, ev := range events {
				if ev.Tipo == domain.TipoRegistroNota {
					regCount++
				} else {
					modCount++
				}
			}

			fmt.Println("Volumen y resumen de eventos generados:")
			fmt.Printf("    - Total eventos: %d\n", n)
			fmt.Printf("    - REGISTRO_NOTA: %d evento(s)\n", regCount)
			fmt.Printf("    - MODIFICACION_NOTA: %d evento(s)\n\n", modCount)

			nav, err := promptStepNavigation("Paso 2")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 1
				printStepHeader(1, "MODIFICAR VOLUMEN DE EVENTOS (N)")
				newN, err := promptEventVolume()
				if err != nil {
					return nil
				}
				n = newN
				events = nil
				continue
			}
			currentStep = 3

		case 3:
			printStepHeader(3, "HASH DEL PAYLOAD DE CADA OPERACIÓN")
			if eventsWithHash == nil {
				ewh, err := crypto.ComputePayloadHashes(events)
				if err != nil {
					return fmt.Errorf("error en Paso 3: %w", err)
				}
				eventsWithHash = ewh
			}

			fmt.Println("Hash de la operaciones:")
			for i, ev := range eventsWithHash {
				preview := ev.PayloadHash
				if len(preview) > 24 {
					preview = preview[:24] + "..."
				}
				fmt.Printf("    - Evento %d [%s] Hash: %s\n", i+1, ev.Tipo, preview)
			}
			fmt.Println()

			nav, err := promptStepNavigation("Paso 3")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 2
				currentStep = 2
				continue
			}
			currentStep = 4

		case 4:
			printStepHeader(4, "CONSOLIDACIÓN DEL LOG ESTRUCTURADO EN JSON")
			if err := storage.SaveLogsJSON(originalLogPath, eventsWithHash); err != nil {
				return fmt.Errorf("error en Paso 4: %w", err)
			}

			fmt.Printf("Eventos consolidados en %s\n\n", originalLogPath)
			fmt.Println("Parte inicial del documento " + originalLogPath + ":")
			showJSONPreview(originalLogPath, 22)
			fmt.Println()

			nav, err := promptStepNavigation("Paso 4")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 3
				currentStep = 3
				continue
			}
			currentStep = 5

		case 5:
			printStepHeader(5, "HASH DEL LOG EN CLARO (ORIGINAL)")
			h, err := crypto.ComputeFileHash(originalLogPath)
			if err != nil {
				return fmt.Errorf("error en Paso 5: %w", err)
			}
			origHash = h
			if err := storage.SaveHash(originalHashPath, origHash); err != nil {
				return fmt.Errorf("error en Paso 5: %w", err)
			}

			fmt.Println("Paso 5: Hash del log en claro (original):")
			fmt.Printf("    -> SHA-256: %s\n", origHash)
			fmt.Printf("    -> Guardado en: %s\n\n", originalHashPath)

			nav, err := promptStepNavigation("Paso 5")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 4
				currentStep = 4
				continue
			}
			currentStep = 6

		case 6:
			printStepHeader(6, "CIFRADO EN REPOSO (AES-256-GCM)")
			if err := crypto.EncryptLog(originalLogPath, encryptedLogPath, keystorePath); err != nil {
				return fmt.Errorf("error en Paso 6: %w", err)
			}

			fmt.Println("Paso 6: Cifrado en reposo (AES-256-GCM):")
			fmt.Printf("    -> Archivo cifrado: %s\n", encryptedLogPath)
			fmt.Printf("    -> Almacén de claves: %s\n\n", keystorePath)

			nav, err := promptStepNavigation("Paso 6")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 5
				currentStep = 5
				continue
			}
			currentStep = 7

		case 7:
			printStepHeader(7, "HASH DEL ARCHIVO CIFRADO")
			h, err := crypto.ComputeFileHash(encryptedLogPath)
			if err != nil {
				return fmt.Errorf("error en Paso 7: %w", err)
			}
			encHash = h
			if err := storage.SaveHash(encryptedHashPath, encHash); err != nil {
				return fmt.Errorf("error en Paso 7: %w", err)
			}

			fmt.Println("Paso 7: Hash del archivo cifrado:")
			fmt.Printf("    -> SHA-256: %s\n", encHash)
			fmt.Printf("    -> Guardado en: %s\n\n", encryptedHashPath)

			nav, err := promptStepNavigation("Paso 7")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 6
				currentStep = 6
				continue
			}
			currentStep = 8

		case 8:
			printStepHeader(8, "DESCIFRADO DEL LOG EN REPOSO")
			if err := crypto.DecryptLog(encryptedLogPath, keystorePath, decryptedLogPath); err != nil {
				return fmt.Errorf("error en Paso 8: %w", err)
			}

			fmt.Println("Paso 8: Descifrado completado:")
			fmt.Printf("    -> Archivo recuperado: %s\n\n", decryptedLogPath)

			nav, err := promptStepNavigation("Paso 8")
			if err != nil || nav == 2 {
				fmt.Println("Volviendo al menú principal...")
				return nil
			}
			if nav == 1 { // Atrás al Paso 7
				currentStep = 7
				continue
			}
			currentStep = 9

		case 9:
			printStepHeader(9, "COMPARACIÓN DE INTEGRIDAD BIT A BIT")
			h, err := crypto.ComputeFileHash(decryptedLogPath)
			if err != nil {
				return fmt.Errorf("error en Paso 9: %w", err)
			}
			decHash = h
			isMatch = crypto.CompareHashes(origHash, decHash)

			fmt.Println("Paso 9: Comparación de integridad bit a bit:")
			fmt.Printf("    - Hash Original:   %s\n", origHash)
			fmt.Printf("    - Hash Descifrado: %s\n", decHash)

			line := strings.Repeat("-", 70)
			fmt.Println(line)
			if isMatch {
				fmt.Println("[RESULTADO]: COINCIDENCIA EXACTA (MATCH)")
				fmt.Println("[CONCLUSIÓN]: El proceso de cifrado y descifrado preservó la")
				fmt.Println("              integridad y confidencialidad de la información original")
				fmt.Println("              sin alteraciones en reposo.")
			} else {
				fmt.Println("[RESULTADO]: DISCREPANCIA (MISMATCH)")
				fmt.Println("[CONCLUSIÓN]: Se detectó alteración o corrupción en los datos.")
			}
			fmt.Println(strings.Repeat("=", 70))
			fmt.Println()

			// Menú de opciones del Paso 9
			for {
				nav9, err := promptStep9Navigation()
				if err != nil || nav9 == 3 { // Salir de la aplicación
					fmt.Println("Saliendo de la aplicación...")
					os.Exit(0)
				}
				if nav9 == 0 { // Continuar (volver al menú principal)
					fmt.Println("Flujo paso a paso finalizado con éxito. Volviendo al menú principal...")
					return nil
				}
				if nav9 == 1 { // Ver lista de artefactos generados en disco
					printArtifactsList(originalLogPath, originalHashPath, encryptedLogPath, keystorePath, encryptedHashPath, decryptedLogPath)
					continue
				}
				if nav9 == 2 { // Atrás (volver al Paso 8)
					currentStep = 8
					break
				}
			}
		}
	}

	return nil
}

// promptEventVolume solicita y valida la cantidad entera positiva N.
func promptEventVolume() (int, error) {
	validate := func(input string) error {
		val, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("debe ingresar un número entero válido")
		}
		if val <= 0 {
			return fmt.Errorf("la cantidad de eventos debe ser mayor a 0")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Paso 1: Ingrese la cantidad de eventos a simular (N)",
		Default:  "8",
		Validate: validate,
	}

	input, err := prompt.Run()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(input)
}

// promptStepNavigation despliega las opciones de control entre pasos 2 al 8.
func promptStepNavigation(stepName string) (int, error) {
	prompt := promptui.Select{
		Label: fmt.Sprintf("Acciones disponibles en %s", stepName),
		Items: []string{
			"1. Continuar",
			"2. Atrás",
			"3. Salir (al menú principal)",
		},
	}
	idx, _, err := prompt.Run()
	return idx, err
}

// promptStep9Navigation despliega las opciones especiales requeridas para el Paso 9.
func promptStep9Navigation() (int, error) {
	prompt := promptui.Select{
		Label: "Acciones disponibles en Paso 9 (Finalización)",
		Items: []string{
			"1. Continuar (volver al menú principal)",
			"2. Ver lista de artefactos generados en disco",
			"3. Atrás (volver al Paso 8)",
			"4. Salir de la aplicación",
		},
	}
	idx, _, err := prompt.Run()
	return idx, err
}

// printStepHeader imprime de forma destacada el paso actual en la consola.
func printStepHeader(step int, title string) {
	divider := strings.Repeat("=", 70)
	fmt.Println()
	fmt.Println(divider)
	fmt.Printf("   [ESTÁS EN EL PASO %d / 9]: %s\n", step, title)
	fmt.Println(divider)
}

// showJSONPreview imprime las primeras líneas del archivo JSON formateado.
func showJSONPreview(path string, maxLines int) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("    (No fue posible leer el archivo: %v)\n", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	total := len(lines)
	displayCount := total
	if displayCount > maxLines {
		displayCount = maxLines
	}

	divider := strings.Repeat("-", 65)
	fmt.Println("    " + divider)
	for i := 0; i < displayCount; i++ {
		fmt.Println("    " + lines[i])
	}
	if total > maxLines {
		fmt.Printf("    ... [mostrando las primeras %d de %d líneas de %s]\n", maxLines, total, path)
	}
	fmt.Println("    " + divider)
}

// printArtifactsList imprime de forma formateada las rutas de los artefactos generados.
func printArtifactsList(origLog, origHash, encLog, keystore, encHash, decLog string) {
	fmt.Println("\n--- Artefactos Creados en /artifacts ---")
	fmt.Printf("1. %-32s (Log consolidado texto plano)\n", origLog)
	fmt.Printf("2. %-32s (Hash SHA-256 original)\n", origHash)
	fmt.Printf("3. %-32s (Log binario cifrado AES-256-GCM)\n", encLog)
	fmt.Printf("4. %-32s (Insumos criptográficos Clave y Nonce)\n", keystore)
	fmt.Printf("5. %-32s (Hash SHA-256 del binario cifrado)\n", encHash)
	fmt.Printf("6. %-32s (Log recuperado tras descifrado)\n\n", decLog)
}
