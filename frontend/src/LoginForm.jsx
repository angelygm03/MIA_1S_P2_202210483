import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './LoginForm.css';

function LoginForm() {
  const [id, setId] = useState('');
  const [user, setUser] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async () => {
    try {
      const response = await fetch('http://localhost:8080/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id, user, password }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        setError(`Error: ${errorText}`);
        return;
      }

      const message = await response.text();
      alert(message); // Mostrar mensaje de éxito
      navigate('/'); // Redirigir a la página principal
    } catch (err) {
      setError(`Error al iniciar sesión: ${err.message}`);
    }
  };

  return (
    <div className="login-form-container">
      <h2>Iniciar Sesión</h2>
      {error && <p className="error-message">{error}</p>}
      <div className="form-group">
        <label htmlFor="id">ID Partición</label>
        <input
          type="text"
          id="id"
          value={id}
          onChange={(e) => setId(e.target.value)}
          placeholder="Ingrese el ID de la partición"
        />
      </div>
      <div className="form-group">
        <label htmlFor="user">Usuario</label>
        <input
          type="text"
          id="user"
          value={user}
          onChange={(e) => setUser(e.target.value)}
          placeholder="Ingrese su usuario"
        />
      </div>
      <div className="form-group">
        <label htmlFor="password">Contraseña</label>
        <input
          type="password"
          id="password"
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