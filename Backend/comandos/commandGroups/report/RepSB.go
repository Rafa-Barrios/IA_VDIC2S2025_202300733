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

// RepSB genera el reporte del SuperBloque en HTML
func RepSB(id string, fileName string) (string, bool) {

	// 1️⃣ Obtener partición montada
	mount := disk.GetMountedPartition(id)
	if mount == nil {
		return "ID de partición no encontrado", true
	}

	// 2️⃣ Abrir disco
	file, err := os.Open(mount.Path)
	if err != nil {
		return "No se pudo abrir el disco", true
	}
	defer file.Close()

	// 3️⃣ Leer SuperBloque
	var sb structures.SuperBlock
	if err := disk.ReadSuperBlock(file, int64(mount.Start), &sb); err != nil {
		return "No se pudo leer el SuperBloque", true
	}

	// 4️⃣ Normalizar nombre del archivo
	if !strings.HasSuffix(strings.ToLower(fileName), ".html") {
		fileName += ".html"
	}

	// 5️⃣ Crear carpeta de reportes
	reportDir := "C:/Users/Rafael Barrios/Downloads/Rep"
	_ = os.MkdirAll(reportDir, os.ModePerm)

	reportPath := filepath.Join(reportDir, fileName)

	html, err := os.Create(reportPath)
	if err != nil {
		return "No se pudo crear el reporte", true
	}
	defer html.Close()

	// 6️⃣ Generar HTML
	fmt.Fprintln(html, "<html><body>")
	fmt.Fprintln(html, "<h1>Reporte del SuperBloque</h1>")
	fmt.Fprintln(html, "<table border='1'>")

	writeRow := func(name string, value interface{}) {
		fmt.Fprintf(html, "<tr><td>%s</td><td>%v</td></tr>", name, value)
	}

	writeRow("s_filesystem_type", sb.S_filesystem_type)
	writeRow("s_inodes_count", sb.S_inodes_count)
	writeRow("s_blocks_count", sb.S_blocks_count)
	writeRow("s_free_blocks_count", sb.S_free_blocks_count)
	writeRow("s_free_inodes_count", sb.S_free_inodes_count)
	writeRow("s_mtime", utils.IntFechaToStr(sb.S_mtime))
	writeRow("s_umtime", utils.IntFechaToStr(sb.S_umtime))
	writeRow("s_mnt_count", sb.S_mnt_count)
	writeRow("s_magic", sb.S_magic)
	writeRow("s_inode_s", sb.S_inode_s)
	writeRow("s_block_s", sb.S_block_s)
	writeRow("s_first_ino", sb.S_first_ino)
	writeRow("s_first_blo", sb.S_first_blo)
	writeRow("s_bm_inode_start", sb.S_bm_inode_start)
	writeRow("s_bm_block_start", sb.S_bm_block_start)
	writeRow("s_inode_start", sb.S_inode_start)
	writeRow("s_block_start", sb.S_block_start)

	fmt.Fprintln(html, "</table>")
	fmt.Fprintln(html, "</body></html>")

	return "[REP SB]: Reporte generado correctamente", false
}
