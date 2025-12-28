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

// Ejecuta lista de comandos
func GlobalCom(lista []string) ([]string, int) {

	var errores []string
	contErrores := 0

	for _, comm := range lista {

		comm = strings.TrimSpace(comm)
		if comm == "" {
			continue
		}

		group, command, blnError, strError := detectGroup(comm)
		if blnError {
			color.Red("[ERROR] %s -> %s", comm, strError)
			errores = append(errores, strError)
			contErrores++
			continue
		}

		parametros := ObtenerParametros(comm)

		switch group {

		case "disk":
			color.Cyan("Administración de discos: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				color.Red("[ERROR] %s", msg)
				errores = append(errores, msg)
				contErrores++
			}

		case "groups":
			color.White("Administración de grupos: %s", command)

			// 🔥 AQUÍ ESTABA EL PROBLEMA
			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				color.Red("[ERROR] %s", msg)
				errores = append(errores, msg)
				contErrores++
			}

		case "users":
			color.Yellow("Administración de usuarios: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				color.Red("[ERROR] %s", msg)
				errores = append(errores, msg)
				contErrores++
			}

		case "reports":
			color.Magenta("Administración de reportes: %s", command)

		case "files":
			color.Green("Administración de archivos: %s", command)

			msg, err := disk.DiskExecuteCommanWithProps(command, parametros)
			if err {
				color.Red("[ERROR] %s", msg)
				errores = append(errores, msg)
				contErrores++
			} else if msg != "" {
				color.Green(msg)
			}

		case "cat":
			color.Blue("Comando CAT")
		}
	}

	return errores, contErrores
}
