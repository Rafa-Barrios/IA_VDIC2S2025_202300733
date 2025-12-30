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

type MountedPartition struct {
	Id       string
	DiskName string
	Path     string
	Name     string
	Start    int32
	Size     int32
}

var mountedPartitions []MountedPartition

// obtiene la letra del disco a partir del nombre (VDIC-A.mia -> A)
func obtenerLetraDisco(diskName string) byte {
	base := strings.ToUpper(diskName)

	for i := 0; i < len(base)-1; i++ {
		if base[i] == '-' && base[i+1] >= 'A' && base[i+1] <= 'Z' {
			return base[i+1]
		}
	}
	return 'A'
}

func obtenerCorrelativoGlobal() int32 {
	var max int32 = 0
	for _, part := range mountedPartitions {
		if part.Id != "" {
			num := part.Id[2 : len(part.Id)-1]
			var n int32
			fmt.Sscanf(num, "%d", &n)
			if n > max {
				max = n
			}
		}
	}
	return max + 1
}

func mountExecute(_ string, props map[string]string) (string, bool) {

	diskName := strings.TrimSpace(props["diskname"])
	partName := strings.TrimSpace(props["name"])

	if diskName == "" || partName == "" {
		return "Error: diskname y name son obligatorios", true
	}

	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		diskName += ".mia"
	}

	path := utils.DirectorioDisco + diskName

	file, err := os.Open(path)
	if err != nil {
		return fmt.Sprintf("No se pudo abrir el disco: %s", diskName), true
	}
	defer file.Close()

	var mbr structures.MBR
	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return "Error al leer el MBR", true
	}

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

	for _, mp := range mountedPartitions {
		if strings.EqualFold(mp.Path, path) &&
			strings.EqualFold(mp.Name, partName) {
			return "La partición ya se encuentra montada", true
		}
	}

	part := mbr.Mbr_partitions[partIndex]

	// Generar ID
	correlativo := obtenerCorrelativoGlobal()
	letra := obtenerLetraDisco(diskName)
	id := fmt.Sprintf("21%d%c", correlativo, letra)

	mountedPartitions = append(mountedPartitions, MountedPartition{
		Id:       id,
		DiskName: diskName,
		Path:     path,
		Name:     partName,
		Start:    part.Part_start,
		Size:     part.Part_s,
	})

	color.Green("-----------------------------------------------------------")
	color.Blue("Partición montada correctamente")
	color.Blue("Disco: %s", diskName)
	color.Blue("Partición: %s", partName)
	color.Blue("ID asignado: %s", id)
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("Partición montada correctamente con ID %s", id), false
}

func GetMountedPartition(id string) *MountedPartition {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}

	for i := range mountedPartitions {
		if strings.EqualFold(mountedPartitions[i].Id, id) {
			return &mountedPartitions[i]
		}
	}
	return nil
}
