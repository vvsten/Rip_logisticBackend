import type { User } from './User';

/**
 * TransportService внутри заявки приходит в snake_case (как в Go JSON tags).
 * Нам здесь достаточно полей name/description/price для отображения.
 */
export interface OrderTransportService {
  id: number;
  name: string;
  description: string;
  price: number;
  image_url?: string;
  delivery_days?: number;
  max_weight?: number;
  max_volume?: number;
}

export interface LogisticRequestServiceItem {
  id: number;
  logistic_request_id: number;
  transport_service_id: number;
  quantity: number;
  comment?: string;
  sort_order?: number;
  service?: OrderTransportService;
}

export interface LogisticRequest {
  id: number;
  session_id?: string;
  is_draft?: boolean;
  from_city: string;
  to_city: string;
  weight: number;
  length: number;
  width: number;
  height: number;
  services: LogisticRequestServiceItem[];
  total_cost: number;
  total_days: number;
  status: string; // draft | formed | completed | rejected | deleted
  creator_id: number;
  moderator_id?: number | null;
  created_at: string;
  formed_at?: string | null;
  completed_at?: string | null;
  updated_at: string;
  creator?: User;
  moderator?: User | null;
}

export interface OrdersListResponse {
  status: string;
  logistic_requests: LogisticRequest[];
}

export interface OrderResponse {
  status: string;
  logistic_request: LogisticRequest;
}
