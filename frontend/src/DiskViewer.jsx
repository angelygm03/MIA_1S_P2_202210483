import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./DiskViewer.css";
import diskImage from "./assets/hdd.png";

function DiskViewer({ onSelectDisk }) {
  const [disks, setDisks] = useState([]);
  const [expandedDisk, setExpandedDisk] = useState(null); // State to track which disk is expanded
  const navigate = useNavigate();

  useEffect(() => {
    const fetchDisks = async () => {
      try {
        const response = await fetch("http://localhost:8080/list-disks", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
        });
  
        if (!response.ok) {
          throw new Error("Error al obtener la lista de discos");
        }
  
        const data = await response.json();
        setDisks(data); // Actualiza la lista de discos con los datos del servidor
      } catch (error) {
        console.error("Error al obtener discos:", error);
        setDisks([]); // Limpia la lista de discos si hay un error
      }
    };
  
    fetchDisks();
  
    const interval = setInterval(fetchDisks, 3000);
  
    return () => clearInterval(interval);
  }, []);

  const handleDiskClick = (diskPath) => {
    setExpandedDisk(expandedDisk === diskPath ? null : diskPath); // Alternar entre expandido y colapsado
  };

  const handlePartitionClick = (partition) => {
    onSelectDisk(partition);
    navigate("/");
  };

  return (
    <div className="disk-viewer-container">
      <h1>Visualizador de Discos</h1>
      {Object.keys(disks).length === 0 ? (
        <p>No hay discos creados.</p>
      ) : (
        <div className="disks-grid">
          {Object.entries(disks).map(([diskPath, info]) => {
            const diskName = diskPath.split("/").pop();
            const isExpanded = expandedDisk === diskPath;

            return (
              <div key={diskPath} className="disk-card">
                {/* Button with image and disk name */}
                <button
                  className="disk-button"
                  onClick={() => handleDiskClick(diskPath)}
                >
                  <img src={diskImage} alt="Disco" className="disk-image" />
                  <span className="disk-name">{diskName}</span>
                </button>

                {/* Show disk info if it is expanded*/}
                {isExpanded && (
                  <div className="disk-info">
                    <p>
                      <strong>Path:</strong> {info.Path}
                    </p>
                    <p>
                      <strong>Size:</strong> {info.SizeWithUnit || `${info.Size} ${info.Unit}`}
                    </p>
                    <p>
                      <strong>Fit:</strong> {info.Fit}
                    </p>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

export default DiskViewer;