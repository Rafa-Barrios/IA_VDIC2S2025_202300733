package disk

import (
	"Proyecto/Estructuras/size"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// P = Primario
func fdiskExecute(comando string, parametros map[string]string) (string, bool) {

	tamanio, er, strError := utils.TieneSize(comando, parametros["size"])
	if er {
		return strError, er
	}

	unidad, er, strError := utils.TieneUnit(comando, parametros["unit"])
	if er {
		return strError, er
	}

	diskName, er, strError := utils.TieneDiskName(parametros["diskname"])
	if er {
		return strError, er
	}

	tipo, er, strError := utils.TieneType(parametros["type"])
	if er {
		return strError, er
	}

	tipoFit, er, strError := utils.TieneFit("fdisk", parametros["fit"])
	if er {
		return strError, er
	}

	nombreParticion, er, strError := utils.TieneName(parametros["name"])
	if er {
		return strError, er
	}

	return fdiskCreate(tamanio, unidad, diskName, tipo, tipoFit, nombreParticion)
}

func fdiskCreate(tamanio int32, unidad byte, diskName string, tipo byte, tipoFit byte, nombreParticion string) (string, bool) {

	diskName = strings.TrimSpace(diskName)

	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		diskName += ".mia"
	}

	path := utils.DirectorioDisco + diskName

	switch tipo {
	case 'P':
		return particionPrimaria(path, nombreParticion, tipo, tamanio, tipoFit, unidad)

	case 'E':
		return "Particiones extendidas aún no implementadas", true

	case 'L':
		return "Particiones lógicas aún no implementadas", true

	default:
		return "Tipo de partición desconocido", true
	}
}

func particionPrimaria(ubicacionArchivo string, nombreParticion string, tipo byte, tamanioDisco int32, tipoFit byte, unidad byte) (string, bool) {

	if !utils.ExisteArchivo("FDISK", ubicacionArchivo) {
		color.Yellow("[FDISK]: Disco <<" + ubicacionArchivo + ">> no encontrado")
		return "Disco no encontrado", true
	}

	if len(nombreParticion) > 16 {
		return "El nombre de la partición no puede exceder 16 caracteres", true
	}

	mbr, er, strError := utils.ObtenerEstructuraMBR(ubicacionArchivo)
	if er {
		return strError, er
	}

	pos := -1
	for i := range mbr.Mbr_partitions {
		if mbr.Mbr_partitions[i].Part_start == -1 {
			pos = i
			break
		}
	}

	if pos == -1 {
		return "No hay espacio para más particiones primarias", true
	}

	// Nombre duplicado
	nombreExistente, msg := utils.ExisteNombreParticion(ubicacionArchivo, nombreParticion)
	if nombreExistente {
		return msg, true
	}

	// Espacio
	if !utils.ExisteEspacioDisponible(tamanioDisco, ubicacionArchivo, unidad, int32(pos)) {
		return "Espacio insuficiente en el disco", true
	}

	particion := utils.NuevaPartitionVacia()
	particion.Part_type = tipo
	particion.Part_fit = tipoFit
	particion.Part_status = 0
	particion.Part_name = [16]byte(utils.ConvertirStringAByte(nombreParticion, 16))
	particion.Part_correlative = utils.ObtenerDiskSignature()
	particion.Part_s = utils.ObtenerTamanioDisco(tamanioDisco, unidad)

	if pos == 0 {
		particion.Part_start = size.SizeMBR()
	} else {
		particion.Part_start = mbr.Mbr_partitions[pos-1].Part_start +
			mbr.Mbr_partitions[pos-1].Part_s
	}

	mbr.Mbr_partitions[pos] = particion

	file, err := os.OpenFile(ubicacionArchivo, os.O_RDWR, 0666)
	if err != nil {
		return "Error al abrir el disco", true
	}
	defer file.Close()

	file.Seek(0, 0)
	if err := binary.Write(file, binary.LittleEndian, &mbr); err != nil {
		return "Error al escribir el MBR", true
	}

	color.Green("-----------------------------------------------------------")
	color.Blue("Partición primaria creada exitosamente")
	color.Blue("Nombre: " + nombreParticion)
	color.Blue("Inicio: " + strconv.Itoa(int(particion.Part_start)))
	color.Blue("Tamaño: " + strconv.Itoa(int(particion.Part_s)))
	color.Green("-----------------------------------------------------------")

	return "", false
}
