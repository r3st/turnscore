import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';

export function HomePage() {
  const { t } = useTranslation();

  return (
    <AppLayout>
      <h1 className="font-heading text-3xl font-bold mb-8">{t('app.name')}</h1>
      <section aria-labelledby="upcoming-heading">
        <h2 id="upcoming-heading" className="text-xl font-semibold mb-4">
          {t('home.upcoming_tournaments')}
        </h2>
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('home.no_tournaments')}
        </p>
      </section>
    </AppLayout>
  );
}
