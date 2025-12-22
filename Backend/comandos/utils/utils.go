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
		return 0, true, "Valor entero menor o igual a 0"
	}

	return int32(i), false, ""
}

func TieneSize(comando string, size string) (int32, bool, string) {
	salida, er, msg := esEntero(size)
	if er {
		return salida, true, fmt.Sprintf("%s - Comando: %s", msg, comando)
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
		// El comando no maneja unit, se ignora sin error
		return 0, false, ""
	}

	if strings.TrimSpace(unit) == "" {
		return rule.Default, false, ""
	}

	u := strings.ToUpper(unit)
	if !rule.Allowed[u] {
		return rule.Default, true,
			fmt.Sprintf("[%s] unidad inválida: %s", command, u)
	}

	return u[0], false, ""
}

/* =========================
   FIT
========================= */

func TieneFit(command string, fit string) (byte, bool, string) {

	fit = strings.ToUpper(strings.TrimSpace(fit))

	if fit == "" {
		return 'F', false, "" // FF por defecto
	}

	switch fit {
	case "FF":
		return 'F', false, ""
	case "WF":
		return 'W', false, ""
	case "BF":
		return 'B', false, ""
	default:
		return 0, true, fmt.Sprintf("Fit inválido: %s", fit)
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
	p.Part_fit = 'F'
	p.Part_start = -1
	p.Part_s = -1
	p.Part_correlative = -1
	return p
}

func TieneType(tipo string) (byte, bool, string) {
	switch strings.ToUpper(tipo) {
	case "P":
		return 'P', false, ""
	case "E":
		return 'E', false, ""
	case "L":
		return 'L', false, ""
	default:
		return 0, true, "Tipo no reconocido"
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
   MBR / EBR
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
	mbr, err, strMensajeErr := ObtenerEstructuraMBR(pathDisco)
	if err {
		fmt.Println(strMensajeErr)
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
