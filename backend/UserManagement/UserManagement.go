package UserManagement

import (
	"Proyecto1/backend/DiskControl"
	"Proyecto1/backend/DiskStruct"
	"Proyecto1/backend/FileManagement"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Login(user string, password string, id string) {
	fmt.Println("======Start LOGIN======")
	fmt.Println("User:", user)
	fmt.Println("Password:", password)
	fmt.Println("Id:", id)

	// Verify if the user is already logged in a partition
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool
	var login bool = false // Nobody is logged in

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.ID == id && partition.LoggedIn { //Find the user in the mounted partitions
				fmt.Println("Ya existe un usuario logueado!")
				return
			}
			if partition.ID == id { // Find the partition with the given id
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No se encontró ninguna partición montada con el ID proporcionado")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return
	}
	defer file.Close()

	var TempMBR DiskStruct.MRB
	// Read the MBR from the binary file
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	DiskStruct.PrintMBR(TempMBR)
	fmt.Println("-------------")

	var index int = -1
	// Find the correct partition in the MBR
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				fmt.Println("Partition found")
				if TempMBR.Partitions[i].Status[0] == '1' {
					fmt.Println("Partition is mounted")
					index = i
				} else {
					fmt.Println("Partition is not mounted")
					return
				}
				break
			}
		}
	}

	if index != -1 {
		DiskStruct.PrintPartition(TempMBR.Partitions[index])
	} else {
		fmt.Println("Partition not found")
		return
	}

	var tempSuperblock DiskStruct.Superblock
	// Read the Superblock from the binary file
	if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error: No se pudo leer el Superblock:", err)
		return
	}

	// Find users.txt and returns the index of the Inode
	indexInode := InitSearch("/users.txt", file, tempSuperblock)

	var crrInode DiskStruct.Inode
	// Read the Inode from the binary file
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo:", err)
		return
	}

	// Read the data from the file
	data := GetInodeFileData(crrInode, file, tempSuperblock)

	// Split the data by lines
	lines := strings.Split(data, "\n")

	// Iterate over the lines to find the user and password
	for _, line := range lines {
		// Ignore empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			fmt.Println("Línea ignorada (vacía)")
			continue
		}

		fmt.Printf("Procesando línea: %s\n", line)

		words := strings.Split(line, ",")

		fmt.Printf("Campos parseados: %v\n", words)

		// U = User, G = Group
		if len(words) == 5 && strings.TrimSpace(words[1]) == "U" {
			userFromFile := strings.TrimSpace(words[2])     // index 2 for the user
			passwordFromFile := strings.TrimSpace(words[4]) // index 4 for the password

			fmt.Printf("Comparando usuario: '%s' con '%s', contraseña: '%s' con '%s'\n", userFromFile, user, passwordFromFile, password)

			if userFromFile == user && passwordFromFile == password {
				login = true
				break
			}
		} else {
			fmt.Printf("Línea ignorada (no válida): %s\n", line)
		}
	}

	fmt.Println("Inode", crrInode.I_block)

	// If the login was successful, mark the partition as logged in
	if login {
		fmt.Println("Usuario logueado con exito")
		DiskControl.MarkPartitionAsLoggedIn(id)
	}

	fmt.Println("======End LOGIN======")
}

// Returned value is the index of the Inode
func InitSearch(path string, file *os.File, tempSuperblock DiskStruct.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INICIAL ======")
	fmt.Println("path:", path)

	//Search and split the path (we need users.txt)
	TempStepsPath := strings.Split(path, "/")
	StepsPath := TempStepsPath[1:]

	fmt.Println("StepsPath:", StepsPath, "len(StepsPath):", len(StepsPath))
	for _, step := range StepsPath {
		fmt.Println("step:", step)
	}

	var Inode0 DiskStruct.Inode
	// Read object from bin file
	if err := FileManagement.ReadObject(file, &Inode0, int64(tempSuperblock.S_inode_start)); err != nil {
		return -1
	}

	fmt.Println("======End BUSQUEDA INICIAL======")

	return SearchInodeByPath(StepsPath, Inode0, file, tempSuperblock)
}

// stack (pila) to store logged in users
func pop(s *[]string) string {
	lastIndex := len(*s) - 1
	last := (*s)[lastIndex]
	*s = (*s)[:lastIndex]
	return last
}

