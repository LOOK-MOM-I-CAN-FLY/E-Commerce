import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const Products = () => {
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [addingToCart, setAddingToCart] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/products');
        if (!response.ok) {
          throw new Error('Не удалось загрузить товары');
        }
        const data = await response.json();
        setProducts(data);
      } catch (err) {
        setError(err.message);
        console.error('Ошибка при загрузке товаров:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchProducts();
  }, []);

  const addToCart = async (productId) => {
    try {
      setAddingToCart(productId);
      const response = await fetch('/api/cart/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ product_id: productId })
      });
      
      if (response.status === 401) {
        // Пользователь не аутентифицирован, перенаправляем на страницу входа
        navigate('/login');
        return;
      }
      
      if (!response.ok) {
        throw new Error('Не удалось добавить товар в корзину');
      }
      
      const text = await response.text();
      alert(text);
    } catch (err) {
      console.error('Ошибка при добавлении в корзину:', err);
      alert('Ошибка: ' + err.message);
    } finally {
      setAddingToCart(null);
    }
  };

  if (loading) {
    return <div className="card">Загрузка товаров...</div>;
  }

  if (error) {
    return <div className="card" style={{ color: 'var(--error-color)' }}>Ошибка: {error}</div>;
  }

  return (
    <div>
      <h2 style={{ marginBottom: '20px', textAlign: 'center' }}>Фреймворки программирования</h2>
      <div className="product-grid">
        {products.length === 0 ? (
          <div className="card">Товары отсутствуют</div>
        ) : (
          products.map(product => (
            <div key={product.id} className="card" style={{ 
              display: 'flex', 
              flexDirection: 'column',
              height: '100%'
            }}>
              <img 
                src={product.image_url} 
                alt={product.name} 
                style={{ 
                  maxWidth: '100%', 
                  maxHeight: '200px', 
                  objectFit: 'cover',
                  borderRadius: '6px',
                  marginBottom: '10px'
                }} 
              />
              <h3 style={{ margin: '10px 0' }}>{product.name}</h3>
              <p style={{ 
                flex: '1', 
                color: 'var(--light-text)',
                fontSize: '14px',
                marginBottom: '15px'
              }}>
                {product.description}
              </p>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <p style={{ 
                  fontWeight: 'bold', 
                  color: 'var(--text-color)',
                  fontSize: '18px',
                  margin: '0'
                }}>
                  {product.price} ₽
                </p>
                <button 
                  onClick={() => addToCart(product.id)}
                  disabled={addingToCart === product.id}
                  style={{ opacity: addingToCart === product.id ? 0.7 : 1 }}
                >
                  {addingToCart === product.id ? 'Добавляем...' : 'В корзину'}
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Products;