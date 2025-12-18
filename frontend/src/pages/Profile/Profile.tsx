import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../shared/store/hooks';
import { fetchUserProfile, updateUserProfile, clearError, logoutUser } from '../../shared/store/slices/authSlice';
import { clearFilters } from '../../shared/store/slices/filtersSlice';
import { clearOrdersState } from '../../shared/store/slices/ordersSlice';
import { clearUserDraft, resetDraftState } from '../../shared/store/slices/draftSlice';
import { LoadingSpinner } from '../../shared/components/LoadingSpinner/LoadingSpinner';

/**
 * Страница личного кабинета пользователя.
 * Позволяет просматривать и обновлять профиль, а также демонстрирует работу Redux Toolkit + axios.
 */
export function Profile() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { user, isLoading, error, isAuthenticated } = useAppSelector((state) => state.auth);

  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
  });

  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
    }
  }, [isAuthenticated, navigate]);

  useEffect(() => {
    if (isAuthenticated) {
      dispatch(fetchUserProfile());
    }
  }, [dispatch, isAuthenticated]);

  useEffect(() => {
    if (user) {
      setFormData({
        name: user.name || '',
        email: user.email || '',
        phone: user.phone || '',
      });
    }
  }, [user]);

  useEffect(() => {
    return () => {
      dispatch(clearError());
    };
  }, [dispatch]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleUpdateProfile = async () => {
    setMessage(null);
    const result = await dispatch(updateUserProfile(formData));
    if (updateUserProfile.fulfilled.match(result)) {
      setIsEditing(false);
      setMessage('Профиль успешно обновлен');
    }
  };

  const handleLogout = async () => {
    // Сначала пытаемся очистить черновик на сервере, пока токен ещё валиден
    await dispatch(clearUserDraft());
    // Сбрасываем UI-состояния согласно требованию лаб7
    dispatch(clearFilters());
    dispatch(clearOrdersState());
    dispatch(resetDraftState());
    await dispatch(logoutUser());
    navigate('/');
  };

  if (!isAuthenticated) {
    return null;
  }

  if (isLoading && !user) {
    return <LoadingSpinner text="Загрузка профиля..." />;
  }

  if (!user) {
    return (
      <div className="container" style={{ margin: '2rem auto', textAlign: 'center' }}>
        <p>Профиль не найден</p>
      </div>
    );
  }

  return (
    <div className="container" style={{ margin: '2rem auto', maxWidth: '600px' }}>
      <h2 style={{ marginBottom: '2rem' }}>Личный кабинет</h2>

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

      {message && (
        <div style={{
          background: '#d1e7dd',
          color: '#0f5132',
          padding: '1rem',
          borderRadius: '4px',
          marginBottom: '2rem',
        }}>
          {message}
        </div>
      )}

      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: '2rem',
      }}>
        <h3 style={{ marginBottom: '1rem' }}>Профиль</h3>

        {!isEditing ? (
          <div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>Логин</label>
              <p>{user.login}</p>
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>Имя</label>
              <p>{user.name}</p>
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>Email</label>
              <p>{user.email}</p>
            </div>
            {user.phone && (
              <div style={{ marginBottom: '1rem' }}>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: 'bold' }}>Телефон</label>
                <p>{user.phone}</p>
              </div>
            )}

            <button
              onClick={() => setIsEditing(true)}
              style={{
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
        ) : (
          <div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Имя *</label>
              <input
                type="text"
                name="name"
                value={formData.name}
                onChange={handleInputChange}
                required
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                }}
              />
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Email *</label>
              <input
                type="email"
                name="email"
                value={formData.email}
                onChange={handleInputChange}
                required
                style={{
                  width: '100%',
                  padding: '0.5rem',
                  border: '1px solid #ddd',
                  borderRadius: '4px',
                }}
              />
            </div>
            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem' }}>Телефон</label>
              <input
                type="tel"
                name="phone"
                value={formData.phone}
                onChange={handleInputChange}
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
                onClick={handleUpdateProfile}
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
                  setFormData({
                    name: user.name || '',
                    email: user.email || '',
                    phone: user.phone || '',
                  });
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
        )}
      </div>

      <div style={{
        backgroundColor: 'white',
        padding: '1.5rem',
        borderRadius: '8px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
      }}>
        <h3 style={{ marginBottom: '1rem' }}>Аккаунт</h3>
        <button
          type="button"
          onClick={handleLogout}
          disabled={isLoading}
          aria-disabled={isLoading}
          style={{
            padding: '0.75rem 1.25rem',
            backgroundColor: '#dc3545',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            opacity: isLoading ? 0.7 : 1,
          }}
        >
          {isLoading ? 'Выходим...' : '🚪 Выход'}
        </button>
      </div>
    </div>
  );
}
