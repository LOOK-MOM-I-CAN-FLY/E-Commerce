import React from 'react';
import { useNavigate } from 'react-router-dom';

const Home = () => {
  const navigate = useNavigate();

  return (
    <div style={{ textAlign: 'center', padding: '40px 20px' }}>
      <div style={{ 
        maxWidth: '800px', 
        margin: '0 auto',
        backgroundColor: 'white',
        borderRadius: '12px',
        padding: '40px',
        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.05)'
      }}>
        <h1 style={{ 
          fontSize: '36px', 
          color: 'var(--text-color)',
          marginBottom: '20px'
        }}>
          FrameStore
        </h1>
        
        <p style={{ 
          fontSize: '18px', 
          lineHeight: '1.6',
          color: 'var(--light-text)',
          marginBottom: '30px'
        }}>
          Добро пожаловать в наш магазин фреймворков для программирования! 
          Мы предлагаем уникальные фотографии популярных фреймворков, 
          которые будут отправлены на ваш email после покупки.
        </p>
        
        <div style={{ display: 'flex', justifyContent: 'center', gap: '20px' }}>
          <button 
            onClick={() => navigate('/products')} 
            style={{ 
              padding: '12px 30px', 
              fontSize: '18px' 
            }}
          >
            Смотреть фреймворки
          </button>
          
          <button 
            onClick={() => navigate('/login')} 
            style={{ 
              backgroundColor: 'var(--secondary-bg)', 
              color: 'var(--text-color)' 
            }}
          >
            Войти в аккаунт
          </button>
        </div>
      </div>
      
      <div style={{ 
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
        gap: '30px',
        margin: '40px auto',
        maxWidth: '1000px'
      }}>
        <div className="card">
          <h3 style={{ marginTop: '0' }}>Большой выбор</h3>
          <p>Мы предлагаем фотографии самых популярных фреймворков для различных языков программирования</p>
        </div>
        
        <div className="card">
          <h3 style={{ marginTop: '0' }}>Мгновенная доставка</h3>
          <p>После оформления заказа фотографии сразу отправляются на ваш email</p>
        </div>
        
        <div className="card">
          <h3 style={{ marginTop: '0' }}>Высокое качество</h3>
          <p>Все фотографии выполнены в отличном качестве и отображают сущность каждого фреймворка</p>
        </div>
      </div>
    </div>
  );
};

export default Home;