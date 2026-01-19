/* global axios */
(function () {
  if (!window.axios) {
    // eslint-disable-next-line no-console
    console.error('axios is not loaded. Add <script src="...axios.min.js"></script> before /static/js/http.js');
    return;
  }

  const AUTH_STORAGE_KEYS = {
    accessToken: 'access_token',
    refreshToken: 'refresh_token',
    expiresAt: 'expires_at',
    user: 'user',
  };

  function getAccessToken() {
    return localStorage.getItem(AUTH_STORAGE_KEYS.accessToken);
  }

  function getRefreshToken() {
    return localStorage.getItem(AUTH_STORAGE_KEYS.refreshToken);
  }

  function applyAuthResponse(data) {
    if (!data) return;
    if (data.access_token) localStorage.setItem(AUTH_STORAGE_KEYS.accessToken, data.access_token);
    if (data.refresh_token) localStorage.setItem(AUTH_STORAGE_KEYS.refreshToken, data.refresh_token);
    if (data.expires_at) localStorage.setItem(AUTH_STORAGE_KEYS.expiresAt, data.expires_at);
    if (data.user) localStorage.setItem(AUTH_STORAGE_KEYS.user, JSON.stringify(data.user));
  }

  const bare = window.axios.create({
    baseURL: '',
    timeout: 30000,
    withCredentials: true,
  });

  async function refreshAccessTokenIfPossible() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return null;

    try {
      const res = await bare.post('/refresh', { refresh_token: refreshToken });
      const data = res?.data;
      if (!data?.access_token) return null;
      applyAuthResponse(data);
      return data.access_token;
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error('refresh token failed:', e);
      return null;
    }
  }

  async function getValidAccessToken() {
    const token = getAccessToken();
    if (token) return token;
    return await refreshAccessTokenIfPossible();
  }

  const http = window.axios.create({
    baseURL: '',
    timeout: 30000,
    withCredentials: true,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  http.interceptors.request.use(
    (config) => {
      const accessToken = getAccessToken();
      if (accessToken) {
        config.headers = config.headers || {};
        config.headers.Authorization = `Bearer ${accessToken}`;
      }
      return config;
    },
    (error) => Promise.reject(error),
  );

  http.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error?.config;
      const status = error?.response?.status;
      const url = originalRequest?.url || '';

      // Не зацикливаемся на /refresh и не повторяем бесконечно.
      if (status === 401 && originalRequest && !originalRequest._retry && url !== '/refresh') {
        originalRequest._retry = true;
        const newToken = await refreshAccessTokenIfPossible();
        if (newToken) {
          originalRequest.headers = originalRequest.headers || {};
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return http.request(originalRequest);
        }
      }

      return Promise.reject(error);
    },
  );

  // Экспортируем минимум, чтобы шаблоны были простыми.
  window.http = http;
  window.auth = {
    getAccessToken,
    getRefreshToken,
    refreshAccessTokenIfPossible,
    getValidAccessToken,
  };
})();