// Search Inode by path
func SearchInodeByPath(StepsPath []string, Inode DiskStruct.Inode, file *os.File, tempSuperblock DiskStruct.Superblock) int32 {
	fmt.Println("======Start BUSQUEDA INODO POR PATH======")
	index := int32(0)
	SearchedName := strings.Replace(pop(&StepsPath), " ", "", -1)

	fmt.Println("========== SearchedName:", SearchedName)

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			if index < 13 {

				//==== DIRECT CASE ====
				var crrFolderBlock DiskStruct.Folderblock
				// Read object from bin file
				if err := FileManagement.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
					return -1
				}

				for _, folder := range crrFolderBlock.B_content {
					fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

					if strings.Contains(string(folder.B_name[:]), SearchedName) {

						fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
						if len(StepsPath) == 0 {
							fmt.Println("Folder found======")
							return folder.B_inodo
						} else {
							fmt.Println("NextInode======")
							var NextInode DiskStruct.Inode
							// Read object from bin file
							if err := FileManagement.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
								return -1
							}
							return SearchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
						}
					}
				}

			} else if index == 13 {
				// ==== INDIRECT CASE ====
				fmt.Println("Indirect case: Simple Indirect Block")

				// Read the Pointerblock
				var pointerBlock DiskStruct.Pointerblock
				if err := FileManagement.ReadObject(file, &pointerBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
					fmt.Println("Error reading Pointerblock:", err)
					return -1
				}

				// Iterate over the pointers in the Pointerblock
				for _, pointer := range pointerBlock.B_pointers {
					if pointer != -1 {
						var crrFolderBlock DiskStruct.Folderblock
						// Read the Folderblock pointed by the current pointer
						if err := FileManagement.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+pointer*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
							fmt.Println("Error reading Folderblock:", err)
							return -1
						}

						// Iterate over the contents of the Folderblock
						for _, folder := range crrFolderBlock.B_content {
							fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

							if strings.Contains(string(folder.B_name[:]), SearchedName) {
								fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
								if len(StepsPath) == 0 {
									fmt.Println("Folder found======")
									return folder.B_inodo
								} else {
									fmt.Println("NextInode======")
									var NextInode DiskStruct.Inode
									// Read the next Inode
									if err := FileManagement.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
										fmt.Println("Error reading NextInode:", err)
										return -1
									}
									return SearchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
								}
							}
						}
					}
				}
			} else if index == 14 {
				// ==== DOUBLE INDIRECT CASE ====
				fmt.Println("Indirect case: Double Indirect Block")

				// Read the first-level Pointerblock
				var firstLevelPointerBlock DiskStruct.Pointerblock
				if err := FileManagement.ReadObject(file, &firstLevelPointerBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
					fmt.Println("Error reading first-level Pointerblock:", err)
					return -1
				}

				// Iterate over the first-level pointers
				for _, firstPointer := range firstLevelPointerBlock.B_pointers {
					if firstPointer != -1 {
						// Read the second-level Pointerblock
						var secondLevelPointerBlock DiskStruct.Pointerblock
						if err := FileManagement.ReadObject(file, &secondLevelPointerBlock, int64(tempSuperblock.S_block_start+firstPointer*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
							fmt.Println("Error reading second-level Pointerblock:", err)
							return -1
						}

						// Iterate over the second-level pointers
						for _, secondPointer := range secondLevelPointerBlock.B_pointers {
							if secondPointer != -1 {
								var crrFolderBlock DiskStruct.Folderblock
								// Read the Folderblock pointed by the second-level pointer
								if err := FileManagement.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+secondPointer*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
									fmt.Println("Error reading Folderblock:", err)
									return -1
								}

								// Iterate over the contents of the Folderblock
								for _, folder := range crrFolderBlock.B_content {
									fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

									if strings.Contains(string(folder.B_name[:]), SearchedName) {
										fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
										if len(StepsPath) == 0 {
											fmt.Println("Folder found======")
											return folder.B_inodo
										} else {
											fmt.Println("NextInode======")
											var NextInode DiskStruct.Inode
											// Read the next Inode
											if err := FileManagement.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
												fmt.Println("Error reading NextInode:", err)
												return -1
											}
											return SearchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
										}
									}
								}
							}
						}
					}
				}
			}
		} else if index == 15 {
			// ==== TRIPLE INDIRECT CASE ====
			fmt.Println("Indirect case: Triple Indirect Block")

			// Read the first-level Pointerblock
			var firstLevelPointerBlock DiskStruct.Pointerblock
			if err := FileManagement.ReadObject(file, &firstLevelPointerBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
				fmt.Println("Error reading first-level Pointerblock:", err)
				return -1
			}

			// Iterate over the first-level pointers
			for _, firstPointer := range firstLevelPointerBlock.B_pointers {
				if firstPointer != -1 {
					// Read the second-level Pointerblock
					var secondLevelPointerBlock DiskStruct.Pointerblock
					if err := FileManagement.ReadObject(file, &secondLevelPointerBlock, int64(tempSuperblock.S_block_start+firstPointer*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
						fmt.Println("Error reading second-level Pointerblock:", err)
						return -1
					}

					// Iterate over the second-level pointers
					for _, secondPointer := range secondLevelPointerBlock.B_pointers {
						if secondPointer != -1 {
							// Read the third-level Pointerblock
							var thirdLevelPointerBlock DiskStruct.Pointerblock
							if err := FileManagement.ReadObject(file, &thirdLevelPointerBlock, int64(tempSuperblock.S_block_start+secondPointer*int32(binary.Size(DiskStruct.Pointerblock{})))); err != nil {
								fmt.Println("Error reading third-level Pointerblock:", err)
								return -1
							}

							// Iterate over the third-level pointers
							for _, thirdPointer := range thirdLevelPointerBlock.B_pointers {
								if thirdPointer != -1 {
									var crrFolderBlock DiskStruct.Folderblock
									// Read the Folderblock pointed by the third-level pointer
									if err := FileManagement.ReadObject(file, &crrFolderBlock, int64(tempSuperblock.S_block_start+thirdPointer*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
										fmt.Println("Error reading Folderblock:", err)
										return -1
									}

									// Iterate over the contents of the Folderblock
									for _, folder := range crrFolderBlock.B_content {
										fmt.Println("Folder === Name:", string(folder.B_name[:]), "B_inodo", folder.B_inodo)

										if strings.Contains(string(folder.B_name[:]), SearchedName) {
											fmt.Println("len(StepsPath)", len(StepsPath), "StepsPath", StepsPath)
											if len(StepsPath) == 0 {
												fmt.Println("Folder found======")
												return folder.B_inodo
											} else {
												fmt.Println("NextInode======")
												var NextInode DiskStruct.Inode
												// Read the next Inode
												if err := FileManagement.ReadObject(file, &NextInode, int64(tempSuperblock.S_inode_start+folder.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
													fmt.Println("Error reading NextInode:", err)
													return -1
												}
												return SearchInodeByPath(StepsPath, NextInode, file, tempSuperblock)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		index++
	}

	fmt.Println("======End BUSQUEDA INODO POR PATH======")
	return 0
}

// Logout function
func Logout() string {
	fmt.Println("====== Start LOGOUT ======")

	// Get the mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var sessionActive bool
	var activePartitionID string

	// Verify if there is an active session
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
				sessionActive = true // There is an active session
				activePartitionID = partition.ID
				break
			}
		}
		if sessionActive {
			break
		}
	}

	// No logout if there is no active session
	if !sessionActive {
		fmt.Println("Error: No hay ninguna sesión activa.")
		fmt.Println("====== End LOGOUT ======")
		return "No hay ninguna sesión activa."
	}

	// Logout the active session
	DiskControl.MarkPartitionAsLoggedOut(activePartitionID)
	fmt.Println("Sesión cerrada con éxito en la partición:", activePartitionID)

	fmt.Println("====== End LOGOUT ======")
	return "Logged out successfully."
}

// Get the data from an Inode
func GetInodeFileData(Inode DiskStruct.Inode, file *os.File, tempSuperblock DiskStruct.Superblock) string {
	fmt.Println("======Start CONTENIDO DEL BLOQUE======")
	index := int32(0)

	var content string

	// Iterate over i_blocks from Inode
	for _, block := range Inode.I_block {
		if block != -1 {
			//Inside of direct ones
			if index < 13 {
				var crrFileBlock DiskStruct.Fileblock
				// Read object from bin file
				if err := FileManagement.ReadObject(file, &crrFileBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Fileblock{})))); err != nil {
					return ""
				}
				content += string(crrFileBlock.B_content[:])
			} else {
				fmt.Print("indirectos")
			}
		}
		index++
	}

	fmt.Println("======End CONTENIDO DEL BLOQUE======")
	return content
}

//===== MKUSER =====

func AppendToFileBlock(inode *DiskStruct.Inode, newData string, file *os.File, superblock DiskStruct.Superblock) error {
	// read existing data
	existingData := GetInodeFileData(*inode, file, superblock)
	fmt.Println("Datos existentes en el bloque:", existingData)

	// Concatenate the new data
	fullData := existingData + newData
	fmt.Println("Datos completos después de agregar:", fullData)

	// Verify that the file size does not exceed the total capacity
	blockSize := binary.Size(DiskStruct.Fileblock{})
	totalCapacity := len(inode.I_block) * blockSize
	if len(fullData) > totalCapacity {
		fmt.Printf("Error: El tamaño del archivo (%d bytes) excede la capacidad total asignada (%d bytes).\n", len(fullData), totalCapacity)
		return fmt.Errorf("el tamaño del archivo excede la capacidad total asignada y no se ha implementado la creación de bloques adicionales")
	}

	// Split the full data into blocks
	remainingData := fullData
	blockIndex := 0

	for len(remainingData) > 0 {
		// If no block is assigned, find a free block
		if blockIndex >= len(inode.I_block) || inode.I_block[blockIndex] == -1 {
			newBlockIndex := FindFreeBlock(superblock, file)
			if newBlockIndex == -1 {
				return fmt.Errorf("no hay bloques libres disponibles")
			}
			inode.I_block[blockIndex] = int32(newBlockIndex)
			fmt.Printf("Asignando nuevo bloque: %d\n", newBlockIndex)
		}

		// Create a new file block with the data
		var updatedFileBlock DiskStruct.Fileblock
		copy(updatedFileBlock.B_content[:], remainingData[:min(len(remainingData), blockSize)])

		// Write the updated block to the file
		position := int64(superblock.S_block_start + inode.I_block[blockIndex]*int32(blockSize))
		fmt.Printf("Escribiendo bloque en la posición: %d\n", position)
		if err := FileManagement.WriteObject(file, updatedFileBlock, position); err != nil {
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		// Update the rest of the data
		remainingData = remainingData[min(len(remainingData), blockSize):]
		blockIndex++
	}

	// Update inode size
	inode.I_size = int32(len(fullData))
	inodePosition := int64(superblock.S_inode_start + inode.I_block[0]*int32(binary.Size(DiskStruct.Inode{})))
	fmt.Printf("Actualizando inodo en la posición: %d\n", inodePosition)
	fmt.Printf("Nuevo tamaño del inodo (I_size): %d\n", inode.I_size)
	if err := FileManagement.WriteObject(file, *inode, inodePosition); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	fmt.Println("Bloque e inodo actualizados correctamente.")
	return nil
}

// Aux func to find a free block of the superblock and
func FindFreeBlock(superblock DiskStruct.Superblock, file *os.File) int {
	// Read the bitmap of blocks
	bitmap := make([]byte, superblock.S_blocks_count)
	if _, err := file.ReadAt(bitmap, int64(superblock.S_bm_block_start)); err != nil {
		fmt.Println("Error al leer el bitmap de bloques:", err)
		return -1
	}

	// Find the first free block
	for i, b := range bitmap {
		if b == 0 {
			// Mark the block as used
			bitmap[i] = 1
			if _, err := file.WriteAt(bitmap, int64(superblock.S_bm_block_start)); err != nil {
				fmt.Println("Error al actualizar el bitmap de bloques:", err)
				return -1
			}
			return i
		}
	}

	return -1 //If there is no free block
}

// Aux fun copy the new data to the block
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Mkusr(user string, pass string, grp string) string {
	fmt.Printf("Parámetros recibidos: user=%s, pass=%s, grp=%s\n", user, pass, grp)

	// Validate that the user is root
	if !IsRootUser() {
		fmt.Println("Error: Solo el usuario root puede ejecutar este comando.")
		fmt.Println("====== End MKUSR ======")
		return "Solo el usuario root puede ejecutar este comando."
	}

	// Validate the length of the parameters
	if len(user) > 10 || len(pass) > 10 || len(grp) > 10 {
		fmt.Println("Error: Los parámetros 'user', 'pass' y 'grp' no pueden exceder los 10 caracteres.")
		fmt.Println("====== End MKUSR ======")
		return "Los parámetros 'user', 'pass' y 'grp' no pueden exceder los 10 caracteres."
	}

	// Clean the parameters
	user = strings.TrimSpace(strings.ReplaceAll(user, "\"", ""))
	pass = strings.TrimSpace(strings.ReplaceAll(pass, "\"", ""))
	grp = strings.TrimSpace(strings.ReplaceAll(grp, "\"", ""))

	// Get mounted partitions and find the active partition
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Find the active partition
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End MKUSR ======")
		return "No hay ninguna partición activa."
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End MKUSR ======")
		return "No se pudo abrir el archivo."
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return "Error al leer el MBR."
	}

	// Read the Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return "Error al leer el Superblock."
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End MKUSR ======")
		return "No se encontró el archivo users.txt."
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return "Error al leer el Inodo del archivo users.txt."
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido actual del archivo users.txt:")
	fmt.Println(data)

	// Verify if the group exists
	lines := strings.Split(data, "\n")
	groupExists := false
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 3 && words[1] == "G" && words[2] == grp {
			groupExists = true
			fmt.Printf("Grupo encontrado: %s\n", grp)
			break
		}
	}

	if !groupExists {
		fmt.Println("Error: El grupo especificado no existe.")
		fmt.Println("====== End MKUSR ======")
		return "El grupo especificado no existe."
	}

	// If the user already exists, return an error
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 5 && strings.TrimSpace(words[1]) == "U" && strings.TrimSpace(words[2]) == user {
			return fmt.Sprintf("El usuario '%s' ya existe.", user)
		}
	}

	// Global counter for the user IDs
	if nextUserID == 0 {
		// Init the counter
		if err := InitializeUserIDCounter(file, tempSuperblock); err != nil {
			fmt.Println(err)
			fmt.Println("====== End MKUSR ======")
			return "Error al inicializar el contador de ID de usuario."
		}
	}

	newUserID := nextUserID
	nextUserID++ // Increase the counter for the next user
	newUser := fmt.Sprintf("%d,U,%s,%s,%s\n", newUserID, user, grp, pass)
	fmt.Printf("Nuevo usuario a agregar: %s\n", newUser)

	// Add the new user to the users.txt file
	if err := AppendToFileBlock(&crrInode, newUser, file, tempSuperblock); err != nil {
		fmt.Println("Error: No se pudo agregar el nuevo usuario al archivo users.txt:", err)
		fmt.Println("====== End MKUSR ======")
		return "Error al agregar el nuevo usuario al archivo users.txt."
	}

	fmt.Println("Usuario creado exitosamente.")
	fmt.Println("====== End MKUSR ======")
	return "User created successfully."
}

// Aux fun to verify if the user is root or not
func IsRootUser() bool {
	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
				// If there is an active session, verify if the user is root
				file, err := FileManagement.OpenFile(partition.Path)
				if err != nil {
					fmt.Println("Error: No se pudo abrir el archivo:", err)
					return false
				}
				defer file.Close()

				var TempMBR DiskStruct.MRB
				if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
					fmt.Println("Error: No se pudo leer el MBR:", err)
					return false
				}

				var tempSuperblock DiskStruct.Superblock
				if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[0].Start)); err != nil {
					fmt.Println("Error: No se pudo leer el Superblock:", err)
					return false
				}

				// Find the users.txt file
				indexInode := InitSearch("/users.txt", file, tempSuperblock)
				if indexInode == -1 {
					fmt.Println("Error: No se encontró el archivo users.txt.")
					return false
				}

				var crrInode DiskStruct.Inode
				if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
					fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
					return false
				}

				// Read the content of the users.txt file
				data := GetInodeFileData(crrInode, file, tempSuperblock)

				// Verify if the logged user is root
				lines := strings.Split(data, "\n")
				for _, line := range lines {
					words := strings.Split(line, ",")
					if len(words) == 5 && words[1] == "U" && words[3] == "root" {
						return true
					}
				}
			}
		}
	}

	// If not active session, return false
	return false
}

