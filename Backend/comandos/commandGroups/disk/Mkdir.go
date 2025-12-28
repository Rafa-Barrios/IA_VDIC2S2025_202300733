package disk

import (
	"fmt"
	"os"
	"path"
	"strings"

	"Proyecto/Estructuras/structures"

	"github.com/fatih/color"
)

/* =========================
   MKDIR
========================= */

func mkdirExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de carpetas: mkdir")
	color.Green("-----------------------------------------------------------")

	/* 1️⃣ Validar sesión */
	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	/* 2️⃣ Parámetros */
	dirPath := strings.TrimSpace(props["path"])
	pFlag := false

	// Detectar el flag -p
	if _, ok := props["p"]; ok {
		if props["p"] != "" {
			return "❌ Error: el parámetro -p no recibe valores", true
		}
		pFlag = true
	}

	if dirPath == "" {
		return "❌ Error: el parámetro path es obligatorio", true
	}

	/* 3️⃣ Partición montada */
	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		return "❌ Error: la partición no está montada", true
	}

	/* 4️⃣ Abrir disco */
	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	/* 5️⃣ Leer SuperBloque */
	var sb structures.SuperBlock
	if err := ReadSuperBlock(file, int64(part.Start), &sb); err != nil {
		return "❌ Error al leer el SuperBloque", true
	}

	/* 6️⃣ Procesar ruta */
	cleanPath := path.Clean(dirPath)
	if cleanPath == "/" {
		return "❌ Error: no se puede crear la raíz", true
	}

	dirs := strings.Split(cleanPath, "/")
	currentInode := int32(0) // raíz

	for i, dir := range dirs {
		if dir == "" {
			continue
		}

		inode, err := ReadInode(file, sb, currentInode)
		if err != nil {
			return err.Error(), true
		}

		found := false
		var nextInode int32

		/* Buscar carpeta existente en los bloques del inode actual */
		for _, blk := range inode.I_block {
			if blk == -1 {
				continue
			}

			var block structures.BloqueCarpeta
			if err := ReadBlock(file, sb, blk, &block); err != nil {
				return err.Error(), true
			}

			for _, content := range block.B_content {
				name := strings.TrimRight(string(content.B_name[:]), "\x00")
				if name == dir {
					nextInode = content.B_inodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		isLast := i == len(dirs)-1

		/* Crear carpeta si no existe */
		if !found {
			if !pFlag && !isLast {
				// ❌ Error si falta un padre y no hay -p
				return fmt.Sprintf("❌ Error: la carpeta '%s' no existe", dir), true
			}

			// ✅ Crear carpeta intermedia o final
			newInode, err := createDirectory(file, sb, currentInode, dir)
			if err != nil {
				return err.Error(), true
			}
			nextInode = newInode
		}

		currentInode = nextInode
	}

	return fmt.Sprintf("✅ Carpeta '%s' creada correctamente", dirPath), false
}
