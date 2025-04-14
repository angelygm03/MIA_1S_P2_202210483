package DiskControl

import (
	"Proyecto1/backend/DiskStruct"
	"Proyecto1/backend/FileManagement"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Structure for mounted partitions
type MountedPartition struct {
	Path     string
	Name     string
	ID       string
	Status   byte // 0: unmounted, 1: mounted
	LoggedIn bool //true: logged in, false: not logged in
}

// Map to store the mounted partitions by disk
var mountedPartitions = make(map[string][]MountedPartition)

// Second validation of the command
func Mkdisk(size int, fit string, unit string, path string) {
	fmt.Println("======INICIO MKDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Fit:", fit)
	fmt.Println("Unit:", unit)
	fmt.Println("Path:", path)

	// Validate fit bf/ff/wf
	if fit != "bf" && fit != "wf" && fit != "ff" {
		fmt.Println("Error: Fit debe ser bf, wf or ff")
		return
	}

	// If fit is empty
	if fit == "" {
		fit = "ff"
	}

	// Validate size > 0
	if size <= 0 {
		fmt.Println("Error: Size debe ser mayor a 0")
		return
	}

	// Validate k - m
	if unit != "k" && unit != "m" {
		fmt.Println("Error: Las unidades válidas son k o m")
		return
	}

	// If unit is empty
	if unit == "" {
		unit = "m"
	}

	// Create file
	err := FileManagement.CreateFile(path)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	// Convert size to bytes
	if unit == "k" {
		size = size * 1024 // 1 KB = 1024 bytes
	} else {
		size = size * 1024 * 1024 // 1 MB = 1024 * 1024 bytes
	}

	// Open bin file
	file, err := FileManagement.OpenFile(path)
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
		return
	}
	defer file.Close()

	// === Write MBR ===
	// Create a buffer of 1 MB filled with zeros
	zeroBlock := make([]byte, 1024*1024) // 1 MB block
	remainingSize := size

	// Write the zero blocks to the file
	for remainingSize > 0 {
		if remainingSize >= len(zeroBlock) {
			_, err = file.Write(zeroBlock)
			if err != nil {
				fmt.Println("Error al escribir en el archivo:", err)
				return
			}
			remainingSize -= len(zeroBlock)
		} else {
			// Write the remaining bytes
			_, err = file.Write(zeroBlock[:remainingSize])
			if err != nil {
				fmt.Println("Error al escribir en el archivo:", err)
				return
			}
			remainingSize = 0
		}
	}

	// Create MBR
	var newMRB DiskStruct.MRB       // Create a new MBR
	newMRB.MbrSize = int32(size)    // Set the size
	newMRB.Signature = rand.Int31() // Set the signature to a random number
	copy(newMRB.Fit[:], fit)        // Set the fit

	// Date format yyyy-mm-dd
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")
	copy(newMRB.CreationDate[:], formattedDate)

	// Write the MBR
	if err := FileManagement.WriteObject(file, newMRB, 0); err != nil {
		fmt.Println("Error al escribir el MBR:", err)
		return
	}

	// === Read MBR ===
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error al leer el MBR:", err)
		return
	}

	// Print object
	DiskStruct.PrintMBR(TempMBR)

	fmt.Println("======FIN MKDISK======")
}

// Fuction to remove a disk
func Rmdisk(path string) {
	fmt.Println("======INICIO RMDISK======")
	fmt.Println("Path:", path)

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Error: El archivo no existe en la ruta especificada")
		return
	}

	// Confirm deletion
	fmt.Println("¿Está seguro de que desea eliminar el archivo? (yes/no):")
	var confirmation string
	fmt.Scanln(&confirmation)

	if strings.ToLower(confirmation) == "yes" {
		// Remove the file
		err := os.Remove(path)
		if err != nil {
			fmt.Println("Error al eliminar el archivo:", err)
			return
		}
		fmt.Println("Archivo eliminado exitosamente")
	} else {
		fmt.Println("Operación cancelada")
	}

	fmt.Println("======FIN RMDISK======")
}

