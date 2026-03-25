import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components, operations } from '@/api/generated/api';

export type TournamentSummary = components['schemas']['TournamentSummary'];
export type TournamentDetail = components['schemas']['TournamentDetail'];
export type CreateTournamentRequest = components['schemas']['CreateTournamentRequest'];
export type UpdateTournamentRequest = components['schemas']['UpdateTournamentRequest'];
export type ResultConfig = components['schemas']['ResultConfig'];
export type CriteriaKey = components['schemas']['CriteriaKey'];

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

export function useCreateTournament() {
  const qc = useQueryClient();
  return useMutation<TournamentDetail, Error, CreateTournamentRequest>({
    mutationFn: (body) => apiClient.post('/tournaments', body).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tournaments'] });
      qc.invalidateQueries({ queryKey: ['me'] });
    },
  });
}

export function useUpdateTournament(slug: string) {
  const qc = useQueryClient();
  return useMutation<TournamentDetail, Error, UpdateTournamentRequest>({
    mutationFn: (body) =>
      apiClient.put(`/tournaments/${slug}`, body).then((r) => r.data),
    onSuccess: (updated) => {
      qc.setQueryData(['tournaments', slug], updated);
      qc.invalidateQueries({ queryKey: ['tournaments'] });
    },
  });
}

export function useDeleteTournament() {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: (slug) => apiClient.delete(`/tournaments/${slug}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tournaments'] });
      qc.invalidateQueries({ queryKey: ['me'] });
    },
  });
}

export function useUpdateResultConfig(slug: string) {
  const qc = useQueryClient();
  return useMutation<ResultConfig, Error, ResultConfig>({
    mutationFn: (body) =>
      apiClient.put(`/tournaments/${slug}/result-config`, body).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tournaments', slug] }),
  });
}
