package disk

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"Proyecto/Estructuras/structures"
)

/* =========================
   COMANDO MKFS
========================= */

type MKFS struct {
	Id   string
	Type string
}

/* =========================
   EJECUCIÓN PRINCIPAL
========================= */

func (mkfs *MKFS) Execute() {

	// 1️⃣ Validar tipo (solo EXT2)
	if mkfs.Type != "" && mkfs.Type != "full" && mkfs.Type != "fast" {
		fmt.Println("❌ Error: tipo de formato no válido")
		return
	}

	// 2️⃣ Verificar ID montado
	part := GetMountedPartition(mkfs.Id)
	if part == nil {
		fmt.Println("❌ Error: No existe una partición montada con ese ID")
		return
	}

	// 3️⃣ Abrir disco
	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error al abrir el disco")
		return
	}
	defer file.Close()

	// 4️⃣ Limpiar área de la partición (FULL FORMAT)
	file.Seek(int64(part.Start), 0)
	zero := make([]byte, part.Size)
	file.Write(zero)

	// 5️⃣ Calcular estructuras EXT2
	size := part.Size
	sbSize := int32(binary.Size(structures.SuperBlock{}))
	inodeSize := int32(binary.Size(structures.Inode{}))
	blockSize := int32(64)

	n := (size - sbSize) / (inodeSize + 3*blockSize)
	if n <= 0 {
		fmt.Println("❌ Error: espacio insuficiente para EXT2")
		return
	}

	sb := structures.SuperBlock{
		S_filesystem_type:   2,
		S_inodes_count:      n,
		S_blocks_count:      n * 3,
		S_free_inodes_count: n - 2, // root y users.txt ocupan 2
		S_free_blocks_count: (n * 3) - 2,
		S_mtime:             int32(time.Now().Unix()),
		S_umtime:            0,
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_s:           inodeSize,
		S_block_s:           blockSize,
		S_first_ino:         2,
		S_first_blo:         2,
	}

	// 6️⃣ Posiciones físicas
	sb.S_bm_inode_start = part.Start + sbSize
	sb.S_bm_block_start = sb.S_bm_inode_start + n
	sb.S_inode_start = sb.S_bm_block_start + (n * 3)
	sb.S_block_start = sb.S_inode_start + (n * inodeSize)

	// 7️⃣ Escribir SuperBloque
	file.Seek(int64(part.Start), 0)
	if err := binary.Write(file, binary.LittleEndian, &sb); err != nil {
		fmt.Println("❌ Error al escribir el SuperBloque")
		return
	}

	// 8️⃣ Inicializar Bitmaps
	initBitmap(file, sb.S_bm_inode_start, n)
	initBitmap(file, sb.S_bm_block_start, n*3)

	// 9️⃣ Crear raíz y users.txt con bloques asignados y bitmaps marcados
	createRootAndUsers(file, sb)

	fmt.Println("✅ MKFS realizado correctamente en EXT2")
}

/* =========================
   FUNCIONES AUXILIARES
========================= */

func initBitmap(file *os.File, start int32, size int32) {
	file.Seek(int64(start), 0)
	for i := int32(0); i < size; i++ {
		file.Write([]byte{0})
	}
}

func markBitmap(file *os.File, start int32, index int32) {
	file.Seek(int64(start+index), 0)
	file.Write([]byte{1})
}

func createRootAndUsers(file *os.File, sb structures.SuperBlock) {

	now := int32(time.Now().Unix())

	// ---- INODO RAÍZ ----
	root := structures.Inode{
		I_uid:   0,
		I_gid:   0,
		I_s:     sb.S_block_s,
		I_atime: now,
		I_ctime: now,
		I_mtime: now,
		I_type:  0,
		I_perm:  [3]byte{7, 7, 7},
	}
	for i := 0; i < 15; i++ {
		root.I_block[i] = -1
	}
	root.I_block[0] = 0 // primer bloque para root

	// ---- INODO users.txt ----
	users := root
	users.I_type = 1
	users.I_perm = [3]byte{6, 6, 4}
	users.I_block[0] = 1
	users.I_s = int32(len("1,G,root\n1,U,root,root,123\n"))

	// ---- BLOQUE CARPETA RAÍZ ----
	var folder structures.BloqueCarpeta
	copy(folder.B_content[0].B_name[:], ".")
	folder.B_content[0].B_inodo = 0
	copy(folder.B_content[1].B_name[:], "..")
	folder.B_content[1].B_inodo = 0
	copy(folder.B_content[2].B_name[:], "users.txt")
	folder.B_content[2].B_inodo = 1
	folder.B_content[3].B_inodo = -1

	// ---- BLOQUE users.txt ----
	content := "1,G,root\n1,U,root,root,123\n"
	var fileBlock structures.BloqueArchivo
	copy(fileBlock.B_content[:], content)

	// ---- ESCRITURA INODOS ----
	file.Seek(int64(sb.S_inode_start), 0)
	binary.Write(file, binary.LittleEndian, &root)
	binary.Write(file, binary.LittleEndian, &users)

	// ---- ESCRITURA BLOQUES ----
	file.Seek(int64(sb.S_block_start), 0)
	binary.Write(file, binary.LittleEndian, &folder)
	binary.Write(file, binary.LittleEndian, &fileBlock)

	// ---- BITMAPS ----
	markBitmap(file, sb.S_bm_inode_start, 0) // root
	markBitmap(file, sb.S_bm_inode_start, 1) // users.txt
	markBitmap(file, sb.S_bm_block_start, 0) // root
	markBitmap(file, sb.S_bm_block_start, 1) // users.txt
}
