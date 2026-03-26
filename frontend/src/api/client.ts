import axios from 'axios';
import { QueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';

export const apiClient = axios.create({
  baseURL: `${import.meta.env.VITE_API_URL ?? ''}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach JWT token to every request
apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401: clear auth and redirect to appropriate login page
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const role = useAuthStore.getState().role;
      useAuthStore.getState().logout();
      if (role === 'rater') {
        // Raters go back to rater login — keep path so slug can be extracted if needed
        window.location.href = '/rate-login';
      } else {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  },
);

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
      retry: 1,
    },
  },
});
