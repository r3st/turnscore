import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

export type UserProfile = components['schemas']['UserProfile'];
export type TournamentMembership = components['schemas']['TournamentMembership'];

export function useMe() {
  return useQuery<UserProfile>({
    queryKey: ['me'],
    queryFn: () => apiClient.get('/me').then((r) => r.data),
  });
}
