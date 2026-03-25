import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';

export function PrivacyPage() {
  const { t } = useTranslation();

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8 max-w-2xl">
        <h1 className="font-heading text-2xl font-bold mb-4">{t('nav.privacy')}</h1>
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          Privacy policy content to be added.
        </p>
      </main>
    </AppLayout>
  );
}
