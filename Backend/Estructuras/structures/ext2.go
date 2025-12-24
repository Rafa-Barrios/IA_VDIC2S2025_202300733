package structures

/* =========================
   SUPER BLOQUE (EXT2)
========================= */

type SuperBlock struct {
	S_filesystem_type   int32 // 2 = EXT2
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             int32
	S_umtime            int32
	S_mnt_count         int32
	S_magic             int32 // 0xEF53
	S_inode_s           int32
	S_block_s           int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

/* =========================
   INODO
========================= */

type Inode struct {
	I_uid   int32
	I_gid   int32
	I_s     int32
	I_atime int32
	I_ctime int32
	I_mtime int32
	I_block [15]int32
	I_type  byte    // 0 = carpeta | 1 = archivo
	I_perm  [3]byte // permisos UGO
}
