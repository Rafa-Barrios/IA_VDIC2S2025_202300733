package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"Proyecto/Estructuras/structures"
)

/* =========================
   SUPER BLOQUE
========================= */

func ReadSuperBlock(file *os.File, start int64, sb *structures.SuperBlock) error {
	if _, err := file.Seek(start, 0); err != nil {
		return fmt.Errorf("error al posicionar el SuperBloque")
	}
	if err := binary.Read(file, binary.LittleEndian, sb); err != nil {
		return fmt.Errorf("error al leer el SuperBloque")
	}
	return nil
}

/* =========================
   INODOS
========================= */

func ReadInode(file *os.File, sb structures.SuperBlock, inodeIndex int32) (structures.Inode, error) {
	var inode structures.Inode
	pos := sb.S_inode_start + inodeIndex*sb.S_inode_s

	if _, err := file.Seek(int64(pos), 0); err != nil {
		return inode, fmt.Errorf("error al posicionar inodo %d", inodeIndex)
	}
	if err := binary.Read(file, binary.LittleEndian, &inode); err != nil {
		return inode, fmt.Errorf("error al leer inodo %d", inodeIndex)
	}
	return inode, nil
}

func WriteInode(file *os.File, sb structures.SuperBlock, inodeIndex int32, inode structures.Inode) error {
	pos := sb.S_inode_start + inodeIndex*sb.S_inode_s

	if _, err := file.Seek(int64(pos), 0); err != nil {
		return fmt.Errorf("error al posicionar inodo %d", inodeIndex)
	}
	if err := binary.Write(file, binary.LittleEndian, &inode); err != nil {
		return fmt.Errorf("error al escribir inodo %d", inodeIndex)
	}
	return nil
}

/* =========================
   BLOQUES
========================= */

func ReadBlock(file *os.File, sb structures.SuperBlock, blockIndex int32, out interface{}) error {
	pos := sb.S_block_start + blockIndex*sb.S_block_s

	if _, err := file.Seek(int64(pos), 0); err != nil {
		return fmt.Errorf("error al posicionar bloque %d", blockIndex)
	}
	if err := binary.Read(file, binary.LittleEndian, out); err != nil {
		return fmt.Errorf("error al leer bloque %d", blockIndex)
	}
	return nil
}

func WriteBlock(file *os.File, sb structures.SuperBlock, blockIndex int32, data interface{}) error {
	pos := sb.S_block_start + blockIndex*sb.S_block_s

	if _, err := file.Seek(int64(pos), 0); err != nil {
		return fmt.Errorf("error al posicionar bloque %d", blockIndex)
	}
	if err := binary.Write(file, binary.LittleEndian, data); err != nil {
		return fmt.Errorf("error al escribir bloque %d", blockIndex)
	}
	return nil
}

/* =========================
   BITMAPS
========================= */

func FindFreeInode(file *os.File, sb structures.SuperBlock) int32 {
	for i := int32(0); i < sb.S_inodes_count; i++ {
		pos := sb.S_bm_inode_start + i
		file.Seek(int64(pos), 0)
		b := []byte{0}
		file.Read(b)
		if b[0] == 0 {
			return i
		}
	}
	return -1
}

func FindFreeBlock(file *os.File, sb structures.SuperBlock) int32 {
	for i := int32(0); i < sb.S_blocks_count; i++ {
		pos := sb.S_bm_block_start + i
		file.Seek(int64(pos), 0)
		b := []byte{0}
		file.Read(b)
		if b[0] == 0 {
			return i
		}
	}
	return -1
}

func MarkBitmap(file *os.File, bmStart int32, index int32) {
	pos := bmStart + index
	file.Seek(int64(pos), 0)
	file.Write([]byte{1})
}

func UnmarkBitmap(file *os.File, bmStart int32, index int32) {
	pos := bmStart + index
	file.Seek(int64(pos), 0)
	file.Write([]byte{0})
}

/* =========================
   DIRECTORIOS
========================= */

func findEntryInDirectory(
	file *os.File,
	sb structures.SuperBlock,
	dirInode int32,
	name string,
) (bool, int32) {

	if name == "" {
		return false, -1
	}

	inode, err := ReadInode(file, sb, dirInode)
	if err != nil || inode.I_type != 0 {
		return false, -1
	}

	for _, blk := range inode.I_block {
		if blk == -1 {
			continue
		}

		var folder structures.BloqueCarpeta
		ReadBlock(file, sb, blk, &folder)

		for _, entry := range folder.B_content {
			entryName := strings.TrimRight(string(entry.B_name[:]), "\x00")
			if entryName == name {
				return true, entry.B_inodo
			}
		}
	}

	return false, -1
}

func traversePath(
	file *os.File,
	sb structures.SuperBlock,
	p string,
	create bool,
) (int32, error) {

	if p == "/" {
		return 0, nil
	}

	parts := strings.Split(strings.Trim(p, "/"), "/")
	current := int32(0)

	for _, dir := range parts {
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
	}

	return current, nil
}

func createDirectory(
	file *os.File,
	sb structures.SuperBlock,
	parent int32,
	name string,
) (int32, error) {

	newInode := FindFreeInode(file, sb)
	newBlock := FindFreeBlock(file, sb)

	if newInode == -1 || newBlock == -1 {
		return -1, fmt.Errorf("❌ Error: no hay espacio para crear carpeta")
	}

	var inode structures.Inode
	inode.I_type = 0
	inode.I_perm = [3]byte{7, 7, 5}
	inode.I_block[0] = newBlock
	for i := 1; i < 15; i++ {
		inode.I_block[i] = -1
	}

	WriteInode(file, sb, newInode, inode)
	MarkBitmap(file, sb.S_bm_inode_start, newInode)
	MarkBitmap(file, sb.S_bm_block_start, newBlock)

	var folder structures.BloqueCarpeta
	for i := 0; i < 4; i++ {
		folder.B_content[i].B_inodo = -1
	}

	WriteBlock(file, sb, newBlock, &folder)
	addEntryToDirectory(file, sb, parent, name, newInode)

	return newInode, nil
}

func addEntryToDirectory(
	file *os.File,
	sb structures.SuperBlock,
	parentInode int32,
	name string,
	childInode int32,
) error {

	if name == "" {
		return fmt.Errorf("nombre de entrada vacío")
	}

	parent, err := ReadInode(file, sb, parentInode)
	if err != nil {
		return err
	}

	for _, blk := range parent.I_block {
		if blk == -1 {
			continue
		}

		var folder structures.BloqueCarpeta
		ReadBlock(file, sb, blk, &folder)

		for i := 0; i < 4; i++ {
			if folder.B_content[i].B_inodo == -1 {
				copy(folder.B_content[i].B_name[:], name)
				folder.B_content[i].B_inodo = childInode
				return WriteBlock(file, sb, blk, &folder)
			}
		}
	}

	return fmt.Errorf("❌ Error: no hay espacio en el directorio")
}
