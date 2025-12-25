package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"Proyecto/Estructuras/structures"
)

/* =========================
   SESIÓN ACTIVA
========================= */

type Session struct {
	User  string
	Group string
	Uid   int32
	Gid   int32
	Id    string
}

var currentSession *Session = nil

/* =========================
   LOGIN
========================= */

func loginExecute(_ string, props map[string]string) (string, bool) {

	// 1️⃣ Verificar sesión activa
	if currentSession != nil {
		return "❌ Ya existe una sesión activa, primero ejecute logout", true
	}

	user := props["user"]
	pass := props["pass"]
	id := props["id"]

	// 2️⃣ Verificar partición montada
	part := GetMountedPartition(id)
	if part == nil {
		return "❌ No existe una partición montada con ese ID", true
	}

	// 3️⃣ Abrir disco
	file, err := os.OpenFile(part.Path, os.O_RDONLY, 0666)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	// 4️⃣ Leer SuperBloque
	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		return "❌ Error al leer el SuperBloque", true
	}

	// 5️⃣ Leer inodo users.txt (siempre es el #1)
	var inode structures.Inode
	file.Seek(int64(sb.S_inode_start+sb.S_inode_s), 0)
	binary.Read(file, binary.LittleEndian, &inode)

	// 6️⃣ Leer bloque users.txt
	var block structures.BloqueArchivo
	file.Seek(int64(sb.S_block_start+sb.S_block_s), 0)
	binary.Read(file, binary.LittleEndian, &block)

	content := strings.Trim(string(block.B_content[:]), "\x00")
	lines := strings.Split(content, "\n")

	// 7️⃣ Buscar usuario
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		if fields[1] == "U" {
			u := strings.TrimSpace(fields[3])
			p := strings.TrimSpace(fields[4])

			if u == user {
				if p != pass {
					return "❌ Contraseña incorrecta", true
				}

				// LOGIN OK
				currentSession = &Session{
					User:  u,
					Group: strings.TrimSpace(fields[2]),
					Uid:   1,
					Gid:   1,
					Id:    id,
				}

				return fmt.Sprintf("✅ Sesión iniciada como %s", user), false
			}
		}
	}

	return "❌ Usuario no encontrado", true
}

/* =========================
   LOGOUT
========================= */

func logoutExecute(_ string, _ map[string]string) (string, bool) {

	if currentSession == nil {
		return "❌ No hay ninguna sesión activa", true
	}

	user := currentSession.User
	currentSession = nil

	return fmt.Sprintf("✅ Sesión cerrada correctamente (%s)", user), false
}
