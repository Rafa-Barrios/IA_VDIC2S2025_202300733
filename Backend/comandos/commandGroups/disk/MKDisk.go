package disk

import (
	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/utils"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func mkdiskExecute(comando string, parametros map[string]string) (string, bool) {

	// SIZE
	tamanio, er, msg := utils.TieneSize(comando, parametros["size"])
	if er || tamanio <= 0 {
		return "El parámetro -size debe ser mayor que 0", true
	}

	// UNIT
	unidad, er, msg := utils.TieneUnit(comando, parametros["unit"])
	if er {
		return msg, true
	}

	// FIT
	fit, er, msg := utils.TieneFit(comando, parametros["fit"])
	if er {
		return msg, true
	}

	msg, err := mkdisk_Create(tamanio, unidad, fit)
	if err {
		return msg, true
	}

	return "Disco creado correctamente", false
}

func mkdisk_Create(_size int32, _unit byte, _fit byte) (string, bool) {

	// Asegurar directorio
	if err := os.MkdirAll(utils.DirectorioDisco, 0755); err != nil {
		return "No se pudo crear el directorio de discos", true
	}

	for i := 0; i < 26; i++ {
		nombreDisco := fmt.Sprintf("VDIC-%c.mia", 'A'+i)
		archivo := utils.DirectorioDisco + nombreDisco

		if _, err := os.Stat(archivo); os.IsNotExist(err) {

			er, strmsg := createDiskFile(archivo, _size, _fit, _unit)
			if er {
				return strmsg, true
			}

			color.Green("[MKDISK]: Disco %s creado correctamente", nombreDisco)
			return "", false
		}
	}

	return "No hay letras disponibles para crear más discos", true
}

func createDiskFile(archivo string, tamanio int32, fit byte, unidad byte) (bool, string) {

	file, err := os.Create(archivo)
	if err != nil {
		return true, "Error al crear el archivo del disco"
	}
	defer file.Close()

	tamanioDisco := utils.ObtenerTamanioDisco(tamanio, unidad)
	if tamanioDisco <= 0 {
		return true, "El tamaño del disco es inválido"
	}

	var estructura structures.MBR
	estructura.Mbr_tamano = tamanioDisco
	estructura.Mbr_fecha_creacion = utils.ObFechaInt()
	estructura.Mbr_disk_signature = utils.ObtenerDiskSignature()
	estructura.Dsk_fit = fit

	for i := 0; i < 4; i++ {
		estructura.Mbr_partitions[i] = utils.NuevaPartitionVacia()
	}

	// Llenar con ceros
	buffer := make([]byte, 1024)
	restante := tamanioDisco

	for restante > 0 {
		escribir := int32(len(buffer))
		if restante < escribir {
			escribir = restante
		}
		if _, err := file.Write(buffer[:escribir]); err != nil {
			return true, "Error al escribir ceros en el disco"
		}
		restante -= escribir
	}

	// Escribir MBR al inicio
	if _, err := file.Seek(0, 0); err != nil {
		return true, "Error al posicionar puntero del archivo"
	}

	if err := binary.Write(file, binary.LittleEndian, &estructura); err != nil {
		return true, "Error al escribir el MBR"
	}

	return false, ""
}
