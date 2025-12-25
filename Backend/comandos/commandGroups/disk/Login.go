package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"Proyecto/Estructuras/structures"
)

/* =========================
   SESIÓN GLOBAL
========================= */

type Session struct {
	User  string
	Group string
	Id    string
	Uid   int32
	Gid   int32
}

var currentSession *Session = nil

/* =========================
   LOGIN
========================= */

func loginExecute(_ string, props map[string]string) (string, bool) {

	// 1️⃣ No permitir login si ya hay sesión activa
	if currentSession != nil {
		return "❌ Error: ya existe una sesión activa, debe cerrar sesión primero", true
	}

	user := strings.TrimSpace(props["user"])
	pass := strings.TrimSpace(props["pass"])
	id := strings.TrimSpace(props["id"])

	// Validar parámetros obligatorios
	if user == "" || pass == "" || id == "" {
		return "❌ Error: faltan parámetros obligatorios (user, pass, id)", true
	}

	// 2️⃣ Validar que la partición esté montada
	part := GetMountedPartition(id)
	if part == nil {
		return "❌ Error: la partición no está montada", true
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

	// 5️⃣ Leer contenido de users.txt (primer bloque de datos)
	usersBlockPos := sb.S_block_start
	buffer := make([]byte, sb.S_block_s)

	file.Seek(int64(usersBlockPos), 0)
	file.Read(buffer)

	lines := strings.Split(string(buffer), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		// Formato: UID,U,grupo,user,pass
		if fields[1] == "U" && fields[3] == user {

			if fields[4] != pass {
				return "❌ Error: contraseña incorrecta", true
			}

			// LOGIN EXITOSO
			currentSession = &Session{
				User:  fields[3],
				Group: fields[2],
				Id:    id,
				Uid:   1,
				Gid:   1,
			}

			return fmt.Sprintf("✅ Sesión iniciada correctamente como %s", user), false
		}
	}

	return "❌ Error: usuario no existe", true
}

/* =========================
   LOGOUT
========================= */

func logoutExecute(_ string, _ map[string]string) (string, bool) {

	// 3️⃣ No permitir logout si no hay sesión activa
	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	currentSession = nil
	return "✅ Sesión cerrada correctamente", false
}
