import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getSystemStats, listUsers, getUserDetails, updateUser, ListUsersParams } from '../api/admin';

export const useSystemStats = () => {
  return useQuery({
    queryKey: ['systemStats'],
    queryFn: getSystemStats,
    refetchInterval: 30000, // 每30秒刷新一次
  });
};

export const useUsers = (params: ListUsersParams = {}) => {
  return useQuery({
    queryKey: ['users', params],
    queryFn: () => listUsers(params),
  });
};

export const useUserDetails = (userId: string) => {
  return useQuery({
    queryKey: ['userDetails', userId],
    queryFn: () => getUserDetails(userId),
    enabled: !!userId,
  });
};

export const useUpdateUser = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ userId, data }: { userId: string; data: { role?: string; status?: string } }) =>
      updateUser(userId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['userDetails', variables.userId] });
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
  });
};