import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const Checkout = () => {
  const [email, setEmail] = useState('');
  const [name, setName] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const navigate = useNavigate();

  const handleCheckout = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    
    try {
      const response = await fetch('/api/checkout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, name })
      });
      
      if (response.status === 401) {
        navigate('/login');
        return;
      }
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Не удалось оформить заказ');
      }
      
      setSuccess(true);
      // Очистить форму после успешного оформления
      setEmail('');
      setName('');
    } catch (err) {
      setError(err.message);
      console.error('Ошибка при оформлении заказа:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: '20px', textAlign: 'center' }}>Оформление заказа</h2>
      
      {success ? (
        <div className="card" style={{ 
          backgroundColor: 'var(--success-color)', 
          padding: '30px',
          textAlign: 'center' 
        }}>
          <h3 style={{ margin: '0 0 15px 0', color: '#2a5738' }}>Заказ успешно оформлен!</h3>
          <p style={{ color: '#2a5738' }}>Фреймворки будут отправлены на указанный email.</p>
          <button 
            onClick={() => navigate('/products')} 
            style={{ marginTop: '20px', backgroundColor: 'white', color: 'var(--text-color)' }}
          >
            Вернуться к покупкам
          </button>
        </div>
      ) : (
        <div className="card">
          {error && (
            <div style={{ 
              backgroundColor: 'var(--error-color)', 
              padding: '10px',
              borderRadius: '4px',
              marginBottom: '20px',
              color: '#721c24'
            }}>
              {error}
            </div>
          )}
          
          <form onSubmit={handleCheckout}>
            <div style={{ marginBottom: '20px' }}>
              <label htmlFor="name" style={{ 
                display: 'block', 
                marginBottom: '8px', 
                fontWeight: '500' 
              }}>
                Ваше имя:
              </label>
              <input 
                type="text" 
                id="name"
                value={name} 
                onChange={e => setName(e.target.value)} 
                required 
                style={{ width: '100%' }}
              />
            </div>
            
            <div style={{ marginBottom: '20px' }}>
              <label htmlFor="email" style={{ 
                display: 'block', 
                marginBottom: '8px', 
                fontWeight: '500' 
              }}>
                Email для отправки фреймворков:
              </label>
              <input 
                type="email" 
                id="email"
                value={email} 
                onChange={e => setEmail(e.target.value)} 
                required 
                style={{ width: '100%' }}
              />
              <p style={{ 
                margin: '10px 0 0 0', 
                fontSize: '14px', 
                color: 'var(--light-text)' 
              }}>
                На этот адрес будут отправлены приобретенные фреймворки в виде фотографий
              </p>
            </div>
            
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '30px' }}>
              <button 
                type="button" 
                onClick={() => navigate('/cart')}
                style={{ 
                  backgroundColor: 'var(--secondary-bg)', 
                  color: 'var(--text-color)' 
                }}
              >
                Назад в корзину
              </button>
              <button 
                type="submit" 
                disabled={isSubmitting}
                style={{ opacity: isSubmitting ? 0.7 : 1 }}
              >
                {isSubmitting ? 'Оформляем...' : 'Оформить заказ'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
};

export default Checkout;