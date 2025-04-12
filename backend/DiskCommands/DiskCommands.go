package DiskCommands

import (
	"Proyecto1/backend/DiskControl"
	"Proyecto1/backend/DiskStruct"
	"Proyecto1/backend/FileManagement"
	"Proyecto1/backend/UserManagement"
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`-(\w+)=("[^"]+"|\S+)`)

// Fuction to get the command and its parameters
func GetCommand(input string) (string, string) {
	parts := strings.Fields(input) // Split the input into parts
	if len(parts) > 0 {
		command := strings.ToLower(parts[0])   // Get the command in lowercase
		params := strings.Join(parts[1:], " ") // Join the rest of the parts into a string
		return command, params
	}
	return "", input
}

func Analyze() {

	for true {
		var input string
		fmt.Println("======================")
		fmt.Println("Ingrese comando: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input = scanner.Text()

		command, params := GetCommand(input)

		fmt.Println("Comando: ", command, " - ", "Parametro: ", params)

		AnalyzeCommand(command, params)

		//mkdisk -size=3000 -unit=K -fit=BF -path="/home/angely-gmartinez/Disks/disk1.bin"
	}
}

func AnalyzeCommand(command string, params string) {
	// Check the command
	if strings.Contains(command, "mkdisk") {
		fn_mkdisk(params) // Call the function mkdisk
	} else if strings.Contains(command, "rmdisk") {
		fn_rmdisk(params) // Call the function rmdisk
	} else if strings.Contains(command, "fdisk") {
		fn_fdisk(params) // Call the function fdisk
	} else if strings.Contains(command, "mount") {
		fn_mount(params) // Call the function mount
	} else if strings.Contains(command, "rep") {
		Fn_Rep(params) // Call the function rep
	} else if strings.Contains(command, "login") {
		fn_login(params) // Call the function login
	} else if strings.Contains(command, "mkusr") {
		fn_mkusr(params) // Call the function mkusr
	} else {
		fmt.Println("Error: Comando inválido o no encontrado")
	}
}

func fn_mkdisk(params string) {
	// Definir flag
	fs := flag.NewFlagSet("mkdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	fit := fs.String("fit", "ff", "Ajuste")
	unit := fs.String("unit", "m", "Unidad")
	path := fs.String("path", "", "Ruta")

	// Parse flag
	fs.Parse(os.Args[1:])

	// Find all the flags
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Get the flag name: size, fit, unit, or path
		flagValue := strings.ToLower(match[2]) // match[2]: Get the flag value in lowercase

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// ====== Check the flags ======

	// Check the size: positive and greater than 0
	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	// Check the fit: bf, ff, or wf
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	//If fit is empty, set it to "ff"
	if *fit == "" {
		*fit = "ff"
	}

	// Check the unit: k or m
	if *unit != "k" && *unit != "m" {
		fmt.Println("Error: Unit must be 'k' or 'm'")
		return
	}

	//If unit is empty, set it to "m"
	if *unit == "" {
		*unit = "m"
	}

	// Check the path: not empty
	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	DiskControl.Mkdisk(*size, *fit, *unit, *path)
}

// Function to remove a disk
func fn_rmdisk(params string) {
	// Define flag
	fs := flag.NewFlagSet("rmdisk", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")

	// Parse flag
	fs.Parse(os.Args[1:])

	// Find all the flags
	matches := re.FindAllStringSubmatch(params, -1)

	// Process the input
	for _, match := range matches {
		flagName := match[1]                   // match[1]: Get the flag name: path
		flagValue := strings.ToLower(match[2]) // match[2]: Get the flag value in lowercase

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "path":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	// Check the path: not empty
	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	DiskControl.Rmdisk(*path)
}

func fn_fdisk(input string) {
	fs := flag.NewFlagSet("fdisk", flag.ExitOnError)
	size := fs.Int("size", 0, "Tamaño")
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre")
	unit := fs.String("unit", "k", "Unidad")
	type_ := fs.String("type", "p", "Tipo")
	fit := fs.String("fit", "wf", "Ajuste")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "size", "fit", "unit", "path", "name", "type":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
		}
	}

	if *size <= 0 {
		fmt.Println("Error: Size must be greater than 0")
		return
	}

	if *path == "" {
		fmt.Println("Error: Path is required")
		return
	}

	// If fit is empty, set it to "w"
	if *fit == "" {
		*fit = "wf"
	}

	//Fit must be 'bf', 'ff', or 'ww'
	if *fit != "bf" && *fit != "ff" && *fit != "wf" {
		fmt.Println("Error: Fit must be 'bf', 'ff', or 'wf'")
		return
	}

	// If unit is empty, set it to "k"
	if *unit == "" {
		*unit = "k"
	}

	//Unit must be 'k', 'm'or 'b'
	if *unit != "k" && *unit != "m" && *unit != "b" {
		fmt.Println("Error: Unit must be 'b', 'm' or 'k'")
		return
	}

	// If type is empty, set it to "p"
	if *type_ == "" {
		*type_ = "p"
	}

	// Type must be 'p', 'e', or 'l'
	if *type_ != "p" && *type_ != "e" && *type_ != "l" {
		fmt.Println("Error: Type must be 'p', 'e', or 'l'")
		return
	}

	// Call the function
	DiskControl.Fdisk(*size, *path, *name, *unit, *type_, *fit)
}

func fn_mount(params string) {
	fs := flag.NewFlagSet("mount", flag.ExitOnError)
	path := fs.String("path", "", "Ruta")
	name := fs.String("name", "", "Nombre de la partición")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(params, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := strings.ToLower(match[2])
		flagValue = strings.Trim(flagValue, "\"")
		fs.Set(flagName, flagValue)
	}

	if *path == "" || *name == "" {
		fmt.Println("Error: Path y Name son obligatorios")
		return
	}

	// Convertir el nombre a minúsculas antes de pasarlo al Mount
	lowercaseName := strings.ToLower(*name)
	DiskControl.Mount(*path, lowercaseName)
}

func Fn_Rep(input string) string {
	fmt.Println("======Start REP======")
	fs := flag.NewFlagSet("rep", flag.ExitOnError)
	name := fs.String("name", "", "Nombre del reporte a generar (mbr, disk, inode, block, bm_inode, bm_block, sb, file, ls)")
	path := fs.String("path", "", "Ruta donde se generará el reporte")
	id := fs.String("id", "", "ID de la partición")
	pathFileLs := fs.String("path_file_ls", "", "Nombre del archivo o carpeta para reportes file o ls")

	matches := re.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		flagName := strings.ToLower(match[1])
		flagValue := strings.Trim(match[2], "\"")

		switch flagName {
		case "name", "path", "id", "path_file_ls":
			err := fs.Set(flagName, flagValue)
			if err != nil {
				fmt.Printf("Error al procesar el parámetro '%s': %v\n", flagName, err)
				return fmt.Sprintf("Error al procesar el parámetro '%s': %v", flagName, err)
			}
		default:
			fmt.Println("Error: Flag no encontrada:", flagName)
			fmt.Println("======FIN REP======")
			return "Error al procesar el parámetro " + flagName
		}
		fmt.Printf("Parámetros procesados: name=%s, path=%s, id=%s, path_file_ls=%s\n", *name, *path, *id, *pathFileLs)
	}

	// Name, path and id are required
	if *name == "" || *path == "" || *id == "" {
		fmt.Println("Error: 'name', 'path' y 'id' son parámetros obligatorios.")
		fmt.Println("======FIN REP======")
		return "Error: 'name', 'path' y 'id' son parámetros obligatorios."
	}

	// Parameter path_file_ls is required for file report
	if *name == "file" && *pathFileLs == "" {
		fmt.Println("Error: 'path_file_ls' es obligatorio para el reporte 'file'.")
		fmt.Println("======FIN REP======")
		return "Error: 'path_file_ls' es obligatorio para el reporte 'file'."
	}

	// ID must be valid
	if strings.Contains(*id, "x") || len(*id) < 3 {
		fmt.Printf("Error: El ID '%s' no es válido. Debe seguir el formato correcto.\n", *id)
		fmt.Println("======FIN REP======")
		return "El ID no es válido. Debe seguir el formato correcto."
	}

	// Verifying if the partition is mounted
	mounted := false
	var diskPath string
	for _, partitions := range DiskControl.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == *id {
				mounted = true
				diskPath = partition.Path
				break
			}
		}
	}

	if !mounted {
		fmt.Println("Error: La partición con ID", *id, "no está montada.")
		fmt.Println("======FIN REP======")
		return "La partición con ID " + *id + " no está montada."
	}

	// Creating the reports directory if it doesn't exist
	reportsDir := filepath.Dir(*path)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error al crear la carpeta:", reportsDir)
		fmt.Println("======FIN REP======")
		return "Error al crear la carpeta: " + reportsDir
	}

	switch *name {

	// ===== MBR REPORT =====
	case "mbr":
		// Create the dir if it doesnt exist
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755) // Dir with 755 permission
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}

		// Open bin file
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo"
		}

		// Content of dot file
		var dot bytes.Buffer
		fmt.Fprintln(&dot, "digraph G {")
		fmt.Fprintln(&dot, "node [shape=plaintext];")
		fmt.Fprintln(&dot, "fontname=\"Courier New\";")
		fmt.Fprintln(&dot, "mbrTable [label=<")
		fmt.Fprintln(&dot, "<table border='1' cellborder='1' cellspacing='0'>")
		fmt.Fprintln(&dot, "<tr><td bgcolor=\"Mediumslateblue\" colspan='2'>MBR</td></tr>")
		fmt.Fprintf(&dot, "<tr><td>Tamaño</td><td>%d</td></tr>\n", TempMBR.MbrSize)
		fmt.Fprintf(&dot, "<tr><td>Fecha De Creación</td><td>%s</td></tr>\n", string(TempMBR.CreationDate[:]))
		fmt.Fprintf(&dot, "<tr><td>Ajuste</td><td>%s</td></tr>\n", string(TempMBR.Fit[:]))
		fmt.Fprintf(&dot, "<tr><td>Signature</td><td>%d</td></tr>\n", TempMBR.Signature)

		// Add details of each partition
		for i, Particion := range TempMBR.Partitions {
			if Particion.Size != 0 {
				fmt.Fprintf(&dot, "<tr><td colspan='2' bgcolor='Slategray'>Partición %d</td></tr>\n", i+1)
				fmt.Fprintf(&dot, "<tr><td>Estado</td><td>%s</td></tr>\n", string(Particion.Status[:]))
				fmt.Fprintf(&dot, "<tr><td>Tipo</td><td>%s</td></tr>\n", string(Particion.Type[:]))
				fmt.Fprintf(&dot, "<tr><td>Ajuste</td><td>%s</td></tr>\n", string(Particion.Fit[:]))
				fmt.Fprintf(&dot, "<tr><td>Inicio</td><td>%d</td></tr>\n", Particion.Start)
				fmt.Fprintf(&dot, "<tr><td>Tamaño</td><td>%d</td></tr>\n", Particion.Size)
				fmt.Fprintf(&dot, "<tr><td>Nombre</td><td>%s</td></tr>\n", strings.Trim(string(Particion.Name[:]), "\x00"))

				// If it's an extended partition, read the EBRs
				if Particion.Type[0] == 'e' {
					var EBR DiskStruct.EBR
					if err := FileManagement.ReadObject(file, &EBR, int64(Particion.Start)); err != nil {
						fmt.Println("Error al leer el EBR desde el archivo")
						fmt.Println("======FIN REP======")
						return "Error al leer el EBR desde el archivo"
					}
					if EBR.PartSize != 0 {
						fmt.Fprintln(&dot, "<tr><td colspan='2' bgcolor='Lightcoral'>EBRs</td></tr>")
						for {
							fmt.Fprintf(&dot, "<tr><td bgcolor='LightPink'>Nombre</td><td bgcolor='LightPink'>%s</td></tr>\n", strings.Trim(string(EBR.PartName[:]), "\x00"))
							fmt.Fprintf(&dot, "<tr><td>Tipo</td><td>%s</td></tr>\n", "l")
							fmt.Fprintf(&dot, "<tr><td>Ajuste</td><td>%c</td></tr>\n", EBR.PartFit)
							fmt.Fprintf(&dot, "<tr><td>Inicio</td><td>%d</td></tr>\n", EBR.PartStart)
							fmt.Fprintf(&dot, "<tr><td>Tamaño</td><td>%d</td></tr>\n", EBR.PartSize)
							fmt.Fprintf(&dot, "<tr><td>Siguiente</td><td>%d</td></tr>\n", EBR.PartNext)

							if EBR.PartNext == -1 {
								break
							}
							if err := FileManagement.ReadObject(file, &EBR, int64(EBR.PartNext)); err != nil {
								fmt.Println("Error al leer el siguiente EBR")
								fmt.Println("======FIN REP======")
								return "Error al leer el siguiente EBR"
							}
						}
					}
				}
			}
		}

		fmt.Fprintln(&dot, "</table>")
		fmt.Fprintln(&dot, ">];")
		fmt.Fprintln(&dot, "}")

		// Save dot file
		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = os.WriteFile(dotFilePath, dot.Bytes(), 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo DOT.")
			fmt.Println("======FIN REP======")
			return "Error al escribir el archivo DOT."
		}

		// Generate the report
		cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", *path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al ejecutar Graphviz.")
			fmt.Println("======FIN REP======")
			return "Error al ejecutar Graphviz"
		}

		fmt.Printf("Reporte MBR generado con éxito en la ruta: %s\n", *path)
		fmt.Println("======FIN REP======")

	// ===== DISK REPORT =====
	case "disk":
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755) // Crear el directorio con permisos 0755
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}
		// Open the binary file of the mounted disk
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Println("Error: No se pudo abrir el archivo en la ruta:", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR object from the binary file
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo"
		}

		// Read and process the EBRs if there are extended partitions
		var ebrs []DiskStruct.EBR
		for i := 0; i < 4; i++ {
			if string(TempMBR.Partitions[i].Type[:]) == "e" { // Partición extendida
				ebrPosition := TempMBR.Partitions[i].Start
				for ebrPosition != -1 {
					var tempEBR DiskStruct.EBR
					if err := FileManagement.ReadObject(file, &tempEBR, int64(ebrPosition)); err != nil {
						break
					}
					ebrs = append(ebrs, tempEBR)   // Add the EBR to the slice
					ebrPosition = tempEBR.PartNext // Move to the next EBR
				}
			}
		}

		// Calculate the total disk size
		totalDiskSize := TempMBR.MbrSize

		// Generates the .dot file
		reportPath := *path
		if err := FileManagement.GenerateDiskReport(TempMBR, ebrs, reportPath, file, totalDiskSize); err != nil {
			fmt.Println("Error al generar el reporte DISK:", err)
			fmt.Println("======FIN REP======")
		} else {
			fmt.Println("Reporte DISK generado exitosamente en:", reportPath)

			dotFile := strings.TrimSuffix(reportPath, filepath.Ext(reportPath)) + ".dot"
			outputJpg := reportPath
			cmd := exec.Command("dot", "-Tjpg", dotFile, "-o", outputJpg)
			err = cmd.Run()
			if err != nil {
				fmt.Println("Error al renderizar el archivo .dot a imagen:", err)
				fmt.Println("======FIN REP======")
			} else {
				fmt.Println("Imagen generada exitosamente en:", outputJpg)
			}
		}
	case "bm_inode":
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}

		// Verify if the partition is mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open the binary file of the mounted disk
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR object from the bin file
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Find the partition with the given ID
		var index int = -1
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Size != 0 {
				if strings.Contains(string(TempMBR.Partitions[i].Id[:]), *id) {
					if TempMBR.Partitions[i].Status[0] == '1' {
						index = i
					} else {
						fmt.Printf("La partición con el ID:%s no está montada.\n", *id)
						fmt.Println("======FIN REP======")
						return "Error: La partición con el ID:" + *id + " no está montada."
					}
					break
				}
			}
		}

		if index == -1 {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Read the SuperBlock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[index].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Check the values of the SuperBlock
		if TemporalSuperBloque.S_inodes_count <= 0 || TemporalSuperBloque.S_bm_inode_start <= 0 {
			fmt.Println("Valores inválidos en el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: Valores inválidos en el SuperBloque."
		}

		// Read the bitmap of inodes
		BitMapInode := make([]byte, TemporalSuperBloque.S_inodes_count)
		if _, err := file.ReadAt(BitMapInode, int64(TemporalSuperBloque.S_bm_inode_start)); err != nil {
			fmt.Println("No se pudo leer el bitmap de inodos:", err)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el bitmap de inodos."
		}

		// Create the report file
		SalidaArchivo, err := os.Create(*path)
		if err != nil {
			fmt.Println("No se pudo crear el archivo de reporte:", err)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo crear el archivo de reporte."
		}

		// Close the file
		defer SalidaArchivo.Close()

		// Write the bitmap of inodes to the report file
		for i, bit := range BitMapInode {
			if bit != 0 && bit != 1 {
				fmt.Printf("Advertencia: Valor inesperado en el bitmap de inodos: %d\n", bit)
				fmt.Println("======FIN REP======")
				continue
			}
			if i > 0 && i%20 == 0 {
				fmt.Fprintln(SalidaArchivo)
			}
			fmt.Fprintf(SalidaArchivo, "%d ", bit)
		}

		fmt.Printf("Reporte de BITMAP INODE de la partición:%s generado con éxito en la ruta: %s\n", *id, *path)
		fmt.Println("======FIN REP======")

	case "bm_block":
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}

		// Verify if the partition is mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open bin file of the mounted disk
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR object from the bin file
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Find the partition with the given ID
		var index int = -1
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Size != 0 {
				if strings.Contains(string(TempMBR.Partitions[i].Id[:]), *id) {
					if TempMBR.Partitions[i].Status[0] == '1' {
						index = i
					} else {
						fmt.Printf("La partición con el ID:%s no está montada.\n", *id)
						fmt.Println("======FIN REP======")
						return "Error: La partición con el ID:" + *id + " no está montada."
					}
					break
				}
			}
		}

		if index == -1 {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Read the SuperBlock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[index].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Read the bitmap of blocks
		BitMapBlock := make([]byte, TemporalSuperBloque.S_blocks_count)
		if _, err := file.ReadAt(BitMapBlock, int64(TemporalSuperBloque.S_bm_block_start)); err != nil {
			fmt.Println("No se pudo leer el bitmap de bloques:", err)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el bitmap de bloques."
		}

		// Create the report file
		SalidaArchivo, err := os.Create(*path)
		if err != nil {
			fmt.Println("No se pudo crear el archivo de reporte:", err)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo crear el archivo de reporte."
		}
		defer SalidaArchivo.Close()

		// Write the bitmap of blocks to the report file
		for i, bit := range BitMapBlock {
			if i > 0 && i%20 == 0 {
				fmt.Fprintln(SalidaArchivo)
			}
			fmt.Fprintf(SalidaArchivo, "%d ", bit)
		}

		fmt.Printf("Reporte de la partición:%s generado con éxito en la ruta: %s\n", *id, *path)
		fmt.Println("======FIN REP======")

	case "inode":
		// Verify if the partition is mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open bin file
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// read the MBR
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Find the partition with the given ID
		var index int = -1
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Size != 0 {
				if strings.Contains(string(TempMBR.Partitions[i].Id[:]), *id) {
					if TempMBR.Partitions[i].Status[0] == '1' {
						index = i
					} else {
						fmt.Printf("La partición con el ID:%s no está montada.\n", *id)
						fmt.Println("======FIN REP======")
						return "Error: La partición con el ID:" + *id + " no está montada."
					}
					break
				}
			}
		}

		if index == -1 {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Read the SuperBlock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[index].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Content of the dot file
		var dot bytes.Buffer
		fmt.Fprintln(&dot, "digraph G {")
		fmt.Fprintln(&dot, "node [shape=none];")
		fmt.Fprintln(&dot, "fontname=\"Courier New\";")

		// Read the inodes
		for i := 0; i < int(TemporalSuperBloque.S_inodes_count); i++ {
			var inode DiskStruct.Inode
			offset := int64(TemporalSuperBloque.S_inode_start) + int64(i)*int64(TemporalSuperBloque.S_inode_size)
			if err := FileManagement.ReadObject(file, &inode, offset); err != nil {
				fmt.Println("Error al leer el inodo:", err)
				fmt.Println("======FIN REP======")
				continue
			}

			// Add the inode to the DOT file
			if inode.I_size > 0 {
				fmt.Fprintf(&dot, "inode%d [label=<\n", i)
				fmt.Fprintf(&dot, "<table border='0' cellborder='1' cellspacing='0' cellpadding='10'>\n")
				fmt.Fprintf(&dot, "<tr><td colspan='2' bgcolor='lightgreen'>Inode %d</td></tr>\n", i)
				fmt.Fprintf(&dot, "<tr><td>UID</td><td>%d</td></tr>\n", inode.I_uid)
				fmt.Fprintf(&dot, "<tr><td>GID</td><td>%d</td></tr>\n", inode.I_gid)
				fmt.Fprintf(&dot, "<tr><td>Size</td><td>%d</td></tr>\n", inode.I_size)
				fmt.Fprintf(&dot, "<tr><td>ATime</td><td>%s</td></tr>\n", string(inode.I_atime[:]))
				fmt.Fprintf(&dot, "<tr><td>CTime</td><td>%s</td></tr>\n", string(inode.I_ctime[:]))
				fmt.Fprintf(&dot, "<tr><td>MTime</td><td>%s</td></tr>\n", string(inode.I_mtime[:]))
				fmt.Fprintf(&dot, "<tr><td>Blocks</td><td>%v</td></tr>\n", inode.I_block)
				fmt.Fprintf(&dot, "<tr><td>Perms</td><td>%s</td></tr>\n", string(inode.I_perm[:]))
				fmt.Fprintf(&dot, "</table>\n")
				fmt.Fprintf(&dot, " >];\n")
			}
		}
		fmt.Fprintln(&dot, "}")

		// Save the dot file
		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = os.WriteFile(dotFilePath, dot.Bytes(), 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo DOT.")
			fmt.Println("======FIN REP======")
			return "Error al escribir el archivo DOT."
		}

		// Create the directory if it doesn't exist
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}

		// Generate the image
		cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", *path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al ejecutar Graphviz.")
			fmt.Println("======FIN REP======")
			return "Error al ejecutar Graphviz"
		}

		fmt.Printf("Reporte de INODE de la partición:%s generado con éxito en la ruta: %s\n", *id, *path)
		fmt.Println("======FIN REP======")

	case "block":
		dir := filepath.Dir(*path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("Error al crear el directorio: %v\n", err)
				fmt.Println("======FIN REP======")
				return "Error al crear el directorio: " + dir
			}
		}

		// Verify if the partition is mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open bin file
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Find the partition with the given ID
		var index int = -1
		for i := 0; i < 4; i++ {
			if TempMBR.Partitions[i].Size != 0 {
				if strings.Contains(string(TempMBR.Partitions[i].Id[:]), *id) {
					if TempMBR.Partitions[i].Status[0] == '1' {
						index = i
					} else {
						fmt.Printf("La partición con el ID:%s no está montada.\n", *id)
						fmt.Println("======FIN REP======")
						return "Error: La partición con el ID:" + *id + " no está montada."
					}
					break
				}
			}
		}

		if index == -1 {
			fmt.Printf("No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Read the superblock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[index].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Read the bitmap of blocks
		BitMapBlock := make([]byte, TemporalSuperBloque.S_blocks_count)
		if _, err := file.ReadAt(BitMapBlock, int64(TemporalSuperBloque.S_bm_block_start)); err != nil {
			fmt.Println("No se pudo leer el bitmap de bloques:", err)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el bitmap de bloques."
		}

		// Create the report file
		totalBlocks := len(BitMapBlock)
		usedBlocks := 0
		unusedBlocks := 0

		// Create the dot file
		var dot bytes.Buffer
		fmt.Fprintln(&dot, "digraph G {")
		fmt.Fprintln(&dot, "node [shape=plaintext];")
		fmt.Fprintln(&dot, "fontname=\"Courier New\";")

		fmt.Fprintln(&dot, "blockTable [label=<")
		fmt.Fprintln(&dot, "<table border='1' cellborder='1' cellspacing='0'>")
		fmt.Fprintln(&dot, "<tr><td bgcolor=\"LightBlue\" colspan='2'>BLOCKS</td></tr>")

		// Read the used blocks
		for i, bit := range BitMapBlock {
			if bit == 1 { // Show only used blocks
				usedBlocks++ // Increment the counter of used blocks

				var block DiskStruct.Fileblock
				offset := int64(TemporalSuperBloque.S_block_start) + int64(i)*int64(TemporalSuperBloque.S_block_size)
				fmt.Printf("Leyendo bloque %d en offset %d\n", i, offset) // Depuración: Verificar el bloque y el offset

				if err := FileManagement.ReadObject(file, &block, offset); err != nil {
					fmt.Printf("Error al leer el bloque %d: %v\n", i, err)
					continue
				}

				fmt.Printf("Contenido crudo del bloque %d: %v\n", i, block.B_content)

				blockContent := strings.Trim(string(block.B_content[:]), "\x00") // Delete null characters
				if blockContent == "" {                                          // If the block is empty, skip it
					continue
				}

				cleanedContent := ""
				for _, char := range blockContent {
					if char >= 32 && char <= 126 {
						cleanedContent += string(char)
					}
				}
				cleanedContent = strings.TrimLeft(cleanedContent, ".")
				cleanedContent = strings.ReplaceAll(cleanedContent, "&", "&amp;")
				cleanedContent = strings.ReplaceAll(cleanedContent, "<", "&lt;")
				cleanedContent = strings.ReplaceAll(cleanedContent, ">", "&gt;")

				fmt.Fprintf(&dot, "<tr><td>Bloque %d</td><td>%s</td></tr>\n", i, cleanedContent)
			} else {
				unusedBlocks++
			}
		}

		fmt.Fprintln(&dot, "</table>")
		fmt.Fprintln(&dot, ">];")

		fmt.Fprintln(&dot, "summaryTable [label=<")
		fmt.Fprintln(&dot, "<table border='1' cellborder='1' cellspacing='0'>")
		fmt.Fprintln(&dot, "<tr><td bgcolor=\"LightGreen\" colspan='2'>Resumen de Bloques</td></tr>")
		fmt.Fprintf(&dot, "<tr><td>Total de Bloques</td><td>%d</td></tr>\n", totalBlocks)
		fmt.Fprintf(&dot, "<tr><td>Bloques Utilizados</td><td>%d</td></tr>\n", usedBlocks)
		fmt.Fprintf(&dot, "<tr><td>Bloques No Utilizados</td><td>%d</td></tr>\n", unusedBlocks)
		fmt.Fprintln(&dot, "</table>")
		fmt.Fprintln(&dot, ">];")

		fmt.Fprintln(&dot, "}")

		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = os.WriteFile(dotFilePath, dot.Bytes(), 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo DOT.")
			fmt.Println("======FIN REP======")
			return "Error al escribir el archivo DOT."
		}

		// Generate the report
		cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", *path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al ejecutar Graphviz.")
			fmt.Println("======FIN REP======")
			return "Error al ejecutar Graphviz"
		}

		fmt.Printf("Reporte de BLOQUES generado con éxito en la ruta: %s\n", *path)
		fmt.Println("======FIN REP======")

	case "sb":
		GenerateSuperblockReport(*id, *path)
		fmt.Println("======FIN REP======")

		// ===== FILE -LS REPORT =====
	case "file":
		// Param path_file_ls is required
		if *pathFileLs == "" {
			fmt.Println("Error: 'path_file_ls' es obligatorio para el reporte 'file'.")
			fmt.Println("======FIN REP======")
			return "Error: 'path_file_ls' es obligatorio para el reporte 'file'."
		}

		// Verify if the partition is mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("Error: No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open bin file
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("Error: No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Read the SuperBlock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[0].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Find the file of path_file_ls
		indexInode := UserManagement.InitSearch(*pathFileLs, file, TemporalSuperBloque)
		if indexInode == -1 {
			fmt.Printf("Error: No se encontró el archivo especificado en el path: %s.\n", *pathFileLs)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró el archivo especificado en el path: " + *pathFileLs
		}

		var crrInode DiskStruct.Inode
		if err := FileManagement.ReadObject(file, &crrInode, int64(TemporalSuperBloque.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
			fmt.Println("Error al leer el inodo del archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el inodo del archivo."
		}

		// Get content of the file
		fileContent := UserManagement.GetInodeFileData(crrInode, file, TemporalSuperBloque)
		fmt.Printf("Contenido leído del archivo: %s\n", fileContent)

		// Content of the report
		reportContent := fmt.Sprintf("Reporte FILE\n\nNombre del archivo: %s\nContenido:\n%s", *pathFileLs, fileContent)

		// Create the report
		err = os.WriteFile(*path, []byte(reportContent), 0644)
		if err != nil {
			fmt.Printf("Error al escribir el archivo de reporte: %s\n", err)
			fmt.Println("======FIN REP======")
			return "Error al escribir el archivo de reporte: " + err.Error()
		}

		fmt.Printf("Reporte FILE generado exitosamente en la ruta: %s\n", *path)
		fmt.Println("======FIN REP======")

	case "ls":
		// Parameter path_file_ls is required
		if *pathFileLs == "" {
			fmt.Println("Error: 'path_file_ls' es obligatorio para el reporte 'ls'.")
			fmt.Println("======FIN REP======")
			return "Error: 'path_file_ls' es obligatorio para el reporte 'ls'."
		}

		// Partition must be mounted
		mounted := false
		var diskPath string
		for _, partitions := range DiskControl.GetMountedPartitions() {
			for _, partition := range partitions {
				if partition.ID == *id {
					mounted = true
					diskPath = partition.Path
					break
				}
			}
			if mounted {
				break
			}
		}

		if !mounted {
			fmt.Printf("Error: No se encontró la partición con el ID: %s.\n", *id)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró la partición con el ID: " + *id
		}

		// Open bin file
		file, err := FileManagement.OpenFile(diskPath)
		if err != nil {
			fmt.Printf("Error: No se pudo abrir el archivo en la ruta: %s\n", diskPath)
			fmt.Println("======FIN REP======")
			return "Error: No se pudo abrir el archivo en la ruta: " + diskPath
		}
		defer file.Close()

		// Read the MBR
		var TempMBR DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
			fmt.Println("Error: No se pudo leer el MBR desde el archivo.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el MBR desde el archivo."
		}

		// Read the Superblock
		var TemporalSuperBloque DiskStruct.Superblock
		if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[0].Start)); err != nil {
			fmt.Println("Error al leer el SuperBloque.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el SuperBloque."
		}

		// Find the inode of path_file_ls
		indexInode := UserManagement.InitSearch(*pathFileLs, file, TemporalSuperBloque)
		if indexInode == -1 {
			fmt.Printf("Error: No se encontró el directorio especificado en el path: %s.\n", *pathFileLs)
			fmt.Println("======FIN REP======")
			return "Error: No se encontró el directorio especificado en el path: " + *pathFileLs
		}

		var crrInode DiskStruct.Inode
		if err := FileManagement.ReadObject(file, &crrInode, int64(TemporalSuperBloque.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
			fmt.Println("Error al leer el inodo del directorio.")
			fmt.Println("======FIN REP======")
			return "Error: No se pudo leer el inodo del directorio."
		}

		// Content of dot file
		var dot bytes.Buffer
		fmt.Fprintln(&dot, "digraph G {")
		fmt.Fprintln(&dot, "node [shape=plaintext];")
		fmt.Fprintln(&dot, "fontname=\"Courier New\";")
		fmt.Fprintln(&dot, "lsTable [label=<")
		fmt.Fprintln(&dot, "<table border='1' cellborder='1' cellspacing='0'>")
		fmt.Fprintln(&dot, "<tr><td bgcolor=\"PowderBlue\" colspan='7'>Reporte LS</td></tr>")
		fmt.Fprintln(&dot, "<tr><td>Permisos</td><td>Propietario</td><td>Grupo</td><td>Tamaño (en bytes) </td><td>Fecha Modificación</td><td>Nombre</td></tr>")

		// Read the blocks of files and dir
		for _, block := range crrInode.I_block {
			if block != -1 {
				var folderBlock DiskStruct.Folderblock
				if err := FileManagement.ReadObject(file, &folderBlock, int64(TemporalSuperBloque.S_block_start+block*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
					fmt.Println("Error al leer el bloque de carpeta:", err)
					continue
				}

				for _, folder := range folderBlock.B_content {
					if folder.B_inodo != -1 {
						var inode DiskStruct.Inode
						if err := FileManagement.ReadObject(file, &inode, int64(TemporalSuperBloque.S_inode_start+folder.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
							fmt.Println("Error al leer el inodo:", err)
							continue
						}

						nombre := strings.Trim(string(folder.B_name[:]), "\x00")
						if nombre == "" { // Name is empty
							continue
						}

						// Get the details of files and directories
						propietario := GetUserNameByID(int(inode.I_uid))
						grupo := GetGroupNameByID(int(inode.I_gid))

						permisos := string(inode.I_perm[:])
						tamaño := inode.I_size
						fechaMod := string(inode.I_mtime[:])

						fmt.Fprintf(&dot, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td><td>%s</td></tr>\n",
							permisos, propietario, grupo, tamaño, fechaMod, nombre)
					}
				}
			}
		}

		fmt.Fprintln(&dot, "</table>")
		fmt.Fprintln(&dot, ">];")
		fmt.Fprintln(&dot, "}")

		dotFilePath := strings.TrimSuffix(*path, filepath.Ext(*path)) + ".dot"
		err = os.WriteFile(dotFilePath, dot.Bytes(), 0644)
		if err != nil {
			fmt.Println("Error al escribir el archivo DOT.")
			fmt.Println("======FIN REP======")
			return "Error al escribir el archivo DOT."
		}

		// Generate the report
		cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", *path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error al ejecutar Graphviz.")
			fmt.Println("======FIN REP======")
			return "Error al ejecutar Graphviz"
		}

		fmt.Printf("Reporte LS generado con éxito en la ruta: %s\n", *path)
		fmt.Println("======FIN REP======")

	default:
		fmt.Println("Error: Tipo de reporte no válido.")
		fmt.Println("======FIN REP======")
	}
	fmt.Println("======FIN REP======")
	return "Reporte generado con éxito en la ruta: " + *path
}

func GetUserNameByID(userID int) string {
	if userID == 1 {
		return "root"
	}
	return "unknown"
}

func GetGroupNameByID(groupID int) string {
	if groupID == 1 {
		return "root"
	}
	return "unknown"
}

func fn_login(input string) {
	fmt.Println("======Start LOGIN======")
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	user := fs.String("user", "", "Usuario")
	pass := fs.String("pass", "", "Contraseña")
	id := fs.String("id", "", "Id")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "id":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag not found")
			fmt.Println("======FIN LOGIN======")
		}
	}

	UserManagement.Login(*user, *pass, *id)

}

