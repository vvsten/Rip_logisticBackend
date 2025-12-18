import axios from 'axios';
import { API_BASE_URL } from '../config/apiConfig';

/**
 * Единый HTTP-клиент приложения.
 *
 * Почему именно так (для защиты на лабе):
 * - **axios** используется для всех запросов к API.
 * - Bearer-токен берём из **localStorage** и автоматически добавляем в заголовок Authorization.
 * - В WEB режиме baseURL пустой, поэтому запросы идут через Vite proxy (см. vite.config.ts).
 * - В Tauri режиме baseURL = IP сервера из apiConfig.ts.
 */
export const httpClient = axios.create({
  baseURL: API_BASE_URL || '',
  withCredentials: true,
});

httpClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers = config.headers ?? {};
    // Формат: "Bearer <token>"
    (config.headers as any).Authorization = `Bearer ${token}`;
  }
  return config;
});

/**
 * Авто-refresh access токена при 401 (JWT-only auth).
 * Нужно, чтобы UI не ломался сообщением "Invalid token", когда access_token истёк.
 *
 * Важно: /refresh проксируется Vite'ом на бэкенд в dev (см. vite.config.ts).
 */
const refreshClient = axios.create({
  baseURL: API_BASE_URL || '',
  withCredentials: true,
});

httpClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error?.config as any;
    const status = error?.response?.status;

    // Не зацикливаемся
    if (status !== 401 || originalRequest?._retry) {
      return Promise.reject(error);
    }

    // Не пытаемся рефрешить сам рефреш
    if (typeof originalRequest?.url === 'string' && originalRequest.url.includes('/refresh')) {
      return Promise.reject(error);
    }

    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      const res = await refreshClient.post('/refresh', { refresh_token: refreshToken });
      const newAccessToken = res?.data?.access_token;
      const newRefreshToken = res?.data?.refresh_token;
      const expiresAt = res?.data?.expires_at;
      const user = res?.data?.user;

      if (newAccessToken) localStorage.setItem('access_token', newAccessToken);
      if (newRefreshToken) localStorage.setItem('refresh_token', newRefreshToken);
      if (expiresAt) localStorage.setItem('expires_at', expiresAt);
      if (user) localStorage.setItem('user', JSON.stringify(user));

      originalRequest.headers = originalRequest.headers ?? {};
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;

      return httpClient(originalRequest);
    } catch (refreshErr) {
      return Promise.reject(refreshErr);
    }
  },
);
