package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Proyecto/comandos/commandGroups/disk"
	"Proyecto/comandos/utils"
)

// RepMBR genera el reporte MBR en HTML
func RepMBR(id string, fileName string) (string, bool) {

	// 1️⃣ Obtener partición montada
	mount := disk.GetMountedPartition(id)
	if mount == nil {
		return "ID de partición no encontrado", true
	}

	// 2️⃣ Leer MBR usando utils (FUNCIÓN REAL)
	mbr, err, msg := utils.ObtenerEstructuraMBR(mount.Path)
	if err {
		return msg, true
	}

	// 3️⃣ Crear carpeta de reportes
	reportDir := "C:/Users/Rafael Barrios/Downloads/Rep"
	_ = os.MkdirAll(reportDir, os.ModePerm)

	// 3️⃣a Normalizar nombre de archivo: agregar ".html" si no lo tiene
	if !strings.HasSuffix(strings.ToLower(fileName), ".html") {
		fileName += ".html"
	}

	reportPath := filepath.Join(reportDir, fileName)

	html, errFile := os.Create(reportPath)
	if errFile != nil {
		return "No se pudo crear el reporte", true
	}
	defer html.Close()

	// 4️⃣ Generar HTML
	fmt.Fprintln(html, "<html><body>")
	fmt.Fprintln(html, "<h1>Reporte de MBR</h1>")

	fmt.Fprintln(html, "<h2>Datos del MBR</h2>")
	fmt.Fprintln(html, "<table border='1'>")
	fmt.Fprintf(html, "<tr><td>Tamaño</td><td>%d</td></tr>", mbr.Mbr_tamano)
	fmt.Fprintf(html, "<tr><td>Fecha creación</td><td>%s</td></tr>",
		utils.IntFechaToStr(mbr.Mbr_fecha_creacion))
	fmt.Fprintf(html, "<tr><td>Disk Signature</td><td>%d</td></tr>", mbr.Mbr_disk_signature)
	fmt.Fprintln(html, "</table>")

	// 5️⃣ Particiones
	fmt.Fprintln(html, "<h2>Particiones</h2>")
	fmt.Fprintln(html, "<table border='1'>")
	fmt.Fprintln(html,
		"<tr><th>Status</th><th>Type</th><th>Fit</th><th>Start</th><th>Size</th><th>Name</th></tr>",
	)

	for _, p := range mbr.Mbr_partitions {
		if p.Part_start != -1 {
			fmt.Fprintf(
				html,
				"<tr><td>%d</td><td>%c</td><td>%c</td><td>%d</td><td>%d</td><td>%s</td></tr>",
				p.Part_status,
				p.Part_type,
				p.Part_fit,
				p.Part_start,
				p.Part_s,
				utils.ConvertirByteAString(p.Part_name[:]),
			)
		}
	}

	fmt.Fprintln(html, "</table>")
	fmt.Fprintln(html, "</body></html>")

	return "[REP MBR]: Reporte generado correctamente", false
}
