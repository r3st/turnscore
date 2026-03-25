import { useParams } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';
import { useTranslation } from 'react-i18next';

export function RatePage() {
  const { slug, tableNum } = useParams<{ slug: string; tableNum: string }>();
  const { t } = useTranslation();

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8">
        <h1 className="font-heading text-2xl font-bold mb-4">
          {t('rating.title')} — {t('table.number', { number: tableNum })}
        </h1>
        <p className="text-sm text-muted-foreground">Tournament: {slug}</p>
      </main>
    </AppLayout>
  );
}
