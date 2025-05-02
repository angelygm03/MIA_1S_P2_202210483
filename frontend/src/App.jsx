import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { useState } from 'react';
import './App.css';
import Navbar from "./Navbar";
import LoginForm from "./LoginForm";
import DiskViewer from "./DiskViewer";

function App() {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [selectedDisk, setSelectedDisk] = useState(null);

  const handleExecute = async () => {
    try {
      const commands = input.trim().split("\n"); // Commands are separated by lines
      let results = [];
      
      for (const command of commands) {
        const trimmedCommand = command.trim();

        // Ignore empty lines
        if (trimmedCommand === "") {
          results.push(""); // Add an empty line to the results
          continue; // Skip processing this line
        }

        // Check if the line is a comment
        if (trimmedCommand.startsWith("#") || trimmedCommand.startsWith("/")) {
          results.push(trimmedCommand); // Add the comment directly to the results
          continue; // Skip processing this line as a command
        }
  
        const params = trimmedCommand.split(" ");
        let requestBody = {};
        let endpoint = "";

        if (command.trim().toLowerCase() === "mounted") {
          try {
            const response = await fetch(`http://localhost:8080/list-mounted`, {
              method: "GET",
              headers: { "Content-Type": "application/json" },
            });

            if (!response.ok) {
              throw new Error(`Error del servidor: ${response.status} ${response.statusText}`);
            }

            const partitions = await response.json();

            let partitionList = "Particiones montadas:\n";

            for (const [disk, parts] of Object.entries(partitions)) {
              partitionList += `Disco: ${disk}\n`;
              parts.forEach((part) => {
                partitionList += `- Nombre: ${part.Name}, ID: ${part.ID}\n`;
              });
            }

            console.log("Lista de particiones procesada:", partitionList);

            if (!results.some((res) => res.includes("Particiones montadas"))) {
              results.push(`======================================================\nComando: ${command}\n${partitionList}======================================================\n`);
            }
          } catch (error) {
            console.error("Error al ejecutar el comando 'mounted':", error);
            results.push(`Error al obtener particiones montadas: ${error.message}`);
          }

        } else if (command.startsWith("mount ")) {
          let path = "", name = "";
          params.forEach(param => {
            if (param.startsWith("-path=")) path = param.split("=")[1].replace(/"/g, '');
            if (param.startsWith("-name=")) name = param.split("=")[1].replace(/"/g, '');
          });

          requestBody = { path, name };
          endpoint = "mount";

        } else if (command.toLowerCase().startsWith("mkdisk")) {
          let size = null, unit = "m", fit = "ff", path = "";
          let errors = [];
        
          // List of valid parameters
          const validParams = ["-size", "-unit", "-fit", "-path"];
        
          // Regular expression to match parameters
          const paramRegex = /(-\w+=("[^"]*"|[^\s]+))/g;
          const matches = command.match(paramRegex);
        
          if (matches) {
            matches.forEach(param => {
              const [key, value] = param.split("=");
              const lowerKey = key.toLowerCase(); 
        
              if (!validParams.includes(lowerKey)) {
                errors.push(`Parámetro no reconocido: ${lowerKey}`);
              }
        
              if (lowerKey === "-size") size = parseInt(value);
              if (lowerKey === "-unit") unit = value.toLowerCase();
              if (lowerKey === "-fit") fit = value.toLowerCase();
              if (lowerKey === "-path") path = value.replace(/"/g, ''); // Eliminar comillas
            });
          }
        
          // Validate required parameters
          if (size === null || isNaN(size) || size <= 0) {
            errors.push("El parámetro '-size' es obligatorio y debe ser mayor a 0.");
          }
          if (!path) {
            errors.push("El parámetro '-path' es obligatorio.");
          }
          // Validate optional parameters
          if (unit !== "k" && unit !== "m") {
            errors.push("El parámetro '-unit' debe ser 'k' o 'm'.");
          }
          if (fit !== "ff" && fit !== "bf" && fit !== "wf") {
            errors.push("El parámetro '-fit' debe ser 'ff', 'bf' o 'wf'.");
          }
        
          // If errors are found, add them to the results and skip this command
          if (errors.length > 0) {
            results.push(`======================================================\nComando: ${command}\nErrores:\n- ${errors.join("\n- ")}\n======================================================\n`);
            continue; // Stop processing this command
          }
        
          requestBody = { size, unit, fit, path };
          endpoint = "mkdisk";

        } else if (command.startsWith("rmdisk")) {
          // Get parameters for rmdisk
          let path = "";
          params.forEach(param => {
            if (param.startsWith("-path=")) path = param.split("=")[1].replace(/"/g, '');
          });
          
          // Set request body and endpoint for rmdisk
          requestBody = { path };
          endpoint = "rmdisk";

        } else if (command.startsWith("fdisk")) {
          let size = 0, unit = "k", path = "", type = "p", fit = "wf", name = "", delete_ = "", add = 0;
          let errors = [];
        
          params.forEach(param => {
            if (param.startsWith("-size=")) size = parseInt(param.split("=")[1]);
            if (param.startsWith("-path=")) path = param.split("=")[1].replace(/"/g, '');
            if (param.startsWith("-name=")) name = param.split("=")[1].replace(/"/g, '');
            if (param.startsWith("-unit=")) unit = param.split("=")[1].toLowerCase();
            if (param.startsWith("-type=")) type = param.split("=")[1].toLowerCase();
            if (param.startsWith("-fit=")) fit = param.split("=")[1].toLowerCase();
            if (param.startsWith("-delete=")) delete_ = param.split("=")[1].toLowerCase();
            if (param.startsWith("-add=")) add = parseInt(param.split("=")[1]);
          });
        
          // Validate required parameters
          if (!path) errors.push("El parámetro '-path' es obligatorio.");
          if (!name) errors.push("El parámetro '-name' es obligatorio.");
          
          if (errors.length > 0) {
            results.push(`Errores:\n- ${errors.join("\n- ")}`);
            return;
          }

          // Validate -size for partition creation
          if (!delete_ && add === 0 && size <= 0) {
            errors.push("El parámetro '-size' es obligatorio y debe ser mayor a 0 al crear una partición.");
          }
        
          // Validate -delete values
          if (delete_ && delete_ !== "fast" && delete_ !== "full") {
            errors.push("El parámetro '-delete' debe ser 'fast' o 'full'.");
          }
        
          // Validate -unit for add or delete
          if ((add !== 0 || delete_) && unit !== "k" && unit !== "m") {
            errors.push("El parámetro '-unit' debe ser 'k' o 'm' cuando se utiliza '-add' o '-delete'.");
          }
        
          // If errors are found, add them to the results and skip this command
          if (errors.length > 0) {
            results.push(`======================================================\nComando: ${command}\nErrores:\n- ${errors.join("\n- ")}\n======================================================\n`);
            return;
          }
        
          // Show confirmation dialog for delete
          if (delete_) {
            const confirmDelete = window.confirm(`¿Está seguro de que desea eliminar la partición '${name}'?`);
            if (!confirmDelete) {
              results.push("Operación cancelada por el usuario.");
              return;
            }
          }
        
          requestBody = { size, unit, path, type, fit, name, delete: delete_, add };
          endpoint = "fdisk";
        
        } else if (command.startsWith("login")) {
          let user = "", pass = "", id="";
          params.forEach(param => {
            if (param.startsWith("-user=")) user = param.split("=")[1];
            if (param.startsWith("-pass=")) pass = param.split("=")[1];
            if (param.startsWith("-id=")) id = param.split("=")[1].toLowerCase();
          });
        
          requestBody = { user, password: pass, id };
          endpoint = "login";
        
        } else if (command.startsWith("mkfs")) {
          let id = "", type = "full", fs = "2fs";
          params.forEach(param => {
            if (param.startsWith("-id=")) id = param.split("=")[1].toLowerCase();
            if (param.startsWith("-type=")) type = param.split("=")[1].toLowerCase();
            if (param.startsWith("-fs=")) fs = param.split("=")[1].toLowerCase();
          });
        
          requestBody = { id, type, fs };
          endpoint = "mkfs";
        
        } else if (command.startsWith("logout")) {
          requestBody = {}; 
          endpoint = "logout";    
        
        } else if (command.startsWith("rep")) {
          let path = "", name = "", id = "", pathFileLs = ""; 
        
          params.forEach(param => {
            const [key, value] = param.split("=");
            const normalizedKey = key.toLowerCase(); // Convertir el nombre del parámetro a minúsculas
        
            if (normalizedKey === "-path") path = value.replace(/"/g, '');
            if (normalizedKey === "-name") name = value.replace(/"/g, '');
            if (normalizedKey === "-id") id = value.toLowerCase();
            if (normalizedKey === "-path_file_ls") pathFileLs = value.replace(/"/g, '');
          });
        
          requestBody = { path, name, id, pathFileLs };
          endpoint = "report";
        
        } else if (command.startsWith("mkusr")) {
          let user = "", pass = "", grp = "";
          params.forEach(param => {
            if (param.startsWith("-user=")) user = param.split("=")[1].trim();
            if (param.startsWith("-pass=")) pass = param.split("=")[1].trim();
            if (param.startsWith("-grp=")) grp = param.split("=")[1].trim();
          });
        
          // Validate that all parameters are present
          if (!user || !pass || !grp) {
            results.push(`Error: Los parámetros 'user', 'pass' y 'grp' son obligatorios para el comando 'mkusr'.`);
            continue;
          }
        
          // Validate that all parameters are not longer than 10 characters
          if (user.length > 10 || pass.length > 10 || grp.length > 10) {
            results.push(`Error: Los valores de 'user', 'pass' y 'grp' no pueden exceder los 10 caracteres.`);
            continue;
          }
      
          requestBody = { user, pass, grp };
          endpoint = "mkusr";
        
        } else if (command.startsWith("mkgrp")) {
          let name = "";
          params.forEach(param => {
            if (param.startsWith("-name=")) name = param.split("=")[1].trim();
          });
        
          // Name is required
          if (!name) {
            results.push(`Error: El parámetro 'name' es obligatorio para el comando 'mkgrp'.`);
            continue;
          }
        
          requestBody = { name };
          endpoint = "mkgrp";

        } else if (command.startsWith("rmgrp")) {
          let name = "";
          params.forEach(param => {
            if (param.startsWith("-name=")) name = param.split("=")[1].trim();
          });

          // Name is required
          if (!name) {
            results.push(`Error: El parámetro 'name' es obligatorio para el comando 'rmgrp'.`);
            continue;
          }

          requestBody = { name };
          endpoint = "rmgrp";

        } else if (command.startsWith("rmusr")) {
          let user = "";
          params.forEach(param => {
            if (param.startsWith("-user=")) user = param.split("=")[1].trim();
          });

          // User is required
          if (!user) {
            results.push(`Error: El parámetro 'user' es obligatorio para el comando 'rmusr'.`);
            continue;
          }

          requestBody = { user };
          endpoint = "rmusr";
        
        } else if (command.startsWith("chgrp")) {
          let user = "", grp = "";
          params.forEach(param => {
            if (param.startsWith("-user=")) user = param.split("=")[1].trim();
            if (param.startsWith("-grp=")) grp = param.split("=")[1].trim();
          });
        
          // User and group are required
          if (!user || !grp) {
            results.push(`Error: Los parámetros 'user' y 'grp' son obligatorios para el comando 'chgrp'.`);
            continue;
          }
        
          requestBody = { user, grp };
          endpoint = "chgrp";
        
        } else if (command.startsWith("mkfile")) {
          let path = "", recursive = false, size = 0, contentPath = "";
          params.forEach(param => {
            if (param.startsWith("-path=")) path = param.split("=")[1].replace(/"/g, '');
            if (param === "-r") recursive = true;
            if (param.startsWith("-size=")) size = parseInt(param.split("=")[1]);
            if (param.startsWith("-cont=")) contentPath = param.split("=")[1].replace(/"/g, '');
          });
        
          // Path is required
          if (!path) {
            results.push(`Error: El parámetro 'path' es obligatorio para el comando 'mkfile'.`);
            continue;
          }
        
          // size must be a positive number
          if (size < 0) {
            results.push(`Error: El tamaño del archivo no puede ser negativo.`);
            continue;
          }
        
          requestBody = { path, recursive, size, contentPath };
          endpoint = "mkfile";
        
        } else if (command.startsWith("cat")) {
          let files = {};
          params.forEach(param => {
            const match = param.match(/^-file(\d+)=(.+)$/);
            if (match) {
              const key = `file${match[1]}`; // file followed by a number
              const value = match[2].replace(/"/g, ''); 
              files[key] = value;
            }
          });

          // At least one file is required
          if (Object.keys(files).length === 0) {
            results.push("Error: Debe proporcionar al menos un archivo para el comando 'cat'.");
            continue;
          }

          requestBody = files; 
          endpoint = "cat";
        
        } else if (command.trim().toLowerCase().startsWith("unmount")) {
          let id = "";
          params.forEach(param => {
            if (param.toLowerCase().startsWith("-id=")) {
              id = param.split("=")[1].toLowerCase();
            }
          });
        
          if (!id) {
            results.push(`Error: El parámetro '-id' es obligatorio para el comando 'unmount'.`);
            continue;
          }
        
          requestBody = { id };
          endpoint = "unmount";
        
        } else if (command.startsWith("recovery")) {
          let id = "";
          params.forEach(param => {
            if (param.startsWith("-id=")) id = param.split("=")[1].toLowerCase();
          });
        
          if (!id) {
            results.push(`Error: El parámetro '-id' es obligatorio para el comando 'recovery'.`);
            return;
          }
        
          requestBody = { id };
          endpoint = "recovery";
        
        } else if (command.startsWith("loss")) {
          let id = "";
          params.forEach(param => {
            if (param.startsWith("-id=")) id = param.split("=")[1].toLowerCase();
          });
        
          if (!id) {
            results.push(`Error: El parámetro '-id' es obligatorio para el comando 'loss'.`);
            return;
          }
        
          requestBody = { id };
          endpoint = "loss";

        } else if (command.toLowerCase().startsWith("mkdir ")) {
          let path = "";
          let p = false;
        
          params.forEach(param => {
            if (param.startsWith("-path=")) {
              path = param.split("=")[1].replace(/"/g, '');
            }
            if (param === "-p") {
              p = true; // If "-p" is present, set p to true
            }
          });
        
          if (!path) {
            results.push(`Error: El parámetro '-path' es obligatorio para el comando mkdir.`);
            continue;
          }
        
          requestBody = { path, p };
          endpoint = "mkdir";

        } else {
          results.push(`======================================================\nComando no reconocido: ${command}\n======================================================\n`);
          continue;
        }
  
        if (endpoint) {
          try {
            const response = await fetch(`http://localhost:8080/${endpoint}`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify(requestBody),
            });

            if (!response.ok) {
              const errorText = await response.text();
              results.push(`======================================================\nError: ${errorText}\n======================================================\n\n`);
            } else {
              const text = await response.text();
              results.push(`======================================================\nComando: ${command}\n\nRespuesta: ${text}\n======================================================\n\n`);
            }
          } catch (error) {
            results.push(`======================================================\nError al comunicarse con el servidor: ${error.message}\n======================================================\n\n`);
          }
        }
      }

      // Show results in the output
      setOutput(results.join("\n"));

    } catch (error) {
      setOutput(`======================================================\nError al ejecutar comandos: ${error.message}\n======================================================\n\n`);
    }
  };
  
  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (file) {
      // Validate file extension
      if (!file.name.endsWith(".smia")) {
        alert("Solo se permiten archivos con la extensión .smia");
        return;
      }
  
      const reader = new FileReader();
      reader.onload = (e) => setInput(e.target.result); // Set the file content to the input textarea
      reader.readAsText(file);
    }
  };

    // Function to clear both textareas
    const handleClear = () => {
      setInput('');
      setOutput('');
    };

    const handleSelectDisk = (disk) => {
      setSelectedDisk(disk);
      console.log(`Disco seleccionado: ${disk}`);
    };

    return (
      <Router>
        <div className="container">
          <Navbar /> {/* Navbar fija en la parte superior */}
          <Routes>
            {/* Ruta principal */}
            <Route
              path="/"
              element={
                <>
                  <h1>Sistema de Archivos EXT2</h1>
                  {selectedDisk && <p>Disco seleccionado: {selectedDisk}</p>}
                  <div className="textarea-container">
                    <textarea
                      className="input-area"
                      placeholder="Ingrese comandos aquí..."
                      value={input}
                      onChange={(e) => setInput(e.target.value)}
                    ></textarea>
                    <textarea
                      className="output-area"
                      placeholder="Salida..."
                      value={output}
                      readOnly
                    ></textarea>
                  </div>
                  <div className="buttons">
                    <input type="file" accept=".smia" onChange={handleFileUpload} />
                    <button onClick={handleExecute}>Ejecutar</button>
                    <button onClick={handleClear}>Limpiar</button>
                  </div>
                </>
              }
            />
  
            {/* Ruta para el visualizador de discos */}
            <Route
              path="/disks"
              element={<DiskViewer onSelectDisk={handleSelectDisk} />}
            />
  
            {/* Ruta para el formulario de login */}
            <Route path="/login" element={<LoginForm />} />
          </Routes>
        </div>
      </Router>
    );
  }
  
  export default App;