func PrintUsersFile() {
	fmt.Println("====== Start Print Users File ======")

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	// Find the active partition
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Active partition
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End Print Users File ======")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End Print Users File ======")
		return
	}
	defer file.Close()

	// Read the Superblock
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End Print Users File ======")
		return
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido del archivo users.txt:")
	fmt.Println(data)

	fmt.Println("====== End Print Users File ======")
}

// ==== GLOBAL COUNTERS ====
var nextUserID int = 0
var nextGroupID int = 0

// Func to initialize the user ID counter
func InitializeUserIDCounter(file *os.File, tempSuperblock DiskStruct.Superblock) error {
	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Printf("error: No se encontró el archivo users.txt.")

	}

	// Read the Inode of the users.txt file
	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Printf("error al leer el Inodo del archivo users.txt: %v", err)
	}

	// Read the data from the file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	lines := strings.Split(data, "\n")

	// Calculating the max ID
	maxID := 0
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) > 0 {
			// Get the ID from the first column
			if id, err := strconv.Atoi(strings.TrimSpace(words[0])); err == nil {
				if id > maxID {
					maxID = id
				}
			}
		}
	}

	// Update global counter
	nextUserID = maxID + 1
	fmt.Printf("Contador de IDs inicializado en: %d\n", nextUserID)
	return nil
}

