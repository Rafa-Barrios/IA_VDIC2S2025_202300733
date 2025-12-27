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

/* =========================
   MKUSR
========================= */

func mkusrExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de usuarios: mkusr")
	color.Green("-----------------------------------------------------------")

	// 1️⃣ Verificar sesión activa
	if currentSession == nil {
		color.Red("❌ Error: no hay una sesión activa")
		return "❌ Error: no hay una sesión activa", true
	}

	// 2️⃣ Solo root puede crear usuarios
	if currentSession.User != "root" {
		color.Red("❌ Error: usuario no autorizado (%s)", currentSession.User)
		return "❌ Error: solo el usuario root puede crear usuarios", true
	}

	// 3️⃣ Leer parámetros
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

	// 4️⃣ Obtener partición de la sesión
	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		color.Red("❌ Error: partición de la sesión no montada")
		return "❌ Error: la partición de la sesión no está montada", true
	}

	color.Cyan("✔ Partición activa: %s", part.Id)

	// 5️⃣ Abrir disco
	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		color.Red("❌ Error al abrir disco")
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	// 6️⃣ Leer SuperBloque
	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		color.Red("❌ Error al leer SuperBloque")
		return "❌ Error al leer el SuperBloque", true
	}

	// 7️⃣ Leer inodo de users.txt (inodo 1)
	var usersInode structures.Inode
	inodePos := sb.S_inode_start + sb.S_inode_s // inodo 1
	file.Seek(int64(inodePos), 0)
	if err := binary.Read(file, binary.LittleEndian, &usersInode); err != nil {
		color.Red("❌ Error al leer inodo de users.txt")
		return "❌ Error al leer el inodo de users.txt", true
	}

	blockIndex := usersInode.I_block[0]
	if blockIndex == -1 {
		color.Red("❌ Error: users.txt no tiene bloques asignados")
		return "❌ Error: users.txt sin bloques", true
	}

	usersBlockPos := sb.S_block_start + (blockIndex * sb.S_block_s)
	buffer := make([]byte, usersInode.I_s)
	file.Seek(int64(usersBlockPos), 0)
	file.Read(buffer)

	content := string(buffer)
	lines := strings.Split(strings.TrimSpace(content), "\n")

	// 8️⃣ Verificar existencia de grupo y duplicados de usuario
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

	// 9️⃣ Crear nueva línea y verificar espacio
	newID := maxID + 1
	newLine := fmt.Sprintf("%d,U,%s,%s,%s\n", newID, groupName, userName, password)
	newContent := content + newLine

	if len(newContent) > int(sb.S_block_s) {
		color.Red("❌ Error: users.txt sin espacio")
		return "❌ Error: no hay espacio suficiente en users.txt", true
	}

	// 10️⃣ Limpiar bloque y escribir actualizado
	file.Seek(int64(usersBlockPos), 0)
	file.Write(make([]byte, sb.S_block_s))
	file.Seek(int64(usersBlockPos), 0)
	file.Write([]byte(newContent))

	// 🔹 11️⃣ Actualizar tamaño del inodo para login
	usersInode.I_s = int32(len(newContent))
	file.Seek(int64(inodePos), 0)
	binary.Write(file, binary.LittleEndian, &usersInode)

	// 12️⃣ Actualizar tiempo en SuperBloque
	sb.S_mtime = int32(time.Now().Unix())
	file.Seek(int64(part.Start), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	color.Green("-----------------------------------------------------------")
	color.Green("✅ Usuario creado correctamente")
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("✅ Usuario '%s' creado correctamente", userName), false
}
