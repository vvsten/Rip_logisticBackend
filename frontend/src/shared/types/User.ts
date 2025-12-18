/**
 * Пользователь (формат JSON как приходит из Go-бэкенда).
 *
 * Ключевые моменты для защиты:
 * - объект пользователя сохраняем в localStorage вместе с access_token
 * - после перезагрузки страницы состояние восстанавливается (см. authSlice.restoreAuth)
 */
export interface User {
  id: number;
  uuid?: string;
  login: string;
  name: string;
  email: string;
  phone?: string;
  role: string; // buyer | manager | admin
  created_at?: string;
  updated_at?: string;
}
