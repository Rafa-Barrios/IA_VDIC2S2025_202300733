package report

import "strings"

type Result struct {
	Mensaje string
	Error   bool
}

// Rep es el punto de entrada para el comando REP
func Rep(params map[string]string) Result {

	// =========================
	// Validación de parámetros
	// =========================
	id, okID := params["id"]
	name, okName := params["name"]
	nameReport, okFile := params["namereport"]

	if !okID || !okName || !okFile {
		return Result{
			Mensaje: "Error: parámetros obligatorios faltantes (-id, -name, -namereport)",
			Error:   true,
		}
	}

	name = strings.ToLower(strings.TrimSpace(name))

	// =========================
	// Enrutamiento de reportes
	// =========================
	switch name {

	case "mbr":
		msg, err := RepMBR(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	default:
		return Result{
			Mensaje: "Error: tipo de reporte no válido",
			Error:   true,
		}
	}
}
