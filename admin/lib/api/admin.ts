import api from '../api';

export interface ModelStats {
  model_id: string;
  model_name: string;
  requests: number;
  revenue: number;
  success_rate: number;
}

export interface PaymentItem {
  id: string;
  amount: number;
  currency: string;
  payment_method: string;
  status: string;
  transaction_id?: string;
  created_at: string;
  paid_at?: string;
}

export interface ServerStatus {
  database: boolean;
  redis: boolean;
  uptime: string;
  memory_usage: number;
  cpu_usage: number;
}

export interface SystemStats {
  total_users: number;
  active_users: number;
  total_requests: number;
  total_revenue: number;
  daily_requests: number;
  daily_revenue: number;
  top_models: ModelStats[];
  recent_payments: PaymentItem[];
  server_status: ServerStatus;
}

export interface AdminUserInfo {
  id: string;
  email: string;
  username: string;
  role: string;
  status: string;
  created_at: string;
  total_paid: number;
  total_used: number;
  current_balance: number;
  api_keys_count: number;
  last_activity?: string;
}

export interface ListUsersResponse {
  success: boolean;
  data: {
    users: AdminUserInfo[];
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
}

export interface ListUsersParams {
  page?: number;
  limit?: number;
  search?: string;
  role?: string;
  status?: string;
  sort_by?: string;
  sort_order?: string;
}

export async function getSystemStats(): Promise<{ success: boolean; data: SystemStats }> {
  const response = await api.get('/admin/stats');
  return response.data;
}

export async function listUsers(params: ListUsersParams = {}): Promise<ListUsersResponse> {
  const response = await api.get('/admin/users', { params });
  return response.data;
}

export async function getUserDetails(userId: string) {
  const response = await api.get(`/admin/users/${userId}`);
  return response.data;
}

export async function updateUser(userId: string, data: { role?: string; status?: string }) {
  const response = await api.put(`/admin/users/${userId}`, data);
  return response.data;
}