// Func to init the grup counter
func InitializeGroupIDCounter(file *os.File, tempSuperblock DiskStruct.Superblock) error {
	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Printf("Error: No se encontró el archivo users.txt.\n")
		return fmt.Errorf("archivo users.txt no encontrado")
	}

	// Read the Inode of the users.txt file
	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Printf("Error al leer el Inodo del archivo users.txt: %v\n", err)
		return err
	}

	// Read the data from the file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	lines := strings.Split(data, "\n")

	// Calculating the max group ID
	maxGroupID := 0
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) > 0 && len(words) >= 3 && strings.TrimSpace(words[1]) == "G" {
			// Get the ID from the first column
			if id, err := strconv.Atoi(strings.TrimSpace(words[0])); err == nil {
				if id > maxGroupID {
					maxGroupID = id
				}
			}
		}
	}

	// Update global counter
	nextGroupID = maxGroupID + 1
	fmt.Printf("Contador de IDs de grupos inicializado en: %d\n", nextGroupID)
	return nil
}

func Mkgrp(name string) string {
	fmt.Printf("Parámetro recibido: name=%s\n", name)

	// User must be root
	if !IsRootUser() {
		fmt.Println("Error: Solo el usuario root puede ejecutar este comando.")
		fmt.Println("====== End MKGRP ======")
		return "Error: Solo el usuario root puede ejecutar este comando."
	}

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Find the active partition
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End MKGRP ======")
		return "Error: No hay ninguna partición activa."
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End MKGRP ======")
		return "Error: No se pudo abrir el archivo:"
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return "Error: No se pudo leer el MBR:"
	}

	// Read the Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return "Error: No se pudo leer el Superblock"
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End MKGRP ======")
		return "Error: No se encontró el archivo users.txt."
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return "Error: No se pudo leer el Inodo del archivo users.txt"
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido actual del archivo users.txt:")
	fmt.Println(data)

	// Verify if the group already exists
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 3 && words[1] == "G" && words[2] == name {
			fmt.Println("Error: El grupo especificado ya existe.")
			fmt.Println("====== End MKGRP ======")
			return "El grupo especificado ya existe."
		}
	}

	// Init the global counter for the group IDs
	if nextGroupID == 0 {
		if err := InitializeGroupIDCounter(file, tempSuperblock); err != nil {
			fmt.Println(err)
			fmt.Println("====== End MKGRP ======")
			return "Error: No se pudo inicializar el contador de IDs de grupos."
		}
	}

	// Create the new group
	newGroupID := nextGroupID
	nextGroupID++ //Increase the counter for the next group
	newGroup := fmt.Sprintf("%d,G,%s\n", newGroupID, name)
	fmt.Printf("Nuevo grupo a agregar: %s\n", newGroup)

	// Add the new group to the users.txt file
	if err := AppendToFileBlock(&crrInode, newGroup, file, tempSuperblock); err != nil {
		fmt.Println("Error: No se pudo agregar el nuevo grupo al archivo users.txt:", err)
		fmt.Println("====== End MKGRP ======")
		return "Error: No se pudo agregar el nuevo grupo al archivo users.txt"
	}

	fmt.Println("Grupo creado exitosamente.")
	fmt.Println("====== End MKGRP ======")
	return "Group created successfully."
}