func GenerateSuperblockReport(id string, path string) {
	// Find the mounted partition
	mounted := false
	var diskPath string
	for _, partitions := range DiskControl.GetMountedPartitions() {
		for _, partition := range partitions {
			if partition.ID == id {
				mounted = true
				diskPath = partition.Path
				break
			}
		}
		if mounted {
			break
		}
	}

	if !mounted {
		fmt.Printf("Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	// Open bin file of the mounted disk
	file, err := FileManagement.OpenFile(diskPath)
	if err != nil {
		fmt.Printf("Error REP SB: No se pudo abrir el archivo en la ruta: %s\n", diskPath)
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		fmt.Println("Error REP SB: No se pudo leer el MBR desde el archivo.")
		return
	}

	// Find the partition with the given ID
	var index int = -1
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Size != 0 {
			if strings.Contains(string(TempMBR.Partitions[i].Id[:]), id) {
				if TempMBR.Partitions[i].Status[0] == '1' {
					index = i
				} else {
					fmt.Printf("Error REP SB: La partición con el ID:%s no está montada.\n", id)
					return
				}
				break
			}
		}
	}

	if index == -1 {
		fmt.Printf("Error REP SB: No se encontró la partición con el ID: %s.\n", id)
		return
	}

	// Read the SuperBlock
	var TemporalSuperBloque DiskStruct.Superblock
	if err := FileManagement.ReadObject(file, &TemporalSuperBloque, int64(TempMBR.Partitions[index].Start)); err != nil {
		fmt.Println("Error REP SB: Error al leer el SuperBloque.")
		return
	}

	// Content of the dot file
	var dot bytes.Buffer
	fmt.Fprintln(&dot, "digraph G {")
	fmt.Fprintln(&dot, "node [shape=plaintext];")
	fmt.Fprintln(&dot, "fontname=\"Courier New\";")
	fmt.Fprintln(&dot, "SBTable [label=<")
	fmt.Fprintln(&dot, "<table border='1' cellborder='1' cellspacing='0'>")
	fmt.Fprintln(&dot, "<tr><td bgcolor=\"RosyBrown\" colspan='2'>Super Bloque</td></tr>")
	fmt.Fprintf(&dot, "<tr><td>S_filesystem_type</td><td>%d</td></tr>\n", TemporalSuperBloque.S_filesystem_type)
	fmt.Fprintf(&dot, "<tr><td>S_inodes_count</td><td>%d</td></tr>\n", TemporalSuperBloque.S_inodes_count)
	fmt.Fprintf(&dot, "<tr><td>S_blocks_count</td><td>%d</td></tr>\n", TemporalSuperBloque.S_blocks_count)
	fmt.Fprintf(&dot, "<tr><td>S_free_blocks_count</td><td>%d</td></tr>\n", TemporalSuperBloque.S_free_blocks_count)
	fmt.Fprintf(&dot, "<tr><td>S_free_inodes_count</td><td>%d</td></tr>\n", TemporalSuperBloque.S_free_inodes_count)
	fmt.Fprintf(&dot, "<tr><td>S_mtime</td><td>%s</td></tr>\n", string(TemporalSuperBloque.S_mtime[:]))
	fmt.Fprintf(&dot, "<tr><td>S_umtime</td><td>%s</td></tr>\n", string(TemporalSuperBloque.S_umtime[:]))
	fmt.Fprintf(&dot, "<tr><td>S_mnt_count</td><td>%d</td></tr>\n", TemporalSuperBloque.S_mnt_count)
	fmt.Fprintf(&dot, "<tr><td>S_magic</td><td>0x%X</td></tr>\n", TemporalSuperBloque.S_magic)
	fmt.Fprintf(&dot, "<tr><td>S_inode_size</td><td>%d</td></tr>\n", TemporalSuperBloque.S_inode_size)
	fmt.Fprintf(&dot, "<tr><td>S_block_size</td><td>%d</td></tr>\n", TemporalSuperBloque.S_block_size)
	fmt.Fprintf(&dot, "<tr><td>S_fist_ino</td><td>%d</td></tr>\n", TemporalSuperBloque.S_fist_ino)
	fmt.Fprintf(&dot, "<tr><td>S_first_blo</td><td>%d</td></tr>\n", TemporalSuperBloque.S_first_blo)
	fmt.Fprintf(&dot, "<tr><td>S_bm_inode_start</td><td>%d</td></tr>\n", TemporalSuperBloque.S_bm_inode_start)
	fmt.Fprintf(&dot, "<tr><td>S_bm_block_start</td><td>%d</td></tr>\n", TemporalSuperBloque.S_bm_block_start)
	fmt.Fprintf(&dot, "<tr><td>S_inode_start</td><td>%d</td></tr>\n", TemporalSuperBloque.S_inode_start)
	fmt.Fprintf(&dot, "<tr><td>S_block_start</td><td>%d</td></tr>\n", TemporalSuperBloque.S_block_start)
	fmt.Fprintln(&dot, "</table>")
	fmt.Fprintln(&dot, ">];")
	fmt.Fprintln(&dot, "}")

	// Create the dot file
	dotFilePath := strings.TrimSuffix(path, filepath.Ext(path)) + ".dot"
	err = os.WriteFile(dotFilePath, dot.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error REP SB: Error al escribir el archivo DOT.")
		return
	}

	// New folder if it doesn't exist
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println("Error REP SB: Error al crear el directorio.")
			return
		}
	}

	// Generate the image
	cmd := exec.Command("dot", "-Tjpg", dotFilePath, "-o", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error REP SB: Error al ejecutar Graphviz.")
		return
	}

	fmt.Printf("Reporte de SB de la partición:%s generado con éxito en la ruta: %s\n", id, path)
}

