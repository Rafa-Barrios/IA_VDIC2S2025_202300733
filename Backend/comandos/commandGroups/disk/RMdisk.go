package disk

import (
	"fmt"
	"os"
	"path/filepath"

	"Proyecto/comandos/utils"

	"github.com/fatih/color"
)

/* =========================
   RMDISK
========================= */

func rmdiskExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de discos: rmdisk")
	color.Green("-----------------------------------------------------------")

	// 🔴 Regla obligatoria: no debe haber sesión activa
	if currentSession != nil {
		return "❌ Error: no se puede eliminar un disco con una sesión activa", true
	}

	diskName := props["diskname"]
	if diskName == "" {
		return "❌ Error: el parámetro diskName es obligatorio", true
	}

	// 📌 Ruta REAL donde mkdisk crea los discos
	diskPath := filepath.Join(utils.DirectorioDisco, diskName)

	// 🔍 Verificar existencia
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		return fmt.Sprintf("❌ Error: el disco '%s' no existe", diskName), true
	}

	// 🗑 Eliminar disco
	if err := os.Remove(diskPath); err != nil {
		return fmt.Sprintf("❌ Error al eliminar el disco '%s'", diskName), true
	}

	color.Green("🗑 Disco eliminado correctamente: %s", diskPath)
	return fmt.Sprintf("✅ Disco '%s' eliminado correctamente", diskName), false
}
