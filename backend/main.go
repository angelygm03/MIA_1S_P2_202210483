package main

import (
	"Proyecto2/backend/DiskCommands"
	"Proyecto2/backend/DiskControl"
	"Proyecto2/backend/DiskStruct"
	"Proyecto2/backend/FileManagement"
	"Proyecto2/backend/FileSystem"
	"Proyecto2/backend/UserManagement"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ====== JSON Request ======
type MKDISKRequest struct {
	Path string `json:"path"`
	Size int    `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
}

type RMDISKRequest struct {
	Path string `json:"path"`
}

type FDISKRequest struct {
	Size   int    `json:"size"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Unit   string `json:"unit"`
	Type   string `json:"type"`
	Fit    string `json:"fit"`
	Delete string `json:"delete"`
	Add    int    `json:"add"`
}

type MountRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type ReportRequest struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Id         string `json:"id"`
	PathFileLs string `json:"pathFileLs"`
}

type MkfsRequest struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Fs   string `json:"fs"`
}

type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Id       string `json:"id"`
}

type MkusrRequest struct {
	User string `json:"user"`
	Pass string `json:"pass"`
	Grp  string `json:"grp"`
}

type MkgrpRequest struct {
	Name string `json:"name"`
}

type MkdirRequest struct {
	Path string `json:"path"`
	P    bool   `json:"p"`
}

// ====== CORS ======
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ====== Handlers ======
func createDisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	//Decodify the JSON
	var req MKDISKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para crear disco:", req)

	// Call the function to create the disk
	DiskControl.Mkdisk(req.Size, req.Fit, req.Unit, req.Path)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Disk created successfully at %s", req.Path)))
}

func removeDisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req RMDISKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("Error al decodificar JSON:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify if the file exists
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		fmt.Println("Error: El archivo no existe en la ruta especificada")
		http.Error(w, "Disk not found", http.StatusNotFound)
		return
	}

	// Remove the file
	err := os.Remove(req.Path)
	if err != nil {
		fmt.Println("Error al eliminar el archivo:", err)
		http.Error(w, "Error deleting disk", http.StatusInternalServerError)
		return
	}

	fmt.Println("Archivo eliminado exitosamente:", req.Path)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Disk removed successfully at %s", req.Path)))
}

func createPartition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req FDISKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para crear/modificar/eliminar partición:", req)

	// Check the parameters
	if req.Path == "" || req.Name == "" {
		http.Error(w, "Error: Los parámetros 'path' y 'name' son obligatorios.", http.StatusBadRequest)
		return
	}

	// Delete partition
	if req.Delete != "" {
		if req.Delete != "fast" && req.Delete != "full" {
			http.Error(w, "Error: El parámetro 'delete' debe ser 'fast' o 'full'.", http.StatusBadRequest)
			return
		}
		result, err := DiskControl.DeletePartition(req.Path, req.Name, req.Delete)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error al eliminar la partición: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
		return
	}

	// Modify partition
	if req.Add != 0 {
		if req.Unit != "k" && req.Unit != "m" {
			http.Error(w, "Error: El parámetro 'unit' debe ser 'k' o 'm' cuando se utiliza 'add'.", http.StatusBadRequest)
			return
		}
		result, err := DiskControl.ModifyPartition(req.Path, req.Name, req.Add, req.Unit)
		if err != nil {
			http.Error(w, fmt.Sprintf("No se ha podido modificar la partición: %v", err), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
		return
	}

	// Create partition
	if req.Size <= 0 {
		http.Error(w, "Error: El parámetro 'size' es obligatorio y debe ser mayor a 0 al crear una partición.", http.StatusBadRequest)
		return
	}

	if req.Fit == "" {
		req.Fit = "wf"
	}
	if req.Unit == "" {
		req.Unit = "k"
	}
	if req.Type == "" {
		req.Type = "p"
	}

	result := DiskControl.Fdisk(req.Size, req.Path, req.Name, req.Unit, req.Type, req.Fit, "", 0)
	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func mountPartition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req MountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para montar partición:", req)

	//Call the function to mount the partition
	message := DiskControl.Mount(req.Path, req.Name)

	// Si el mensaje comienza con "Error:", envíalo como un error
	if strings.HasPrefix(message, "Error") {
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func generateReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error: Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para generar reporte:")
	fmt.Printf("Name: %s, Path: %s, ID: %s, PathFileLs: %s\n", req.Name, req.Path, req.Id, req.PathFileLs)

	reportCommand := fmt.Sprintf("-name=%s -path=%s -id=%s", req.Name, req.Path, req.Id)
	if req.PathFileLs != "" {
		reportCommand += fmt.Sprintf(" -path_file_ls=%s", req.PathFileLs)
	}

	result := DiskCommands.Fn_Rep(reportCommand)

	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func formatMkfs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req MkfsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para formatear partición:", req)

	if req.Fs == "" {
		req.Fs = "2fs"
	}

	if req.Fs != "2fs" && req.Fs != "3fs" {
		http.Error(w, "Error: El parámetro 'fs' debe ser '2fs' o '3fs'.", http.StatusBadRequest)
		return
	}

	FileSystem.Mkfs(req.Id, req.Type, req.Fs)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Partition formatted successfully with id %s using %s", req.Id, req.Fs)))
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para loggear usuario:", req)

	// Verify if the user is already logged in
	if UserManagement.IsUserLoggedIn() {
		http.Error(w, "Ya hay un usuario logueado.", http.StatusConflict)
		return
	}

	// Call the function to log in the user
	UserManagement.Login(req.User, req.Password, req.Id)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("User logged in successfully with id %s", req.Id)))
}

func logoutUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	fmt.Println("Solicitud recibida para desloguear usuario")

	result := UserManagement.Logout()

	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func getMountedPartitionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	mountedPartitions := DiskControl.GetMountedPartitions()

	// Convert the data to JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(mountedPartitions); err != nil {
		http.Error(w, "Error al generar JSON", http.StatusInternalServerError)
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req MkusrRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para crear usuario:", req)

	// Validate the parameters
	if req.User == "" || req.Pass == "" || req.Grp == "" {
		http.Error(w, "Error: Los parámetros 'user', 'pass' y 'grp' son obligatorios.", http.StatusBadRequest)
		return
	}

	if len(req.User) > 10 || len(req.Pass) > 10 || len(req.Grp) > 10 {
		http.Error(w, "Error: Los valores de 'user', 'pass' y 'grp' no pueden exceder los 10 caracteres.", http.StatusBadRequest)
		return
	}

	result := UserManagement.Mkusr(req.User, req.Pass, req.Grp)
	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func createGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req MkgrpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error: Solicitud inválida.", http.StatusBadRequest)
		return
	}

	result := UserManagement.Mkgrp(req.Name)
	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func removeUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		User string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para eliminar usuario:", req)

	if req.User == "" {
		http.Error(w, "Error: El parámetro 'user' es obligatorio.", http.StatusBadRequest)
		return
	}

	UserManagement.Rmusr(req.User)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Usuario '%s' eliminado exitosamente.", req.User)))
}

func removeGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para eliminar grupo:", req)

	if req.Name == "" {
		http.Error(w, "Error: El parámetro 'name' es obligatorio.", http.StatusBadRequest)
		return
	}

	UserManagement.Rmgrp(req.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Grupo '%s' eliminado exitosamente.", req.Name)))
}

func changeGroupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		User string `json:"user"`
		Grp  string `json:"grp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para cambiar grupo:", req)

	if req.User == "" || req.Grp == "" {
		http.Error(w, "Error: Los parámetros 'user' y 'grp' son obligatorios.", http.StatusBadRequest)
		return
	}

	UserManagement.Chgrp(req.User, req.Grp)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Grupo del usuario '%s' cambiado exitosamente a '%s'.", req.User, req.Grp)))
}

func createFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Path        string `json:"path"`
		Recursive   bool   `json:"recursive"`
		Size        int    `json:"size"`
		ContentPath string `json:"contentPath"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para crear archivo:", req)

	UserManagement.Mkfile(req.Path, req.Recursive, req.Size, req.ContentPath)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Archivo '%s' creado exitosamente.", req.Path)))
}

func catFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Decodify the JSON
	var files map[string]string
	if err := json.NewDecoder(r.Body).Decode(&files); err != nil {
		fmt.Println("Error al decodificar JSON:", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para leer archivos:", files)

	// Process the files
	var results []string
	for _, path := range files {
		fmt.Printf("Leyendo archivo: %s\n", path)
		result := UserManagement.Cat(path)
		results = append(results, fmt.Sprintf("Archivo %s:\n%s", path, result))
	}

	response := strings.Join(results, "\n\n")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func mkdirHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req MkdirRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error al decodificar la solicitud: "+err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Solicitud recibida para mkdir: path=%s, p=%t\n", req.Path, req.P)

	UserManagement.Mkdir(req.Path, req.P)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Directorio creado exitosamente"))
}

