import { useQuery, useInfiniteQuery } from '@tanstack/react-query';
import { listModels, searchModels, getModelDetails, getModelProviders, getModelCategories, ListModelsParams } from '../api/models';

export const useModels = (params: ListModelsParams = {}) => {
  return useQuery({
    queryKey: ['models', params],
    queryFn: () => listModels(params),
  });
};

export const useModelDetails = (modelId: string) => {
  return useQuery({
    queryKey: ['model', modelId],
    queryFn: () => getModelDetails(modelId),
    enabled: !!modelId,
  });
};

export const useModelSearch = (query: string, filters?: any) => {
  return useQuery({
    queryKey: ['modelSearch', query, filters],
    queryFn: () => searchModels(query, filters),
    enabled: !!query,
  });
};

export const useModelProviders = () => {
  return useQuery({
    queryKey: ['modelProviders'],
    queryFn: getModelProviders,
  });
};

export const useModelCategories = () => {
  return useQuery({
    queryKey: ['modelCategories'],
    queryFn: getModelCategories,
  });
};