import { Configuration } from './generated/configuration';
import { AuthApi } from './generated/apis/auth-api';
import { LogisticRequestsApi } from './generated/apis/logistic-requests-api';
import { httpClient } from './httpClient';
import { API_BASE_URL } from '../config/apiConfig';

/**
 * Пример использования сгенерированного (swagger → openapi-generator → typescript-axios) кода.
 *
 * На защите можно показать:
 * - папку `src/shared/api/generated` (сгенерирована из `docs/swagger.json`)
 * - как мы подключаем её к нашему axios instance
 * - как Bearer-токен берётся из localStorage (см. httpClient.ts)
 */
const configuration = new Configuration({
  basePath: API_BASE_URL || '',
  apiKey: async (name: string) => {
    // В swagger securityDefinitions.apikey: name=Authorization
    if (name === 'Authorization') {
      const token = localStorage.getItem('access_token');
      return token ? `Bearer ${token}` : '';
    }
    return '';
  },
});

export const logisticRequestsApi = new LogisticRequestsApi(
  configuration,
  API_BASE_URL || undefined,
  httpClient,
);

export const authApi = new AuthApi(
  configuration,
  API_BASE_URL || undefined,
  httpClient,
);