func Rmusr(user string) {
	fmt.Printf("Parámetro recibido: user='%s'\n", user)

	// Validate that the user is root
	if !IsRootUser() {
		fmt.Println("Error: Solo el usuario root puede ejecutar este comando.")
		fmt.Println("====== End RMUSR ======")
		return
	}

	// Get mounted partitions and find the active partition
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { //active sesion
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End RMUSR ======")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End RMUSR ======")
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Read the superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // active session
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End RMUSR ======")
		return
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido actual del archivo users.txt:")
	fmt.Println(data)

	// Find the user to remove
	lines := strings.Split(data, "\n")
	var updatedLines []string
	userFound := false

	// Clean the user parameter
	cleanedUser := strings.TrimSpace(user)
	cleanedUser = strings.ReplaceAll(cleanedUser, "\u200B", "") // Remove invisible characters

	for _, line := range lines {
		// Eliminar espacios en blanco adicionales
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\u200B", "") // Remove invisible characters
		if line == "" {
			continue // Ignorar líneas vacías
		}

		words := strings.Split(line, ",")
		fmt.Printf("Campos de la línea: %v\n", words)

		if len(words) == 5 {
			// Clean the user field
			for i := range words {
				words[i] = strings.TrimSpace(words[i])
				words[i] = strings.ReplaceAll(words[i], "\u200B", "")
			}

			// Compare the user
			fmt.Printf("Comparando usuario: '%s' con '%s'\n", words[2], cleanedUser)
			if words[1] == "U" && words[2] == cleanedUser {
				// Change the status of the user to 0
				words[0] = "0"
				userFound = true
			}
			updatedLines = append(updatedLines, strings.Join(words, ","))
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !userFound {
		fmt.Printf("Error: El usuario '%s' no existe.\n", cleanedUser)
		fmt.Println("====== End RMUSR ======")
		return
	}

	// Update the content of the users.txt file
	newData := strings.Join(updatedLines, "\n")

	// Overwrite the file block with the updated data
	if err := OverwriteFileBlock(&crrInode, newData, file, tempSuperblock, indexInode); err != nil {
		fmt.Println("Error: No se pudo actualizar el archivo users.txt:", err)
		fmt.Println("====== End RMUSR ======")
		return
	}

	// Updated content of the users.txt file
	fmt.Println("Contenido actualizado del archivo users.txt:")
	fmt.Println(newData)

	fmt.Println("Usuario eliminado exitosamente.")
	fmt.Println("====== End RMUSR ======")
}

func OverwriteFileBlock(inode *DiskStruct.Inode, newData string, file *os.File, superblock DiskStruct.Superblock, indexInode int32) error {
	// Split the new data into blocks
	blockSize := binary.Size(DiskStruct.Fileblock{})
	remainingData := newData
	blockIndex := 0

	for len(remainingData) > 0 {
		// If no block is assigned, find a free block
		if blockIndex >= len(inode.I_block) || inode.I_block[blockIndex] == -1 {
			newBlockIndex := FindFreeBlock(superblock, file)
			if newBlockIndex == -1 {
				return fmt.Errorf("no hay bloques libres disponibles")
			}
			inode.I_block[blockIndex] = int32(newBlockIndex)
		}

		// New file block with the data
		var updatedFileBlock DiskStruct.Fileblock
		copy(updatedFileBlock.B_content[:], remainingData[:min(len(remainingData), blockSize)])

		// Write the updated block to the file
		position := int64(superblock.S_block_start + inode.I_block[blockIndex]*int32(blockSize))
		if err := FileManagement.WriteObject(file, updatedFileBlock, position); err != nil {
			return fmt.Errorf("error al escribir el bloque actualizado: %v", err)
		}

		// Update the rest of the data
		remainingData = remainingData[min(len(remainingData), blockSize):]
		blockIndex++
	}

	//Update inode size
	inode.I_size = int32(len(newData))
	inodePosition := int64(superblock.S_inode_start + indexInode*int32(binary.Size(DiskStruct.Inode{})))
	if err := FileManagement.WriteObject(file, *inode, inodePosition); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	return nil
}

func Rmgrp(name string) {
	fmt.Printf("Parámetro recibido: name='%s'\n", name)

	// VUser must be root
	if !IsRootUser() {
		fmt.Println("Error: Solo el usuario root puede ejecutar este comando.")
		fmt.Println("====== End RMGRP ======")
		return
	}

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // active session
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End RMGRP ======")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End RMGRP ======")
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Read the Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End RMGRP ======")
		return
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido actual del archivo users.txt:")
	fmt.Println(data)

	// find the group to remove
	lines := strings.Split(data, "\n")
	var updatedLines []string
	groupFound := false

	// Clean the group parameter
	cleanedName := strings.TrimSpace(name)
	cleanedName = strings.ReplaceAll(cleanedName, "\u200B", "") // Delete invisible characters

	for _, line := range lines {
		// Eliminar espacios en blanco adicionales
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\u200B", "") // Delete invisible characters
		if line == "" {
			continue // Ignorar líneas vacías
		}

		words := strings.Split(line, ",")
		fmt.Printf("Campos de la línea: %v\n", words)

		if len(words) == 3 {
			// Clean the group field
			for i := range words {
				words[i] = strings.TrimSpace(words[i])
				words[i] = strings.ReplaceAll(words[i], "\u200B", "")
			}

			// Compare the group
			fmt.Printf("Comparando grupo: '%s' con '%s'\n", words[2], cleanedName)
			if words[1] == "G" && words[2] == cleanedName {
				// Change the id of the group to 0
				words[0] = "0"
				groupFound = true
			}
			updatedLines = append(updatedLines, strings.Join(words, ","))
		} else {
			updatedLines = append(updatedLines, line)
		}
	}

	if !groupFound {
		fmt.Printf("Error: El grupo '%s' no existe.\n", cleanedName)
		fmt.Println("====== End RMGRP ======")
		return
	}

	// Update the content of the users.txt file
	newData := strings.Join(updatedLines, "\n")

	// Overwrite the file block with the updated data
	if err := OverwriteFileBlock(&crrInode, newData, file, tempSuperblock, indexInode); err != nil {
		fmt.Println("Error: No se pudo actualizar el archivo users.txt:", err)
		fmt.Println("====== End RMGRP ======")
		return
	}

	fmt.Println("Contenido actualizado del archivo users.txt:")
	fmt.Println(newData)

	fmt.Println("Grupo eliminado exitosamente.")
	fmt.Println("====== End RMGRP ======")
}

func Chgrp(user string, grp string) {
	fmt.Printf("Parámetros recibidos: user='%s', grp='%s'\n", user, grp)

	// User must be root
	if !IsRootUser() {
		fmt.Println("Error: Solo el usuario root puede ejecutar este comando.")
		fmt.Println("====== End CHGRP ======")
		return
	}

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // active session
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End CHGRP ======")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End CHGRP ======")
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Find the users.txt file
	indexInode := InitSearch("/users.txt", file, tempSuperblock)
	if indexInode == -1 {
		fmt.Println("Error: No se encontró el archivo users.txt.")
		fmt.Println("====== End CHGRP ======")
		return
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
		return
	}

	// Read the content of the users.txt file
	data := GetInodeFileData(crrInode, file, tempSuperblock)
	fmt.Println("Contenido actual del archivo users.txt:")
	fmt.Println(data)

	// Verify if the group exists
	lines := strings.Split(data, "\n")
	groupExists := false
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 3 && words[1] == "G" && words[2] == grp {
			groupExists = true
			break
		}
	}

	if !groupExists {
		fmt.Printf("Error: El grupo '%s' no existe.\n", grp)
		fmt.Println("====== End CHGRP ======")
		return
	}

	// Update the group of the user
	var updatedLines []string
	userFound := false
	for _, line := range lines {
		words := strings.Split(line, ",")
		if len(words) == 5 && words[1] == "U" && words[2] == user {
			words[3] = grp // Change the group
			userFound = true
		}
		updatedLines = append(updatedLines, strings.Join(words, ","))
	}

	if !userFound {
		fmt.Printf("Error: El usuario '%s' no existe.\n", user)
		fmt.Println("====== End CHGRP ======")
		return
	}

	// Update the content of the users.txt file
	newData := strings.Join(updatedLines, "\n")

	// Overwrite the file block with the updated data
	if err := OverwriteFileBlock(&crrInode, newData, file, tempSuperblock, indexInode); err != nil {
		fmt.Println("Error: No se pudo actualizar el archivo users.txt:", err)
		fmt.Println("====== End CHGRP ======")
		return
	}

	fmt.Println("Contenido actualizado del archivo users.txt:")
	fmt.Println(newData)

	fmt.Println("Grupo del usuario actualizado exitosamente.")
	fmt.Println("====== End CHGRP ======")
}

