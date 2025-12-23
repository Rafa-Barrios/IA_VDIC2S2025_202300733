package general

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

/* =========================
   RUTAS BASE DEL PROYECTO
========================= */

var NamePath = "VDIC-MIA"
var ReportPath = "VDIC-MIA/Rep"
var DiskPath = "VDIC-MIA/Disks"

/* =========================
   OBTENCIÓN DE PARÁMETROS
========================= */

func ObtenerParametros(x string) []string {

	var comandos []string

	// -param=value | -param="value" | >path=value
	regex := regexp.MustCompile(`(-|>)(\w+)(?:="([^"]+)"|=([^"\s]+))?`)
	matches := regex.FindAllStringSubmatch(x, -1)

	for _, m := range matches {
		atributo := strings.ToLower(strings.TrimSpace(m[2]))

		if m[3] != "" {
			comandos = append(comandos, fmt.Sprintf("%s=%s", atributo, m[3]))
		} else if m[4] != "" {
			comandos = append(comandos, fmt.Sprintf("%s=%s", atributo, m[4]))
		}
		// ❌ Se eliminan parámetros bandera (-p)
		// No se usan en este proyecto y pueden causar errores
	}

	return comandos
}

/* =========================
   CREACIÓN DE ESTRUCTURA
========================= */

func CrearCarpeta() {

	nombreArchivo := "VDIC-MIA/CarpetaImagenes.txt"

	// Carpeta raíz
	if _, err := os.Stat(NamePath); os.IsNotExist(err) {
		if err := os.MkdirAll(NamePath, 0755); err != nil {
			color.Red("Error al crear carpeta VDIC-MIA")
			return
		}
		color.Green("Carpeta VDIC-MIA creada correctamente")
	}

	// Carpeta Rep
	if _, err := os.Stat(ReportPath); os.IsNotExist(err) {
		if err := os.Mkdir(ReportPath, 0755); err != nil {
			color.Red("Error al crear carpeta Rep")
			return
		}
		color.Green("Carpeta Rep creada correctamente")
	}

	// Carpeta Disks
	if _, err := os.Stat(DiskPath); os.IsNotExist(err) {
		if err := os.Mkdir(DiskPath, 0755); err != nil {
			color.Red("Error al crear carpeta Disks")
			return
		}
		color.Green("Carpeta Disks creada correctamente")
	}

	// Archivo informativo
	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		archivo, err := os.Create(nombreArchivo)
		if err != nil {
			color.Red("Error al crear archivo informativo")
			return
		}
		defer archivo.Close()

		archivo.WriteString("Proyecto Único - VDIC-MIA\n")
		color.Green("Archivo informativo creado")
	}

	color.Green("Estructura base verificada")
}

/* =========================
   VALIDACIÓN DE PATH
========================= */

func TienePath(x string) string {

	partes := strings.SplitN(x, "=", 2)
	if len(partes) != 2 {
		return ""
	}

	path := partes[1]
	color.Yellow("Buscando: %s", path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		color.Red("Archivo no encontrado")
		return ""
	}

	color.Green("Archivo encontrado")
	return path
}

/* =========================
   EXECUTE (LECTURA SCRIPT)
========================= */

func ExecuteCommandList(comandos []string) Resultado {

	var lineasValidas []string

	for _, c := range comandos {
		c = strings.TrimSpace(c)
		if c != "" && !strings.HasPrefix(c, "#") {
			lineasValidas = append(lineasValidas, c)
		}
	}

	// Eliminar comentarios finales
	reg := regexp.MustCompile(`^(.*?)\s*(?:#.*)?$`)
	var comandosFinales []string

	for _, l := range lineasValidas {
		match := reg.FindStringSubmatch(l)
		if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			comandosFinales = append(comandosFinales, strings.TrimSpace(match[1]))
		}
	}

	return Resultado{
		Mensaje: "",
		Error:   false,
		Salida: SalidaComandoEjecutado{
			LstComandos: comandosFinales,
		},
	}
}
