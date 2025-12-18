import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { httpClient } from '../../api/httpClient';

/**
 * User-draft (черновик) — отдельное состояние конструктора заявки.
 *
 * Требование лаб7: при выходе сбрасываем конструктор заявки.
 * Поэтому этот slice можно полностью очистить экшеном resetDraftState.
 */
interface DraftState {
  draftId: number | null;
  count: number;
  isLoading: boolean;
  error: string | null;
}

const initialState: DraftState = {
  draftId: null,
  count: 0,
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

export const fetchUserDraftIcon = createAsyncThunk<
  { request_id: number; count: number },
  void,
  { rejectValue: string }
>('draft/fetchUserDraftIcon', async (_, { rejectWithValue }) => {
  try {
    const res = await httpClient.get('/api/logistic-requests/user-draft/icon');
    const requestId = Number(res.data?.request_id ?? 0);
    const count = Number(res.data?.count ?? 0);
    if (!requestId) return rejectWithValue('Черновик не найден');
    return { request_id: requestId, count };
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const addServiceToUserDraft = createAsyncThunk<
  { request_id: number; count: number },
  { serviceId: number },
  { rejectValue: string }
>('draft/addServiceToUserDraft', async ({ serviceId }, { rejectWithValue }) => {
  try {
    const res = await httpClient.post(`/api/logistic-requests/user-draft/services/${serviceId}`);
    const requestId = Number(res.data?.request_id ?? 0);
    const count = Number(res.data?.count ?? 0);
    if (!requestId) return rejectWithValue('Не удалось обновить черновик');
    return { request_id: requestId, count };
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const clearUserDraft = createAsyncThunk<void, void, { rejectValue: string }>(
  'draft/clearUserDraft',
  async (_, { rejectWithValue }) => {
    try {
      await httpClient.delete('/api/logistic-requests/user-draft');
    } catch (e) {
      return rejectWithValue(normalizeError(e));
    }
  },
);

const draftSlice = createSlice({
  name: 'draft',
  initialState,
  reducers: {
    resetDraftState(state) {
      state.draftId = null;
      state.count = 0;
      state.isLoading = false;
      state.error = null;
    },
    clearDraftError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchUserDraftIcon.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchUserDraftIcon.fulfilled, (state, action) => {
        state.isLoading = false;
        state.draftId = action.payload.request_id;
        state.count = action.payload.count;
      })
      .addCase(fetchUserDraftIcon.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось загрузить черновик';
        state.draftId = null;
        state.count = 0;
      })

      .addCase(addServiceToUserDraft.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(addServiceToUserDraft.fulfilled, (state, action) => {
        state.isLoading = false;
        state.draftId = action.payload.request_id;
        state.count = action.payload.count;
      })
      .addCase(addServiceToUserDraft.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось добавить услугу в черновик';
      })

      .addCase(clearUserDraft.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(clearUserDraft.fulfilled, (state) => {
        state.isLoading = false;
        state.draftId = null;
        state.count = 0;
      })
      .addCase(clearUserDraft.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось очистить черновик';
        // Локально всё равно сбрасываем (для UX на выходе)
        state.draftId = null;
        state.count = 0;
      });
  },
});

export const { resetDraftState, clearDraftError } = draftSlice.actions;
export default draftSlice.reducer;
