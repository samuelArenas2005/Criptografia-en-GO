package generator

import (
	"encoding/json"
	"strings"
	"testing"

	"taller-cripto-go/internal/domain"
)

// TestRandomGradeRange verifica que las notas generadas estén siempre en el rango [0.0, 5.0]
func TestRandomGradeRange(t *testing.T) {
	for i := 0; i < 500; i++ {
		grade := RandomGrade(0.0, 5.0)
		if grade < 0.0 || grade > 5.0 {
			t.Fatalf("Nota generada fuera de rango [0.0, 5.0]: %f", grade)
		}
	}
}

// TestZeroGradePreservation verifica que una nota de 0.0 NO sea omitida en el JSON serializado.
func TestZeroGradePreservation(t *testing.T) {
	notaCero := 0.0
	payload := domain.Payload{
		DocenteID:    "usr-prof-1001",
		EstudianteID: "usr-est-2001",
		CursoID:      "cur-mat-101",
		Nota:         &notaCero,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Error al serializar payload: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"nota":0`) && !strings.Contains(jsonStr, `"nota":0.0`) {
		t.Fatalf("FALLO CRÍTICO: La nota 0.0 fue eliminada por omitempty en el JSON: %s", jsonStr)
	}
}

// TestGenerateEventsValidation verifica que rechace cantidades n <= 0
func TestGenerateEventsValidation(t *testing.T) {
	_, err := GenerateEvents(0)
	if err == nil {
		t.Fatal("Se esperaba error con n=0, pero se obtuvo nil")
	}

	_, err = GenerateEvents(-10)
	if err == nil {
		t.Fatal("Se esperaba error con n negativo, pero se obtuvo nil")
	}
}

// TestGenerateEventsChainedLogic comprueba la lógica encadenada e imprime cada evento
func TestGenerateEventsChainedLogic(t *testing.T) {
	n := 10 // Generamos 10 eventos representativos para inspeccionar en consola
	events, err := GenerateEvents(n)
	if err != nil {
		t.Fatalf("Error inesperado al generar eventos: %v", err)
	}

	if len(events) != n {
		t.Fatalf("Se esperaban %d eventos, pero se obtuvieron %d", n, len(events))
	}

	// Regla 1: El primer evento obligatoriamente debe ser REGISTRO_NOTA
	if events[0].Tipo != domain.TipoRegistroNota {
		t.Fatalf("El primer evento debe ser REGISTRO_NOTA, se obtuvo: %s", events[0].Tipo)
	}

	t.Logf("\n=======================================================")
	t.Logf("        LISTA DE EVENTOS GENERADOS (TOTAL: %d)          ", len(events))
	t.Logf("=======================================================")

	registeredKeys := make(map[string]float64)

	for i, ev := range events {
		// Validar IDs y timestamp
		if ev.IDEvento == "" {
			t.Errorf("Evento %d sin IDEvento", i)
		}
		if ev.Timestamp == "" {
			t.Errorf("Evento %d sin Timestamp", i)
		}
		if ev.Payload.DocenteID == "" || ev.Payload.EstudianteID == "" || ev.Payload.CursoID == "" {
			t.Errorf("Evento %d con datos de payload incompletos", i)
		}

		key := ev.Payload.EstudianteID + "@" + ev.Payload.CursoID

		switch ev.Tipo {
		case domain.TipoRegistroNota:
			if ev.Payload.Nota == nil || *ev.Payload.Nota < 0.0 || *ev.Payload.Nota > 5.0 {
				t.Errorf("Evento %d REGISTRO con nota inválida o fuera de rango", i)
			}
			registeredKeys[key] = *ev.Payload.Nota

			t.Logf("[%02d] %-17s | Estudiante: %-12s | Curso: %-13s | Nota: %.1f | Docente: %s",
				i+1, ev.Tipo, ev.Payload.EstudianteID, ev.Payload.CursoID, *ev.Payload.Nota, ev.Payload.DocenteID)

		case domain.TipoModificacionNota:
			lastKnownGrade, exists := registeredKeys[key]
			if !exists {
				t.Errorf("Evento %d MODIFICACION intenta modificar una nota que nunca fue registrada", i)
			}
			if ev.Payload.NotaAnterior == nil || *ev.Payload.NotaAnterior != lastKnownGrade {
				t.Errorf("Evento %d MODIFICACION: NotaAnterior no coincide con previa (%f)", i, lastKnownGrade)
			}
			if ev.Payload.NotaNueva == nil || *ev.Payload.NotaNueva == *ev.Payload.NotaAnterior {
				t.Errorf("Evento %d MODIFICACION: NotaNueva es idéntica a NotaAnterior", i)
			}
			if ev.Payload.Motivo == "" {
				t.Errorf("Evento %d MODIFICACION sin motivo", i)
			}

			registeredKeys[key] = *ev.Payload.NotaNueva

			t.Logf("[%02d] %-17s | Estudiante: %-12s | Curso: %-13s | Nota: %.1f -> %.1f | Motivo: %s",
				i+1, ev.Tipo, ev.Payload.EstudianteID, ev.Payload.CursoID, *ev.Payload.NotaAnterior, *ev.Payload.NotaNueva, ev.Payload.Motivo)

		default:
			t.Errorf("Evento %d con tipo desconocido: %s", i, ev.Tipo)
		}
	}

	// También imprimimos el primer evento en formato JSON estructurado como demostración
	jsonDemo, _ := json.MarshalIndent(events[0], "", "  ")
	t.Logf("\n--- Ejemplo JSON estructurado (Evento 1) ---\n%s\n", string(jsonDemo))
}
