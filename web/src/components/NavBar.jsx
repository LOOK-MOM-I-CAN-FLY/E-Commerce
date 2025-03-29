import React from 'react';
import { Link } from 'react-router-dom';

const NavBar = () => {
  return (
    <nav style={{ 
      backgroundColor: 'var(--secondary-bg)', 
      padding: '15px 0',
      borderBottom: '1px solid var(--border-color)'
    }}>
      <div style={{ 
        maxWidth: '1200px', 
        margin: '0 auto', 
        display: 'flex', 
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '0 20px'
      }}>
        <Link to="/" style={{ 
          fontSize: '24px', 
          fontWeight: 'bold',
          color: 'var(--text-color)',
          textDecoration: 'none'
        }}>
          FrameStore
        </Link>
        <div style={{ display: 'flex', gap: '20px' }}>
          <Link to="/products" style={{ color: 'var(--text-color)', textDecoration: 'none' }}>Товары</Link>
          <Link to="/cart" style={{ color: 'var(--text-color)', textDecoration: 'none' }}>Корзина</Link>
          <Link to="/login" style={{ color: 'var(--text-color)', textDecoration: 'none' }}>Вход</Link>
          <Link to="/register" style={{ color: 'var(--text-color)', textDecoration: 'none' }}>Регистрация</Link>
        </div>
      </div>
    </nav>
  );
};

export default NavBar;