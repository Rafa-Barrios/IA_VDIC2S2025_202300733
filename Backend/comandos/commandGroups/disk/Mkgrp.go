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
   MKGRP
========================= */

func mkgrpExecute(_ string, props map[string]string) (string, bool) {

	color.Green("-----------------------------------------------------------")
	color.Blue("Administración de grupos: mkgrp")
	color.Green("-----------------------------------------------------------")

	// 1️⃣ Verificar sesión activa
	if currentSession == nil {
		color.Red("❌ Error: no hay una sesión activa")
		return "❌ Error: no hay una sesión activa", true
	}

	// 2️⃣ Solo root
	if currentSession.User != "root" {
		color.Red("❌ Error: usuario no autorizado (%s)", currentSession.User)
		return "❌ Error: solo el usuario root puede crear grupos", true
	}

	groupName := strings.TrimSpace(props["name"])
	if groupName == "" {
		color.Red("❌ Error: nombre de grupo vacío")
		return "❌ Error: el nombre del grupo es obligatorio", true
	}

	// 3️⃣ Obtener partición de la sesión
	part := GetMountedPartition(currentSession.Id)
	if part == nil {
		color.Red("❌ Error: partición de la sesión no montada")
		return "❌ Error: la partición de la sesión no está montada", true
	}

	color.Cyan("✔ Partición activa: %s", part.Id)

	// 4️⃣ Abrir disco
	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		color.Red("❌ Error al abrir disco")
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	// 5️⃣ Leer SuperBloque
	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		color.Red("❌ Error al leer SuperBloque")
		return "❌ Error al leer el SuperBloque", true
	}

	// 6️⃣ Leer users.txt
	usersBlockPos := sb.S_block_start
	buffer := make([]byte, sb.S_block_s)

	file.Seek(int64(usersBlockPos), 0)
	file.Read(buffer)

	content := strings.TrimRight(string(buffer), "\x00")
	lines := strings.Split(content, "\n")

	// 7️⃣ Verificar duplicados y obtener nuevo ID
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
			color.Red("❌ Error: el grupo '%s' ya existe", groupName)
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

	color.Green("✔ Creando grupo: %s (ID=%d)", groupName, newID)

	newContent := content + newLine
	if len(newContent) > int(sb.S_block_s) {
		color.Red("❌ Error: users.txt sin espacio")
		return "❌ Error: no hay espacio suficiente en users.txt", true
	}

	// Limpiar bloque
	file.Seek(int64(usersBlockPos), 0)
	file.Write(make([]byte, sb.S_block_s))

	// Escribir actualizado
	file.Seek(int64(usersBlockPos), 0)
	file.Write([]byte(newContent))

	// Actualizar tiempo
	sb.S_mtime = int32(time.Now().Unix())
	file.Seek(int64(part.Start), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	color.Green("-----------------------------------------------------------")
	color.Green("✅ Grupo creado correctamente")
	color.Green("-----------------------------------------------------------")

	return fmt.Sprintf("✅ Grupo '%s' creado correctamente", groupName), false
}
