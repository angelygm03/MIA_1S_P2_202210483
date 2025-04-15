package FileManagement

import (
	"Proyecto2/backend/DiskStruct"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Function to create a file
func CreateFile(name string) error {
	// Check if the file exists
	dir := filepath.Dir(name) // Get the directory
	fmt.Println("Intentando crear directorio:", dir)
	// if the directory does not exist, create it
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Println("Error creating file dir==", err)
		return err
	}

	// Create the file
	if _, err := os.Stat(name); os.IsNotExist(err) {
		file, err := os.Create(name)
		if err != nil {
			fmt.Println("Error creating file==", err)
			return err
		}
		defer file.Close()
	}
	return nil
}

// Function to open a file
func OpenFile(name string) (*os.File, error) {
	file, err := os.OpenFile(name, os.O_RDWR, 0644) // Permission 644: Read and write
	if err != nil {
		fmt.Println("Error opening file==", err)
		return nil, err
	}
	return file, nil
}

// Function to write an object to a file
func WriteObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0) // Move the pointer to the position
	err := binary.Write(file, binary.LittleEndian, data)
	if err != nil {
		fmt.Println("Error writing object==", err)
		return err
	}
	return nil
}

// Function to read an object from a file
func ReadObject(file *os.File, data interface{}, position int64) error {
	file.Seek(position, 0)
	err := binary.Read(file, binary.LittleEndian, data) // Read the object
	if err != nil {
		fmt.Println("Error reading object==", err)
		return err
	}
	return nil
}

// Función para generar el reporte del MBR y los EBRs en formato Graphviz y guardarlo en un archivo .dot
func GenerateMBRReport(mbr DiskStruct.MRB, ebrs []DiskStruct.EBR, outputPath string, file *os.File) error {
	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	// Crear el archivo .dot donde se generará el reporte
	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	// Iniciar el contenido del archivo en formato Graphviz (.dot)
	content := "digraph G {\n"
	content += "\tnode [fillcolor=lightyellow style=filled]\n"

	// Subgrafo del MBR
	content += fmt.Sprintf("\tsubgraph cluster_MBR {\n\t\tcolor=lightgrey fillcolor=lightblue label=\"MBR\nTamaño: %d\nFecha Creación: %s\nDisk Signature: %d\" style=filled\n",
		mbr.MbrSize, string(mbr.CreationDate[:]), mbr.Signature)

	// Recorrer las particiones del MBR en orden
	lastPartId := ""
	for i := 0; i < 4; i++ {
		part := mbr.Partitions[i]
		if part.Size > 0 { // Si la partición tiene un tamaño válido
			partName := strings.TrimRight(string(part.Name[:]), "\x00") // Limpiar el nombre de la partición
			partId := fmt.Sprintf("PART%d", i+1)
			content += fmt.Sprintf("\t\t%s [label=\"Partición %d\nStatus: %s\nType: %s\nFit: %s\nStart: %d\nSize: %d\nName: %s\" fillcolor=green shape=box style=filled]\n",
				partId, i+1, string(part.Status[:]), string(part.Type[:]), string(part.Fit[:]), part.Start, part.Size, partName)

			// Conectar la partición actual con la anterior de manera invisible para mantener el orden
			if lastPartId != "" {
				content += fmt.Sprintf("\t\t%s -> %s [style=invis]\n", lastPartId, partId)
			}
			lastPartId = partId

			// Si la partición es extendida, leer los EBRs
			if string(part.Type[:]) == "e" {
				content += fmt.Sprintf("\tsubgraph cluster_EBR%d {\n\t\tcolor=black fillcolor=lightpink label=\"Partición Extendida %d\" style=dashed\n", i+1, i+1)

				// Recolectamos todos los EBRs en orden
				ebrPos := part.Start
				var ebrList []DiskStruct.EBR
				for {
					var ebr DiskStruct.EBR
					err := ReadObject(file, &ebr, int64(ebrPos)) // Asegúrate de que la función ReadObject proviene de Utilities
					if err != nil {
						fmt.Println("Error al leer EBR:", err)
						break
					}
					ebrList = append(ebrList, ebr)

					// Si no hay más EBRs, salir del bucle
					if ebr.PartNext == -1 {
						break
					}

					// Mover a la siguiente posición de EBR
					ebrPos = ebr.PartNext
				}

				// Ahora agregamos los EBRs en orden correcto
				lastEbrId := ""
				for j, ebr := range ebrList {
					ebrName := strings.TrimRight(string(ebr.PartName[:]), "\x00") // Limpiar el nombre del EBR
					ebrId := fmt.Sprintf("EBR%d", j+1)
					content += fmt.Sprintf("\t\t%s [label=\"EBR\nStart: %d\nSize: %d\nNext: %d\nName: %s\" fillcolor=lightpink shape=box style=filled]\n",
						ebrId, ebr.PartStart, ebr.PartSize, ebr.PartNext, ebrName)

					// Conectar el EBR actual con el anterior de manera invisible para mantener el orden
					if lastEbrId != "" {
						content += fmt.Sprintf("\t\t%s -> %s [style=invis]\n", lastEbrId, ebrId)
					}
					lastEbrId = ebrId
				}

				content += "\t}\n" // Cerrar el subgrafo de la partición extendida
			}
		}
	}

	content += "\t}\n" // Cerrar el subgrafo del MBR

	content += "}\n" // Cerrar el grafo principal

	// Escribir el contenido en el archivo .dot
	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte MBR generado exitosamente en:", dotFilePath)
	return nil
}

