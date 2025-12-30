package disk

import (
	"Proyecto/Estructuras/structures"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fatih/color"
)

func mkfileExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de archivos: mkfile")
	color.Green("-----------------------------------------------------------")

	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	filePath := strings.TrimSpace(props["path"])
	rFlag := false
	size := int32(0)

	if filePath == "" {
		return "❌ Error: el parámetro path es obligatorio", true
	}

	if _, ok := props["r"]; ok {
		rFlag = true
	}

	if val, ok := props["size"]; ok {
		var s int
		_, err := fmt.Sscanf(val, "%d", &s)
		if err != nil || s < 0 {
			return "❌ Error: el parámetro size debe ser >= 0", true
		}
		size = int32(s)
	}

	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		return "❌ Error: la partición no está montada", true
	}

	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	var sb structures.SuperBlock
	if err := ReadSuperBlock(file, int64(part.Start), &sb); err != nil {
		return err.Error(), true
	}

	cleanPath := path.Clean(filePath)
	parentPath := path.Dir(cleanPath)
	fileName := path.Base(cleanPath)

	parentInode, err := traversePath(file, sb, parentPath, rFlag)
	if err != nil {
		return err.Error(), true
	}

	exists, inodeIndex := findEntryInDirectory(file, sb, parentInode, fileName)
	if exists {
		color.Yellow("⚠ El archivo ya existe, será sobrescrito")
		cleanFileBlocks(file, sb, inodeIndex)
		writeFileContentSafe(file, sb, inodeIndex, size)
		return fmt.Sprintf("✅ Archivo '%s' sobrescrito correctamente", filePath), false
	}

	inodeIndex = FindFreeInode(file, sb)
	if inodeIndex == -1 {
		return "❌ Error: no hay inodos libres", true
	}

	now := int32(time.Now().Unix())

	inode := structures.Inode{
		I_uid:   currentSession.Uid,
		I_gid:   currentSession.Gid,
		I_s:     0,
		I_atime: now,
		I_ctime: now,
		I_mtime: now,
		I_type:  1,
		I_perm:  [3]byte{6, 6, 4},
	}

	for i := 0; i < 15; i++ {
		inode.I_block[i] = -1
	}

	WriteInode(file, sb, inodeIndex, inode)
	MarkBitmap(file, sb.S_bm_inode_start, inodeIndex)

	writeFileContentSafe(file, sb, inodeIndex, size)

	if err := addEntryToDirectory(file, sb, parentInode, fileName, inodeIndex); err != nil {
		return err.Error(), true
	}

	return fmt.Sprintf("✅ Archivo '%s' creado correctamente", filePath), false
}

// LIMPIAR BLOQUES
func cleanFileBlocks(file *os.File, sb structures.SuperBlock, inodeIndex int32) {
	inode, err := ReadInode(file, sb, inodeIndex)
	if err != nil {
		return
	}

	for i, blk := range inode.I_block {
		if blk != -1 {
			UnmarkBitmap(file, sb.S_bm_block_start, blk)
			inode.I_block[i] = -1
		}
	}

	inode.I_s = 0
	WriteInode(file, sb, inodeIndex, inode)
}

func writeFileContentSafe(file *os.File, sb structures.SuperBlock, inodeIndex int32, size int32) {

	inode, err := ReadInode(file, sb, inodeIndex)
	if err != nil || inode.I_type != 1 {
		return
	}

	var written int32 = 0

	for i := 0; i < 15 && written < size; i++ {

		if inode.I_block[i] == -1 {
			blk := FindFreeBlock(file, sb)
			if blk == -1 {
				break
			}
			inode.I_block[i] = blk
			MarkBitmap(file, sb.S_bm_block_start, blk)
		}

		var block structures.BloqueArchivo

		for j := 0; j < 64 && written < size; j++ {
			block.B_content[j] = byte('0' + (written % 10))
			written++
		}

		WriteBlock(file, sb, inode.I_block[i], &block)

		fmt.Printf("DEBUG MKFILE: bloque=%d escritos=%d\n", inode.I_block[i], written)
	}

	inode.I_s = size
	inode.I_mtime = int32(time.Now().Unix())
	WriteInode(file, sb, inodeIndex, inode)
}
