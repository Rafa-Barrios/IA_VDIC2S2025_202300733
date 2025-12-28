package disk

import (
	"encoding/binary"
	"fmt"
	"os"

	"Proyecto/Estructuras/structures"
)

/* =========================
   SUPER BLOQUE
========================= */

func ReadSuperBlock(file *os.File, start int64, sb *structures.SuperBlock) error {
	// Posicionarse en el inicio real de la partición
	if _, err := file.Seek(start, 0); err != nil {
		return fmt.Errorf("error al posicionar el SuperBloque")
	}

	// Leer SuperBloque
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

	inodePos := sb.S_inode_start + inodeIndex*sb.S_inode_s
	if _, err := file.Seek(int64(inodePos), 0); err != nil {
		return inode, fmt.Errorf("error al posicionar inodo %d", inodeIndex)
	}

	if err := binary.Read(file, binary.LittleEndian, &inode); err != nil {
		return inode, fmt.Errorf("error al leer inodo %d", inodeIndex)
	}

	return inode, nil
}

func WriteInode(file *os.File, sb structures.SuperBlock, inodeIndex int32, inode structures.Inode) error {
	inodePos := sb.S_inode_start + inodeIndex*sb.S_inode_s
	if _, err := file.Seek(int64(inodePos), 0); err != nil {
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
	blockPos := sb.S_block_start + blockIndex*sb.S_block_s
	if _, err := file.Seek(int64(blockPos), 0); err != nil {
		return fmt.Errorf("error al posicionar bloque %d", blockIndex)
	}

	if err := binary.Read(file, binary.LittleEndian, out); err != nil {
		return fmt.Errorf("error al leer bloque %d", blockIndex)
	}

	return nil
}

func WriteBlock(file *os.File, sb structures.SuperBlock, blockIndex int32, data interface{}) error {
	blockPos := sb.S_block_start + blockIndex*sb.S_block_s
	if _, err := file.Seek(int64(blockPos), 0); err != nil {
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
		if _, err := file.Seek(int64(pos), 0); err != nil {
			continue
		}

		b := make([]byte, 1)
		if _, err := file.Read(b); err != nil {
			continue
		}

		if b[0] == 0 {
			return i
		}
	}
	return -1
}

func FindFreeBlock(file *os.File, sb structures.SuperBlock) int32 {
	for i := int32(0); i < sb.S_blocks_count; i++ {
		pos := sb.S_bm_block_start + i
		if _, err := file.Seek(int64(pos), 0); err != nil {
			continue
		}

		b := make([]byte, 1)
		if _, err := file.Read(b); err != nil {
			continue
		}

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

/* =========================
   DIRECTORIOS
========================= */

func addEntryToDirectory(
	file *os.File,
	sb structures.SuperBlock,
	parentInode int32,
	name string,
	childInode int32,
) error {

	parent, err := ReadInode(file, sb, parentInode)
	if err != nil {
		return err
	}

	// 1️⃣ Buscar espacio en bloques existentes
	for _, blk := range parent.I_block {
		if blk == -1 {
			continue
		}

		var folder structures.BloqueCarpeta
		if err := ReadBlock(file, sb, blk, &folder); err != nil {
			return err
		}

		for i := 0; i < 4; i++ {
			if folder.B_content[i].B_inodo == -1 {

				copy(folder.B_content[i].B_name[:], name)
				folder.B_content[i].B_inodo = childInode

				return WriteBlock(file, sb, blk, &folder)
			}
		}
	}

	// 2️⃣ No hay espacio → crear nuevo bloque
	newBlock := FindFreeBlock(file, sb)
	if newBlock == -1 {
		return fmt.Errorf("❌ Error: no hay bloques disponibles")
	}

	var newFolder structures.BloqueCarpeta
	for i := 0; i < 4; i++ {
		newFolder.B_content[i].B_inodo = -1
	}

	copy(newFolder.B_content[0].B_name[:], name)
	newFolder.B_content[0].B_inodo = childInode

	// 3️⃣ Asignar bloque al inode padre
	for i := 0; i < 15; i++ {
		if parent.I_block[i] == -1 {
			parent.I_block[i] = newBlock
			break
		}
	}

	if err := WriteInode(file, sb, parentInode, parent); err != nil {
		return err
	}

	MarkBitmap(file, sb.S_bm_block_start, newBlock)

	return WriteBlock(file, sb, newBlock, &newFolder)
}
