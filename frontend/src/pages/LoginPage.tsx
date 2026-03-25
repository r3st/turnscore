import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';

export function LoginPage() {
  const { t } = useTranslation();

  const handleGoogleLogin = () => {
    window.location.href = `${import.meta.env.VITE_API_URL ?? ''}/api/v1/auth/google`;
  };

  return (
    <AppLayout>
      <main className="flex min-h-[60vh] items-center justify-center">
        <div className="w-full max-w-sm space-y-6 p-8 rounded" style={{ backgroundColor: 'var(--color-surface)' }}>
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
