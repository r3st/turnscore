import { useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import { useDeleteTournament } from '@/api/hooks/useTournaments';
import type { TournamentMembership } from '@/api/hooks/useUser';

export function MyTournamentsPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: profile, isLoading } = useMe();
  const deleteTournament = useDeleteTournament();

  const memberships = profile?.memberships.filter((m) => m.role === 'organizer') ?? [];

  const handleDelete = async (m: TournamentMembership) => {
    if (!window.confirm(t('dashboard.delete_confirm'))) return;
    await deleteTournament.mutateAsync(m.tournament_slug);
  };

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  return (
    <section aria-labelledby="my-tournaments-heading">
      <h1 id="my-tournaments-heading" className="font-heading text-2xl font-bold mb-6">
        {t('dashboard.my_tournaments')}
      </h1>

      {memberships.length === 0 ? (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('dashboard.no_tournaments')}
        </p>
      ) : (
        <ul className="space-y-2" role="list">
          {memberships.map((m) => (
            <li
              key={m.tournament_id}
              className="flex flex-wrap items-center justify-between gap-2 p-4 rounded"
              style={{ backgroundColor: 'var(--color-surface)' }}
            >
              <span className="font-medium">{m.tournament_name}</span>
              <div className="flex gap-2 flex-wrap">
                <button
                  onClick={() => navigate(`/dashboard/tournaments/${m.tournament_slug}`)}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                >
                  {t('dashboard.manage')}
                </button>
                <Link
                  to={`/dashboard/tournaments/${m.tournament_slug}/tables`}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                >
                  {t('dashboard.tables')}
                </Link>
                <Link
                  to={`/dashboard/tournaments/${m.tournament_slug}/raters`}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                >
                  {t('rater.section_title')}
                </Link>
                <Link
                  to={`/dashboard/tournaments/${m.tournament_slug}/helpers`}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                >
                  {t('helper.section_title')}
                </Link>
                <Link
                  to={`/dashboard/tournaments/${m.tournament_slug}/results`}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
                >
                  {t('results.title')}
                </Link>
                <button
                  onClick={() => handleDelete(m)}
                  disabled={deleteTournament.isPending}
                  className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
                  style={{ borderColor: '#dc2626', color: '#dc2626' }}
                  aria-label={`${t('dashboard.delete_tournament')} ${m.tournament_name}`}
                >
                  {t('dashboard.delete_tournament')}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
