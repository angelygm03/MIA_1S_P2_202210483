import React from "react";
import { Link } from "react-router-dom";
import "./Navbar.css";

function Navbar() {
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
          <Link to="/login" className="navbar-link">
            Login
          </Link>
          <Link to="/partitions" className="navbar-link">
            Particiones
          </Link>
        </div>
      </div>
    </nav>
  );
}

export default Navbar;