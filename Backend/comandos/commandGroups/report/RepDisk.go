package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Proyecto/comandos/commandGroups/disk"
	"Proyecto/comandos/utils"
)

func RepDISK(id string, fileName string) (string, bool) {

	mount := disk.GetMountedPartition(id)
	if mount == nil {
		return "ID de partición no encontrado", true
	}

	mbr, err, msg := utils.ObtenerEstructuraMBR(mount.Path)
	if err {
		return msg, true
	}

	reportDir := "C:/Users/Rafael Barrios/Downloads/Rep"
	_ = os.MkdirAll(reportDir, os.ModePerm)

	if !strings.HasSuffix(strings.ToLower(fileName), ".html") {
		fileName += ".html"
	}

	reportPath := filepath.Join(reportDir, fileName)

	html, errFile := os.Create(reportPath)
	if errFile != nil {
		return "No se pudo crear el reporte", true
	}
	defer html.Close()

	fmt.Fprintln(html, "<html><body>")
	fmt.Fprintln(html, "<h1>Reporte DISK</h1>")
	fmt.Fprintln(html, "<div style='display:flex; width:100%; border:1px solid black;'>")

	totalDisk := float64(mbr.Mbr_tamano)
	var usado int32 = 0

	mbrSize := int32(512)
	mbrPercent := (float64(mbrSize) / totalDisk) * 100

	fmt.Fprintf(html,
		"<div style='width:%f%%; border-right:1px solid black; text-align:center;'>MBR<br/>%.2f%%</div>",
		mbrPercent, mbrPercent)

	usado += mbrSize

	for _, p := range mbr.Mbr_partitions {

		if p.Part_start == -1 {
			continue
		}

		// Espacio libre antes de la partición
		if p.Part_start > usado {
			libre := p.Part_start - usado
			librePercent := (float64(libre) / totalDisk) * 100

			fmt.Fprintf(html,
				"<div style='width:%f%%; border-right:1px solid black; text-align:center;'>Libre<br/>%.2f%%</div>",
				librePercent, librePercent)

			usado += libre
		}

		// Partición primaria
		partPercent := (float64(p.Part_s) / totalDisk) * 100
		fmt.Fprintf(html,
			"<div style='width:%f%%; border-right:1px solid black; text-align:center;'>Primaria<br/>%.2f%%</div>",
			partPercent, partPercent)

		usado += p.Part_s
	}

	if usado < mbr.Mbr_tamano {
		libre := mbr.Mbr_tamano - usado
		librePercent := (float64(libre) / totalDisk) * 100

		fmt.Fprintf(html,
			"<div style='width:%f%%; text-align:center;'>Libre<br/>%.2f%%</div>",
			librePercent, librePercent)
	}

	fmt.Fprintln(html, "</div>")
	fmt.Fprintln(html, "</body></html>")

	return "[REP DISK]: Reporte generado correctamente", false
}
