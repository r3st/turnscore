import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type Role = 'user' | 'rater';

interface JWTClaims {
  sub: string;
  role: Role;
  exp: number;
}

function parseJWT(token: string): JWTClaims | null {
  try {
    const payload = token.split('.')[1];
    return JSON.parse(atob(payload)) as JWTClaims;
  } catch {
    return null;
  }
}

interface AuthState {
  token: string | null;
  refreshToken: string | null;
  userId: string | null;
  role: Role | null;
  login: (token: string, refreshToken: string | null) => void;
  logout: () => void;
  isAuthenticated: () => boolean;
  hasRole: (...roles: string[]) => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      refreshToken: null,
      userId: null,
      role: null,
      login: (token, refreshToken) => {
        const claims = parseJWT(token);
        if (claims) {
          set({ token, refreshToken, userId: claims.sub, role: claims.role });
        }
      },
      logout: () =>
        set({ token: null, refreshToken: null, userId: null, role: null }),
      isAuthenticated: () => {
        const { token } = get();
        if (!token) return false;
        const claims = parseJWT(token);
        if (!claims) return false;
        return claims.exp * 1000 > Date.now();
      },
      hasRole: (...roles: string[]) => roles.includes(get().role ?? ''),
    }),
    { name: 'turnscore-auth' },
  ),
);