// Función para generar el reporte DISK en formato .dot
func GenerateDiskReport(mbr DiskStruct.MRB, ebrs []DiskStruct.EBR, outputPath string, file *os.File, totalDiskSize int32) error {
	// Crear la carpeta si no existe
	reportsDir := filepath.Dir(outputPath)
	err := os.MkdirAll(reportsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error al crear la carpeta de reportes: %v", err)
	}

	// Crear el archivo .dot donde se generara el reporte
	dotFilePath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".dot"
	fileDot, err := os.Create(dotFilePath)
	if err != nil {
		return fmt.Errorf("Error al crear el archivo .dot de reporte: %v", err)
	}
	defer fileDot.Close()

	// Iniciar el contenido del archivo en formato Graphviz (.dot)
	content := "digraph G {\n"
	content += "\tnode [shape=none];\n"
	content += "\tgraph [splines=false];\n"
	content += "\tsubgraph cluster_disk {\n"
	content += "\t\tlabel=\"Reporte Disco.mia\";\n"
	content += "\t\tstyle=rounded;\n"
	content += "\t\tcolor=black;\n"

	// Iniciar tabla para las particiones
	content += "\t\ttable [label=<\n\t\t\t<TABLE BORDER=\"1\" CELLBORDER=\"1\" CELLSPACING=\"0\" CELLPADDING=\"10\">\n"
	content += "\t\t\t<TR>\n"
	content += "\t\t\t<TD>MBR (159 bytes)</TD>\n"

	// Variables para el porcentaje y espacio libre
	var usedSpace int32 = 159 // Tamaño del MBR en bytes
	var freeSpace int32 = totalDiskSize - usedSpace

	for i := 0; i < 4; i++ {
		part := mbr.Partitions[i]
		if part.Size > 0 { // Si la partición tiene un tamaño valido
			percentage := float64(part.Size) / float64(totalDiskSize) * 100
			partName := strings.TrimRight(string(part.Name[:]), "\x00") // Limpiar el nombre de la partición

			if string(part.Type[:]) == "p" { // Partición primaria
				content += fmt.Sprintf("\t\t\t<TD>Primaria<br/>%s<br/>%.2f%% del disco</TD>\n", partName, percentage)
				usedSpace += part.Size
			} else if string(part.Type[:]) == "e" { // Partición extendida
				content += "\t\t\t<TD>\n"
				content += "\t\t\t\t<TABLE BORDER=\"0\" CELLBORDER=\"1\" CELLSPACING=\"0\">\n"
				content += fmt.Sprintf("\t\t\t\t<TR><TD COLSPAN=\"5\">Extendida</TD></TR>\n")

				// Leer los EBRs y agregar las particiones lógicas
				content += "\t\t\t\t<TR>\n"
				for _, ebr := range ebrs {
					logicalPercentage := float64(ebr.PartSize) / float64(totalDiskSize) * 100
					content += fmt.Sprintf("\t\t\t\t<TD>EBR (32 bytes)</TD>\n\t\t\t\t<TD>Lógica<br/>%.2f%% del disco</TD>\n", logicalPercentage)
					usedSpace += ebr.PartSize + 32 // Añadir el tamaño de la partición lógica y el EBR
				}
				content += "\t\t\t\t</TR>\n"
				content += "\t\t\t\t</TABLE>\n"
				content += "\t\t\t</TD>\n"
			}
		}
	}

	// Recalcular el espacio libre
	freeSpace = totalDiskSize - usedSpace
	freePercentage := float64(freeSpace) / float64(totalDiskSize) * 100

	// Agregar el espacio libre restante
	content += fmt.Sprintf("\t\t\t<TD>Libre<br/>%.2f%% del disco</TD>\n", freePercentage)
	content += "\t\t\t</TR>\n"
	content += "\t\t\t</TABLE>\n>];\n"
	content += "\t}\n"
	content += "}\n"

	// Escribir el contenido en el archivo .dot
	_, err = fileDot.WriteString(content)
	if err != nil {
		return fmt.Errorf("Error al escribir en el archivo .dot: %v", err)
	}

	fmt.Println("Reporte DISK generado exitosamente en:", dotFilePath)
	return nil
}

// Func to fill a section of the file with zeros
func FillWithZeros(file *os.File, start int32, size int32) error {
	// Point to the start of the section to fill
	file.Seek(int64(start), 0)

	// Create a buffer of zeros
	buffer := make([]byte, size)

	// Write the zeros to the file
	_, err := file.Write(buffer)
	if err != nil {
		fmt.Println("Error al llenar el espacio con ceros:", err)
		return err
	}

	fmt.Println("Espacio llenado con ceros desde el byte", start, "por", size, "bytes.")
	return nil
}

// Func to verify if a section of the file is filled with zeros
func VerifyZeros(file *os.File, start int32, size int32) {
	zeros := make([]byte, size)
	_, err := file.ReadAt(zeros, int64(start))
	if err != nil {
		fmt.Println("Error al leer la sección eliminada:", err)
		return
	}

	// Check if the bytes are all zeros
	isZeroFilled := true
	for _, b := range zeros {
		if b != 0 {
			isZeroFilled = false
			break
		}
	}

	if isZeroFilled {
		fmt.Println("La partición eliminada está completamente llena de ceros.")
	} else {
		fmt.Println("Advertencia: La partición eliminada no está completamente llena de ceros.")
	}
}
