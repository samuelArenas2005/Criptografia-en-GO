package domain

// Payload representa los datos sensibles de negocio de una calificación académica.
type Payload struct {
	DocenteID    string  `json:"docente_id"`
	EstudianteID string  `json:"estudiante_id"`
	CursoID      string  `json:"curso_id"`
	Nota         float64 `json:"nota,omitempty"`
	NotaAnterior float64 `json:"nota_anterior,omitempty"`
	NotaNueva    float64 `json:"nota_nueva,omitempty"`
	Motivo       string  `json:"motivo,omitempty"`
}

// LogEvent representa un registro estructurado de auditoría en el sistema.
type LogEvent struct {
	IDEvento    string  `json:"id_evento"`
	Tipo        string  `json:"tipo"`
	Timestamp   string  `json:"timestamp"`
	Payload     Payload `json:"payload"`
	PayloadHash string  `json:"payload_hash"`
}

