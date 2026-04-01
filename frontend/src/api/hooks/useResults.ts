import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

export type TournamentResults = components['schemas']['TournamentResults'];
export type TableResult = components['schemas']['TableResult'];

export function useTournamentResults(slug: string) {
  return useQuery<TournamentResults>({
    queryKey: ['results', slug],
    queryFn: () => apiClient.get(`/tournaments/${slug}/results`).then((r) => r.data),
    enabled: !!slug,
  });
}
