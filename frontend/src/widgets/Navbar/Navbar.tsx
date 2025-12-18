import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../../shared/store/hooks';
import { logoutUser } from '../../shared/store/slices/authSlice';
import { clearFilters } from '../../shared/store/slices/filtersSlice';
import { clearOrdersState } from '../../shared/store/slices/ordersSlice';
import { clearUserDraft, resetDraftState } from '../../shared/store/slices/draftSlice';

/**
 * Компонент навигационной панели
 * Использует существующие стили из style.css (header, logo, home-btn)
 * 
 * Props: не требуются (использует useLocation из react-router-dom для определения активной страницы)
 */
export function Navbar() {
  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { isAuthenticated, user, isLoading } = useAppSelector((state) => state.auth);

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

  return (
    <header className="header">
      <Link to="/" className="logo">
        <div className="logo-icon">🚚</div>
        GruzDelivery
      </Link>
      <div className="header-actions">
        {/* Кнопки навигации */}
        {location.pathname !== '/' && (
          <Link to="/" className="home-btn">🏠 Главная</Link>
        )}
        {location.pathname !== '/transport-services' && (
          <Link to="/transport-services" className="home-btn">📦 Услуги</Link>
        )}
        {location.pathname !== '/about' && (
          <Link to="/about" className="home-btn">ℹ️ О компании</Link>
        )}

        {/* Меню пользователя */}
        {isAuthenticated ? (
          <>
            {location.pathname !== '/orders' && (
              <Link to="/orders" className="home-btn">📋 Мои заявки</Link>
            )}
            {location.pathname !== '/profile' && (
              <Link to="/profile" className="home-btn">👤 ЛК</Link>
            )}
            <span className="home-btn" style={{ cursor: 'default', opacity: 0.9 }}>
              {user?.name || user?.login || 'Пользователь'}
            </span>
            <button
              type="button"
              className="home-btn"
              onClick={handleLogout}
              disabled={isLoading}
              aria-disabled={isLoading}
            >
              {isLoading ? 'Выход...' : '🚪 Выход'}
            </button>
          </>
        ) : (
          <>
            {location.pathname !== '/login' && (
              <Link to="/login" className="home-btn">🔐 Вход</Link>
            )}
          </>
        )}
      </div>
    </header>
  );
}
