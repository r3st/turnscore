import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

type CreateEventRatingRequest = components['schemas']['CreateEventRatingRequest'];

export function useEventRatingStatus(slug: string, enabled: boolean) {
  return useQuery<{ submitted: boolean }>({
    queryKey: ['event-rating-status', slug],
    queryFn: () =>
      apiClient.get(`/tournaments/${slug}/event-rating`).then((r) => r.data),
    enabled: enabled && !!slug,
  });
}

export function useMyRatings(slug: string, enabled: boolean) {
  return useQuery<{ rated_table_numbers: number[] }>({
    queryKey: ['my-ratings', slug],
    queryFn: () =>
      apiClient.get(`/tournaments/${slug}/my-ratings`).then((r) => r.data),
    enabled: enabled && !!slug,
  });
}

export function useSubmitEventRating(slug: string) {
  const queryClient = useQueryClient();
  return useMutation<void, Error, CreateEventRatingRequest>({
    mutationFn: (body) =>
      apiClient.post(`/tournaments/${slug}/event-rating`, body).then(() => undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['event-rating-status', slug] });
    },
  });
}
