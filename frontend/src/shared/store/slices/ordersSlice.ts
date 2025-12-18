import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import type { LogisticRequest, OrderResponse, OrdersListResponse } from '../../types/Order';
import { httpClient } from '../../api/httpClient';
import { logisticRequestsApi } from '../../api/generatedApi';

interface OrdersState {
  orders: LogisticRequest[];
  currentOrder: LogisticRequest | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: OrdersState = {
  orders: [],
  currentOrder: null,
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

/**
 * Загрузка списка заявок.
 *
 * Здесь специально используется сгенерированный swagger-клиент `logisticRequestsApi`
 * (см. `src/shared/api/generated`). Это ключевая часть требования по кодогенерации.
 */
export const fetchOrders = createAsyncThunk<
  LogisticRequest[],
  { status?: string; date_from?: string; date_to?: string } | undefined,
  { rejectValue: string }
>('orders/fetchOrders', async (params, { rejectWithValue }) => {
  try {
    const res = await logisticRequestsApi.apiLogisticRequestsGet({
      status: params?.status,
      dateFrom: params?.date_from,
      dateTo: params?.date_to,
    });

    const data = res.data as unknown as OrdersListResponse;
    return Array.isArray(data?.logistic_requests) ? data.logistic_requests : [];
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const fetchOrder = createAsyncThunk<LogisticRequest, number, { rejectValue: string }>(
  'orders/fetchOrder',
  async (id, { rejectWithValue }) => {
    try {
      const res = await httpClient.get(`/api/logistic-requests/${id}`);
      const data = res.data as OrderResponse;
      if (!data?.logistic_request) return rejectWithValue('Заявка не найдена');
      return data.logistic_request;
    } catch (e) {
      return rejectWithValue(normalizeError(e));
    }
  },
);

export const updateOrder = createAsyncThunk<
  LogisticRequest,
  { id: number; data: { from_city: string; to_city: string; weight: number; length: number; width: number; height: number } },
  { rejectValue: string }
>('orders/updateOrder', async ({ id, data }, { rejectWithValue }) => {
  try {
    await httpClient.put(`/api/logistic-requests/${id}/update`, data);
    const res = await httpClient.get(`/api/logistic-requests/${id}`);
    const out = res.data as OrderResponse;
    return out.logistic_request;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const removeServiceFromOrder = createAsyncThunk<
  LogisticRequest,
  { orderId: number; transportServiceId: number },
  { rejectValue: string }
>('orders/removeServiceFromOrder', async ({ orderId, transportServiceId }, { rejectWithValue }) => {
  try {
    await httpClient.delete(`/api/logistic-requests/${orderId}/services/${transportServiceId}`);
    const res = await httpClient.get(`/api/logistic-requests/${orderId}`);
    const out = res.data as OrderResponse;
    return out.logistic_request;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const updateServiceInOrder = createAsyncThunk<
  LogisticRequest,
  { orderId: number; transportServiceId: number; quantity: number },
  { rejectValue: string }
>('orders/updateServiceInOrder', async ({ orderId, transportServiceId, quantity }, { rejectWithValue }) => {
  try {
    await httpClient.put(`/api/logistic-requests/${orderId}/services/${transportServiceId}`, {
      quantity,
      sort_order: 0,
      comment: '',
    });
    const res = await httpClient.get(`/api/logistic-requests/${orderId}`);
    const out = res.data as OrderResponse;
    return out.logistic_request;
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

export const formOrder = createAsyncThunk<
  void,
  { id: number; data: { from_city: string; to_city: string; weight: number; length: number; width: number; height: number } },
  { rejectValue: string }
>('orders/formOrder', async ({ id, data }, { rejectWithValue }) => {
  try {
    await httpClient.put(`/api/logistic-requests/${id}/form`, data);
  } catch (e) {
    return rejectWithValue(normalizeError(e));
  }
});

const ordersSlice = createSlice({
  name: 'orders',
  initialState,
  reducers: {
    clearOrdersState(state) {
      state.orders = [];
      state.currentOrder = null;
      state.isLoading = false;
      state.error = null;
    },
    clearOrdersError(state) {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchOrders.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchOrders.fulfilled, (state, action) => {
        state.isLoading = false;
        state.orders = action.payload;
      })
      .addCase(fetchOrders.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось загрузить заявки';
      })

      .addCase(fetchOrder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchOrder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentOrder = action.payload;
      })
      .addCase(fetchOrder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось загрузить заявку';
      })

      .addCase(updateOrder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(updateOrder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentOrder = action.payload;
      })
      .addCase(updateOrder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось обновить заявку';
      })

      .addCase(removeServiceFromOrder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(removeServiceFromOrder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentOrder = action.payload;
      })
      .addCase(removeServiceFromOrder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось удалить услугу';
      })

      .addCase(updateServiceInOrder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(updateServiceInOrder.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentOrder = action.payload;
      })
      .addCase(updateServiceInOrder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось обновить количество';
      })

      .addCase(formOrder.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(formOrder.fulfilled, (state) => {
        state.isLoading = false;
      })
      .addCase(formOrder.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload ?? 'Не удалось сформировать заявку';
      });
  },
});

export const { clearOrdersState, clearOrdersError } = ordersSlice.actions;
export default ordersSlice.reducer;
