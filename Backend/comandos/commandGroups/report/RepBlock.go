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

// RepBlock genera el reporte de BLOQUES en HTML
func RepBlock(id string, fileName string) (string, bool) {

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

	// 4️⃣ Crear carpeta de reportes
	reportDir := "C:/Users/Rafael Barrios/Downloads/Rep"
	_ = os.MkdirAll(reportDir, os.ModePerm)

	// 4️⃣a Normalizar nombre de archivo
	if !strings.HasSuffix(strings.ToLower(fileName), ".html") {
		fileName += ".html"
	}

	reportPath := filepath.Join(reportDir, fileName)

	html, errFile := os.Create(reportPath)
	if errFile != nil {
		return "No se pudo crear el reporte", true
	}
	defer html.Close()

	// 5️⃣ Inicio HTML
	fmt.Fprintln(html, "<html><body>")
	fmt.Fprintln(html, "<h1>Reporte de Bloques</h1>")

	// 6️⃣ Recorrer inodos
	for i := int32(0); i < sb.S_inodes_count; i++ {

		inode, err := disk.ReadInode(file, sb, i)
		if err != nil {
			continue
		}

		// Solo inodos usados
		if inode.I_type == 0xFF {
			continue
		}

		// 7️⃣ Recorrer bloques del inodo
		for _, blk := range inode.I_block {

			if blk == -1 {
				continue
			}

			// =========================
			// BLOQUE DE CARPETA
			// =========================
			if inode.I_type == 0 {
				var folder structures.BloqueCarpeta
				if err := disk.ReadBlock(file, sb, blk, &folder); err != nil {
					continue
				}

				fmt.Fprintf(html, "<h2>Bloque Carpeta %d</h2>", blk)
				fmt.Fprintln(html, "<table border='1'>")
				fmt.Fprintln(html, "<tr><th>b_name</th><th>b_inodo</th></tr>")

				for _, entry := range folder.B_content {
					if entry.B_inodo != -1 {
						fmt.Fprintf(
							html,
							"<tr><td>%s</td><td>%d</td></tr>",
							utils.ConvertirByteAString(entry.B_name[:]),
							entry.B_inodo,
						)
					}
				}

				fmt.Fprintln(html, "</table>")
			}

			// =========================
			// BLOQUE DE ARCHIVO
			// =========================
			if inode.I_type == 1 {
				var fileBlock structures.BloqueArchivo
				if err := disk.ReadBlock(file, sb, blk, &fileBlock); err != nil {
					continue
				}

				fmt.Fprintf(html, "<h2>Bloque Archivo %d</h2>", blk)
				fmt.Fprintln(html, "<pre>")
				fmt.Fprintln(html, utils.ConvertirByteAString(fileBlock.B_content[:]))
				fmt.Fprintln(html, "</pre>")
			}

			// =========================
			// BLOQUE DE APUNTADORES
			// =========================
			if inode.I_type == 2 {
				var pointerBlock structures.BloqueApuntador
				if err := disk.ReadBlock(file, sb, blk, &pointerBlock); err != nil {
					continue
				}

				fmt.Fprintf(html, "<h2>Bloque Apuntadores %d</h2>", blk)
				fmt.Fprintln(html, "<p>")

				for i, p := range pointerBlock.B_pointers {
					if p != -1 {
						if i > 0 {
							fmt.Fprint(html, ", ")
						}
						fmt.Fprint(html, p)
					}
				}

				fmt.Fprintln(html, "</p>")
			}
		}
	}

	// 8️⃣ Cierre HTML
	fmt.Fprintln(html, "</body></html>")

	return "[REP BLOCK]: Reporte generado correctamente", false
}
