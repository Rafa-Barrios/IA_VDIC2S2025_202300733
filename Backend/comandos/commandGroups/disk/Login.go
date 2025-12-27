package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"Proyecto/Estructuras/structures"

	"github.com/fatih/color"
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

/*
	=========================
	  LOGIN

=========================
*/
func loginExecute(_ string, props map[string]string) (string, bool) {

	// 1️⃣ No permitir login si ya hay sesión activa
	if currentSession != nil {
		return "❌ Error: ya existe una sesión activa, debe cerrar sesión primero", true
	}

	user := strings.TrimSpace(props["user"])
	pass := strings.TrimSpace(props["pass"])
	id := strings.TrimSpace(props["id"])

	// 2️⃣ Validar parámetros obligatorios
	if user == "" || pass == "" || id == "" {
		return "❌ Error: faltan parámetros obligatorios (user, pass, id)", true
	}

	// 2.1️⃣ Validar máximo 10 caracteres
	if len(user) > 10 || len(pass) > 10 {
		return "❌ Error: usuario o contraseña exceden 10 caracteres", true
	}

	// 3️⃣ Validar que la partición esté montada
	part := GetMountedPartition(id)
	if part == nil {
		return "❌ Error: la partición no existe o no está montada", true
	}

	// 4️⃣ Validar que el disco exista físicamente
	if _, err := os.Stat(part.Path); err != nil {
		return "❌ Error: el disco asociado a la partición no existe", true
	}

	// 5️⃣ Abrir disco
	file, err := os.Open(part.Path)
	if err != nil {
		return "❌ Error al abrir el disco", true
	}
	defer file.Close()

	// 6️⃣ Leer SuperBloque
	var sb structures.SuperBlock
	file.Seek(int64(part.Start), 0)
	if err := binary.Read(file, binary.LittleEndian, &sb); err != nil {
		return "❌ Error al leer el SuperBloque", true
	}

	/* =========================
	   LEER users.txt CORRECTAMENTE
	========================= */

	// Inodo 1 = users.txt (según mkfs)
	var usersInode structures.Inode
	inodePos := sb.S_inode_start + sb.S_inode_s // inodo 1
	file.Seek(int64(inodePos), 0)
	if err := binary.Read(file, binary.LittleEndian, &usersInode); err != nil {
		return "❌ Error al leer el inodo de users.txt", true
	}

	blockIndex := usersInode.I_block[0]
	if blockIndex == -1 {
		return "❌ Error: users.txt no tiene bloques asignados", true
	}

	// Posición física real del bloque
	usersBlockPos := sb.S_block_start + (blockIndex * sb.S_block_s)

	buffer := make([]byte, usersInode.I_s)
	file.Seek(int64(usersBlockPos), 0)
	file.Read(buffer)

	// DEBUG REAL
	color.Yellow("----------- DEBUG users.txt -----------")
	color.White(string(buffer))
	color.Yellow("---------------------------------------")

	/* =========================
	   PROCESAR CONTENIDO
	========================= */

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

	if currentSession == nil {
		return "❌ Error: no hay una sesión activa", true
	}

	currentSession = nil
	return "✅ Sesión cerrada correctamente", false
}
