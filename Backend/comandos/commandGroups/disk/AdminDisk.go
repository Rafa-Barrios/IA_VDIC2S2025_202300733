package disk

import (
	"fmt"
	"strings"
)

type Handler func(comando string, props map[string]string) (string, bool)

type CommandDef struct {
	Allowed  map[string]bool
	Required []string
	Defaults map[string]string
	Run      Handler
}

var commands = map[string]CommandDef{
	"mkdisk": {
		Allowed: map[string]bool{
			"size": true, "fit": true, "unit": true,
		},
		Required: []string{"size"},
		Defaults: map[string]string{"fit": "FF", "unit": "M"},
		Run:      mkdiskExecute,
	},
	"rmdisk": {
		Allowed: map[string]bool{
			"diskname": true,
		},
		Required: []string{"diskname"},
		Defaults: map[string]string{},
		Run:      nil,
	},
	"fdisk": {
		Allowed: map[string]bool{
			"size": true, "unit": true, "diskname": true, "type": true, "fit": true, "name": true,
		},
		Required: []string{"size", "diskname", "name"},
		Defaults: map[string]string{"unit": "K", "type": "P", "fit": "FF"},
		Run:      fdiskExecute,
	},
	"mount": {
		Allowed: map[string]bool{
			"diskname": true, "name": true,
		},
		Required: []string{"diskname", "name"},
		Defaults: map[string]string{},
		Run:      mountExecute,
	},
	"mounted": {
		Allowed:  map[string]bool{},
		Required: []string{},
		Defaults: map[string]string{},
		Run:      nil,
	},
	"mkfs": {
		Allowed: map[string]bool{
			"id": true, "type": true,
		},
		Required: []string{"id"},
		Defaults: map[string]string{"type": "FULL"},
		Run:      nil,
	},
}

func diskCommandProps(comando string, instrucciones []string) (string, bool) {
	cmd := strings.ToLower(comando)
	def, ok := commands[cmd]

	if !ok {
		return fmt.Sprintf("Comando no reconocido: %s", comando), true
	}

	allowedLower := make(map[string]bool, len(def.Allowed))
	for k := range def.Allowed {
		allowedLower[strings.ToLower(k)] = true
	}

	props := make(map[string]string)
	for k, v := range def.Defaults {
		props[strings.ToLower(k)] = v
	}

	seen := make(map[string]bool)

	for _, token := range instrucciones {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if !strings.Contains(token, "=") {
			return fmt.Sprintf("Parámetro inválido: '%v'", token), true
		}

		parts := strings.SplitN(token, "=", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if key == "" {
			return fmt.Sprintf("Parámetro inválido: '%s'", token), true
		}

		if !allowedLower[key] {
			return fmt.Sprintf("Parámetro no permitido para '%s': '%s'", cmd, key), true
		}

		if seen[key] {
			return fmt.Sprintf("Parámetro duplicado no permitido: %s", key), true
		}

		seen[key] = true
		props[key] = val
	}

	for _, req := range def.Required {
		reqLower := strings.ToLower(req)
		if strings.TrimSpace(props[reqLower]) == "" {
			return fmt.Sprintf("Parámetro obligatorio faltante: %s", req), true
		}
	}

	if def.Run == nil {
		return fmt.Sprintf("Comando que no tiene handler: %s", cmd), true
	}

	return def.Run(comando, props)
}

func DiskExecuteCommanWithProps(command string, parameters []string) (string, bool) {
	return diskCommandProps(command, parameters)
}
