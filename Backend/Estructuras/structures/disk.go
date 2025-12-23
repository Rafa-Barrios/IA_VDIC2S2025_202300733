package structures

type EBR struct {
	Part_mount int8
	Part_fit   byte
	Part_start int32
	Part_s     int32
	Part_next  int32
	Name       [16]byte
}

// PARTICIONES PRIMARIAS
type Partition struct {
	Part_status      int8
	Part_type        byte
	Part_fit         byte
	Part_start       int32
	Part_s           int32
	Part_name        [16]byte
	Part_correlative int32
	Part_id          [4]byte
}

// MBR
type MBR struct {
	Mbr_tamano         int32
	Mbr_fecha_creacion int32
	Mbr_disk_signature int32
	Dsk_fit            byte
	Mbr_partitions     [4]Partition
}

// ESTRUCTURAS AUXILIARES PARA RESPUESTA / FRONTEND
type Particion_Enviar struct {
	Particion  string
	Type       string
	Status     int8
	Id_mounted string
}

type MBR_Obtener struct {
	Disco          string
	Disco_Path     string
	Mbr_partitions [4]Particion_Enviar
}
