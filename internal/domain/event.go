package domain

// Tipos de eventos auditables soportados en el sistema académico
const (
	TipoRegistroNota     = "REGISTRO_NOTA"
	TipoModificacionNota = "MODIFICACION_NOTA"
)

// Payload representa los datos sensibles de negocio de una calificación académica.
// Los campos de nota se definen como punteros (*float64) con omitempty para que un
// valor legítimo de 0.0 sea serializado y preservado en el JSON (un float64 primitivo
// con omitempty se elimina si es 0.0 al coincidir con su valor por defecto).
type Payload struct {
	DocenteID    string   `json:"docente_id"`
	EstudianteID string   `json:"estudiante_id"`
	CursoID      string   `json:"curso_id"`
	Nota         *float64 `json:"nota,omitempty"`
	NotaAnterior *float64 `json:"nota_anterior,omitempty"`
	NotaNueva    *float64 `json:"nota_nueva,omitempty"`
	Motivo       string   `json:"motivo,omitempty"`
}

// LogEvent representa un registro estructurado de auditoría en el sistema.
type LogEvent struct {
	IDEvento    string  `json:"id_evento"`
	Tipo        string  `json:"tipo"`
	Timestamp   string  `json:"timestamp"`
	Payload     Payload `json:"payload"`
	PayloadHash string  `json:"payload_hash"`
}
