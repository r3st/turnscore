import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export function TableManagementPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();

  return (
    <div>
      <h1 className="font-heading text-2xl font-bold mb-6">{slug}</h1>
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('common.loading')}
      </p>
    </div>
  );
}
