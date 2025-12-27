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

	if currentSession != nil {
		return "❌ Error: ya existe una sesión activa, debe cerrar sesión primero", true
	}

	user := strings.TrimSpace(props["user"])
	pass := strings.TrimSpace(props["pass"])
	id := strings.TrimSpace(props["id"])

	if user == "" || pass == "" || id == "" {
		return "❌ Error: faltan parámetros obligatorios (user, pass, id)", true
	}

	if len(user) > 10 || len(pass) > 10 {
		return "❌ Error: usuario o contraseña exceden 10 caracteres", true
	}

	part := GetMountedPartition(id)
	if part == nil {
		return "❌ Error: la partición no existe o no está montada", true
	}

	if _, err := os.Stat(part.Path); err != nil {
		return "❌ Error: el disco asociado a la partición no existe", true
	}

	file, err := os.Open(part.Path)
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
	if err := binary.Read(file, binary.LittleEndian, &usersInode); err != nil {
		return "❌ Error al leer el inodo de users.txt", true
	}

	// =========================
	// Leer todos los bloques asignados a users.txt
	// =========================
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

	// DEBUG
	color.Yellow("----------- DEBUG users.txt -----------")
	color.White(content.String())
	color.Yellow("---------------------------------------")

	lines := strings.Split(strings.TrimSpace(content.String()), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			continue
		}

		if fields[1] == "U" && fields[3] == user {
			if fields[4] != pass {
				return "❌ Error: contraseña incorrecta", true
			}

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
