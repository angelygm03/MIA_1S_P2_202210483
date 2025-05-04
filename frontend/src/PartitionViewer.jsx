import React, { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import "./PartitionViewer.css";
import partitionImage from "./assets/disco-duro.png";

function PartitionViewer() {
  const [partitions, setPartitions] = useState([]);
  const [expandedPartition, setExpandedPartition] = useState(null);
  const location = useLocation();

  useEffect(() => {
    const fetchPartitions = async () => {
      try {
        const params = new URLSearchParams(location.search);
        const diskPath = params.get("disk");

        const url = diskPath
          ? `http://localhost:8080/list-partitions?disk=${encodeURIComponent(diskPath)}`
          : "http://localhost:8080/list-partitions";

        const response = await fetch(url, {
          method: "GET",
          headers: { "Content-Type": "application/json" },
        });

        if (!response.ok) {
          throw new Error("Error al obtener la lista de particiones");
        }

        const data = await response.json();
        setPartitions(data);
      } catch (error) {
        console.error("Error al obtener particiones:", error);
      }
    };

    fetchPartitions();
  }, [location.search]);

  const handlePartitionClick = (index) => {
    setExpandedPartition(expandedPartition === index ? null : index);
  };

  return (
    <div className="partition-viewer-container">
      <h1>Visualizador de Particiones</h1>
      {partitions.length === 0 ? (
        <p>No hay particiones disponibles.</p>
      ) : (
        <div className="partitions-grid">
          {partitions.map((partition, index) => {
            const isExpanded = expandedPartition === index;

            const diskName = partition.Path.split("/").pop();

            return (
              <div key={index} className="partition-card">
                <button
                  className="partition-button"
                  onClick={() => handlePartitionClick(index)}
                >
                  <img
                    src={partitionImage}
                    alt="Partición"
                    className="partition-image"
                  />
                  <span className="partition-name">
                    {partition.Name || "Sin nombre"}
                  </span>
                </button>

                {isExpanded && (
                  <div className="partition-info">
                    <p>
                      <strong>Disco:</strong> {diskName}
                    </p>
                    <p>
                      <strong>Path:</strong> {partition.Path}
                    </p>
                    <p>
                      <strong>Tamaño:</strong> {partition.Size} bytes
                    </p>
                    <p>
                      <strong>Fit:</strong> {partition.Fit + "f"}
                    </p>
                    <p>
                      <strong>Tipo:</strong> {partition.Type || "Desconocido"}
                    </p>
                    <p>
                      <strong>Estado:</strong>{" "}
                      {partition.Status === "1" ? "Montada" : "No montada"}
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

export default PartitionViewer;