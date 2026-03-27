import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { TurnScoreLogo } from '@/components/ui/TurnScoreLogo';
import { useAuthStore } from '@/stores/authStore';

const DEV_USERS = ['alice', 'bob', 'carol', 'dave', 'eve'];
const isDevMode = import.meta.env.VITE_DEV_LOGIN === 'true';

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);

  const handleGoogleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_URL ?? ''}/api/v1/auth/google`;
  };

  const handleDevLogin = async (password: string) => {
    const res = await fetch(`${import.meta.env.VITE_API_URL ?? ''}/api/v1/auth/dev-login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password }),
    });
    if (!res.ok) return;
    const data = await res.json();
    login(data.access_token, data.refresh_token);
    navigate('/dashboard');
  };

  return (
    <AppLayout>
      <main className="flex min-h-[60vh] items-center justify-center">
        <div className="w-full max-w-sm space-y-6 p-8 rounded" style={{ backgroundColor: 'var(--color-surface)' }}>
          <div className="flex justify-center">
            <TurnScoreLogo height={72} />
          </div>
          <h1 className="font-heading text-2xl font-bold text-center">
            {t('auth.login_title')}
          </h1>
          <button
            onClick={handleGoogleLogin}
            className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
          >
            {t('auth.login_with_google')}
          </button>
          {isDevMode && (
            <div className="space-y-2 pt-2 border-t" style={{ borderColor: 'color-mix(in srgb, var(--color-text) 20%, transparent)' }}>
              <p className="text-xs text-center" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
                {t('auth.dev_login_hint')}
              </p>
              <div className="flex flex-wrap gap-2 justify-center">
                {DEV_USERS.map((name) => (
                  <button
                    key={name}
                    onClick={() => handleDevLogin(name)}
                    className="px-3 py-1 rounded text-xs font-medium border"
                    style={{ borderColor: 'color-mix(in srgb, var(--color-text) 20%, transparent)', color: 'var(--color-text)' }}
                  >
                    {name.charAt(0).toUpperCase() + name.slice(1)}
                  </button>
                ))}
              </div>
            </div>
          )}
          <div className="text-center">
            <Link to="/rate-login" className="text-sm hover:underline" style={{ color: 'var(--color-primary)' }}>
              {t('auth.rater_login_link')}
            </Link>
          </div>
        </div>
      </main>
    </AppLayout>
  );
}