func fn_mkusr(input string) {
	fmt.Println("======Start MKUSR======")
	fs := flag.NewFlagSet("mkusr", flag.ExitOnError)
	user := fs.String("user", "", "Nombre del usuario")
	pass := fs.String("pass", "", "Contraseña del usuario")
	grp := fs.String("grp", "", "Grupo al que pertenece el usuario")

	fs.Parse(os.Args[1:])
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		flagName := match[1]
		flagValue := match[2]

		flagValue = strings.Trim(flagValue, "\"")

		switch flagName {
		case "user", "pass", "grp":
			fs.Set(flagName, flagValue)
		default:
			fmt.Println("Error: Flag no encontrada:", flagName)
			fmt.Println("======FIN MKUSR======")
			return
		}
	}

	// Validate that the parameters 'user', 'pass' and 'grp' are not empty
	if *user == "" || *pass == "" || *grp == "" {
		fmt.Println("Error: Los parámetros 'user', 'pass' y 'grp' son obligatorios.")
		fmt.Println("======FIN MKUSR======")
		return
	}

	// Values cannot exceed 10 characters
	if len(*user) > 10 || len(*pass) > 10 || len(*grp) > 10 {
		fmt.Println("Error: Los valores de 'user', 'pass' y 'grp' no pueden exceder los 10 caracteres.")
		fmt.Println("======FIN MKUSR======")
		return
	}

	// Call the function to create the user
	UserManagement.Mkusr(*user, *pass, *grp)
	fmt.Println("======FIN MKUSR======")
}
