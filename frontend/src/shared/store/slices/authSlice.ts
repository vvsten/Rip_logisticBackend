import { createAsyncThunk, createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { User } from '../../types/User';
import { httpClient } from '../../api/httpClient';
import { authApi } from '../../api/generatedApi';
import type { ServiceAuthResponse, ServiceLoginRequest, ServiceRegisterRequest } from '../../api/generated/models';

/**
 * AuthState хранит **состояние интерфейса после авторизации**.
 *
 * Для защиты (контрольные вопросы):
 * - reducer: описывает, как меняется state (ниже в createSlice)
 * - store: подключает reducer'ы (см. shared/store/store.ts)
 * - middleware: thunk встроен в Redux Toolkit по умолчанию; createAsyncThunk создаёт thunk-экшены
 * - localStorage: persist токена/пользователя между перезагрузками
 */
interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  expiresAt: string | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  accessToken: null,
  refreshToken: null,
  expiresAt: null,
  isLoading: false,
  error: null,
};

function normalizeError(err: unknown): string {
  if (typeof err === 'string') return err;
  if (err && typeof err === 'object') {
    const anyErr = err as any;
    const msg = anyErr?.response?.data?.error || anyErr?.response?.data?.message;
    if (typeof msg === 'string' && msg.trim()) return msg;
    if (typeof anyErr?.message === 'string') return anyErr.message;
  }
  return 'Произошла ошибка';
}

function persistAuth(payload: { access_token?: string; refresh_token?: string; expires_at?: string; user?: User }) {
  if (payload.access_token) localStorage.setItem('access_token', payload.access_token);
  if (payload.refresh_token) localStorage.setItem('refresh_token', payload.refresh_token);
  if (payload.expires_at) localStorage.setItem('expires_at', payload.expires_at);
  if (payload.user) localStorage.setItem('user', JSON.stringify(payload.user));
}

function clearPersistedAuth() {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('expires_at');
  localStorage.removeItem('user');
}

export const loginUser = createAsyncThunk<
  ServiceAuthResponse,
  { login: string; password: string },
  { rejectValue: string }
