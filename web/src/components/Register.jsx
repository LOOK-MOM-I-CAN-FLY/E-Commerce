import React, { useState } from 'react';

const Register = () => {
  const [email, setEmail]     = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm]   = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    const response = await fetch('/api/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, confirm })
    });
    const text = await response.text();
    alert(text);
  };

  return (
    <div>
      <h2>Регистрация</h2>
      <form onSubmit={handleSubmit}>
        <label>Email:</label>
        <br />
        <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
        <br />
        <label>Пароль:</label>
        <br />
        <input type="password" value={password} onChange={e => setPassword(e.target.value)} required />
        <br />
        <label>Подтверждение пароля:</label>
        <br />
        <input type="password" value={confirm} onChange={e => setConfirm(e.target.value)} required />
        <br />
        <button type="submit">Зарегистрироваться</button>
      </form>
    </div>
  );
};

export default Register;