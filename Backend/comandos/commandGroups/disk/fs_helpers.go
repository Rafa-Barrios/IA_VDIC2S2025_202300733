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

func ReadSuperBlock(file *os.File, sb *structures.SuperBlock) error {
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
	file.Seek(int64(inodePos), 0)

	if err := binary.Read(file, binary.LittleEndian, &inode); err != nil {
		return inode, fmt.Errorf("error al leer inodo %d", inodeIndex)
	}

	return inode, nil
}

func WriteInode(file *os.File, sb structures.SuperBlock, inodeIndex int32, inode structures.Inode) error {
	inodePos := sb.S_inode_start + inodeIndex*sb.S_inode_s
	file.Seek(int64(inodePos), 0)

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
	file.Seek(int64(blockPos), 0)

	if err := binary.Read(file, binary.LittleEndian, out); err != nil {
		return fmt.Errorf("error al leer bloque %d", blockIndex)
	}

	return nil
}

func WriteBlock(file *os.File, sb structures.SuperBlock, blockIndex int32, data interface{}) error {
	blockPos := sb.S_block_start + blockIndex*sb.S_block_s
	file.Seek(int64(blockPos), 0)

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

		b := make([]byte, 1)
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

		b := make([]byte, 1)
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

	for i, blk := range parent.I_block {

		if blk == -1 {
			newBlock := FindFreeBlock(file, sb)
			if newBlock == -1 {
				return fmt.Errorf("no hay bloques libres para el directorio padre")
			}

			parent.I_block[i] = newBlock
			WriteInode(file, sb, parentInode, parent)
			MarkBitmap(file, sb.S_bm_block_start, newBlock)

			var dirBlock structures.BloqueCarpeta
			for j := 0; j < 4; j++ {
				dirBlock.B_content[j].B_inodo = -1
			}

			copy(dirBlock.B_content[0].B_name[:], name)
			dirBlock.B_content[0].B_inodo = childInode

			return WriteBlock(file, sb, newBlock, &dirBlock)
		}

		var block structures.BloqueCarpeta
		ReadBlock(file, sb, blk, &block)

		for j := 0; j < 4; j++ {
			if block.B_content[j].B_inodo == -1 {
				copy(block.B_content[j].B_name[:], name)
				block.B_content[j].B_inodo = childInode
				return WriteBlock(file, sb, blk, &block)
			}
		}
	}

	return fmt.Errorf("no hay espacio en el directorio padre")
}
