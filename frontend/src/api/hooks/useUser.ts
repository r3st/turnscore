import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

export type UserProfile = components['schemas']['UserProfile'];
export type TournamentMembership = components['schemas']['TournamentMembership'];
export type UpdatePreferencesRequest = components['schemas']['UpdatePreferencesRequest'];

export function useMe() {
  return useQuery<UserProfile>({
    queryKey: ['me'],
    queryFn: () => apiClient.get('/me').then((r) => r.data),
  });
}

export function useUpdatePreferences() {
  const qc = useQueryClient();
  return useMutation<UserProfile, Error, UpdatePreferencesRequest>({
    mutationFn: (body) => apiClient.patch('/me', body).then((r) => r.data),
    onSuccess: (data) => {
      qc.setQueryData(['me'], data);
    },
  });
}

export function useDeleteAccount() {
  return useMutation<void, Error, { confirm_email: string }>({
    mutationFn: (body) => apiClient.delete('/me', { data: body }),
  });
}
