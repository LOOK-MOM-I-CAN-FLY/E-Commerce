import React, { useState } from 'react';

const Login = () => {
  const [email, setEmail]     = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    const response = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const text = await response.text();
    alert(text);
  };

  return (
    <div>
      <h2>Вход</h2>
      <form onSubmit={handleSubmit}>
        <label>Email:</label>
        <br />
        <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
        <br />
        <label>Пароль:</label>
        <br />
        <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
        <br />
        <button type="submit">Войти</button>
      </form>
    </div>
  );
};

export default Login;