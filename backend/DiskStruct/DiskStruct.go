package DiskStruct

import (
	"fmt"
)

type MRB struct {
	MbrSize      int32    // 4 bytes
	CreationDate [10]byte // YYYY-MM-DD
	Signature    int32    // 4 bytes
	Fit          [1]byte  // 1 byte: B, F, W
	Partitions   [4]Partition
}

func PrintMBR(data MRB) {
	fmt.Println(fmt.Sprintf("CreationDate: %s, fit: %s, size: %d", string(data.CreationDate[:]), string(data.Fit[:]), data.MbrSize))
	for i := 0; i < 4; i++ {
		PrintPartition(data.Partitions[i])
	}
}

type Partition struct {
	Status      [1]byte // Mounted, Unmounted
	Type        [1]byte // P, E
	Fit         [1]byte // B, F, W
	Start       int32   // Where the partition starts
	Size        int32
	Name        [16]byte
	Correlative int32
	Id          [4]byte
}

func PrintPartition(data Partition) {
	fmt.Println(fmt.Sprintf("Name: %s, type: %s, start: %d, size: %d, status: %s, id: %s", string(data.Name[:]), string(data.Type[:]), data.Start, data.Size, string(data.Status[:]), string(data.Id[:])))
}

type EBR struct {
	PartMount byte //Mounted, Unmounted
	PartFit   byte //B, F, W
	PartStart int32
	PartSize  int32
	PartNext  int32 //EBR next byte, -1 if it's the last one
	PartName  [16]byte
}

func PrintEBR(data EBR) {
	fmt.Println(fmt.Sprintf("Name: %s, fit: %c, start: %d, size: %d, next: %d, mount: %c",
		string(data.PartName[:]),
		data.PartFit,
		data.PartStart,
		data.PartSize,
		data.PartNext,
		data.PartMount))
}

// ==== STRUCTURES FOR EXT2 FILE SYSTEM ====

type Superblock struct {
	S_filesystem_type   int32    // Number that identifies the file system
	S_inodes_count      int32    // Total number of inodes
	S_blocks_count      int32    // Total number of blocks
	S_free_blocks_count int32    // How many blocks are free
	S_free_inodes_count int32    // How many inodes are free
	S_mtime             [19]byte // Last mount time
	S_umtime            [19]byte // Last unmount time
	S_mnt_count         int32    // How many times the disk has been mounted
	S_magic             int32    // Id of the file system: 0xEF53
	S_inode_size        int32    // Inode size
	S_block_size        int32    // Block size
	S_fist_ino          int32    // First free inode
	S_first_blo         int32    // Fist free block
	S_bm_inode_start    int32    // Starting point of the bitmap of inodes
	S_bm_block_start    int32    // Starting point of the bitmap of blocks
	S_inode_start       int32    // Starting point of the table of inodes
	S_block_start       int32    // Starting point of the table of blocks
}

type Inode struct {
	I_uid   int32     // UID of the user
	I_gid   int32     // GID of the group
	I_size  int32     // Size of the file
	I_atime [19]byte  // Last access time
	I_ctime [19]byte  // Creation time
	I_mtime [19]byte  // Last modification time
	I_block [15]int32 // Pointers to the blocks
	I_type  [1]byte   // File type: 0 for folder, 1 for file
	I_perm  [3]byte   // Permissions
}

type Folderblock struct {
	B_content [4]Content // Array of content
}

type Content struct {
	B_name  [12]byte // Name of the file or folder
	B_inodo int32    // Inodo pointer to the file or folder
}

type Fileblock struct {
	B_content [64]byte // Array of content
}

type Pointerblock struct {
	B_pointers [16]int32 // Array of pointers
}

type User struct {
	UID   int32  // User ID
	GID   int32  // Group ID
	Name  string // Username
	Group string // Group name
}

// ==== PRINTS ====

func PrintSuperblock(sb Superblock) {
	fmt.Println("====== Superblock ======")
	fmt.Printf("S_filesystem_type: %d\n", sb.S_filesystem_type)
	fmt.Printf("S_inodes_count: %d\n", sb.S_inodes_count)
	fmt.Printf("S_blocks_count: %d\n", sb.S_blocks_count)
	fmt.Printf("S_free_blocks_count: %d\n", sb.S_free_blocks_count)
	fmt.Printf("S_free_inodes_count: %d\n", sb.S_free_inodes_count)
	fmt.Printf("S_mtime: %s\n", string(sb.S_mtime[:]))
	fmt.Printf("S_umtime: %s\n", string(sb.S_umtime[:]))
	fmt.Printf("S_mnt_count: %d\n", sb.S_mnt_count)
	fmt.Printf("S_magic: 0x%X\n", sb.S_magic) // 0x%X for hexadecimal formart
	fmt.Printf("S_inode_size: %d\n", sb.S_inode_size)
	fmt.Printf("S_block_size: %d\n", sb.S_block_size)
	fmt.Printf("S_fist_ino: %d\n", sb.S_fist_ino)
	fmt.Printf("S_first_blo: %d\n", sb.S_first_blo)
	fmt.Printf("S_bm_inode_start: %d\n", sb.S_bm_inode_start)
	fmt.Printf("S_bm_block_start: %d\n", sb.S_bm_block_start)
	fmt.Printf("S_inode_start: %d\n", sb.S_inode_start)
	fmt.Printf("S_block_start: %d\n", sb.S_block_start)
	fmt.Println("========================")
}

func PrintInode(inode Inode) {
	fmt.Println("====== Inode ======")
	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", string(inode.I_atime[:]))
	fmt.Printf("I_ctime: %s\n", string(inode.I_ctime[:]))
	fmt.Printf("I_mtime: %s\n", string(inode.I_mtime[:]))
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Println("===================")
}

func PrintFolderblock(folderblock Folderblock) {
	fmt.Println("====== Folderblock ======")
	for i, content := range folderblock.B_content {
		fmt.Printf("Content %d: Name: %s, Inodo: %d\n", i, string(content.B_name[:]), content.B_inodo)
	}
	fmt.Println("=========================")
}

func PrintFileblock(fileblock Fileblock) {
	fmt.Println("====== Fileblock ======")
	fmt.Printf("B_content: %s\n", string(fileblock.B_content[:]))
	fmt.Println("=======================")
}

func PrintPointerblock(pointerblock Pointerblock) {
	fmt.Println("====== Pointerblock ======")
	for i, pointer := range pointerblock.B_pointers {
		fmt.Printf("Pointer %d: %d\n", i, pointer)
	}
	fmt.Println("=========================")
}