func unmountPartition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para desmontar partición:", req.ID)

	message := DiskControl.Unmount(req.ID)

	if strings.HasPrefix(message, "Error") {
		http.Error(w, message, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

func listCreatedDisksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Get the created disks
	createdDisks := DiskControl.GetCreatedDisks()

	for path, diskInfo := range createdDisks {
		file, err := FileManagement.OpenFile(path)
		if err != nil {
			fmt.Printf("Error al abrir el disco %s: %v\n", path, err)
			continue
		}
		defer file.Close()

		var mbr DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &mbr, 0); err != nil {
			fmt.Printf("Error al leer el MBR del disco %s: %v\n", path, err)
			continue
		}

		var sizeWithUnit string
		if diskInfo.Unit == "k" {
			sizeWithUnit = fmt.Sprintf("%d k", diskInfo.Size/1024)
		} else if diskInfo.Unit == "m" {
			sizeWithUnit = fmt.Sprintf("%d m", diskInfo.Size/(1024*1024))
		} else {
			sizeWithUnit = fmt.Sprintf("%d b", diskInfo.Size)
		}

		// Add the size with unit to the disk info
		diskInfo.SizeWithUnit = sizeWithUnit
		createdDisks[path] = diskInfo
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdDisks); err != nil {
		http.Error(w, "Error al generar JSON", http.StatusInternalServerError)
	}
}

func listPartitionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	partitions := []map[string]interface{}{}

	// Iterate over the created disks
	for path := range DiskControl.GetCreatedDisks() {
		file, err := FileManagement.OpenFile(path)
		if err != nil {
			fmt.Printf("Error al abrir el disco %s: %v\n", path, err)
			continue
		}
		defer file.Close()

		var mbr DiskStruct.MRB
		if err := FileManagement.ReadObject(file, &mbr, 0); err != nil {
			fmt.Printf("Error al leer el MBR del disco %s: %v\n", path, err)
			continue
		}

		// Find the partitions in the MBR
		for _, partition := range mbr.Partitions {
			if partition.Size > 0 {
				var partitionType string
				switch string(partition.Type[:]) {
				case "p":
					partitionType = "Primaria"
				case "e":
					partitionType = "Extendida"
				default:
					partitionType = "Desconocido"
				}

				partitions = append(partitions, map[string]interface{}{
					"Name":   string(partition.Name[:]),
					"Path":   path,
					"Size":   partition.Size,
					"Fit":    string(partition.Fit[:]),
					"Type":   partitionType,
					"Status": string(partition.Status[:]),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(partitions); err != nil {
		http.Error(w, "Error al generar JSON", http.StatusInternalServerError)
	}
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "El parámetro 'path' es obligatorio.", http.StatusBadRequest)
		return
	}

	// Get the mounted partitions
	mountedPartitions := DiskControl.GetMountedPartitions()
	var filepath string
	var partitionFound bool

	for _, partitions := range mountedPartitions {
		for _, partition := range partitions {
			if partition.LoggedIn {
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
		http.Error(w, "No hay ninguna partición activa.", http.StatusBadRequest)
		return
	}

	// Open bin file
	file, err := FileManagement.OpenFile(filepath)
	if err != nil {
		http.Error(w, "No se pudo abrir el archivo del disco.", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Read the MBR
	var TempMBR DiskStruct.MRB
	if err := FileManagement.ReadObject(file, &TempMBR, 0); err != nil {
		http.Error(w, "No se pudo leer el MBR.", http.StatusInternalServerError)
		return
	}

	// Read the superblock
	var tempSuperblock DiskStruct.Superblock
	for i := 0; i < 4; i++ {
		if TempMBR.Partitions[i].Status[0] == '1' { // Active partition
			if err := FileManagement.ReadObject(file, &tempSuperblock, int64(TempMBR.Partitions[i].Start)); err != nil {
				http.Error(w, "No se pudo leer el Superblock.", http.StatusInternalServerError)
				return
			}
			break
		}
	}

	// Find the inode of the folder
	indexInode := UserManagement.InitSearch(path, file, tempSuperblock)
	if indexInode == -1 {
		http.Error(w, "No se encontró la carpeta especificada.", http.StatusNotFound)
		return
	}

	var crrInode DiskStruct.Inode
	if err := FileManagement.ReadObject(file, &crrInode, int64(tempSuperblock.S_inode_start+indexInode*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
		http.Error(w, "No se pudo leer el Inodo de la carpeta.", http.StatusInternalServerError)
		return
	}

	// Read the folder blocks
	items := []map[string]string{}
	for _, block := range crrInode.I_block {
		if block == -1 {
			continue
		}

		var folderBlock DiskStruct.Folderblock
		if err := FileManagement.ReadObject(file, &folderBlock, int64(tempSuperblock.S_block_start+block*int32(binary.Size(DiskStruct.Folderblock{})))); err != nil {
			http.Error(w, "No se pudo leer el bloque de carpeta.", http.StatusInternalServerError)
			return
		}

		for _, content := range folderBlock.B_content {
			if content.B_inodo == -1 {
				continue
			}

			itemName := strings.Trim(string(content.B_name[:]), " \t\n\r\x00")
			fmt.Printf("Elemento encontrado: '%s'\n", itemName)

			if itemName == "." || itemName == ".." || itemName == "" {
				fmt.Printf("Elemento omitido: '%s'\n", itemName)
				continue
			}

			var itemInode DiskStruct.Inode
			if err := FileManagement.ReadObject(file, &itemInode, int64(tempSuperblock.S_inode_start+content.B_inodo*int32(binary.Size(DiskStruct.Inode{})))); err != nil {
				fmt.Printf("Error al leer el inodo del elemento '%s': %v\n", itemName, err)
				continue
			}

			itemType := "file"
			if itemInode.I_type[0] == '0' {
				itemType = "folder"
			}

			fmt.Printf("Elemento agregado: '%s', Tipo: '%s'\n", itemName, itemType)

			items = append(items, map[string]string{
				"name":        itemName,
				"type":        itemType,
				"permissions": string(itemInode.I_perm[:]),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Error al generar JSON.", http.StatusInternalServerError)
	}
}

func recoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para recuperación del sistema de archivos:", req)

	result := FileSystem.Recovery(req.Id)
	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func lossHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	fmt.Println("Solicitud recibida para simular pérdida del sistema de archivos:", req)

	result := FileSystem.Loss(req.Id)
	if strings.HasPrefix(result, "Error:") {
		http.Error(w, result, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

func main() {
	// Load the mounted partitions
	if err := DiskControl.LoadMountedPartitions(); err != nil {
		fmt.Println("Error al cargar el estado de las particiones montadas:", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})
	mux.HandleFunc("/mkdisk", createDisk)
	mux.HandleFunc("/rmdisk", removeDisk)
	mux.HandleFunc("/fdisk", createPartition)
	mux.HandleFunc("/mount", mountPartition)
	mux.HandleFunc("/report", generateReport)
	mux.HandleFunc("/mkfs", formatMkfs)
	mux.HandleFunc("/login", loginUser)
	mux.HandleFunc("/logout", logoutUser)
	mux.HandleFunc("/list-mounted", getMountedPartitionsHandler)
	mux.HandleFunc("/mkusr", createUserHandler)
	mux.HandleFunc("/mkgrp", createGroupHandler)
	mux.HandleFunc("/rmusr", removeUserHandler)
	mux.HandleFunc("/rmgrp", removeGroupHandler)
	mux.HandleFunc("/chgrp", changeGroupHandler)
	mux.HandleFunc("/mkfile", createFileHandler)
	mux.HandleFunc("/cat", catFileHandler)
	mux.HandleFunc("/mkdir", mkdirHandler)
	mux.HandleFunc("/unmount", unmountPartition)
	mux.HandleFunc("/list-disks", listCreatedDisksHandler)
	mux.HandleFunc("/recovery", recoveryHandler)
	mux.HandleFunc("/loss", lossHandler)
	mux.HandleFunc("/list-partitions", listPartitionsHandler)
	mux.HandleFunc("/list-files", listFilesHandler)

	fmt.Println("Servidor corriendo en http://localhost:8080")
	http.ListenAndServe(":8080", enableCORS(mux))
}