func Fdisk(size int, path string, name string, unit string, type_ string, fit string) string {
	fmt.Println("======Start FDISK======")
	fmt.Println("Size:", size)
	fmt.Println("Path:", path)
	fmt.Println("Name:", name)
	fmt.Println("Unit:", unit)
	fmt.Println("Type:", type_)
	fmt.Println("Fit:", fit)

	// Fit bf, ff, wf
	if fit != "bf" && fit != "ff" && fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return "Fit debe ser 'bf', 'ff', o 'wf'"
	}
	// If fit is empty, set it to "w"
	if fit == "" {
		fit = "wf"
	}

	// Size must be greater than 0
	if size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return "Size debe ser mayor a cero"
	}

	// Validate unit b, k or m
	if unit != "b" && unit != "k" && unit != "m" {
		fmt.Println("Error: Unit must be 'b', 'k', or 'm'")
		return "Unit debe ser 'b', 'k', o 'm'"
	}

	// If unit is empty, set it to "k"
	if unit == "" {
		unit = "k"
	}

	// Validate type p, e or l
	if type_ != "p" && type_ != "e" && type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return "Type debe ser 'p', 'e', o 'l'"
	}

	// If type is empty, set it to "p"
	if type_ == "" {
		type_ = "p"
	}

	// Size to bytes
	if unit == "k" {
		size = size * 1024
	} else if unit == "m" {
		size = size * 1024 * 1024
	}

	// Open file in correct path
	file, err := FileManagement.OpenFile(path)
	if err != nil {
		fmt.Println("Error: Could not open file at path:", path)
		return fmt.Sprintf("No se encontró el archivo: %s", path)
	}

	var TempMBR DiskStruct.MRB
	// Read the object from the binary file
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file")
		return "No se pudo leer el MBR del archivo"
	}

	// Print the object
	DiskStruct.PrintMBR(TempMBR)

	fmt.Println("-------------")

	// Partitions validation
	var primaryCount, extendedCount, totalPartitions int
	var usedSpace int32 = 0

	// Count the number of partitions (4 are allowed)
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			totalPartitions++
			usedSpace += TempMBR.Partitions[i].Size

			// Count the number of primary and extended partitions
			if TempMBR.Partitions[i].Type[0] == 'p' {
				primaryCount++
			} else if TempMBR.Partitions[i].Type[0] == 'e' {
				extendedCount++
			}
		}
	}

	// Validate that there are not more than 4 partitions
	if totalPartitions >= 4 {
		fmt.Println("Error: No se pueden crear más de 4 particiones primarias o extendidas en total.")
		return "No se pueden crear más de 4 particiones primarias o extendidas en total."
	}

	// Validate that exits an extended partition
	if type_ == "e" && extendedCount > 0 {
		fmt.Println("Error: Solo se permite una partición extendida por disco.")
		return "Solo se permite una partición extendida por disco."
	}

	// If theres no extended partition, a logical partition can't be created
	if type_ == "l" && extendedCount == 0 {
		fmt.Println("Error: No se puede crear una partición lógica sin una partición extendida.")
		return "No se puede crear una partición lógica sin una partición extendida."
	}

	// Partition size can't be greater than the disk size
	if usedSpace+int32(size) > TempMBR.MbrSize {
		fmt.Println("Error: No hay suficiente espacio en el disco para crear esta partición.")
		return "No hay suficiente espacio en el disco para crear esta partición."
	}

	// Starting position of the new partition
	var gap int32 = int32(binary.Size(TempMBR))
	if totalPartitions > 0 {
		gap = TempMBR.Partitions[totalPartitions-1].Start + TempMBR.Partitions[totalPartitions-1].Size
	}

	// Encontrar una posición vacía para la nueva partición
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size == 0 {
			if type_ == "p" || type_ == "e" {
				// Creating a primary or extended partition
				TempMBR.Partitions[i].Size = int32(size)
				TempMBR.Partitions[i].Start = gap
				copy(TempMBR.Partitions[i].Name[:], name)
				copy(TempMBR.Partitions[i].Fit[:], fit)
				copy(TempMBR.Partitions[i].Status[:], "0")
				copy(TempMBR.Partitions[i].Type[:], type_)
				TempMBR.Partitions[i].Correlative = int32(totalPartitions + 1)

				// if the partition is extended, initialize the EBR
				if type_ == "e" {
					ebrStart := gap // First EBR starts at the beginning of the extended partition
					ebr := DiskStruct.EBR{
						PartFit:   fit[0],
						PartStart: ebrStart,
						PartSize:  0,
						PartNext:  -1,
					}
					copy(ebr.PartName[:], "")
					FileManagement.WriteObject(file, ebr, int64(ebrStart)) // Write the EBR
				}
				break
			}
		}
	}

	// If the partition is logical
	if type_ == "l" {
		for i := 0; i < 4; i++ {
			// Find the extended partition
			if TempMBR.Partitions[i].Type[0] == 'e' {
				ebrPos := TempMBR.Partitions[i].Start
				var ebr DiskStruct.EBR
				//Find the EBR
				for {
					FileManagement.ReadObject(file, &ebr, int64(ebrPos))
					if ebr.PartNext == -1 {
						break
					}
					ebrPos = ebr.PartNext
				}

				// Starting position of the logical partition is calculated
				newEBRPos := ebr.PartStart + ebr.PartSize                    // The new EBR starts right after the last logical partition
				logicalPartitionStart := newEBRPos + int32(binary.Size(ebr)) // The logical partition starts right after the EBR

				// Adjust tbe next EBR
				ebr.PartNext = newEBRPos
				FileManagement.WriteObject(file, ebr, int64(ebrPos))

				// Create and write new EBR
				newEBR := DiskStruct.EBR{
					PartFit:   fit[0],
					PartStart: logicalPartitionStart,
					PartSize:  int32(size),
					PartNext:  -1,
				}
				copy(newEBR.PartName[:], name)
				FileManagement.WriteObject(file, newEBR, int64(newEBRPos))

				fmt.Println("Nuevo EBR creado:")
				DiskStruct.PrintEBR(newEBR)
				break
			}
		}
	}

	// Overwrite the MBR
	if err := FileManagement.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: Could not write MBR to file")
		return "No se pudo escribir el MBR en el archivo"
	}

	var TempMBR2 DiskStruct.MRB
	// Verify the MBR was written correctly
	if err := FileManagement.ReadObject(file, &TempMBR2, 0); err != nil {
		fmt.Println("Error: Could not read MBR from file after writing")
		return "No se pudo leer el MBR del archivo después de escribir"
	}

	DiskStruct.PrintMBR(TempMBR2)

	for i := 0; i < 4; i++ {
		if TempMBR2.Partitions[i].Type[0] == 'e' {
			fmt.Println("Leyendo EBRs dentro de la partición extendida...")
			ebrPos := TempMBR2.Partitions[i].Start
			var ebr DiskStruct.EBR
			for {
				err := FileManagement.ReadObject(file, &ebr, int64(ebrPos))
				if err != nil {
					fmt.Println("Error al leer un EBR:", err)
					break
				}
				fmt.Println("EBR encontrado en la posición:", ebrPos)
				DiskStruct.PrintEBR(ebr)
				if ebr.PartNext == -1 {
					break
				}
				ebrPos = ebr.PartNext
			}
		}
	}

	// Close file to avoid memory leaks
	defer file.Close()

	fmt.Println("======FIN FDISK======")
	return fmt.Sprintf("Partition created successfully at %s.", path)
}

