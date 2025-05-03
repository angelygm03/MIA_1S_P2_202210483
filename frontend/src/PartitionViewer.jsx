import React, { useState, useEffect } from "react";
import "./PartitionViewer.css";
import partitionImage from "./assets/disco-duro.png";

function PartitionViewer() {
  const [partitions, setPartitions] = useState([]);
  const [expandedPartition, setExpandedPartition] = useState(null);

  useEffect(() => {
    const fetchPartitions = async () => {
      try {
        const response = await fetch("http://localhost:8080/list-partitions", {
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
  }, []);

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

            // Get the disk name from the partition path
            const diskName = partition.Path.split("/").pop();

            return (
              <div key={index} className="partition-card">
                {/* Button with the image and disk name */}
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

                {/* Show partition info if it is expanded */}
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