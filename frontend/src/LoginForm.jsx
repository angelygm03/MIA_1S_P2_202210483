import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './LoginForm.css';

function LoginForm() {
  const [id, setId] = useState('');
  const [user, setUser] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleLogin = () => {
    console.log('ID:', id);
    console.log('Usuario:', user);
    console.log('Contraseña:', password);
    alert(`Iniciando sesión con ID: ${id}, Usuario: ${user}`);
    navigate('/'); // Redirige a la página principal después del login
  };

  return (
    <div className="login-form-container">
      <h2>Iniciar Sesión</h2>
      <div className="form-group">
        <label>ID Partición:</label>
        <input
          type="text"
          value={id}
          onChange={(e) => setId(e.target.value)}
          placeholder="Ingrese el ID de la partición"
        />
      </div>
      <div className="form-group">
        <label>Usuario:</label>
        <input
          type="text"
          value={user}
          onChange={(e) => setUser(e.target.value)}
          placeholder="Ingrese su usuario"
        />
      </div>
      <div className="form-group">
        <label>Contraseña:</label>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Ingrese su contraseña"
        />
      </div>
      <div className="form-buttons">
        <button onClick={handleLogin}>Iniciar Sesión</button>
        <button onClick={() => navigate('/')}>Cancelar</button>
      </div>
    </div>
  );
}

export default LoginForm;