import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../shared/store/hooks';
import { fetchOrders } from '../../shared/store/slices/ordersSlice';
import { LoadingSpinner } from '../../shared/components/LoadingSpinner/LoadingSpinner';

/**
 * Страница списка заявок пользователя в виде таблицы
 * Использует Redux Toolkit для управления состоянием заявок
 */
export function OrdersList() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { orders, isLoading, error } = useAppSelector((state) => state.orders);
  const { isAuthenticated } = useAppSelector((state) => state.auth);

  const [statusFilter, setStatusFilter] = useState<string>('');
  const [dateFrom, setDateFrom] = useState<string>('');
  const [dateTo, setDateTo] = useState<string>('');

  // Перенаправляем неавторизованных пользователей
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  // Загружаем заявки при монтировании и изменении фильтров
  useEffect(() => {
    if (isAuthenticated) {
      const params: any = {};
      if (statusFilter) params.status = statusFilter;
      if (dateFrom) params.date_from = dateFrom;
      if (dateTo) params.date_to = dateTo;
      dispatch(fetchOrders(params));
    }
  }, [dispatch, isAuthenticated, statusFilter, dateFrom, dateTo]);

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      draft: 'Черновик',
      formed: 'Сформирован',
      completed: 'Завершен',
      rejected: 'Отклонен',
      deleted: 'Удален',
    };
    return labels[status] || status;
  };

  const getStatusColor = (status: string) => {
    const colors: Record<string, string> = {
      draft: '#6c757d',
      formed: '#0d6efd',
      completed: '#198754',
      rejected: '#dc3545',
      deleted: '#6c757d',
    };
    return colors[status] || '#6c757d';
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="container" style={{ margin: '2rem auto' }}>
      <h2 style={{ marginBottom: '2rem' }}>Мои заявки</h2>

      {/* Фильтры */}
      <div style={{
        display: 'flex',
        gap: '1rem',
        marginBottom: '2rem',
        flexWrap: 'wrap',
        alignItems: 'flex-end',
      }}>
        <div>
          <label htmlFor="status" style={{ display: 'block', marginBottom: '0.5rem' }}>
            Статус
          </label>
          <select
            id="status"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            style={{
              padding: '0.5rem',
              border: '1px solid #ddd',
              borderRadius: '4px',
            }}
          >
            <option value="">Все</option>
            <option value="draft">Черновик</option>
            <option value="formed">Сформирован</option>
            <option value="completed">Завершен</option>
            <option value="rejected">Отклонен</option>
          </select>
        </div>

        <div>
          <label htmlFor="dateFrom" style={{ display: 'block', marginBottom: '0.5rem' }}>
            Дата от
          </label>
          <input
            id="dateFrom"
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            style={{
              padding: '0.5rem',
              border: '1px solid #ddd',
              borderRadius: '4px',
            }}
          />
        </div>

        <div>
          <label htmlFor="dateTo" style={{ display: 'block', marginBottom: '0.5rem' }}>
            Дата до
          </label>
          <input
            id="dateTo"
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            style={{
              padding: '0.5rem',
              border: '1px solid #ddd',
              borderRadius: '4px',
            }}
          />
        </div>

        <button
          onClick={() => {
            setStatusFilter('');
            setDateFrom('');
            setDateTo('');
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
          Сбросить
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

      {isLoading ? (
        <LoadingSpinner text="Загрузка заявок..." />
      ) : (
        <>
          {orders.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '2rem' }}>
              <p>Заявки не найдены</p>
            </div>
          ) : (
            <div style={{ overflowX: 'auto' }}>
              <table style={{
                width: '100%',
                borderCollapse: 'collapse',
                backgroundColor: 'white',
                boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
              }}>
                <thead>
                  <tr style={{ backgroundColor: '#f8f9fa' }}>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>ID</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Статус</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Маршрут</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Услуг</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Стоимость</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Создано</th>
                    <th style={{ padding: '1rem', textAlign: 'left', borderBottom: '2px solid #dee2e6' }}>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map((order) => (
                    <tr key={order.id} style={{ borderBottom: '1px solid #dee2e6' }}>
                      <td style={{ padding: '1rem' }}>{order.id}</td>
                      <td style={{ padding: '1rem' }}>
                        <span style={{
                          padding: '0.25rem 0.5rem',
                          borderRadius: '4px',
                          backgroundColor: getStatusColor(order.status),
                          color: 'white',
                          fontSize: '0.875rem',
                        }}>
                          {getStatusLabel(order.status)}
                        </span>
                      </td>
                      <td style={{ padding: '1rem' }}>
                        {order.from_city && order.to_city ? (
                          `${order.from_city} → ${order.to_city}`
                        ) : (
                          'Не указано'
                        )}
                      </td>
                      <td style={{ padding: '1rem' }}>{order.services?.length || 0}</td>
                      <td style={{ padding: '1rem' }}>
                        {order.total_cost ? `${order.total_cost.toLocaleString('ru-RU')} ₽` : '-'}
                      </td>
                      <td style={{ padding: '1rem' }}>{formatDate(order.created_at)}</td>
                      <td style={{ padding: '1rem' }}>
                        <button
                          onClick={() => navigate(`/orders/${order.id}`)}
                          style={{
                            padding: '0.5rem 1rem',
                            backgroundColor: '#0d6efd',
                            color: 'white',
                            border: 'none',
                            borderRadius: '4px',
                            cursor: 'pointer',
                          }}
                        >
                          Просмотр
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}