func Mkfile(path string, recursive bool, size int, contentPath string) {
	fmt.Printf("Parámetros recibidos: path='%s', recursive=%t, size=%d, contentPath='%s'\n", path, recursive, size, contentPath)

	// Validar que haya una sesión activa
	if !IsUserLoggedIn() {
		fmt.Println("Error: No hay un usuario logueado.")
		return
	}

	// Validar que el path no esté vacío
	if path == "" {
		fmt.Println("Error: El parámetro 'path' es obligatorio.")
		return
	}

	// Obtener el contenido del archivo
	var fileContent string
	if contentPath != "" {
		// Leer el contenido del archivo desde el sistema local
		contentBytes, err := os.ReadFile(contentPath)
		if err != nil {
			fmt.Printf("Error: No se pudo leer el archivo en '%s': %v\n", contentPath, err)
			return
		}
		fileContent = string(contentBytes)
	} else {
		// Generar contenido basado en el tamaño
		fileContent = generateContent(size)
	}

	// Obtener la partición activa
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Buscar la partición activa
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		return
	}

	// Abrir el archivo binario
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		return
	}
	defer file.Close()

	// Leer el MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Leer el Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Partición activa
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Dividir el path en carpetas y el nombre del archivo
	steps := strings.Split(strings.Trim(path, "/"), "/")
	fileName := steps[len(steps)-1]
	parentPath := steps[:len(steps)-1]

	// Buscar el inodo raíz
	var rootInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &rootInode, int64(tempSuperblock.S_inode_start)); err != nil {
		fmt.Println("Error: No se pudo leer el inodo raíz:", err)
		return
	}

	// Navegar por el path para encontrar el inodo del directorio padre
	currentInode := rootInode
	for _, step := range parentPath {
		index := SearchInodeByPath([]string{step}, currentInode, file, tempSuperblock)
		if index == -1 {
			// Si no existe y no se usa -r, mostrar error
			if !recursive {
				fmt.Printf("Error: La carpeta '%s' no existe y no se usó el parámetro '-r'.\n", step)
				return
			}

			// Crear la carpeta si no existe
			newInodeIndex := FindFreeBlock(tempSuperblock, file)
			if newInodeIndex == -1 {
				fmt.Println("Error: No hay bloques libres disponibles para crear la carpeta.")
				return
			}

			if err := AddFolderToInode(&currentInode, step, int32(newInodeIndex), file, tempSuperblock); err != nil {
				fmt.Printf("Error: No se pudo agregar la carpeta '%s': %v\n", step, err)
				return
			}

			// Crear el nuevo inodo para la carpeta
			newInode := DiskStruct.Inode{}
			initInode(&newInode, time.Now().Format("2006-01-02 15:04:05"))
			newInode.I_type[0] = '0' // Carpeta
			if err := FileManagement.WriteObject(file, newInode, int64(tempSuperblock.S_inode_start+int32(newInodeIndex)*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
				fmt.Printf("Error: No se pudo escribir el nuevo inodo para la carpeta '%s': %v\n", step, err)
				return
			}

			currentInode = newInode
		} else {
			// Leer el inodo correspondiente
			if err := FileManagement.ReadObject(file, &currentInode, int64(tempSuperblock.S_inode_start+index*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
				fmt.Printf("Error: No se pudo leer el inodo de la carpeta '%s': %v\n", step, err)
				return
			}
		}
	}

	// Crear el archivo en el directorio padre
	newInodeIndex := FindFreeBlock(tempSuperblock, file)
	if newInodeIndex == -1 {
		fmt.Println("Error: No hay bloques libres disponibles para crear el archivo.")
		return
	}

	if err := AddFolderToInode(&currentInode, fileName, int32(newInodeIndex), file, tempSuperblock); err != nil {
		fmt.Printf("Error: No se pudo agregar el archivo '%s': %v\n", fileName, err)
		return
	}

	// Crear el nuevo inodo para el archivo
	newInode := DiskStruct.Inode{}
	initInode(&newInode, time.Now().Format("2006-01-02 15:04:05"))
	newInode.I_type[0] = '1' // Archivo
	if err := AppendToFileBlock(&newInode, fileContent, file, tempSuperblock); err != nil {
		fmt.Printf("Error: No se pudo escribir el contenido del archivo '%s': %v\n", fileName, err)
		return
	}

	if err := FileManagement.WriteObject(file, newInode, int64(tempSuperblock.S_inode_start+int32(newInodeIndex)*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		fmt.Printf("Error: No se pudo escribir el nuevo inodo para el archivo '%s': %v\n", fileName, err)
		return
	}

	fmt.Printf("Archivo '%s' creado exitosamente.\n", path)
}

// Aux func to generate content based on the size
func generateContent(size int) string {
	content := ""
	for i := 0; i < size; i++ {
		content += fmt.Sprintf("%d", i%10)
	}
	return content
}

func IsUserLoggedIn() bool {
	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()

	// Verify if there is an active session
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // find active session
				return true
			}
		}
	}

	// False if there is no active session
	return false
}

