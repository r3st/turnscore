import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';

export function RaterLoginPage() {
  const { slug } = useParams<{ slug?: string }>();
  const { t } = useTranslation();

  return (
    <AppLayout>
      <main className="flex min-h-[60vh] items-center justify-center">
        <div className="w-full max-w-sm space-y-6 p-8 rounded" style={{ backgroundColor: 'var(--color-surface)' }}>
          <h1 className="font-heading text-2xl font-bold text-center">
            {t('auth.rater_login_title')}
          </h1>
          {slug && (
            <p className="text-sm text-center" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {slug}
            </p>
          )}
          <p className="text-sm text-center">{t('common.loading')}</p>
        </div>
      </main>
    </AppLayout>
  );
}
