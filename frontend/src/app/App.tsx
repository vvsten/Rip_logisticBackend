import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { Provider } from 'react-redux';
import { store } from '../shared/store/store';
import { AuthInitializer } from '../shared/components/AuthInitializer/AuthInitializer';
import { ProtectedRoute } from '../shared/components/ProtectedRoute/ProtectedRoute';
import { Navbar } from '../widgets/Navbar/Navbar';
import { CalculatorShortcut } from '../widgets/Cart/CalculatorShortcut';
import { Breadcrumbs } from '../widgets/Breadcrumbs/Breadcrumbs';
import { ServerConfig } from '../widgets/ServerConfig/ServerConfig';
import { Home } from '../pages/Home/Home';
import { Services } from '../pages/Services/Services';
import { About } from '../pages/About/About';
import { Login } from '../pages/Login/Login';
import { Register } from '../pages/Register/Register';
import { OrdersList } from '../pages/OrdersList/OrdersList';
import { OrderDetails } from '../pages/OrderDetails/OrderDetails';
import { Profile } from '../pages/Profile/Profile';
import '../css/style.css';

/**
 * Главный компонент приложения
 * 
 * Настраивает роутинг, подключает глобальные компоненты (Navbar, Breadcrumbs)
 * Использует BrowserRouter для SPA навигации
 * Подключает Redux Provider для управления состоянием
 * Подключает существующие стили из style.css
 */
export function App() {
  return (
    <Provider store={store}>
    <BrowserRouter>
        {/* Восстанавливаем auth из localStorage при старте приложения */}
        <AuthInitializer />
        {/* Компонент настройки сервера для Tauri (отображается только в Tauri) */}
        <ServerConfig />
        
      {/* Навигационная панель - всегда вверху */}
      <Navbar />
      
      {/* Иконка калькулятора под хедером */}
      <CalculatorShortcut />
      
      {/* Навигационная цепочка - отображается на нужных страницах */}
      <Breadcrumbs />
      
      {/* Маршруты для страниц */}
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/transport-services" element={<Services />} />
        <Route path="/about" element={<About />} />

        {/* Auth */}
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        {/* User */}
        <Route
          path="/logistic-requests"
          element={(
            <ProtectedRoute>
              <OrdersList />
            </ProtectedRoute>
          )}
        />
        <Route
          path="/logistic-requests/:id"
          element={(
            <ProtectedRoute>
              <OrderDetails />
            </ProtectedRoute>
          )}
        />
        <Route
          path="/profile"
          element={(
            <ProtectedRoute>
              <Profile />
            </ProtectedRoute>
          )}
        />
      </Routes>
    </BrowserRouter>
    </Provider>
  );
}