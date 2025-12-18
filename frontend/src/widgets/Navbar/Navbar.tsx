import { Link, useLocation } from 'react-router-dom';
import { useAppSelector } from '../../shared/store/hooks';

/**
 * Компонент навигационной панели
 * Использует существующие стили из style.css (header, logo, home-btn)
 * 
 * Props: не требуются (использует useLocation из react-router-dom для определения активной страницы)
 */
export function Navbar() {
  const location = useLocation();
  const { isAuthenticated } = useAppSelector((state) => state.auth);

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
            {location.pathname !== '/logistic-requests' && (
              <Link to="/logistic-requests" className="home-btn">📋 Мои заявки</Link>
            )}
            {location.pathname !== '/profile' && (
              <Link to="/profile" className="home-btn">👤 ЛК</Link>
            )}
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
