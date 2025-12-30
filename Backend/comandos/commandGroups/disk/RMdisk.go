package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Proyecto/comandos/utils"

	"github.com/fatih/color"
)

func rmdiskExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de discos: rmdisk")
	color.Green("-----------------------------------------------------------")

	if currentSession != nil {
		return "❌ Error: no se puede eliminar un disco con una sesión activa", true
	}

	diskName := strings.TrimSpace(props["diskname"])
	if diskName == "" {
		return "❌ Error: el parámetro diskName es obligatorio", true
	}

	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		diskName += ".mia"
	}

	diskPath := filepath.Join(utils.DirectorioDisco, diskName)

	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Sprintf("❌ Error: el disco '%s' no existe", diskName), true
	}

	if err := os.Remove(diskPath); err != nil {
		return fmt.Sprintf("❌ Error al eliminar el disco '%s'", diskName), true
	}

	color.Green("🗑 Disco eliminado correctamente: %s", diskPath)
	return fmt.Sprintf("✅ Disco '%s' eliminado correctamente", diskName), false
}