func Cat(filePaths ...string) string {
	fmt.Println("====== Start CAT ======")

	// User must be logged in
	if !IsUserLoggedIn() {
		fmt.Println("Error: No hay un usuario logueado.")
		fmt.Println("====== End CAT ======")
		return "Error: No hay un usuario logueado."
	}

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	// Buscar la partición activa
	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn { // Sesión activa
				filepath = partition.Path
				partitionFound = true
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna partición activa.")
		fmt.Println("====== End CAT ======")
		return "Error: No hay ninguna partición activa."
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End CAT ======")
		return fmt.Sprintf("Error: No se pudo abrir el archivo: %v", err)
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		fmt.Println("====== End CAT ======")
		return fmt.Sprintf("Error: No se pudo leer el MBR: %v", err)
	}

	// Read the Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				fmt.Println("====== End CAT ======")
				return fmt.Sprintf("Error: No se pudo leer el Superblock: %v", err)
			}
			break
		}
	}

	// Process each file path
	var result string
	for _, filePath := range filePaths {
		// Find the file in the filesystem
		indexInode := InitSearch(filePath, file, tempSuperblock)
		if indexInode == -1 {
			fmt.Printf("Error: No se encontró el archivo %s.\n", filePath)
			result += fmt.Sprintf("Error: No se encontró el archivo %s.\n", filePath)
			continue
		}

		var crrInode DiskStruct.Inode
		if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
			fmt.Printf("Error: No se pudo leer el Inodo del archivo %s: %v\n", filePath, err)
			result += fmt.Sprintf("Error: No se pudo leer el Inodo del archivo %s: %v\n", filePath, err)
			continue
		}

		// Verify read permissions
		if !HasReadPermission(crrInode) {
			fmt.Printf("Error: No tiene permisos de lectura para el archivo %s.\n", filePath)
			result += fmt.Sprintf("Error: No tiene permisos de lectura para el archivo %s.\n", filePath)
			continue
		}

		// Read the file data
		data := GetInodeFileData(crrInode, file, tempSuperblock)

		// Clean the data
		cleanedData := strings.TrimSpace(data)
		cleanedData = strings.ReplaceAll(cleanedData, "\u0000", "")
		cleanedData = strings.ReplaceAll(cleanedData, "\u200B", "")

		fmt.Printf("Contenido del archivo %s:\n%s\n", filePath, cleanedData)
		result += fmt.Sprintf("Contenido del archivo %s:\n%s\n", filePath, cleanedData)
	}

	fmt.Println("====== End CAT ======")
	return result
}

// Aux function to check read permissions
func HasReadPermission(inode DiskStruct.Inode) bool {
	// Get the current logged in user
	currentUser := GetLoggedInUser()
	if currentUser == nil {
		fmt.Println("Error: No hay un usuario logueado.")
		return false
	}

	// Get permissions from the inode
	permissions := string(inode.I_perm[:])
	if len(permissions) != 3 {
		fmt.Println("Error: Permisos inválidos en el inodo.")
		return false
	}

	// Verify permissions
	userPerm := permissions[0]  // Owner permissions
	groupPerm := permissions[1] // Group permissions
	otherPerm := permissions[2] // Other permissions

	// Check if the user is the owner, belongs to the group, or is "others"
	if inode.I_uid == currentUser.UID {
		// User is the owner
		return userPerm == '4' || userPerm == '5' || userPerm == '6' || userPerm == '7'
	} else if inode.I_gid == currentUser.GID {
		// User is in the group
		return groupPerm == '4' || groupPerm == '5' || groupPerm == '6' || groupPerm == '7'
	} else {
		// User is "others"
		return otherPerm == '4' || otherPerm == '5' || otherPerm == '6' || otherPerm == '7'
	}
}

