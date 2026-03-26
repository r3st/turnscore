import { useMutation } from '@tanstack/react-query';
import { apiClient } from '@/api/client';
import type { components } from '@/api/generated/api';

type RaterAuthRequest = components['schemas']['RaterAuthRequest'];

interface AuthTokenResponse {
  token: string;
  refresh_token?: string;
}

export function useRaterLogin() {
  return useMutation<AuthTokenResponse, Error, RaterAuthRequest>({
    mutationFn: (body) =>
      apiClient.post('/auth/rater', body).then((r) => r.data),
  });
}
