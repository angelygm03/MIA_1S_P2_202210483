import React from "react";
import { Link, useNavigate } from "react-router-dom";
import "./Navbar.css";

function Navbar() {
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      const response = await fetch("http://localhost:8080/logout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
      });

      if (!response.ok) {
        throw new Error("Error al cerrar sesión");
      }

      alert("Sesión cerrada exitosamente");
      navigate("/login");
    } catch (error) {
      console.error("Error al cerrar sesión:", error);
      alert("Error al cerrar sesión");
    }
  };

  return (
    <nav className="navbar">
      <div className="navbar-container">
        <Link to="/" className="navbar-logo">
          File System
        </Link>
        <div className="navbar-links">
          <Link to="/disks" className="navbar-link">
            Discos
          </Link>
          <Link to="/partitions" className="navbar-link">
            Particiones
          </Link>
          <Link to="/files" className="navbar-link">
            Archivos
          </Link>
          <Link to="/login" className="navbar-link">
            Login
          </Link>
          <span className="navbar-link" onClick={handleLogout}>
            Cerrar Sesión
          </span>
        </div>
      </div>
    </nav>
  );
}

export default Navbar;