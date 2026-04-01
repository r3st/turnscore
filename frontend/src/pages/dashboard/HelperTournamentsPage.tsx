import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import { useLeaveTournament } from '@/api/hooks/useMembers';
import type { TournamentMembership } from '@/api/hooks/useUser';

export function HelperTournamentsPage() {
  const { t } = useTranslation();
  const { data: profile, isLoading } = useMe();

  const memberships = profile?.memberships.filter((m) => m.role === 'helper') ?? [];

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  return (
    <section aria-labelledby="helper-heading">
      <h1 id="helper-heading" className="font-heading text-2xl font-bold mb-6">
        {t('dashboard.as_helper')}
      </h1>

      {memberships.length === 0 ? (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('dashboard.no_helper_tournaments')}
        </p>
      ) : (
        <ul className="space-y-2" role="list">
          {memberships.map((m) => (
            <HelperRow key={m.tournament_id} membership={m} />
          ))}
        </ul>
      )}
    </section>
  );
}

function HelperRow({ membership: m }: { membership: TournamentMembership }) {
  const { t } = useTranslation();
  const leave = useLeaveTournament(m.tournament_slug);

  const handleLeave = async () => {
    if (!window.confirm(t('dashboard.leave_confirm'))) return;
    await leave.mutateAsync();
  };

  return (
    <li
      className="flex flex-wrap items-center justify-between gap-2 p-4 rounded"
      style={{ backgroundColor: 'var(--color-surface)' }}
    >
      <span className="font-medium">{m.tournament_name}</span>
      <div className="flex gap-2 flex-wrap">
        <Link
          to={`/dashboard/tournaments/${m.tournament_slug}`}
          className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
          style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
        >
          {t('dashboard.manage')}
        </Link>
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
          to={`/dashboard/tournaments/${m.tournament_slug}/results`}
          className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
          style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
        >
          {t('results.title')}
        </Link>
        <button
          onClick={handleLeave}
          disabled={leave.isPending}
          className="inline-flex items-center justify-center text-xs px-3 py-1 rounded border"
          style={{ borderColor: '#dc2626', color: '#dc2626' }}
          aria-label={`${t('dashboard.leave_tournament')} ${m.tournament_name}`}
        >
          {t('dashboard.leave_tournament')}
        </button>
      </div>
    </li>
  );
}
