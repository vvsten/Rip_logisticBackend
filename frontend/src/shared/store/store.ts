import { configureStore } from '@reduxjs/toolkit';
import filtersReducer from './slices/filtersSlice';
import authReducer from './slices/authSlice';
import ordersReducer from './slices/ordersSlice';
import draftReducer from './slices/draftSlice';

/**
 * Redux store для приложения
 * 
 * Использует Redux Toolkit для управления состоянием
 * Включает:
 * - filters: состояние фильтров услуг
 * - auth: состояние авторизации (user/token/isAuthenticated)
 * - orders: список заявок и текущая заявка
 * - draft: конструктор новой заявки (черновик)
 */
export const store = configureStore({
  reducer: {
    filters: filtersReducer,
    auth: authReducer,
    orders: ordersReducer,
    draft: draftReducer,
  },
  // Включаем Redux DevTools для отладки
  devTools: import.meta.env.DEV,
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

