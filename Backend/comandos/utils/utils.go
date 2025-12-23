package utils

import (
	"Proyecto/Estructuras/size"
	"Proyecto/Estructuras/structures"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

var DirectorioDisco = "VDIC-MIA/Disks/"

/* =========================
   VALIDACIONES BÁSICAS
========================= */

func esEntero(valor string) (int32, bool, string) {
	i, err := strconv.Atoi(valor)
	if err != nil {
		return 0, true, "Error en la conversión a entero"
	}

	if i <= 0 {
		return 0, true, "El tamaño debe ser mayor a 0"
	}

	return int32(i), false, ""
}

func TieneSize(comando string, size string) (int32, bool, string) {
	salida, er, msg := esEntero(size)
	if er {
		return salida, true, fmt.Sprintf("[%s] %s", strings.ToUpper(comando), msg)
	}
	return salida, false, ""
}

/* =========================
   FECHAS
========================= */

func ObFechaInt() int32 {
	return int32(time.Now().Unix())
}

func IntFechaToStr(fecha int32) string {
	formato := "2006/01/02 (15:04:05)"
	return time.Unix(int64(fecha), 0).Format(formato)
}

/* =========================
   UNITS
========================= */

var unitRules = map[string]struct {
	Default byte
	Allowed map[string]bool
}{
	"mkdisk": {Default: 'M', Allowed: map[string]bool{"K": true, "M": true}},
	"fdisk":  {Default: 'K', Allowed: map[string]bool{"B": true, "K": true, "M": true}},
}

func TieneUnit(command string, unit string) (byte, bool, string) {
	command = strings.ToLower(command)

	rule, ok := unitRules[command]
	if !ok {
		return 0, false, ""
	}

	if strings.TrimSpace(unit) == "" {
		return rule.Default, false, ""
	}

	u := strings.ToUpper(unit)
	if !rule.Allowed[u] {
		return rule.Default, true,
			fmt.Sprintf("[%s] unidad inválida: %s", strings.ToUpper(command), u)
	}

	return u[0], false, ""
}

/* =========================
   FIT
========================= */

func TieneFit(command string, fit string) (byte, bool, string) {

	command = strings.ToLower(command)
	fit = strings.ToUpper(strings.TrimSpace(fit))

	// Valor por defecto
	if fit == "" {
		return 'W', false, ""
	}

	switch fit {
	case "BF":
		return 'B', false, ""
	case "FF":
		return 'F', false, ""
	case "WF":
		return 'W', false, ""
	default:
		color.Red("[%s] Fit inválido: %s", strings.ToUpper(command), fit)
		return 0, true,
			fmt.Sprintf("[%s] Fit inválido, solo se permite BF, FF o WF", strings.ToUpper(command))
	}
}

/* =========================
   DISCO
========================= */

func ObtenerTamanioDisco(size int32, unidad byte) int32 {
	switch unidad {
	case 'B':
		return size
	case 'K':
		return size * 1024
	case 'M':
		return size * 1024 * 1024
	default:
		return 0
	}
}

func ObtenerDiskSignature() int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int32(r.Intn(1000000) + 1)
}

/* =========================
   PARTICIONES
========================= */

func NuevaPartitionVacia() structures.Partition {
	p := structures.Partition{}
	p.Part_status = -1
	p.Part_type = 'P'
	p.Part_fit = 'W'
	p.Part_start = -1
	p.Part_s = -1
	p.Part_correlative = -1
	return p
}

/*
SOLO SE PERMITEN PARTICIONES PRIMARIAS
*/
func TieneType(tipo string) (byte, bool, string) {
	switch strings.ToUpper(strings.TrimSpace(tipo)) {
	case "", "P":
		return 'P', false, ""
	default:
		return 0, true, "Solo se permiten particiones primarias (P)"
	}
}

/* =========================
   STRINGS
========================= */

func ConvertirByteAString(arr []byte) string {
	i := bytes.IndexByte(arr, 0)
	if i == -1 {
		return string(arr)
	}
	return string(arr[:i])
}

func ConvertirStringAByte(texto string, size int) []byte {
	arr := make([]byte, size)
	copy(arr, texto)
	return arr
}

/* =========================
   ARCHIVOS
========================= */

func ExisteArchivo(comando string, path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		color.Red("[%s]: Archivo no encontrado", comando)
		return false
	}
	return true
}

/* =========================
   MBR
========================= */

func ObtenerEstructuraMBR(path string) (structures.MBR, bool, string) {
	var mbr structures.MBR

	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return mbr, true, "Error al abrir el disco"
	}
	defer file.Close()

	if err := binary.Read(file, binary.LittleEndian, &mbr); err != nil {
		return mbr, true, "Error al leer el MBR"
	}

	return mbr, false, ""
}

/* =========================
   ESPACIO
========================= */

func ExisteEspacioDisponible(tamanio int32, pathDisco string, unidad byte, posicion int32) bool {
	mbr, err, msg := ObtenerEstructuraMBR(pathDisco)
	if err {
		fmt.Println(msg)
		return false
	}

	if posicion < 0 {
		return false
	}

	tamanioDisco := ObtenerTamanioDisco(tamanio, unidad)
	if tamanioDisco <= 0 {
		return false
	}

	var espacio int32
	if posicion == 0 {
		espacio = mbr.Mbr_tamano - size.SizeMBR()
	} else {
		prev := mbr.Mbr_partitions[posicion-1]
		espacio = mbr.Mbr_tamano - prev.Part_start - prev.Part_s
	}

	return espacio >= tamanioDisco
}

/* =====================================================
   FUNCIONES AGREGADAS PARA FDISK (SIN ROMPER NADA)
===================================================== */

func TieneDiskName(diskName string) (string, bool, string) {
	diskName = strings.TrimSpace(diskName)

	if diskName == "" {
		return "", true, "diskName no puede estar vacío"
	}

	if !strings.HasSuffix(strings.ToLower(diskName), ".mia") {
		return "", true, "El disco debe tener extensión .mia"
	}

	return diskName, false, ""
}

func TieneName(name string) (string, bool, string) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", true, "El nombre de la partición no puede estar vacío"
	}

	if len(name) > 16 {
		return "", true, "El nombre de la partición no puede exceder 16 caracteres"
	}

	return name, false, ""
}

func ExisteNombreParticion(path string, nombre string) (bool, string) {
	mbr, err, msg := ObtenerEstructuraMBR(path)
	if err {
		return true, msg
	}

	for _, part := range mbr.Mbr_partitions {
		if part.Part_start != -1 {
			nombreActual := ConvertirByteAString(part.Part_name[:])
			if strings.EqualFold(nombreActual, nombre) {
				return true, "Ya existe una partición con ese nombre"
			}
		}
	}

	return false, ""
}
