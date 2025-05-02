package FileSystem

import (
	"Proyecto2/backend/DiskControl"
	"Proyecto2/backend/DiskStruct"
	"Proyecto2/backend/FileManagement"
	"Proyecto2/backend/UserManagement"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

func Mkfs(id string, type_ string, fs_ string) {
	fmt.Println("======INICIO MKFS======")
	fmt.Println("Id:", id)
	fmt.Println("Type:", type_)
	fmt.Println("Fs:", fs_)

	// Find the mounted partition with the ID
	var mountedPartition DiskControl.MountedPartition
	var partitionFound bool

	for _, partitions := range DiskControl.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Partición no encontrada")
		return
	}

	if mountedPartition.Status != '1' { // Check if the partition is mounted
		fmt.Println("La partición aún no está montada")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(mountedPartition.Path)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	var TempMBR DiskStruct.MRB
	// Read the MBR
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error al leer el MBR:", err)
		return
	}

	DiskStruct.PrintMBR(TempMBR)

	fmt.Println("-------------")

	var index int = -1
	// Find the partition in the MBR
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				index = i
				break
			}
		}
	}

	if index != -1 {
		DiskStruct.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Partición no encontrada (EBR)")
		return
	}

	// Calculate the number of inodes
	numerador := int32(TempMBR.Partitions[index].Size - int32(binary.Size(DiskStruct.Superblock{})))
	denominador_base := int32(4 + int32(binary.Size(DiskStruct.Inode{})) + 3*int32(binary.Size(DiskStruct.Fileblock{})))
	var temp int32 = 0

	// Add journaling size for EXT3
	if fs_ == "3fs" {
		temp = int32(binary.Size(DiskStruct.Journaling{}))
	} else if fs_ == "2fs" {
		temp = 0
	} else {
		fmt.Println("Error: Sólo están disponibles los sistemas de archivos 2FS y 3FS.")
		return
	}

	denominador := denominador_base + temp
	n := int32(numerador / denominador)

	fmt.Println("INODOS:", n)

	// Create the superblock
	var newSuperblock DiskStruct.Superblock
	if fs_ == "2fs" {
		newSuperblock.S_filesystem_type = 2 // EXT2
	} else if fs_ == "3fs" {
		newSuperblock.S_filesystem_type = 3 // EXT3
	}
	newSuperblock.S_inodes_count = n
	newSuperblock.S_blocks_count = 3 * n
	newSuperblock.S_free_blocks_count = 3*n - 2
	newSuperblock.S_free_inodes_count = n - 2

	// Date and time
	CurrentDate := time.Now()
	DateString := CurrentDate.Format("2006-01-02 15:04:05")
	copy(newSuperblock.S_mtime[:], DateString)
	copy(newSuperblock.S_umtime[:], DateString)

	newSuperblock.S_mnt_count = 1
	newSuperblock.S_magic = 0xEF53
	newSuperblock.S_inode_size = int32(binary.Size(DiskStruct.Inode{}))
	newSuperblock.S_block_size = int32(binary.Size(DiskStruct.Fileblock{}))

	// Calculate the start positions
	newSuperblock.S_bm_inode_start = TempMBR.Partitions[index].Start + int32(binary.Size(DiskStruct.Superblock{}))
	newSuperblock.S_bm_block_start = newSuperblock.S_bm_inode_start + n
	newSuperblock.S_inode_start = newSuperblock.S_bm_block_start + 3*n
	newSuperblock.S_block_start = newSuperblock.S_inode_start + n*newSuperblock.S_inode_size

	// Call the function to create the filesystem
	if fs_ == "2fs" {
		create_ext2(n, TempMBR.Partitions[index], newSuperblock, DateString, file)
	} else if fs_ == "3fs" {
		create_ext3(n, TempMBR.Partitions[index], newSuperblock, DateString, file)
	}

	fmt.Println("======FIN MKFS======")
}

