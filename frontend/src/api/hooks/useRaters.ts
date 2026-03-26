import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

export type Rater = components['schemas']['Rater'];
export type CreateRaterRequest = components['schemas']['CreateRaterRequest'];

export function useListRaters(slug: string) {
  return useQuery<Rater[]>({
    queryKey: ['raters', slug],
    queryFn: () => apiClient.get(`/tournaments/${slug}/raters`).then((r) => r.data),
    enabled: !!slug,
  });
}

export function useCreateRater(slug: string) {
  const qc = useQueryClient();
  return useMutation<Rater, Error, CreateRaterRequest>({
    mutationFn: (body) =>
      apiClient.post(`/tournaments/${slug}/raters`, body).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['raters', slug] }),
  });
}
