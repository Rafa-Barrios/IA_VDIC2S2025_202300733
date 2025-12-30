package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"Proyecto/Estructuras/structures"

	"github.com/fatih/color"
)

func mkgrpExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de grupos: mkgrp")
	color.Green("-----------------------------------------------------------")

	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	if currentSession.User != "root" {
		return "❌ Error: solo el usuario root puede crear grupos", true
	}

	groupName := strings.TrimSpace(props["name"])
	if groupName == "" {
		return "❌ Error: el nombre del grupo es obligatorio", true
	}

	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		return "❌ Error: la partición no está montada", true
	}

	color.Cyan("✔ Partición activa: %s", part.Id)

	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		return "❌ Error al leer el SuperBloque", true
	}

	var usersInode structures.Inode
	inodePos := sb.S_inode_start + sb.S_inode_s
	file.Seek(int64(inodePos), 0)
	binary.Read(file, binary.LittleEndian, &usersInode)

	var content strings.Builder
	for _, blk := range usersInode.I_block {
		if blk == -1 {
			continue
		}
		blockPos := sb.S_block_start + blk*sb.S_block_s
		buffer := make([]byte, sb.S_block_s)
		file.Seek(int64(blockPos), 0)
		file.Read(buffer)
		content.WriteString(strings.TrimRight(string(buffer), "\x00"))
	}

	lines := strings.Split(strings.TrimSpace(content.String()), "\n")

	maxID := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			continue
		}

		if fields[1] == "G" && fields[2] == groupName {
			return "❌ Error: el grupo ya existe", true
		}

		var id int
		fmt.Sscanf(fields[0], "%d", &id)
		if id > maxID {
			maxID = id
		}
	}

	newID := maxID + 1
	newLine := fmt.Sprintf("%d,G,%s\n", newID, groupName)
	newContent := content.String() + newLine

	color.Green("✔ Creando grupo: %s (ID=%d)", groupName, newID)

	blockSize := int(sb.S_block_s)
	requiredBlocks := (len(newContent) + blockSize - 1) / blockSize

	currentBlocks := 0
	for _, blk := range usersInode.I_block {
		if blk != -1 {
			currentBlocks++
		}
	}

	for currentBlocks < requiredBlocks {
		freeBlock := findFreeBlock(file, sb)
		if freeBlock == -1 {
			return "❌ Error: no hay bloques libres disponibles", true
		}

		for i := 0; i < 15; i++ {
			if usersInode.I_block[i] == -1 {
				usersInode.I_block[i] = freeBlock
				markBitmap(file, sb.S_bm_block_start, freeBlock)
				currentBlocks++
				break
			}
		}
	}

	offset := 0
	for _, blk := range usersInode.I_block {
		if blk == -1 {
			continue
		}

		blockPos := sb.S_block_start + blk*sb.S_block_s
		file.Seek(int64(blockPos), 0)

		end := offset + blockSize
		if end > len(newContent) {
			end = len(newContent)
		}

		data := make([]byte, blockSize)
		copy(data, newContent[offset:end])
		file.Write(data)

		offset = end
		if offset >= len(newContent) {
			break
		}
	}

	usersInode.I_s = int32(len(newContent))
	usersInode.I_mtime = int32(time.Now().Unix())

	file.Seek(int64(inodePos), 0)
	binary.Write(file, binary.LittleEndian, &usersInode)

	sb.S_mtime = int32(time.Now().Unix())
	file.Seek(int64(part.Start), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	color.Green("-----------------------------------------------------------")
	color.Green("✅ Grupo creado correctamente")
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("✅ Grupo '%s' creado correctamente", groupName), false
}

func findFreeBlock(file *os.File, sb structures.SuperBlock) int32 {
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
