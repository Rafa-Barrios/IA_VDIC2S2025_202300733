package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Proyecto/Estructuras/structures"
	"Proyecto/comandos/commandGroups/disk"
)

// RepBMInode genera el reporte del bitmap de inodos en TXT
func RepBMInode(id string, fileName string) (string, bool) {

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

	// 4️⃣ Normalizar nombre del archivo (.txt)
	if !strings.HasSuffix(strings.ToLower(fileName), ".txt") {
		fileName += ".txt"
	}

	// 5️⃣ Ruta de salida (Downloads)
	reportDir := "C:/Users/Rafael Barrios/Downloads/Rep"
	_ = os.MkdirAll(reportDir, os.ModePerm)

	reportPath := filepath.Join(reportDir, fileName)

	txt, err := os.Create(reportPath)
	if err != nil {
		return "No se pudo crear el reporte", true
	}
	defer txt.Close()

	// 6️⃣ Leer bitmap de inodos
	fmt.Fprintln(txt, "# Bitmap de Inodos\n")

	count := 0
	for i := int32(0); i < sb.S_inodes_count; i++ {
		pos := sb.S_bm_inode_start + i
		file.Seek(int64(pos), 0)

		b := []byte{0}
		file.Read(b)

		fmt.Fprint(txt, b[0])

		count++
		if count == 20 {
			fmt.Fprintln(txt)
			count = 0
		} else {
			fmt.Fprint(txt, " ")
		}
	}

	fmt.Fprintln(txt)

	return "[REP BM_INODE]: Reporte generado correctamente", false
}
