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

	name = strings.ToLower(name)
	name = strings.Trim(name, " \n\r\t")
	name = strings.Split(name, " ")[0]

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

	case "disk":
		msg, err := RepDISK(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	case "inode":
		msg, err := RepInode(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	case "block":
		msg, err := RepBlock(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	case "bm_inode":
		msg, err := RepBMInode(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	case "bm_bloc":
		msg, err := RepBMBlock(id, nameReport)
		return Result{
			Mensaje: msg,
			Error:   err,
		}

	case "sb":
		msg, err := RepSB(id, nameReport)
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
