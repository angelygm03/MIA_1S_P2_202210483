import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import "./DiskViewer.css";
import diskImage from "./assets/hdd.png";

function DiskViewer({ onSelectDisk }) {
  const [disks, setDisks] = useState([]);
  const [expandedDisk, setExpandedDisk] = useState(null);
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
        setDisks(data);
      } catch (error) {
        console.error("Error al obtener discos:", error);
        setDisks([]);
      }
    };

    fetchDisks();

    const interval = setInterval(fetchDisks, 3000);

    return () => clearInterval(interval);
  }, []);

  const handleDiskClick = (diskPath) => {
    setExpandedDisk(expandedDisk === diskPath ? null : diskPath);
  };

  const handleViewPartitions = (diskPath) => {
    navigate(`/partitions?disk=${encodeURIComponent(diskPath)}`);
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
                <button
                  className="disk-button"
                  onClick={() => handleDiskClick(diskPath)}
                >
                  <img src={diskImage} alt="Disco" className="disk-image" />
                  <span className="disk-name">{diskName}</span>
                </button>

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
                    <button
                      className="view-partitions-button"
                      onClick={() => handleViewPartitions(diskPath)}
                    >
                      Ver Particiones
                    </button>
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