// Aux func to get the logged in user
func GetLoggedInUser() *DiskStruct.User {
	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
				// Read the users.txt file to get the logged in user
				file, err := FileManagement.OpenFile(partition.Path)
				if err != nil {
					fmt.Println("Error: No se pudo abrir el archivo:", err)
					return nil
				}
				defer file.Close()

				// Read the MBR
				var TempMBR DiskStruct.MRB
				if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
					fmt.Println("Error: No se pudo leer el MBR:", err)
					return nil
				}

				// Read the Superblock
				var tempSuperblock DiskStruct.Superblock
				if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[0].Start)); err != nil {
					fmt.Println("Error: No se pudo leer el Superblock:", err)
					return nil
				}

				// Find the users.txt file
				indexInode := InitSearch("/users.txt", file, tempSuperblock)
				if indexInode == -1 {
					fmt.Println("Error: No se encontró el archivo users.txt.")
					return nil
				}

				// Read the Inode of the users.txt file
				var crrInode DiskStruct.Inode
				if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
					fmt.Println("Error: No se pudo leer el Inodo del archivo users.txt:", err)
					return nil
				}

				// Read the data from the file
				data := GetInodeFileData(crrInode, file, tempSuperblock)
				lines := strings.Split(data, "\n")

				// Find the logged in user
				for _, line := range lines {
					words := strings.Split(line, ",")
					if len(words) == 5 && words[1] == "U" && words[3] == "root" {
						return &DiskStruct.User{
							UID: 1,
							GID: 1,
						}
					}
				}
			}
		}
	}

	return nil
}

func Mkdir(path string, p bool) {
	fmt.Println("====== Start MKDIR ======")
	fmt.Printf("Parámetros recibidos: path=%s, p=%t\n", path, p)

	// Get mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
				filepath = partition.Path
				partitionFound = true
				fmt.Printf("Partición activa encontrada: %s\n", filepath)
				break
			}
		}
		if partitionFound {
			break
		}
	}

	if !partitionFound {
		fmt.Println("Error: No hay ninguna sesión activa.")
		fmt.Println("====== End MKDIR ======")
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		fmt.Println("Error: No se pudo abrir el archivo:", err)
		fmt.Println("====== End MKDIR ======")
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error: No se pudo leer el MBR:", err)
		return
	}

	// Read the Superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active session
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				fmt.Println("Error: No se pudo leer el Superblock:", err)
				return
			}
			break
		}
	}

	// Split the path into folders
	steps := strings.Split(strings.Trim(path, "/"), "/")
	fmt.Printf("Ruta dividida en pasos: %v\n", steps)

	// Find the root inode
	var rootInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &rootInode, int64(tempSuperblock.S_inode_start)); err != nil {
		fmt.Println("Error: No se pudo leer el inodo raíz:", err)
		return
	}

	// Create the folders
	currentInode := rootInode
	for i, step := range steps {
		fmt.Printf("Procesando carpeta: %s\n", step)

		// Check if the folder already exists
		index := SearchInodeByPath([]string{step}, currentInode, file, tempSuperblock)
		if index == -1 {
			// Error if the folder does not exist and -p is not used
			if !p && i != len(steps)-1 {
				fmt.Printf("Error: La carpeta '%s' no existe y no se usó el parámetro -p.\n", step)
				fmt.Println("====== End MKDIR ======")
				return
			}

			// Create the folder if it does not exist
			newInodeIndex := FindFreeBlock(tempSuperblock, file)
			if newInodeIndex == -1 {
				fmt.Println("Error: No hay bloques libres disponibles para crear la carpeta.")
				fmt.Println("====== End MKDIR ======")
				return
			}

			// Update the current inode
			if err := AddFolderToInode(&currentInode, step, int32(newInodeIndex), file, tempSuperblock); err != nil {
				fmt.Printf("Error: No se pudo agregar la carpeta '%s': %v\n", step, err)
				fmt.Println("====== End MKDIR ======")
				return
			}

			// Create new inode for the folder
			newInode := DiskStruct.Inode{}
			initInode(&newInode, time.Now().Format("2006-01-02 15:04:05"))
			newInode.I_type[0] = '0' // Carpeta
			if err := FileManagement.WriteObject(file, newInode, int64(tempSuperblock.S_inode_start+int32(newInodeIndex)*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
				fmt.Printf("Error: No se pudo escribir el nuevo inodo para la carpeta '%s': %v\n", step, err)
				fmt.Println("====== End MKDIR ======")
				return
			}

			// Update the current inode
			currentInode = newInode
		} else {
			// If it exists, read the inode
			if err := FileManagement.ReadObject(file, &currentInode, int64(tempSuperblock.S_inode_start+index*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
				fmt.Printf("Error: No se pudo leer el inodo de la carpeta '%s': %v\n", step, err)
				fmt.Println("====== End MKDIR ======")
				return
			}
		}
	}

	fmt.Println("Carpeta(s) creada(s) exitosamente.")
	fmt.Println("====== End MKDIR ======")
}

// Aux func to add a folder to an inode
func AddFolderToInode(inode *DiskStruct.Inode, folderName string, folderInodeIndex int32, file *os.File, superblock DiskStruct.Superblock) error {
	var folderBlock DiskStruct.Folderblock
	blockIndex := -1

	// Find an empty block in the inode
	for i, block := range inode.I_block {
		if block == -1 {
			blockIndex = i
			break
		}
	}

	if blockIndex == -1 {
		return fmt.Errorf("no hay bloques libres en el inodo")
	}

	// Find a free block in the superblock
	newBlockIndex := FindFreeBlock(superblock, file)
	if newBlockIndex == -1 {
		return fmt.Errorf("no hay bloques libres disponibles")
	}

	inode.I_block[blockIndex] = int32(newBlockIndex)

	// Create the folder block
	copy(folderBlock.B_content[0].B_name[:], folderName)
	folderBlock.B_content[0].B_inodo = folderInodeIndex

	// Write the folder block to the file
	position := int64(superblock.S_block_start + int32(newBlockIndex)*int32(binary.Size(DiskStruct.Folderblock{})))
	if err := FileManagement.WriteObject(file, folderBlock, position); err != nil {
		return fmt.Errorf("error al escribir el bloque de carpeta: %v", err)
	}

	// Update the inode size
	inode.I_size += int32(binary.Size(DiskStruct.Folderblock{}))
	inodePosition := int64(superblock.S_inode_start + inode.I_block[0]*int32(binary.Size(DiskStruct.Inode{})))
	if err := FileManagement.WriteObject(file, *inode, inodePosition); err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	return nil
}

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