func create_ext2(n int32, partition DiskStruct.Partition, newSuperblock DiskStruct.Superblock, date string, file *os.File) {
	fmt.Println("======Start CREATE EXT2======")
	fmt.Println("INODOS:", n)

	//Print the Superblock
	DiskStruct.PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	// Write the bitmaps and inodes to the file
	for i := int32(0); i < n; i++ {
		if err := FileManagement.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	for i := int32(0); i < 3*n; i++ {
		if err := FileManagement.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}

	// Initialize inodes and blocks
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Creates the root folder and the users.txt file
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Writes the superblock to the file
	if err := FileManagement.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Mark the inodes and blocks as used
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}

	fmt.Println("====== Imprimiendo Inodos ======")

	for i := int32(0); i < n; i++ {
		var inode DiskStruct.Inode
		offset := int64(newSuperblock.S_inode_start + i*int32(binary.Size(DiskStruct.Inode{})))
		if err := FileManagement.ReadObject(file, &inode, offset); err != nil {
			fmt.Println("Error al leer inodo: ", err)
			return
		}
		DiskStruct.PrintInode(inode)
	}

	fmt.Println("====== Imprimiendo Folderblocks y Fileblocks ======")

	// ----> Print Folderblocks <----
	for i := int32(0); i < 1; i++ {
		var folderblock DiskStruct.Folderblock
		offset := int64(newSuperblock.S_block_start + i*int32(binary.Size(DiskStruct.Folderblock{})))
		if err := FileManagement.ReadObject(file, &folderblock, offset); err != nil {
			fmt.Println("Error al leer Folderblock: ", err)
			return
		}
		DiskStruct.PrintFolderblock(folderblock)
	}

	// ----> Print Fileblocks <----
	for i := int32(0); i < 1; i++ {
		var fileblock DiskStruct.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(DiskStruct.Folderblock{})) + i*int32(binary.Size(DiskStruct.Fileblock{})))
		if err := FileManagement.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return
		}
		DiskStruct.PrintFileblock(fileblock)
	}

	// ----> Print Final Superblock <----
	DiskStruct.PrintSuperblock(newSuperblock)

	fmt.Println("======End CREATE EXT2======")
}

// ===== Auxiliar functions =====

// Aux func to initialize the inodes and blocks
func initInodesAndBlocks(n int32, newSuperblock DiskStruct.Superblock, file *os.File) error {
	var newInode DiskStruct.Inode
	for i := int32(0); i < 15; i++ {
		newInode.I_block[i] = -1
	}

	for i := int32(0); i < n; i++ {
		if err := FileManagement.WriteObject(file, newInode, int64(newSuperblock.S_inode_start+i*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
			return err
		}
	}

	var newFileblock DiskStruct.Fileblock
	for i := int32(0); i < 3*n; i++ {
		if err := FileManagement.WriteObject(file, newFileblock, int64(newSuperblock.S_block_start+i*int32(binary.Size(DiskStruct.Fileblock{})))); err != nil {
			return err
		}
	}

	return nil
}

// Aux func to create the root folder and the users.txt file
func createRootAndUsersFile(newSuperblock DiskStruct.Superblock, date string, file *os.File) error {
	var Inode0, Inode1 DiskStruct.Inode
	initInode(&Inode0, date)
	initInode(&Inode1, date)

	Inode0.I_block[0] = 0
	Inode1.I_block[0] = 1

	// Real size of the content
	data := "1,G,root\n1,U,root,root,123\n"
	actualSize := int32(len(data))
	Inode1.I_size = actualSize // Set the size of the file

	var Fileblock1 DiskStruct.Fileblock
	copy(Fileblock1.B_content[:], data) // Copy the data to the block

	var Folderblock0 DiskStruct.Folderblock
	Folderblock0.B_content[0].B_inodo = 0
	copy(Folderblock0.B_content[0].B_name[:], ".") // Copy the name of the folder

	Folderblock0.B_content[1].B_inodo = 0
	copy(Folderblock0.B_content[1].B_name[:], "..") // Copy the name of the parent folder

	Folderblock0.B_content[2].B_inodo = 1
	copy(Folderblock0.B_content[2].B_name[:], "users.txt")

	// Writing the objects to the file
	if err := FileManagement.WriteObject(file, Inode0, int64(newSuperblock.S_inode_start)); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, Inode1, int64(newSuperblock.S_inode_start+int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, Folderblock0, int64(newSuperblock.S_block_start)); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, Fileblock1, int64(newSuperblock.S_block_start+int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
		return err
	}

	return nil
}

// Aux funct to initialize the inodes
func initInode(inode *DiskStruct.Inode, date string) {
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = 0
	copy(inode.I_atime[:], date)
	copy(inode.I_ctime[:], date)
	copy(inode.I_mtime[:], date)
	copy(inode.I_perm[:], "664")

	for i := int32(0); i < 15; i++ {
		inode.I_block[i] = -1
	}
}

// Aux funct to mark the inodes and blocks as used
func markUsedInodesAndBlocks(newSuperblock DiskStruct.Superblock, file *os.File) error {
	if err := FileManagement.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start)); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, byte(1), int64(newSuperblock.S_bm_inode_start+1)); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start)); err != nil {
		return err
	}
	if err := FileManagement.WriteObject(file, byte(1), int64(newSuperblock.S_bm_block_start+1)); err != nil {
		return err
	}
	return nil
}

