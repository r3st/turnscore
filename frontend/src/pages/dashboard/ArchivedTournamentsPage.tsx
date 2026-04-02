import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import { useTournament } from '@/api/hooks/useTournaments';
import type { TournamentMembership } from '@/api/hooks/useUser';

export function ArchivedTournamentsPage() {
  const { t } = useTranslation();
  const { data: profile, isLoading } = useMe();

  const memberships = profile?.memberships ?? [];

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  return (
    <section aria-labelledby="archived-heading">
      <h1 id="archived-heading" className="font-heading text-2xl font-bold mb-6">
        {t('dashboard.archived')}
      </h1>
      <ArchivedList memberships={memberships} />
    </section>
  );
}

function ArchivedList({ memberships }: { memberships: TournamentMembership[] }) {
  const { t } = useTranslation();

  // Render each membership and filter to archived ones
  const rows = memberships.map((m) => (
    <ArchivedRow key={m.tournament_id} membership={m} />
  ));

  if (memberships.length === 0) {
    return (
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('dashboard.no_archived')}
      </p>
    );
  }

  return <ul className="space-y-2" role="list">{rows}</ul>;
}

function ArchivedRow({ membership: m }: { membership: TournamentMembership }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: tournament } = useTournament(m.tournament_slug);

  // Only render archived tournaments
  if (!tournament || tournament.status !== 'archived') return null;

  return (
    <li
      className="flex flex-wrap items-center justify-between gap-2 p-4 rounded"
      style={{ backgroundColor: 'var(--color-surface)' }}
    >
      <div>
        <span className="font-medium">{m.tournament_name}</span>
        <span
          className="ml-2 text-xs px-2 py-0.5 rounded"
          style={{ backgroundColor: 'color-mix(in srgb, var(--color-text) 10%, transparent)' }}
        >
          {t('tournament.status_archived')}
        </span>
      </div>
      <button
        onClick={() => navigate(`/dashboard/tournaments/${m.tournament_slug}/results`)}
        className="text-xs px-3 py-1 rounded border"
        style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
      >
        {t('results.title')}
      </button>
    </li>
  );
}
