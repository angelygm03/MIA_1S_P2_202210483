import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./DiskViewer.css";
import diskImage from "./assets/hdd.png";

function DiskViewer({ onSelectDisk }) {
  const [disks, setDisks] = useState([]);
  const navigate = useNavigate(); 

  useEffect(() => {
    const fetchDisks = async () => {
      try {
        const response = await fetch("http://localhost:8080/list-disks", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
        });
    
        console.log("Estado de la respuesta:", response.status);
        if (!response.ok) {
          throw new Error("Error al obtener la lista de discos");
        }
    
        const data = await response.json();
        console.log("Datos recibidos del backend:", data); 
        setDisks(data);
      } catch (error) {
        console.error("Error al obtener discos:", error);
      }
    };

    fetchDisks();
  
    const interval = setInterval(fetchDisks, 3000);
  
    return () => clearInterval(interval);
  }, []);

  const handleDiskClick = (disk) => {
    onSelectDisk(disk); 
    navigate("/"); 
  };

  return (
    <div className="disk-viewer-container">
      <h1>Visualizador de Discos</h1>
      {Object.keys(disks).length === 0 ? (
        <p>No hay discos montados.</p>
      ) : (
        <div className="disks-grid">
          {Object.entries(disks).map(([diskPath, info]) => {
            const diskName = diskPath.split("/").pop();
            return (
              <button
                key={diskPath}
                className="disk-card"
                onClick={() => handleDiskClick(diskPath)}
              >
                <img src={diskImage} alt="Disco" className="disk-image" />
                <h2>{diskName}</h2> {/* Muestra solo el nombre del archivo */}
                <p><strong>Path:</strong> {info.Path}</p>
                <p><strong>Size:</strong> {info.Size} {info.Unit}</p>
                <p><strong>Fit:</strong> {info.Fit}</p>
                <ul>
                  {info.Partitions.length > 0 ? (
                    info.Partitions.map((partition, index) => (
                      <li key={index}>
                        <strong>Partición:</strong> {partition.Name} - <strong>ID:</strong> {partition.ID}
                      </li>
                    ))
                  ) : (
                    <li>No hay particiones</li>
                  )}
                </ul>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default DiskViewer;