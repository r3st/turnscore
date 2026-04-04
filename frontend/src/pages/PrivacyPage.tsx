import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { LegalBody } from '@/components/legal/LegalBody';
import { useLegalContent } from '@/hooks/useLegalContent';

export function PrivacyPage() {
  const { t } = useTranslation();
  const { content, status } = useLegalContent('privacy');

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8 max-w-2xl">
        <h1 className="font-heading text-2xl font-bold mb-6" style={{ color: 'var(--color-text)' }}>
          {t('nav.privacy')}
        </h1>
        <LegalBody status={status} content={content} t={t} />
      </main>
    </AppLayout>
  );
}