func PrintMountedPartitions() {
	fmt.Println("Particiones montadas:")

	if len(mountedPartitions) == 0 {
		fmt.Println("No hay particiones montadas.")
		return
	}

	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco ID: %s\n", diskID)
		for _, partition := range partitions {
			loginStatus := "No"
			if partition.LoggedIn {
				loginStatus = "Sí"
			}
			fmt.Printf(" - Partición Name: %s, ID: %s, Path: %s, Status: %c, LoggedIn: %s\n",
				partition.Name, partition.ID, partition.Path, partition.Status, loginStatus)
		}
	}
	fmt.Println("")
}

func Mount(path string, name string) string {
	fmt.Println("======Start MOUNT======")
	file, err := FileManagement.OpenFile(path)
	if err != nil {
		fmt.Println("No se pudo abrir el archivo en la ruta:", path)
		fmt.Println("======FIN MOUNT======")
		return fmt.Sprintf("No se pudo abrir el archivo en la ruta: %s", path)
	}
	defer file.Close()

	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR desde el archivo")
		fmt.Println("======FIN MOUNT======")
		return "No se pudo leer el MBR desde el archivo"
	}

	fmt.Printf("Buscando partición con nombre: '%s'\n", name)

	partitionFound := false // Indicates if the partition was found
	var partition DiskStruct.Partition
	var partitionIndex int

	// Converting the name to a byte array
	nameBytes := [16]byte{} // 16 bytes for the name
	copy(nameBytes[:], []byte(name))

	// Find the partition with the given name
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Type[0] == 'p' && bytes.Equal(TempMBR.Partitions[i].Name[:], nameBytes[:]) {
			partition = TempMBR.Partitions[i]
			partitionIndex = i
			partitionFound = true
			break
		}
	}

	// Just primary partitions can be mounted
	if !partitionFound {
		fmt.Println("Error: Partición no encontrada o no es una partición primaria")
		fmt.Println("======FIN MOUNT======")
		return "Partición no encontrada o no es una partición primaria"
	}

	// Verify if the partition is already mounted (status = 1)
	if partition.Status[0] == '1' {
		fmt.Println("Error: La partición ya está montada")
		fmt.Println("======FIN MOUNT======")
		return "La partición ya está montada, no se puede montar nuevamente."
	}

	// Generate a unique ID for the partition
	diskID := generateDiskID(path)

	// Verify if the partition is already mounted in the same disk
	mountedPartitionsInDisk := mountedPartitions[diskID]
	var letter byte

	if len(mountedPartitionsInDisk) == 0 {
		// If it's the first partition to be mounted in the disk, use the letter 'a'
		if len(mountedPartitions) == 0 {
			letter = 'a'
		} else {
			lastDiskID := getLastDiskID()
			lastLetter := mountedPartitions[lastDiskID][0].ID[len(mountedPartitions[lastDiskID][0].ID)-1]
			letter = lastLetter + 1
		}
	} else {
		// Using the same letter of the last partition mounted in the disk
		letter = mountedPartitionsInDisk[0].ID[len(mountedPartitionsInDisk[0].ID)-1]
	}

	// Increment the correlative number of the partition based on the number of mounted partitions
	carnet := "202210483"
	lastTwoDigits := carnet[len(carnet)-2:]             // Last two digits of my carnet
	partitionNumber := len(mountedPartitionsInDisk) + 1 // Use the count of mounted partitions in the disk
	partitionID := fmt.Sprintf("%s%d%c", lastTwoDigits, partitionNumber, letter)

	// Update the partition status to '1' (mounted)
	partition.Status[0] = '1'
	copy(partition.Id[:], partitionID)
	TempMBR.Partitions[partitionIndex] = partition
	mountedPartitions[diskID] = append(mountedPartitions[diskID], MountedPartition{
		Path:   path,
		Name:   name,
		ID:     partitionID,
		Status: '1',
	})

	// Writing the updated MBR to the file
	if err := FileManagement.WriteObject(file, TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo sobrescribir el MBR en el archivo")
		fmt.Println("======FIN MOUNT======")
		return "No se pudo sobrescribir el MBR en el archivo"
	}

	fmt.Printf("Partición montada con ID: %s\n", partitionID)

	fmt.Println("")
	fmt.Println("MBR actualizado:")
	DiskStruct.PrintMBR(TempMBR)
	fmt.Println("")
	PrintMountedPartitions()
	fmt.Println("======FIN MOUNT======")
	return fmt.Sprintf("Partition mounted successfully with ID: %s", partitionID)
}

