package disk

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"Proyecto/Estructuras/structures"

	"github.com/fatih/color"
)

/* =========================
   MKFILE
========================= */

func mkfileExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de archivos: mkfile")
	color.Green("-----------------------------------------------------------")

	/* 1️⃣ Validar sesión */
	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	/* 2️⃣ Parámetros */
	filePath := strings.TrimSpace(props["path"])
	rFlag := false
	size := int32(0)

	if filePath == "" {
		return "❌ Error: el parámetro path es obligatorio", true
	}

	if _, ok := props["r"]; ok {
		if props["r"] != "" {
			return "❌ Error: el parámetro -r no recibe valores", true
		}
		rFlag = true
	}

	if val, ok := props["size"]; ok {
		var s int
		_, err := fmt.Sscanf(val, "%d", &s)
		if err != nil || s < 0 {
			return "❌ Error: el parámetro size debe ser un entero >= 0", true
		}
		size = int32(s)
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
		return err.Error(), true
	}

	/* 6️⃣ Procesar ruta */
	cleanPath := path.Clean(filePath)
	parentPath := path.Dir(cleanPath)
	fileName := path.Base(cleanPath)

	if fileName == "" || fileName == "/" {
		return "❌ Error: nombre de archivo inválido", true
	}

	/* 7️⃣ Buscar carpeta padre */
	parentInode, err := traversePath(file, sb, parentPath, rFlag)
	if err != nil {
		return err.Error(), true
	}

	/* 8️⃣ Verificar si archivo ya existe */
	exists, inode := findEntryInDirectory(file, sb, parentInode, fileName)
	if exists {
		color.Yellow("⚠ El archivo ya existe, será sobrescrito")
		writeFileContent(file, sb, inode, size)
		return fmt.Sprintf("✅ Archivo '%s' sobrescrito correctamente", filePath), false
	}

	/* 9️⃣ Crear archivo */
	newInode := FindFreeInode(file, sb)
	if newInode == -1 {
		return "❌ Error: no hay inodos libres", true
	}

	newBlock := FindFreeBlock(file, sb)
	if newBlock == -1 {
		return "❌ Error: no hay bloques libres", true
	}

	now := int32(time.Now().Unix())

	var inodeFile structures.Inode
	inodeFile.I_uid = currentSession.Uid
	inodeFile.I_gid = currentSession.Gid
	inodeFile.I_s = size
	inodeFile.I_atime = now
	inodeFile.I_ctime = now
	inodeFile.I_mtime = now
	inodeFile.I_type = 1 // archivo
	inodeFile.I_perm = [3]byte{6, 6, 4}

	for i := 0; i < 15; i++ {
		inodeFile.I_block[i] = -1
	}
	inodeFile.I_block[0] = newBlock

	WriteInode(file, sb, newInode, inodeFile)
	MarkBitmap(file, sb.S_bm_inode_start, newInode)
	MarkBitmap(file, sb.S_bm_block_start, newBlock)

	/* 🔟 Escribir contenido */
	writeFileContent(file, sb, newInode, size)

	/* 1️⃣1️⃣ Agregar al directorio padre */
	if err := addEntryToDirectory(file, sb, parentInode, fileName, newInode); err != nil {
		return err.Error(), true
	}

	return fmt.Sprintf("✅ Archivo '%s' creado correctamente", filePath), false
}

/* =========================
   FUNCIONES AUXILIARES
========================= */

func writeFileContent(file *os.File, sb structures.SuperBlock, inodeIndex int32, size int32) {

	var content structures.BloqueArchivo
	for i := int32(0); i < size && i < 64; i++ {
		content.B_content[i] = byte('0' + (i % 10))
	}

	inode, _ := ReadInode(file, sb, inodeIndex)
	block := inode.I_block[0]

	WriteBlock(file, sb, block, &content)
}

func traversePath(file *os.File, sb structures.SuperBlock, p string, create bool) (int32, error) {

	if p == "/" {
		return 0, nil
	}

	dirs := strings.Split(strings.Trim(p, "/"), "/")
	current := int32(0)

	for i, dir := range dirs {

		found, inode := findEntryInDirectory(file, sb, current, dir)
		if found {
			current = inode
			continue
		}

		if !create {
			return -1, fmt.Errorf("❌ Error: la carpeta '%s' no existe", dir)
		}

		newInode, err := createDirectory(file, sb, current, dir)
		if err != nil {
			return -1, err
		}

		current = newInode

		if i == len(dirs)-1 {
			break
		}
	}

	return current, nil
}

func findEntryInDirectory(file *os.File, sb structures.SuperBlock, dirInode int32, name string) (bool, int32) {

	inode, err := ReadInode(file, sb, dirInode)
	if err != nil {
		return false, -1
	}

	for _, blk := range inode.I_block {
		if blk == -1 {
			continue
		}

		var block structures.BloqueCarpeta
		ReadBlock(file, sb, blk, &block)

		for _, c := range block.B_content {
			n := strings.TrimRight(string(c.B_name[:]), "\x00")
			if n == name {
				return true, c.B_inodo
			}
		}
	}

	return false, -1
}
