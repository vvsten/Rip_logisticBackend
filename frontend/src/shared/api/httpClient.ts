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
