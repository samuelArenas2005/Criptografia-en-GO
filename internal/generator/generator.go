package generator

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	randv2 "math/rand/v2"
	"time"

	"taller-cripto-go/internal/domain"
)

var cursosDisponibles = []string{
	"cur-mat-101",
	"cur-prog-201",
	"cur-cripto-301",
	"cur-bd-102",
	"cur-redes-202",
	"cur-fis-101",
}

var motivosModificacion = []string{
	"Correccion de evaluacion parcial",
	"Revision y recalificacion de taller",
	"Error de digitacion en acta inicial",
	"Sustentacion de reclamo extemporaneo",
	"Recalificacion de proyecto final",
}

// academicRecord mantiene en memoria el estado de notas registradas
// para poder encadenar modificaciones realistas sobre registros previos.
type academicRecord struct {
	docenteID    string
	estudianteID string
	cursoID      string
	currentGrade float64
}

// RandomGrade genera un valor de nota con distribución uniforme entre min y max (0.0 y 5.0),
// redondeado a 1 decimal (ej. 3.5, 4.8) para consistencia con el dominio académico.
func RandomGrade(min, max float64) float64 {
	val := min + randv2.Float64()*(max-min)
	return math.Round(val*10) / 10
}

// generateUUID genera un identificador único universal (UUID v4) compatible con RFC 4122.
func generateUUID() string {
	var b [16]byte
	_, _ = cryptorand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// GenerateEvents simula la creación aleatoria de N eventos de log académico (Paso 2)
// con lógica encadenada: las modificaciones siempre operan sobre registros previos reales.
func GenerateEvents(n int) ([]domain.LogEvent, error) {
	if n <= 0 {
		return nil, fmt.Errorf("el volumen de eventos debe ser mayor a 0 (recibido: %d)", n)
	}

	events := make([]domain.LogEvent, n)
	now := time.Now().UTC()

	// Lista en memoria de calificaciones activas creadas en este lote
	var activeRecords []*academicRecord

	for i := 0; i < n; i++ {
		// Desplazamiento progresivo en tiempo
		eventTime := now.Add(time.Duration(i*30) * time.Second).Format(time.RFC3339)

		var tipo string
		var payload domain.Payload

		// Si no hay registros previos, el evento OBLIGATORIAMENTE es un REGISTRO_NOTA.
		// Si ya existen registros, se decide al azar entre registrar una nueva nota o modificar una existente.
		shouldRegister := len(activeRecords) == 0 || randv2.IntN(2) == 0

		if shouldRegister {
			tipo = domain.TipoRegistroNota
			docenteID := fmt.Sprintf("usr-prof-%04d", randv2.IntN(9000)+1000)
			estudianteID := fmt.Sprintf("usr-est-%04d", randv2.IntN(9000)+1000)
			cursoID := cursosDisponibles[randv2.IntN(len(cursosDisponibles))]
			nota := RandomGrade(0.0, 5.0)

			payload = domain.Payload{
				DocenteID:    docenteID,
				EstudianteID: estudianteID,
				CursoID:      cursoID,
				Nota:         &nota,
			}

			// Se guarda en memoria activa para posibilitar futuras modificaciones encadenadas
			activeRecords = append(activeRecords, &academicRecord{
				docenteID:    docenteID,
				estudianteID: estudianteID,
				cursoID:      cursoID,
				currentGrade: nota,
			})
		} else {
			tipo = domain.TipoModificacionNota
			// Selecciona aleatoriamente una de las calificaciones registradas previamente
			targetIdx := randv2.IntN(len(activeRecords))
			target := activeRecords[targetIdx]

			notaAnterior := target.currentGrade
			notaNueva := RandomGrade(0.0, 5.0)

			// Asegura que la nueva nota sea distinta a la nota actual/anterior
			for notaNueva == notaAnterior {
				notaNueva = RandomGrade(0.0, 5.0)
			}

			// Actualiza la nota del estudiante por si se vuelve a modificar más adelante
			target.currentGrade = notaNueva

			motivo := motivosModificacion[randv2.IntN(len(motivosModificacion))]
			payload = domain.Payload{
				DocenteID:    target.docenteID,
				EstudianteID: target.estudianteID,
				CursoID:      target.cursoID,
				NotaAnterior: &notaAnterior,
				NotaNueva:    &notaNueva,
				Motivo:       motivo,
			}
		}

		events[i] = domain.LogEvent{
			IDEvento:    "evt-" + generateUUID(),
			Tipo:        tipo,
			Timestamp:   eventTime,
			Payload:     payload,
			PayloadHash: "", // Queda vacío aquí porque se calcula en el Paso 3
		}
	}

	return events, nil
}
