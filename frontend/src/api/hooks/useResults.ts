import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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

export function usePublishResults(slug: string) {
  const qc = useQueryClient();
  return useMutation<void, Error, { published: boolean }>({
    mutationFn: (body) => apiClient.post(`/tournaments/${slug}/publish-results`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tournaments', slug] });
      qc.invalidateQueries({ queryKey: ['results', slug] });
    },
  });
}