>('auth/loginUser', async ({ login, password }, { rejectWithValue }) => {
  try {
    const req: ServiceLoginRequest = { login, password };
    const res = await authApi.loginPost({ request: req });
    persistAuth({
      access_token: res.data.access_token,
      refresh_token: res.data.refresh_token,
      expires_at: res.data.expires_at,
      user: res.data.user as unknown as User,
    });
    return res.data;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const registerUser = createAsyncThunk<
  ServiceAuthResponse,
  { login: string; email: string; name: string; password: string; phone?: string },
  { rejectValue: string }
>('auth/registerUser', async ({ login, email, name, password, phone }, { rejectWithValue }) => {
  try {
    const req: ServiceRegisterRequest = {
      login,
      email,
      name,
      password,
      phone,
      role: 'buyer',
    };
    const res = await authApi.signUpPost({ request: req });
    persistAuth({
      access_token: res.data.access_token,
      refresh_token: res.data.refresh_token,
      expires_at: res.data.expires_at,
      user: res.data.user as unknown as User,
    });
    return res.data;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const logoutUser = createAsyncThunk<void, void, { rejectValue: string }>(
  'auth/logoutUser',
  async (_, { rejectWithValue }) => {
    try {
      // Токен подставится автоматически (httpClient interceptor) + через apiKey в generatedApi
      await authApi.logoutPost();
      clearPersistedAuth();
    } catch (e) {
      // Даже если запрос не удался (например, токен истёк), UI всё равно должен выйти.
      clearPersistedAuth();
      return rejectWithValue(normalizeError(e));
    }
  },
);

export const restoreAuth = createAsyncThunk<
  { accessToken: string; refreshToken: string | null; expiresAt: string | null; user: User | null } | null,
  void
>('auth/restoreAuth', async () => {
  const accessToken = localStorage.getItem('access_token');
  if (!accessToken) return null;

  const refreshToken = localStorage.getItem('refresh_token');
  const expiresAt = localStorage.getItem('expires_at');

  const rawUser = localStorage.getItem('user');
  let user: User | null = null;
  if (rawUser) {
    try {
      user = JSON.parse(rawUser) as User;
    } catch {
      user = null;
    }
  }

  return { accessToken, refreshToken, expiresAt, user };
});

export const fetchUserProfile = createAsyncThunk<User, void, { rejectValue: string }>(
  'auth/fetchUserProfile',
  async (_, { rejectWithValue }) => {
    try {
      const res = await httpClient.get('/api/users/profile');
      const user = (res.data?.user ?? null) as User | null;
      if (!user) return rejectWithValue('Профиль не найден');
      localStorage.setItem('user', JSON.stringify(user));
      return user;
    } catch (e) {
      return rejectWithValue(normalizeError(e));
    }
  },
);

export const updateUserProfile = createAsyncThunk<
  User,
  { name: string; email: string; phone: string },
  { rejectValue: string }
>('auth/updateUserProfile', async (data, { rejectWithValue }) => {
  try {
    const res = await httpClient.put('/api/users/profile', data);
    const user = (res.data?.user ?? null) as User | null;
    if (!user) return rejectWithValue('Не удалось обновить профиль');
    localStorage.setItem('user', JSON.stringify(user));
    return user;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError(state) {
      state.error = null;
    },
    hardLogout(state) {
      clearPersistedAuth();
      state.isAuthenticated = false;
      state.user = null;
      state.accessToken = null;
      state.refreshToken = null;
      state.expiresAt = null;
      state.isLoading = false;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.isLoading = false;
        state.isAuthenticated = true;
        state.accessToken = action.payload.access_token ?? null;
        state.refreshToken = action.payload.refresh_token ?? null;
        state.expiresAt = action.payload.expires_at ?? null;
        state.user = (action.payload.user as unknown as User) ?? null;
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось войти';
      })

      .addCase(registerUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(registerUser.fulfilled, (state, action) => {
        state.isLoading = false;
        state.isAuthenticated = true;
        state.accessToken = action.payload.access_token ?? null;
        state.refreshToken = action.payload.refresh_token ?? null;
        state.expiresAt = action.payload.expires_at ?? null;
        state.user = (action.payload.user as unknown as User) ?? null;
      })
      .addCase(registerUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось зарегистрироваться';
      })

      .addCase(logoutUser.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(logoutUser.fulfilled, (state) => {
        state.isLoading = false;
        state.isAuthenticated = false;
        state.user = null;
        state.accessToken = null;
        state.refreshToken = null;
        state.expiresAt = null;
      })
      .addCase(logoutUser.rejected, (state) => {
        // Ошибку на выходе не показываем как критическую: пользователь уже разлогинен локально.
        state.isLoading = false;
        state.isAuthenticated = false;
        state.user = null;
        state.accessToken = null;
        state.refreshToken = null;
        state.expiresAt = null;
      })

      .addCase(restoreAuth.fulfilled, (state, action) => {
        if (!action.payload) return;
        state.isAuthenticated = true;
        state.accessToken = action.payload.accessToken;
        state.refreshToken = action.payload.refreshToken;
        state.expiresAt = action.payload.expiresAt;
        state.user = action.payload.user;
      })

      .addCase(fetchUserProfile.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchUserProfile.fulfilled, (state, action: PayloadAction<User>) => {
        state.isLoading = false;
        state.user = action.payload;
      })
      .addCase(fetchUserProfile.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось загрузить профиль';
      })

      .addCase(updateUserProfile.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(updateUserProfile.fulfilled, (state, action: PayloadAction<User>) => {
        state.isLoading = false;
        state.user = action.payload;
      })
      .addCase(updateUserProfile.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось обновить профиль';
      });
  },
});

export const { clearError, hardLogout } = authSlice.actions;
export default authSlice.reducer;