// Get the last disk ID
func getLastDiskID() string {
	var lastDiskID string
	for diskID := range mountedPartitions {
		lastDiskID = diskID
	}
	return lastDiskID
}

// Generate a unique ID for the disk
func generateDiskID(path string) string {
	return strings.ToLower(path)
}

// Function to get the mounted partitions
func GetMountedPartitions() map[string][]MountedPartition {
	fmt.Println("Devolviendo particiones montadas...")
	for diskID, partitions := range mountedPartitions {
		fmt.Printf("Disco: %s\n", diskID)
		for _, partition := range partitions {
			fmt.Printf(" - ID: %s, Nombre: %s, Estado: %d\n", partition.ID, partition.Name, partition.Status)
		}
	}
	return mountedPartitions
}

// Mark a partition as logged in
func MarkPartitionAsLoggedIn(id string) {
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = true
				fmt.Printf("Partición con ID %s marcada como logueada.\n", id)
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como logueada.\n", id)
}

// Mark a partition as logged out
func MarkPartitionAsLoggedOut(id string) {
	for diskID, partitions := range mountedPartitions {
		for i, partition := range partitions {
			if partition.ID == id {
				mountedPartitions[diskID][i].LoggedIn = false
				fmt.Printf("Partición con ID %s marcada con sesión cerrada.\n", id)
				return
			}
		}
	}
	fmt.Printf("No se encontró la partición con ID %s para marcarla como deslogueada.\n", id)
}
