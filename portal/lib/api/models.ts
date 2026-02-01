import api from '../api';

export interface Model {
  id: string;
  name: string;
  description?: string;
  provider_id: string;
  provider_name: string;
  context_length?: number;
  max_tokens?: number;
  capabilities: Record<string, any>;
  category?: string;
  pricing_tier?: string;
  input_price: number;
  output_price: number;
  is_free: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListModelsResponse {
  success: boolean;
  data: {
    models: Model[];
    total: number;
    page: number;
    limit: number;
    total_pages: number;
  };
}

export interface ListModelsParams {
  page?: number;
  limit?: number;
  category?: string;
  provider?: string;
  search?: string;
  is_free?: boolean;
  sort_by?: string;
  sort_order?: string;
}

export async function listModels(params: ListModelsParams = {}): Promise<ListModelsResponse> {
  const response = await api.get<ListModelsResponse>('/models', { params });
  return response.data;
}

export async function searchModels(query: string, filters?: {
  categories?: string[];
  providers?: string[];
  min_price?: number;
  max_price?: number;
  is_free?: boolean;
}) {
  const params = new URLSearchParams();
  params.append('q', query);
  if (filters?.categories) {
    filters.categories.forEach(cat => params.append('categories', cat));
  }
  if (filters?.providers) {
    filters.providers.forEach(prov => params.append('providers', prov));
  }
  if (filters?.min_price !== undefined) {
    params.append('min_price', filters.min_price.toString());
  }
  if (filters?.max_price !== undefined) {
    params.append('max_price', filters.max_price.toString());
  }
  if (filters?.is_free !== undefined) {
    params.append('is_free', filters.is_free.toString());
  }
  const response = await api.get('/models/search', { params });
  return response.data;
}

export async function getModelDetails(modelId: string) {
  const response = await api.get(`/models/${modelId}`);
  return response.data;
}

export async function getModelProviders() {
  const response = await api.get('/models/providers');
  return response.data;
}

export async function getModelCategories() {
  const response = await api.get('/models/categories');
  return response.data;
}