package disk

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

// mountedExecute muestra todas las particiones montadas
func mountedExecute(_ string, _ map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Particiones montadas en el sistema")
	color.Green("-----------------------------------------------------------")

	encontradas := false

	// Leer todos los discos del directorio
	err := filepath.Walk(utils.DirectorioDisco, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Solo archivos .mia
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".mia") {
			return nil
		}

		file, err := os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			return nil
		}
		defer file.Close()

		var mbr structures.MBR
		if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
			return nil
		}

		// Revisar particiones
		for _, part := range mbr.Mbr_partitions {
			if part.Part_status == 1 {
				id := strings.TrimRight(string(part.Part_id[:]), "\x00")
				if id != "" {
					color.Cyan("• %s", id)
					encontradas = true
				}
			}
		}

		return nil
	})

	if err != nil {
		return "Error al recorrer los discos", true
	}

	if !encontradas {
		color.Yellow("No hay particiones montadas actualmente")
		return "No hay particiones montadas", false
	}

	color.Green("-----------------------------------------------------------")
	return "Listado de particiones montadas mostrado correctamente", false
}
