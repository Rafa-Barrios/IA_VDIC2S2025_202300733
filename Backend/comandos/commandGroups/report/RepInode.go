package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/commandGroups/disk"
	"Proyecto/comandos/utils"
)

// RepInode general
func RepInode(id string, fileName string) (string, bool) {

	mount := disk.GetMountedPartition(id)
	if mount == nil {
		return "ID de partición no encontrado", true
	}

	file, err := os.OpenFile(mount.Path, os.O_RDWR, 0666)
	if err != nil {
		return "No se pudo abrir el disco", true
	}
	defer file.Close()

	var sb structures.SuperBlock
	if err := disk.ReadSuperBlock(file, int64(mount.Start), &sb); err != nil {
		return "No se pudo leer el SuperBloque", true
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
	fmt.Fprintln(html, "<h1>Reporte de Inodos</h1>")

	for i := int32(0); i < sb.S_inodes_count; i++ {

		inode, err := disk.ReadInode(file, sb, i)
		if err != nil {
			continue
		}

		// Solo inodos usados
		if inode.I_type != 0 && inode.I_type != 1 {
			continue
		}

		fmt.Fprintf(html, "<h2>Inodo %d</h2>", i)
		fmt.Fprintln(html, "<table border='1'>")

		fmt.Fprintf(html, "<tr><td>i_uid</td><td>%d</td></tr>", inode.I_uid)
		fmt.Fprintf(html, "<tr><td>i_size</td><td>%d</td></tr>", inode.I_s)
		fmt.Fprintf(
			html,
			"<tr><td>i_atime</td><td>%s</td></tr>",
			utils.IntFechaToStr(inode.I_atime),
		)

		// Bloques directos
		for j, blk := range inode.I_block {
			fmt.Fprintf(
				html,
				"<tr><td>i_block_%d</td><td>%d</td></tr>",
				j+1,
				blk,
			)
		}

		fmt.Fprintf(
			html,
			"<tr><td>i_perm</td><td>%d%d%d</td></tr>",
			inode.I_perm[0],
			inode.I_perm[1],
			inode.I_perm[2],
		)

		fmt.Fprintln(html, "</table><br>")
	}

	fmt.Fprintln(html, "</body></html>")

	return "[REP INODE]: Reporte generado correctamente", false
}
