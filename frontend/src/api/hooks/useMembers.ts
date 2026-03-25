import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

type MemberPreview = components['schemas']['MemberPreview'];

export function useAddMember(slug: string) {
  const qc = useQueryClient();
  return useMutation<MemberPreview, Error, string>({
    mutationFn: (inviteCode) =>
      apiClient
        .post(`/tournaments/${slug}/members`, { invite_code: inviteCode })
        .then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useRemoveMember(slug: string) {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: (userId) =>
      apiClient.delete(`/tournaments/${slug}/members/${userId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['me'] }),
  });
}

export function useLeaveTournament(slug: string) {
  const qc = useQueryClient();
  return useMutation<void, Error, void>({
    mutationFn: () => apiClient.delete(`/tournaments/${slug}/members/me`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['me'] }),
  });
}
