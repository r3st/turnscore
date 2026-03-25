import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export function DashboardPage() {
  const { t } = useTranslation();

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="font-heading text-2xl font-bold">{t('dashboard.my_tournaments')}</h1>
        <Link
          to="/dashboard/tournaments/new"
          className="px-4 py-2 rounded text-sm font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
        >
          {t('dashboard.create_tournament')}
        </Link>
      </div>
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('dashboard.no_tournaments')}
      </p>
    </div>
  );
}
