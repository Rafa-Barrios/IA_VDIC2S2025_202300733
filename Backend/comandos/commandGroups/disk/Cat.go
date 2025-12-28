package disk

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"Proyecto/Estructuras/structures"
)

/* =========================
   CAT
========================= */

func catExecute(_ string, props map[string]string) (string, bool) {
	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		return "❌ Error: no hay partición montada", true
	}

	file, err := os.OpenFile(part.Path, os.O_RDONLY, 0666)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	var sb structures.SuperBlock
	if err := ReadSuperBlock(file, int64(part.Start), &sb); err != nil {
		return "❌ Error al leer el SuperBloque", true
	}

	// Recoger fileN
	filesMap := make(map[int]string)
	for k, v := range props {
		if strings.HasPrefix(strings.ToLower(k), "file") {
			numStr := strings.TrimPrefix(strings.ToLower(k), "file")
			num, err := strconv.Atoi(numStr)
			if err == nil {
				filesMap[num] = v
			}
		}
	}

	if len(filesMap) == 0 {
		return "❌ Error: no se proporcionó ningún archivo", true
	}

	keys := make([]int, 0, len(filesMap))
	for k := range filesMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var output strings.Builder

	for _, k := range keys {
		path := filesMap[k]
		content, err := readFileContent(file, sb, path)
		if err != nil {
			return fmt.Sprintf("❌ Error en '%s': %s", path, err.Error()), true
		}

		// 🔍 DEBUG OPCIONAL
		fmt.Printf("DEBUG CAT [%s] len=%d\n", path, len(content))

		output.WriteString(content)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n"), false
}

/* =========================
   LECTURA DE ARCHIVO
========================= */

func readFileContent(file *os.File, sb structures.SuperBlock, pathStr string) (string, error) {

	pathStr = strings.TrimSpace(pathStr)
	if pathStr == "" {
		return "", fmt.Errorf("ruta vacía")
	}

	// 🔑 NORMALIZAR RUTA
	parts := strings.Split(strings.Trim(pathStr, "/"), "/")
	currentInode := int32(0) // raíz

	for i, name := range parts {

		isLast := i == len(parts)-1

		inode, err := ReadInode(file, sb, currentInode)
		if err != nil {
			return "", err
		}

		// 🔴 Si no es el último, DEBE ser carpeta
		if !isLast && inode.I_type != 0 {
			return "", fmt.Errorf("'%s' no es una carpeta", name)
		}

		found := false
		var nextInode int32

		for _, blk := range inode.I_block {
			if blk == -1 {
				continue
			}

			var block structures.BloqueCarpeta
			if err := ReadBlock(file, sb, blk, &block); err != nil {
				return "", err
			}

			for _, c := range block.B_content {
				entryName := strings.TrimRight(string(c.B_name[:]), "\x00")
				if entryName == name {
					nextInode = c.B_inodo
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return "", fmt.Errorf("no existe '%s'", name)
		}

		currentInode = nextInode

		// 📄 Último → archivo
		if isLast {
			inode, err := ReadInode(file, sb, currentInode)
			if err != nil {
				return "", err
			}
			if inode.I_type != 1 {
				return "", fmt.Errorf("'%s' no es un archivo", name)
			}

			var content strings.Builder
			var readBytes int32 = 0

			for _, blk := range inode.I_block {
				if blk == -1 || readBytes >= inode.I_s {
					continue
				}

				var fb structures.BloqueArchivo
				if err := ReadBlock(file, sb, blk, &fb); err != nil {
					return "", err
				}

				for i := 0; i < 64 && readBytes < inode.I_s; i++ {
					content.WriteByte(fb.B_content[i])
					readBytes++
				}
			}

			return content.String(), nil
		}
	}

	return "", fmt.Errorf("ruta inválida")
}
