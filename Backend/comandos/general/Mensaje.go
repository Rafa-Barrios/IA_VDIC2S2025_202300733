package general

// ============================================
// RESULTADO INTERNO DE EJECUCIÓN DE COMANDOS
// ============================================

// Resultado representa el resultado de ejecutar
// uno o varios comandos a nivel interno (backend).
type Resultado struct {
	// Mensaje general del procesamiento
	Mensaje string

	// Indica si ocurrió un error durante el análisis
	// o la ejecución de los comandos
	Error bool

	// Salida específica generada por los comandos
	Salida SalidaComandoEjecutado
}

// ============================================
// SALIDA DE COMANDOS EJECUTADOS
// ============================================

// SalidaComandoEjecutado contiene la lista de
// mensajes generados por la ejecución real
// de los comandos (uno por línea).
type SalidaComandoEjecutado struct {
	LstComandos []string
}

// ============================================
// RESULTADO PARA RESPUESTAS HTTP / FRONTEND
// ============================================

// ResultadoAPI define el formato estándar
// de respuesta para la API HTTP.
type ResultadoAPI struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ============================================
// HELPER DE RESPUESTA API
// ============================================

// ResultadoSalida construye una respuesta estándar
// para el frontend o clientes HTTP.
func ResultadoSalida(message string, isError bool, data interface{}) ResultadoAPI {
	return ResultadoAPI{
		Message: message,
		Error:   isError,
		Data:    data,
	}
}
