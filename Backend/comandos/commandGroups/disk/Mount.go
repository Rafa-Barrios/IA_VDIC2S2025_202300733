package disk

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// mountExecute monta una partición primaria en el disco
func mountExecute(_ string, props map[string]string) (string, bool) {

	diskName := strings.TrimSpace(props["diskname"])
	partName := strings.TrimSpace(props["name"])

	// 🔹 Construir ruta REAL del disco
	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		return "El disco debe tener extensión .mia", true
	}

	path := utils.DirectorioDisco + diskName

	// 1. Abrir disco
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Sprintf("No se pudo abrir el disco: %s", diskName), true
	}
	defer file.Close()

	// 2. Leer MBR
	var mbr structures.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return "Error al leer el MBR", true
	}

	// 3. Buscar la partición primaria por nombre
	partIndex := -1
	for i := 0; i < 4; i++ {
		part := mbr.Mbr_partitions[i]
		name := utils.ConvertirByteAString(part.Part_name[:])

		if strings.EqualFold(name, partName) {

			if part.Part_type != 'P' {
				return "Solo se pueden montar particiones primarias", true
			}

			partIndex = i
			break
		}
	}

	if partIndex == -1 {
		return fmt.Sprintf("No existe la partición '%s'", partName), true
	}

	part := &mbr.Mbr_partitions[partIndex]

	// 4. Validar si ya está montada
	if part.Part_status == 1 {
		return "La partición ya se encuentra montada", true
	}

	// 5. Calcular correlativo REAL (por disco)
	var correlativo int32 = 1
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partitions[i].Part_status == 1 {
			correlativo++
		}
	}

	// 6. Generar letra fija por disco (A, B, C...)
	letra := byte('A' + (mbr.Mbr_disk_signature % 26))

	// 7. Generar ID -> 21 + número + letra
	id := fmt.Sprintf("21%d%c", correlativo, letra)

	// 8. Actualizar partición
	part.Part_status = 1
	part.Part_correlative = correlativo
	copy(part.Part_id[:], id)

	// 9. Escribir MBR actualizado
	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return "Error al escribir el MBR actualizado", true
	}

	// 🔹 MENSAJES EN CONSOLA
	color.Green("-----------------------------------------------------------")
	color.Blue("Partición montada correctamente")
	color.Blue("Disco: %s", diskName)
	color.Blue("Partición: %s", partName)
	color.Blue("ID asignado: %s", id)
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("Partición montada correctamente con ID %s", id), false
}
