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

	// 1️⃣ Verificar ID montado (VIENE DE mount.go)
	part := GetMountedPartition(mkfs.Id)
	if part == nil {
		fmt.Println("❌ Error: No existe una partición montada con ese ID")
		return
	}

	// 2️⃣ Abrir disco
	file, err := os.OpenFile(part.Path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("❌ Error al abrir el disco")
		return
	}
	defer file.Close()

	// 3️⃣ Calcular estructuras EXT2
	size := part.Size
	sbSize := int32(binary.Size(structures.SuperBlock{}))
	inodeSize := int32(binary.Size(structures.Inode{}))
	blockSize := int32(64)

	// Fórmula EXT2
	n := (size - sbSize) / (inodeSize + 3*blockSize)

	if n <= 0 {
		fmt.Println("❌ Error: espacio insuficiente para EXT2")
		return
	}

	sb := structures.SuperBlock{
		S_filesystem_type:   2,
		S_inodes_count:      n,
		S_blocks_count:      n * 3,
		S_free_inodes_count: n - 2,
		S_free_blocks_count: (n * 3) - 2,
		S_mtime:             int32(time.Now().Unix()),
		S_umtime:            0,
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_s:           inodeSize,
		S_block_s:           blockSize,
	}

	// 4️⃣ Calcular posiciones
	sb.S_bm_inode_start = part.Start + sbSize
	sb.S_bm_block_start = sb.S_bm_inode_start + n
	sb.S_inode_start = sb.S_bm_block_start + (n * 3)
	sb.S_block_start = sb.S_inode_start + (n * inodeSize)

	sb.S_first_ino = sb.S_inode_start + (2 * inodeSize)
	sb.S_first_blo = sb.S_block_start + (2 * blockSize)

	// 5️⃣ Escribir SuperBloque
	file.Seek(int64(part.Start), 0)
	binary.Write(file, binary.LittleEndian, &sb)

	// 6️⃣ Inicializar Bitmaps
	initBitmap(file, sb.S_bm_inode_start, n)
	initBitmap(file, sb.S_bm_block_start, n*3)

	// 7️⃣ Crear raíz y users.txt
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

func createRootAndUsers(file *os.File, sb structures.SuperBlock) {

	now := int32(time.Now().Unix())

	// ---- Inodo raíz ----
	root := structures.Inode{
		I_uid:   0,
		I_gid:   0,
		I_s:     0,
		I_atime: now,
		I_ctime: now,
		I_mtime: now,
		I_type:  0,
		I_perm:  [3]byte{'7', '7', '7'},
	}

	for i := 0; i < 15; i++ {
		root.I_block[i] = -1
	}
	root.I_block[0] = 0

	// ---- Inodo users.txt ----
	users := root
	users.I_type = 1
	users.I_perm = [3]byte{'6', '6', '4'}
	users.I_block[0] = 1

	// ---- Escribir inodos ----
	file.Seek(int64(sb.S_inode_start), 0)
	binary.Write(file, binary.LittleEndian, &root)
	binary.Write(file, binary.LittleEndian, &users)
}
