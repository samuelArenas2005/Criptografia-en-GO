package ui

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

type Presenter struct{}

func NewPresenter() *Presenter {
	return &Presenter{}
}

// Paso 10: Presentación formateada e interactiva de resultados
func (p *Presenter) ShowResults(result *ExecutionResult) error {
	p.printHeader()
	p.printSummary(result)
	p.printCryptoDetails(result)
	p.printVerdict(result)

	// Exploración interactiva de resultados mediante PromptUI
	return p.interactiveExplorer(result)
}

func (p *Presenter) printHeader() {
	divider := strings.Repeat("=", 70)
	fmt.Println()
	fmt.Println(divider)
	fmt.Println("               REPORTE FINAL DE AUDITORÍA Y SEGURIDAD                 ")
	fmt.Println(divider)
}

func (p *Presenter) printSummary(r *ExecutionResult) {
	fmt.Printf("\n[+] Pasos 1 y 2: Volumen y resumen de eventos generados:\n")
	fmt.Printf("    - Total eventos: %d\n", r.TotalEvents)
	fmt.Printf("    - REGISTRO_NOTA: %d evento(s)\n", r.RegistroCount)
	fmt.Printf("    - MODIFICACION_NOTA: %d evento(s)\n", r.ModificacionCount)

	fmt.Printf("\n[+] Pasos 3 y 4: Integridad local y consolidación:\n")
	for i, ev := range r.Events {
		fmt.Printf("    - Evento %d [%s] Hash: %s...\n", i+1, ev.Tipo, ev.PayloadHash[:24])
	}
	fmt.Printf("    -> Archivo consolidado: %s\n", r.OriginalLogPath)
}

func (p *Presenter) printCryptoDetails(r *ExecutionResult) {
	fmt.Printf("\n[+] Paso 5: Hash del log en claro (original):\n")
	fmt.Printf("    -> SHA-256: %s\n", r.OriginalHash)

	fmt.Printf("\n[+] Paso 6: Cifrado en reposo (AES-256-GCM):\n")
	fmt.Printf("    -> Archivo cifrado: %s\n", r.EncryptedLogPath)
	fmt.Printf("    -> Almacén de claves: %s\n", r.KeystorePath)

	fmt.Printf("\n[+] Paso 7: Hash del archivo cifrado:\n")
	fmt.Printf("    -> SHA-256: %s\n", r.EncryptedHash)

	fmt.Printf("\n[+] Paso 8: Descifrado completado:\n")
	fmt.Printf("    -> Archivo recuperado: %s\n", r.DecryptedLogPath)

	fmt.Printf("\n[+] Paso 9: Comparación de integridad bit a bit:\n")
	fmt.Printf("    - Hash Original:   %s\n", r.OriginalHash)
	fmt.Printf("    - Hash Descifrado: %s\n", r.DecryptedHash)
}

func (p *Presenter) printVerdict(r *ExecutionResult) {
	divider := strings.Repeat("-", 70)
	fmt.Println(divider)
	if r.IsIntegrityValid {
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
}

func (p *Presenter) interactiveExplorer(r *ExecutionResult) error {
	for {
		prompt := promptui.Select{
			Label: "¿Desea inspeccionar detalles adicionales?",
			Items: []string{
				"1. Ver detalle completo de eventos individuales",
				"2. Ver lista de artefactos generados en disco",
				"3. Finalizar y salir",
			},
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}

		switch idx {
		case 0:
			fmt.Println("\n--- Detalle de Eventos ---")
			for i, ev := range r.Events {
				fmt.Printf("[%d] ID: %s | Tipo: %s | Fecha: %s\n", i+1, ev.IDEvento, ev.Tipo, ev.Timestamp)
				fmt.Printf("    Profesor: %s | Estudiante: %s | Curso: %s\n", ev.Payload.DocenteID, ev.Payload.EstudianteID, ev.Payload.CursoID)
				if ev.Tipo == "REGISTRO_NOTA" {
					fmt.Printf("    Nota asignada: %.1f\n", ev.Payload.Nota)
				} else {
					fmt.Printf("    Nota anterior: %.1f -> Nueva: %.1f | Motivo: %s\n", ev.Payload.NotaAnterior, ev.Payload.NotaNueva, ev.Payload.Motivo)
				}
				fmt.Printf("    Hash SHA-256: %s\n\n", ev.PayloadHash)
			}
		case 1:
			fmt.Println("\n--- Artefactos Creados en /artifacts ---")
			fmt.Printf("1. %-25s (Log consolidado texto plano)\n", r.OriginalLogPath)
			fmt.Printf("2. %-25s (Hash SHA-256 original)\n", r.OriginalHash)
			fmt.Printf("3. %-25s (Log binario cifrado AES-256-GCM)\n", r.EncryptedLogPath)
			fmt.Printf("4. %-25s (Insumos criptográficos Clave y Nonce)\n", r.KeystorePath)
			fmt.Printf("5. %-25s (Hash SHA-256 del binario cifrado)\n", r.EncryptedHash)
			fmt.Printf("6. %-25s (Log recuperado tras descifrado)\n\n", r.DecryptedLogPath)
		case 2:
			fmt.Println("Ejecución finalizada con éxito.")
			return nil
		}
	}
}
