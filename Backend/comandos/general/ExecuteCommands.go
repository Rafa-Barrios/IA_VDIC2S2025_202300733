package general

import (
	"Proyecto/comandos/commandGroups/disk"
	"strings"

	"github.com/fatih/color"
)

var commandGroups = map[string][]string{
	"disk":    {"mkdisk", "fdisk", "rmdisk", "mount", "mounted", "mkfs"},
	"reports": {"rep"},
	"files":   {"mkfile", "mkdir"},
	"cat":     {"cat"},
	"users":   {"login", "logout"},
	"groups":  {"mkgrp", "mkusr"},
}

// Detecta el grupo y el comando exacto
func detectGroup(cmd string) (string, string, bool, string) {

	cmd = strings.TrimSpace(cmd)
	cmdLower := strings.ToLower(cmd)

	for group, cmds := range commandGroups {
		for _, c := range cmds {
			if cmdLower == c ||
				strings.HasPrefix(cmdLower, c+" ") ||
				strings.HasPrefix(cmdLower, c+"-") {

				return group, c, false, ""
			}
		}
	}

	return "", "", true, "Comando no reconocido"
}

// ======================================================
// EJECUTA COMANDOS
// - Consola: mensajes técnicos
// - Frontend: SOLO mensajes finales claros
// ======================================================
func GlobalCom(lista []string) ([]string, int, []string) {

	var errores []string
	var frontendLogs []string // 🔥 SOLO lo que verá el frontend
	contErrores := 0

	for _, comm := range lista {

		comm = strings.TrimSpace(comm)
		if comm == "" {
			continue
		}

		group, command, blnError, strError := detectGroup(comm)
		if blnError {
			msgError := "[ERROR] " + strError
			color.Red(msgError)

			errores = append(errores, strError)
			frontendLogs = append(frontendLogs, msgError)
			contErrores++
			continue
		}

		parametros := ObtenerParametros(comm)

		switch group {

		// =========================
		// DISK
		// =========================
		case "disk":
			color.Cyan("Administración de discos: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				msgError := "[ERROR] " + msg
				color.Red(msgError)

				errores = append(errores, msg)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else if msg != "" {
				color.Cyan(msg)
				frontendLogs = append(frontendLogs, msg) // ✅ SOLO mensaje final
			}

		// =========================
		// GROUPS
		// =========================
		case "groups":
			color.White("Administración de grupos: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				msgError := "[ERROR] " + msg
				color.Red(msgError)

				errores = append(errores, msg)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else if msg != "" {
				frontendLogs = append(frontendLogs, msg)
			}

		// =========================
		// USERS
		// =========================
		case "users":
			color.Yellow("Administración de usuarios: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				msgError := "[ERROR] " + msg
				color.Red(msgError)

				errores = append(errores, msg)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else if msg != "" {
				frontendLogs = append(frontendLogs, msg)
			}

		// =========================
		// FILES
		// =========================
		case "files":
			color.Green("Administración de archivos: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				msgError := "[ERROR] " + msg
				color.Red(msgError)

				errores = append(errores, msg)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else if msg != "" {
				frontendLogs = append(frontendLogs, msg)
			}

		// =========================
		// CAT
		// =========================
		case "cat":
			color.Blue("Comando CAT: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				msgError := "[ERROR] " + msg
				color.Red(msgError)

				errores = append(errores, msg)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else if msg != "" {
				frontendLogs = append(frontendLogs, msg)
			}

		// =========================
		// REPORTS (si luego devuelve msg)
		// =========================
		case "reports":
			color.Magenta("Administración de reportes: %s", command)
		}
	}

	return errores, contErrores, frontendLogs
}
