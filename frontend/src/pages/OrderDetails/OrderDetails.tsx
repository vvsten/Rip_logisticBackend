import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../shared/store/hooks';
import {
  fetchOrder,
  updateOrder,
  removeServiceFromOrder,
  updateServiceInOrder,
  formOrder,
} from '../../shared/store/slices/ordersSlice';
import { LoadingSpinner } from '../../shared/components/LoadingSpinner/LoadingSpinner';

/**
 * Страница просмотра и редактирования заявки
 * В статусе черновик можно редактировать, в других статусах - только просмотр
 * Использует Redux Toolkit для управления состоянием заявок
 */
export function OrderDetails() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { currentOrder, isLoading, error } = useAppSelector((state) => state.orders);
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    from_city: '',
    to_city: '',
    weight: 0,
    length: 0,
    width: 0,
    height: 0,
  });

  // Перенаправляем неавторизованных пользователей
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  // Загружаем заявку при монтировании
  useEffect(() => {
    if (id && isAuthenticated) {
      dispatch(fetchOrder(Number(id)));
    }
  }, [dispatch, id, isAuthenticated]);

  // Обновляем форму при загрузке заявки
  useEffect(() => {
    if (currentOrder) {
      setFormData({
        from_city: currentOrder.from_city || '',
        to_city: currentOrder.to_city || '',
        weight: currentOrder.weight || 0,
        length: currentOrder.length || 0,
        width: currentOrder.width || 0,
        height: currentOrder.height || 0,
      });
    }
  }, [currentOrder]);

  const isDraft = currentOrder?.status === 'draft';
  const canEdit = isDraft && isAuthenticated;

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: name === 'from_city' || name === 'to_city' ? value : parseFloat(value) || 0,
    });
  };

  const handleUpdateOrder = async () => {
    if (!id) return;
    await dispatch(updateOrder({ id: Number(id), data: formData }));
    setIsEditing(false);
  };

  const handleRemoveService = async (transportServiceId: number) => {
    if (!id) return;
    await dispatch(removeServiceFromOrder({ orderId: Number(id), transportServiceId }));
  };

  const handleUpdateServiceQuantity = async (transportServiceId: number, quantity: number) => {
    if (!id || quantity < 1) return;
    await dispatch(updateServiceInOrder({ orderId: Number(id), transportServiceId, quantity }));
  };

  const handleFormOrder = async () => {
    if (!id || !currentOrder) return;
    
    // Проверяем, что все необходимые данные заполнены
    if (!formData.from_city || !formData.to_city || formData.weight <= 0 || 
        formData.length <= 0 || formData.width <= 0 || formData.height <= 0) {
      alert('Заполните все поля для формирования заявки');
      return;
    }
    
    const result = await dispatch(formOrder({
      id: Number(id),
      data: {
        from_city: formData.from_city,
        to_city: formData.to_city,
        weight: formData.weight,
        length: formData.length,
        width: formData.width,
        height: formData.height,
      },
    }));
    if (formOrder.fulfilled.match(result)) {
      navigate('/logistic-requests');
    }
  };

  if (!isAuthenticated) {
    return null;
  }

  if (isLoading && !currentOrder) {
    return <LoadingSpinner text="Загрузка заявки..." />;
  }

  if (!currentOrder) {
    return (
      <div className="container" style={{ margin: '2rem auto', textAlign: 'center' }}>
        <p>Заявка не найдена</p>
        <button
          onClick={() => navigate('/logistic-requests')}
          style={{
            marginTop: '1rem',
            padding: '0.5rem 1rem',
            backgroundColor: '#0d6efd',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Вернуться к списку
        </button>
      </div>
    );
  }

  return (
    <div className="container" style={{ margin: '2rem auto', maxWidth: '800px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h2>Заявка #{currentOrder.id}</h2>
        <button
          onClick={() => navigate('/logistic-requests')}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#6c757d',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
          }}
        >
          Назад к списку
        </button>
      </div>

      {error && (
        <div style={{
          background: '#f8d7da',
          color: '#721c24',
          padding: '1rem',
          borderRadius: '4px',
          marginBottom: '2rem',
        }}>
          {error}
        </div>
      )}

      {/* Информация о заявке */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem',
      }}>
        <h3 style={{ marginBottom: '1rem' }}>Информация о заявке</h3>
        
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Статус
            </label>
            <span style={{
              padding: '0.25rem 0.5rem',
              borderRadius: '4px',
              backgroundColor: isDraft ? '#6c757d' : '#198754',
              color: 'white',
            }}>
              {isDraft ? 'Черновик' : currentOrder.status}
            </span>
          </div>

          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>
              Стоимость
            </label>
            <p>{currentOrder.total_cost ? `${currentOrder.total_cost.toLocaleString('ru-RU')} ₽` : '-'}</p>
          </div>
        </div>

        {canEdit && !isEditing ? (
          <div>
            <p><strong>Город отправления:</strong> {formData.from_city || 'Не указано'}</p>
            <p><strong>Город назначения:</strong> {formData.to_city || 'Не указано'}</p>
            <p><strong>Вес:</strong> {formData.weight} кг</p>
            <p><strong>Размеры:</strong> {formData.length} × {formData.width} × {formData.height} м</p>
            <button
              onClick={() => setIsEditing(true)}
              style={{
                marginTop: '1rem',
                padding: '0.5rem 1rem',
                backgroundColor: '#0d6efd',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
              }}
            >
              Редактировать
            </button>
          </div>
        ) : canEdit && isEditing ? (
          <div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Город отправления</label>
              <input
                type="text"
                name="from_city"
                value={formData.from_city}
                onChange={handleInputChange}
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                }}
              />
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Город назначения</label>
              <input
                type="text"
                name="to_city"
                value={formData.to_city}
                onChange={handleInputChange}
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                }}
              />
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>Вес (кг)</label>
                <input
                  type="number"
                  name="weight"
                  value={formData.weight}
                  onChange={handleInputChange}
                  min="0"
                  style={{
                    width: '100%',
                    padding: '0.5rem',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>Длина (м)</label>
                <input
                  type="number"
                  name="length"
                  value={formData.length}
                  onChange={handleInputChange}
                  min="0"
                  style={{
                    width: '100%',
                    padding: '0.5rem',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem' }}>Ширина (м)</label>
                <input
                  type="number"
                  name="width"
                  value={formData.width}
                  onChange={handleInputChange}
                  min="0"
                  style={{
                    width: '100%',
                    padding: '0.5rem',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
              </div>
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Высота (м)</label>
              <input
                type="number"
                name="height"
                value={formData.height}
                onChange={handleInputChange}
                min="0"
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                }}
              />
            </div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <button
                onClick={handleUpdateOrder}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: '#198754',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                Сохранить
              </button>
              <button
                onClick={() => {
                  setIsEditing(false);
                  if (currentOrder) {
                    setFormData({
                      from_city: currentOrder.from_city || '',
                      to_city: currentOrder.to_city || '',
                      weight: currentOrder.weight || 0,
                      length: currentOrder.length || 0,
                      width: currentOrder.width || 0,
                      height: currentOrder.height || 0,
                    });
                  }
                }}
                style={{
                  padding: '0.5rem 1rem',
                  backgroundColor: '#6c757d',
                  color: 'white',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                }}
              >
                Отмена
              </button>
            </div>
          </div>
        ) : (
          <div>
            <p><strong>Город отправления:</strong> {currentOrder.from_city || 'Не указано'}</p>
            <p><strong>Город назначения:</strong> {currentOrder.to_city || 'Не указано'}</p>
            <p><strong>Вес:</strong> {currentOrder.weight} кг</p>
            <p><strong>Размеры:</strong> {currentOrder.length} × {currentOrder.width} × {currentOrder.height} м</p>
          </div>
        )}
      </div>

      {/* Услуги в заявке */}
      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem',
      }}>
        <h3 style={{ marginBottom: '1rem' }}>Услуги</h3>
        
        {currentOrder.services && currentOrder.services.length > 0 ? (
          <div>
            {currentOrder.services.map((service) => (
              <div
                key={service.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '1rem',
                  borderBottom: '1px solid #dee2e6',
                }}
              >
                <div style={{ flex: 1 }}>
                  <p style={{ fontWeight: 'bold', marginBottom: '0.5rem' }}>
                    {service.service?.name || `Услуга #${service.transport_service_id}`}
                  </p>
                  {service.service && (
                    <p style={{ color: '#6c757d', fontSize: '0.875rem' }}>
                      {service.service.description}
                    </p>
                  )}
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                  {canEdit ? (
                    <>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                        <button
                          onClick={() => handleUpdateServiceQuantity(service.transport_service_id, service.quantity - 1)}
                          disabled={service.quantity <= 1}
                          style={{
                            padding: '0.25rem 0.5rem',
                            backgroundColor: '#6c757d',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: service.quantity <= 1 ? 'not-allowed' : 'pointer',
                            opacity: service.quantity <= 1 ? 0.5 : 1,
                          }}
                        >
                          -
                        </button>
                        <span style={{ minWidth: '2rem', textAlign: 'center' }}>{service.quantity}</span>
                        <button
                          onClick={() => handleUpdateServiceQuantity(service.transport_service_id, service.quantity + 1)}
                          style={{
                            padding: '0.25rem 0.5rem',
                            backgroundColor: '#6c757d',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                          }}
                        >
                          +
                        </button>
                      </div>
                      <button
                        onClick={() => handleRemoveService(service.transport_service_id)}
                        style={{
                          padding: '0.5rem 1rem',
                          backgroundColor: '#dc3545',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer',
                        }}
                      >
                        Удалить
                      </button>
                    </>
                  ) : (
                    <span style={{ fontWeight: 'bold' }}>× {service.quantity}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p style={{ color: '#6c757d' }}>Услуги не добавлены</p>
        )}

        {canEdit && currentOrder.services && currentOrder.services.length > 0 && (
          <div style={{ marginTop: '1.5rem', textAlign: 'right' }}>
            <button
              onClick={handleFormOrder}
              style={{
                padding: '0.75rem 1.5rem',
                backgroundColor: '#198754',
                color: 'white',
                border: 'none',
                borderRadius: '4px',
                cursor: 'pointer',
                fontSize: '1rem',
              }}
            >
              Подтвердить заявку
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

