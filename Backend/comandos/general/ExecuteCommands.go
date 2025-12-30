package general

import (
	"Proyecto/comandos/commandGroups/disk"
	"Proyecto/comandos/commandGroups/report"

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

func parametrosToMap(params []string) map[string]string {
	m := make(map[string]string)
	for _, p := range params {
		partes := strings.SplitN(p, "=", 2)
		if len(partes) == 2 {
			clave := strings.ToLower(partes[0])
			valor := partes[1]
			m[clave] = valor
		}
	}
	return m
}

func GlobalCom(lista []string) ([]string, int, []string) {

	var errores []string
	var frontendLogs []string
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
				frontendLogs = append(frontendLogs, msg)
			}

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

		case "reports":
			color.Magenta("Administración de reportes: %s", command)

			paramsMap := parametrosToMap(parametros)

			result := report.Rep(paramsMap)

			if result.Error {
				msgError := "[ERROR] " + result.Mensaje
				color.Red(msgError)

				errores = append(errores, result.Mensaje)
				frontendLogs = append(frontendLogs, msgError)
				contErrores++
			} else {
				frontendLogs = append(frontendLogs, result.Mensaje)
			}
		}
	}

	return errores, contErrores, frontendLogs
}
