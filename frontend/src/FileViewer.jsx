import React, { useState, useEffect } from "react";
import "./FileViewer.css";
import folderImage from "./assets/carpeta.png";
import fileImage from "./assets/txt.png";

function FileViewer() {
  const [currentPath, setCurrentPath] = useState("/");
  const [items, setItems] = useState([]); 
  const [searchQuery, setSearchQuery] = useState(""); 
  const [expandedItem, setExpandedItem] = useState(null); 
  const [fileContent, setFileContent] = useState(""); 
  const [isPanelOpen, setIsPanelOpen] = useState(false); 

  useEffect(() => {
    console.log("Fetching files and folders for path:", currentPath);
    fetchFilesAndFolders(currentPath);
  }, [currentPath]);

  const fetchFilesAndFolders = async (path) => {
    try {
      const response = await fetch(`http://localhost:8080/list-files?path=${encodeURIComponent(path)}`, {
        method: "GET",
        headers: { "Content-Type": "application/json" },
      });
  
      if (!response.ok) {
        throw new Error("Error al obtener los archivos y carpetas");
      }
  
      const data = await response.json();
      console.log("Archivos y carpetas recibidos:", data);
  
      const filteredData = data.filter(
        (item) => item.name !== "." && item.name !== ".." && item.name.trim() !== ""
      );
  
      setItems(filteredData);
    } catch (error) {
      console.error("Error al obtener archivos y carpetas:", error);
    }
  };

  const handleExpandClick = (index) => {
    setExpandedItem(expandedItem === index ? null : index);
  };

  const handleBackClick = () => {
    if (currentPath !== "/") {
      const newPath = currentPath.slice(0, currentPath.lastIndexOf("/", currentPath.length - 2) + 1);
      setCurrentPath(newPath);
    }
  };

  const handleOpenFile = async (fileName) => {
    try {
      const response = await fetch("http://localhost:8080/cat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ file: `${currentPath}${fileName}` }),
      });
  
      if (!response.ok) {
        throw new Error("Error al abrir el archivo");
      }
  
      const content = await response.text();
      setFileContent(content);
      setIsPanelOpen(true); 
    } catch (error) {
      console.error("Error al abrir el archivo:", error);
    }
  };

  const handleClosePanel = () => {
    setIsPanelOpen(false); 
    setFileContent("");
  };

  const filteredItems = items.filter((item) =>
    item.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="file-viewer-container">
    <h1>Visualizador de Archivos</h1>
    <div className="path-container">
      <span className="current-path">{currentPath}</span>
    </div>
    <input
      type="text"
      placeholder="Buscar..."
      value={searchQuery}
      onChange={(e) => setSearchQuery(e.target.value)}
      className="search-bar"
    />
    <div className="items-grid">
      {filteredItems.map((item, index) => {
        const isExpanded = expandedItem === index;

        return (
          <div key={index} className="item-card">
            <button
              className="item-button"
              onClick={() => handleExpandClick(index)}
            >
              <img
                src={item.type === "folder" ? folderImage : fileImage}
                alt={item.type}
                className="item-image"
              />
              <span className="item-name">{item.name}</span>
            </button>

            {isExpanded && (
              <div className="item-info">
                <p>
                  <strong>Nombre:</strong> {item.name}
                </p>
                <p>
                  <strong>Tipo:</strong> {item.type === "folder" ? "Carpeta" : "Archivo"}
                </p>
                <p>
                  <strong>Permisos:</strong> {item.permissions}
                </p>
                {item.type === "file" && (
                  <button
                    className="open-file-button"
                    onClick={() => handleOpenFile(item.name)}
                  >
                    Abrir
                  </button>
                )}
              </div>
            )}
          </div>
        );
      })}

      {/* Botón de demostración */}
      <div className="item-card">
        <button className="item-button">
          <img
            src={folderImage}
            alt="folder"
            className="item-image"
          />
          <span className="item-name"></span>
        </button>
      </div>
    </div>

    {/* Panel to show the content file */}
    {isPanelOpen && (
      <div className="file-content-panel">
        <div className="panel-header">
          <h2>Contenido del Archivo</h2>
          <button className="close-panel-button" onClick={handleClosePanel}>
            Cerrar
          </button>
        </div>
        <div className="panel-content">
          <pre>{fileContent}</pre>
        </div>
      </div>
    )}
  </div>
  );
}

export default FileViewer;