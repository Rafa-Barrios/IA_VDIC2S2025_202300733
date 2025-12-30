package disk

import (
	"Proyecto/Estructuras/structures"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func mkusrExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de usuarios: mkusr")
	color.Green("-----------------------------------------------------------")

	if currentSession == nil {
		color.Red("❌ Error: no hay una sesión activa")
		return "❌ Error: no hay una sesión activa", true
	}

	if currentSession.User != "root" {
		color.Red("❌ Error: usuario no autorizado (%s)", currentSession.User)
		return "❌ Error: solo el usuario root puede crear usuarios", true
	}

	userName := strings.TrimSpace(props["user"])
	password := strings.TrimSpace(props["pass"])
	groupName := strings.TrimSpace(props["grp"])

	if userName == "" || password == "" || groupName == "" {
		color.Red("❌ Error: faltan parámetros obligatorios")
		return "❌ Error: los parámetros user, pass y grp son obligatorios", true
	}

	if len(userName) > 10 || len(password) > 10 || len(groupName) > 10 {
		color.Red("❌ Error: longitud máxima 10 caracteres por parámetro")
		return "❌ Error: parámetros exceden longitud máxima de 10 caracteres", true
	}

	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		color.Red("❌ Error: partición de la sesión no montada")
		return "❌ Error: la partición de la sesión no está montada", true
	}

	color.Cyan("✔ Partición activa: %s", part.Id)

	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		color.Red("❌ Error al abrir disco")
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		color.Red("❌ Error al leer SuperBloque")
		return "❌ Error al leer el SuperBloque", true
	}

	var usersInode structures.Inode
	inodePos := sb.S_inode_start + sb.S_inode_s
	file.Seek(int64(inodePos), 0)
	if err := binary.Read(file, binary.LittleEndian, &usersInode); err != nil {
		color.Red("❌ Error al leer inodo de users.txt")
		return "❌ Error al leer el inodo de users.txt", true
	}

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
	groupExists := false
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
			groupExists = true
		}

		if fields[1] == "U" && len(fields) >= 5 && fields[3] == userName {
			color.Red("❌ Error: el usuario '%s' ya existe", userName)
			return "❌ Error: el usuario ya existe", true
		}

		var id int
		fmt.Sscanf(fields[0], "%d", &id)
		if id > maxID {
			maxID = id
		}
	}

	if !groupExists {
		color.Red("❌ Error: el grupo '%s' no existe", groupName)
		return "❌ Error: el grupo indicado no existe", true
	}

	newID := maxID + 1
	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", newID, groupName, userName, password)
	newContent := content.String() + newLine

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
			color.Red("❌ Error: no hay bloques libres disponibles")
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

	// 14️⃣ Actualizar SuperBloque
	sb.S_mtime = int32(time.Now().Unix())
	file.Seek(int64(part.Start), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	color.Green("-----------------------------------------------------------")
	color.Green("✅ Usuario creado correctamente")
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("✅ Usuario '%s' creado correctamente", userName), false
}
