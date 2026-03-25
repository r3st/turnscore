import { useQuery } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components, operations } from '@/api/generated/api';

export type TournamentSummary = components['schemas']['TournamentSummary'];
export type TournamentDetail = components['schemas']['TournamentDetail'];

type ListTournamentsResponse =
  operations['listTournaments']['responses']['200']['content']['application/json'];

export function useTournaments() {
  return useQuery<ListTournamentsResponse>({
    queryKey: ['tournaments'],
    queryFn: () => apiClient.get('/tournaments').then((r) => r.data),
  });
}

export function useTournament(slug: string) {
  return useQuery<TournamentDetail>({
    queryKey: ['tournaments', slug],
    queryFn: () => apiClient.get(`/tournaments/${slug}`).then((r) => r.data),
    enabled: !!slug,
  });
}
