import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Cart = () => {
  const [cartItems, setCartItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const fetchCart = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/cart');
      
      if (response.status === 401) {
        navigate('/login');
        return;
      }
      
      if (!response.ok) {
        throw new Error('Не удалось загрузить корзину');
      }
      
      const data = await response.json();
      setCartItems(data);
    } catch (err) {
      setError(err.message);
      console.error('Ошибка при загрузке корзины:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCart();
  }, []);

  const handleCheckout = () => {
    navigate('/checkout');
  };

  if (loading) {
    return <div className="card">Загрузка корзины...</div>;
  }

  if (error) {
    return <div className="card" style={{ color: 'var(--error-color)' }}>Ошибка: {error}</div>;
  }

  const totalPrice = cartItems.reduce((sum, item) => sum + parseFloat(item.price || 0), 0);

  return (
    <div>
      <h2 style={{ marginBottom: '20px', textAlign: 'center' }}>Корзина</h2>
      
      {cartItems.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', padding: '30px' }}>
          <p>Ваша корзина пуста</p>
          <button onClick={() => navigate('/products')} style={{ marginTop: '15px' }}>
            Перейти к товарам
          </button>
        </div>
      ) : (
        <div>
          <div>
            {cartItems.map((item, index) => (
              <div key={index} className="card" style={{ 
                display: 'flex',
                alignItems: 'center',
                gap: '20px'
              }}>
                <img 
                  src={item.image_url} 
                  alt={item.name} 
                  style={{ 
                    width: '100px', 
                    height: '100px', 
                    objectFit: 'cover',
                    borderRadius: '6px'
                  }} 
                />
                <div style={{ flex: 1 }}>
                  <h3 style={{ margin: '0 0 10px 0' }}>{item.name}</h3>
                  <p style={{ 
                    color: 'var(--light-text)',
                    fontSize: '14px',
                    margin: '0 0 10px 0'
                  }}>
                    {item.description}
                  </p>
                </div>
                <div style={{ 
                  fontWeight: 'bold', 
                  fontSize: '18px',
                  minWidth: '80px',
                  textAlign: 'right'
                }}>
                  {item.price} ₽
                </div>
              </div>
            ))}
          </div>
          
          <div className="card" style={{ 
            marginTop: '20px',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center'
          }}>
            <div>
              <p style={{ margin: '0', fontSize: '16px' }}>Итого:</p>
              <p style={{ 
                margin: '5px 0 0 0', 
                fontSize: '24px', 
                fontWeight: 'bold' 
              }}>
                {totalPrice.toFixed(2)} ₽
              </p>
            </div>
            <button onClick={handleCheckout}>Оформить заказ</button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Cart;