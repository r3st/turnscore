import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

export type TableSummary = components['schemas']['TableSummary'];
export type TableDetail = components['schemas']['TableDetail'];
export type UpdateTableRequest = components['schemas']['UpdateTableRequest'];
export type CreateRatingRequest = components['schemas']['CreateRatingRequest'];

export function useGetTable(slug: string, number: number | string) {
  return useQuery<TableDetail>({
    queryKey: ['table', slug, number],
    queryFn: () =>
      apiClient.get(`/tournaments/${slug}/tables/${number}`).then((r) => r.data),
    enabled: !!slug && !!number,
  });
}

export function useSubmitRating(slug: string, tableNumber: number | string) {
  return useMutation<void, Error, CreateRatingRequest>({
    mutationFn: (body) =>
      apiClient.post(`/tournaments/${slug}/tables/${tableNumber}/ratings`, body),
  });
}

export function useListTables(slug: string) {
  return useQuery<TableSummary[]>({
    queryKey: ['tables', slug],
    queryFn: () => apiClient.get(`/tournaments/${slug}/tables`).then((r) => r.data),
    enabled: !!slug,
  });
}

export function useCreateTables(slug: string) {
  const qc = useQueryClient();
  return useMutation<TableSummary[], Error, void>({
    mutationFn: () => apiClient.post(`/tournaments/${slug}/tables`).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useUpdateTable(slug: string, number: number) {
  const qc = useQueryClient();
  return useMutation<TableSummary, Error, UpdateTableRequest>({
    mutationFn: (body) =>
      apiClient.put(`/tournaments/${slug}/tables/${number}`, body).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useUploadPhoto(slug: string, tableNumber: number) {
  const qc = useQueryClient();
  return useMutation<void, Error, FormData>({
    mutationFn: (formData) =>
      apiClient.post(`/tournaments/${slug}/tables/${tableNumber}/photos`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useDeletePhoto(slug: string) {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: (photoId) => apiClient.delete(`/photos/${photoId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useGenerateQRCode(slug: string, tableNumber: number) {
  const qc = useQueryClient();
  return useMutation<void, Error, void>({
    mutationFn: () => apiClient.post(`/tournaments/${slug}/tables/${tableNumber}/qrcode`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useDeleteTable(slug: string) {
  const qc = useQueryClient();
  return useMutation<void, Error, number>({
    mutationFn: (number) => apiClient.delete(`/tournaments/${slug}/tables/${number}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tables', slug] }),
  });
}

export function useGenerateAllQRCodes(slug: string) {
  return useMutation<Blob, Error, void>({
    mutationFn: () =>
      apiClient
        .post(`/tournaments/${slug}/qrcodes`, null, { responseType: 'blob' })
        .then((r) => r.data as Blob),
    onSuccess: (blob) => {
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `qrcodes-${slug}.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    },
  });
}