func create_ext3(n int32, partition DiskStruct.Partition, newSuperblock DiskStruct.Superblock, date string, file *os.File) {
	fmt.Println("======Start CREATE EXT3======")
	fmt.Println("INODOS:", n)

	DiskStruct.PrintSuperblock(newSuperblock)
	fmt.Println("Date:", date)

	// Init journaling
	if err := initJournaling(newSuperblock, file); err != nil {
		fmt.Println("Error al inicializar el Journaling: ", err)
		return
	}
	fmt.Println("Journaling inicializado correctamente.")

	// Write the bitmaps blocks and inodes to the file
	for i := int32(0); i < n; i++ {
		if err := FileManagement.WriteObject(file, byte(0), int64(newSuperblock.S_bm_inode_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}
	fmt.Println("Bitmap de inodos escrito correctamente.")

	for i := int32(0); i < 3*n; i++ {
		if err := FileManagement.WriteObject(file, byte(0), int64(newSuperblock.S_bm_block_start+i)); err != nil {
			fmt.Println("Error: ", err)
			return
		}
	}
	fmt.Println("Bitmap de bloques escrito correctamente.")

	// Init inodes and blocks
	if err := initInodesAndBlocks(n, newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Inodos y bloques inicializados correctamente.")

	// Create the root folder and the users.txt file
	if err := createRootAndUsersFile(newSuperblock, date, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Carpeta raíz y archivo users.txt creados correctamente.")

	// Write the superblock to the file
	if err := FileManagement.WriteObject(file, newSuperblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Superbloque escrito correctamente.")

	// Mark the inodes and blocks as used
	if err := markUsedInodesAndBlocks(newSuperblock, file); err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Inodos y bloques iniciales marcados como usados correctamente.")

	for i := int32(0); i < 1; i++ {
		var fileblock DiskStruct.Fileblock
		offset := int64(newSuperblock.S_block_start + int32(binary.Size(DiskStruct.Folderblock{})) + i*int32(binary.Size(DiskStruct.Fileblock{})))
		if err := FileManagement.ReadObject(file, &fileblock, offset); err != nil {
			fmt.Println("Error al leer Fileblock: ", err)
			return
		}
		DiskStruct.PrintFileblock(fileblock)
	}
	fmt.Println("Fileblocks impresos correctamente.")

	fmt.Println("======End CREATE EXT3======")
}

func initJournaling(newSuperblock DiskStruct.Superblock, file *os.File) error {
	var journaling DiskStruct.Journaling
	journaling.Size = 50
	journaling.Ultimo = 0

	// Position to write the journaling
	journalingStart := newSuperblock.S_inode_start - int32(binary.Size(DiskStruct.Journaling{}))*journaling.Size

	// Write the journaling to the file
	for i := 0; i < 50; i++ {
		if err := FileManagement.WriteObject(file, journaling, int64(journalingStart+int32(i*binary.Size(journaling)))); err != nil {
			return fmt.Errorf("error al inicializar el journaling: %v", err)
		}
	}

	fmt.Println("Journaling inicializado correctamente.")
	return nil
}

func Recovery(id string) string {
	fmt.Println("======Start RECOVERY======")
	fmt.Println("Id:", id)

	// Find the mounted partition with the ID
	var mountedPartition DiskControl.MountedPartition
	var partitionFound bool

	for _, partitions := range DiskControl.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No se encontró la partición con el ID proporcionado.")
		return "Error: No se encontró la partición con el ID proporcionado."
	}

	// Open the file
	file, err := FileManagement.OpenFile(mountedPartition.Path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return "Error: No se pudo abrir el archivo."
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return "Error: No se pudo leer el MBR."
	}

	// Find the partition in the MBR
	var partition DiskStruct.Partition
	var partitionIndex int = -1
	for i := 0; i < 4; i++ {
		if string(TempMBR.Partitions[i].Id[:]) == id {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			break
		}
	}

	if partitionIndex == -1 {
		fmt.Println("Error: No se encontró la partición en el MBR.")
		return "Error: No se encontró la partición en el MBR."
	}

	fmt.Printf("Partición encontrada en el índice: %d\n", partitionIndex)

	if partition.Size == 0 {
		fmt.Println("Error: No se encontró la partición en el MBR.")
		return "Error: No se encontró la partición en el MBR."
	}

	// Read the superblock
	var superblock DiskStruct.Superblock
	if err := FileManagement.ReadObject(file, &superblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: No se pudo leer el superbloque:", err)
		return "Error: No se pudo leer el superbloque."
	}

	// Check if the partition is EXT3
	if superblock.S_filesystem_type != 3 {
		fmt.Println("Error: La partición no utiliza el sistema de archivos EXT3.")
		return "Error: La partición no utiliza el sistema de archivos EXT3."
	}

	// Read the journaling
	var journalingStart = superblock.S_inode_start - int32(binary.Size(DiskStruct.Journaling{}))*50
	for i := 0; i < 50; i++ {
		var journalEntry DiskStruct.Content_J
		offset := int64(journalingStart + int32(i*binary.Size(journalEntry)))
		if err := FileManagement.ReadObject(file, &journalEntry, offset); err != nil {
			fmt.Println("Error al leer el journaling:", err)
			break
		}

		// If the operation is empty, break the loop
		if string(journalEntry.Operation[:]) == "" {
			break
		}

		// Process the journal entry
		fmt.Printf("Recuperando operación: %s, Ruta: %s, Contenido: %s\n",
			string(journalEntry.Operation[:]),
			string(journalEntry.Path[:]),
			string(journalEntry.Content[:]))

		if string(journalEntry.Operation[:]) == "mkfile" {
			UserManagement.Mkfile(string(journalEntry.Path[:]), false, len(string(journalEntry.Content[:])), "")
		} else if string(journalEntry.Operation[:]) == "mkdir" {
			UserManagement.Mkdir(string(journalEntry.Path[:]), false)
		}
	}

	fmt.Println("Recuperación completada exitosamente.")
	return "Recuperación completada exitosamente."
}

func Loss(id string) string {
	fmt.Println("====== Start LOSS ======")
	fmt.Println("Id:", id)

	// Find the mounted partition with the ID
	var mountedPartition DiskControl.MountedPartition
	var partitionFound bool

	for _, partitions := range DiskControl.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mountedPartition = partition
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No se encontró la partición con el ID proporcionado.")
		return "Error: No se encontró la partición con el ID proporcionado."
	}

	// Open bin file
	file, err := FileManagement.OpenFile(mountedPartition.Path)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return "Error: No se pudo abrir el archivo."
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return "Error: No se pudo leer el MBR."
	}

	// Find the partition in the MBR
	var partition DiskStruct.Partition
	var partitionIndex int = -1
	for i := 0; i < 4; i++ {
		if string(TempMBR.Partitions[i].Id[:]) == id {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			break
		}
	}

	if partitionIndex == -1 {
		fmt.Println("Error: No se encontró la partición en el MBR.")
		return "Error: No se encontró la partición en el MBR."
	}

	fmt.Printf("Partición encontrada en el índice: %d\n", partitionIndex)

	// Read the superblock
	var superblock DiskStruct.Superblock
	if err := FileManagement.ReadObject(file, &superblock, int64(partition.Start)); err != nil {
		fmt.Println("Error: No se pudo leer el superbloque:", err)
		return "Error: No se pudo leer el superbloque."
	}

	// Clean the data blocks to simulate file system loss
	fmt.Println("Limpiando bloques de datos para simular pérdida del sistema de archivos...")
	FileManagement.FillWithZeros(file, superblock.S_bm_inode_start, superblock.S_bm_block_start-superblock.S_bm_inode_start) // Bitmap of inodos
	FileManagement.FillWithZeros(file, superblock.S_bm_block_start, superblock.S_inode_start-superblock.S_bm_block_start)    // Bitmap of blocks
	FileManagement.FillWithZeros(file, superblock.S_inode_start, superblock.S_block_start-superblock.S_inode_start)          // Inodes
	FileManagement.FillWithZeros(file, superblock.S_block_start, partition.Size-(superblock.S_block_start-partition.Start))  // Blocks

	fmt.Println("Simulación de pérdida completada.")
	return "Simulación de pérdida completada."
}
