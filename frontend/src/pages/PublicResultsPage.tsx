import { useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { useTournamentResults } from '@/api/hooks/useResults';
import { useTournament } from '@/api/hooks/useTournaments';
import { useTheme } from '@/hooks/useTheme';
import { ResultsView } from '@/components/results/ResultsView';

export function PublicResultsPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data: tournament } = useTournament(slug);
  const { data, isLoading, error } = useTournamentResults(slug);
  const { applyTheme, clearTheme } = useTheme();

  useEffect(() => {
    if (!tournament?.type) return;
    const prev = document.documentElement.getAttribute('data-theme');
    applyTheme(tournament.type as 'fantasy' | 'scifi');
    return () => {
      if (prev) document.documentElement.setAttribute('data-theme', prev);
      else clearTheme();
    };
  }, [tournament?.type]);

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8 max-w-3xl">
        <Link
          to={`/t/${slug}`}
          className="inline-flex items-center gap-1 text-sm mb-6"
          style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
        >
          ← {tournament?.name ?? t('common.back')}
        </Link>

        {isLoading && (
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('common.loading')}
          </p>
        )}

        {(error || (!isLoading && !data)) && (
          <p className="text-sm" style={{ color: '#dc2626' }}>{t('errors.generic')}</p>
        )}

        {data && <ResultsView data={data} />}
      </main>
    </AppLayout>
  );
}
