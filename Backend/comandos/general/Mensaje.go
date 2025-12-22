package general

// Resultado para ejecución interna de comandos
type Resultado struct {
	Mensaje string
	Error   bool
	Salida  SalidaComandoEjecutado
}

// Salida específica para listas de comandos ejecutados
type SalidaComandoEjecutado struct {
	LstComandos []string
}

// Resultado para respuestas de la API (HTTP / Frontend)
type ResultadoAPI struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Helper para construir respuestas API
func ResultadoSalida(message string, isError bool, data interface{}) ResultadoAPI {
	return ResultadoAPI{
		Message: message,
		Error:   isError,
		Data:    data,
	}
}
