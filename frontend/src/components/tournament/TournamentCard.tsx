import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { CalendarDays, MapPin, Grid2x2, Swords } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { TournamentSummary } from '@/api/hooks/useTournaments';

interface TournamentCardProps {
  tournament: TournamentSummary;
}

const TYPE_BADGE: Record<string, { label: string; icon: string; className: string }> = {
  fantasy: { label: 'tournament.type_fantasy', icon: '🏰', className: 'bg-amber-100 text-amber-800' },
  scifi:   { label: 'tournament.type_scifi',   icon: '🚀', className: 'bg-cyan-100 text-cyan-800' },
};

const STATUS_CLASS: Record<string, string> = {
  draft:    'bg-gray-100 text-gray-600',
  active:   'bg-green-100 text-green-700',
  voting:   'bg-blue-100 text-blue-700',
  archived: 'bg-slate-100 text-slate-500',
};

export function TournamentCard({ tournament }: TournamentCardProps) {
  const { t } = useTranslation();
  const badge = TYPE_BADGE[tournament.type] ?? TYPE_BADGE.fantasy;
  const statusClass = STATUS_CLASS[tournament.status] ?? STATUS_CLASS.archived;

  return (
    <article
      className={cn(
        'rounded border p-4 flex flex-col justify-between gap-3 h-full transition-shadow hover:shadow-md',
        'focus-within:ring-2 focus-within:ring-[var(--color-primary)]',
      )}
      style={{ backgroundColor: 'var(--color-surface)' }}
    >
      <div className="flex flex-col gap-3">
        {/* Header row */}
        <div className="flex items-start justify-between gap-2">
          <Link
            to={`/t/${tournament.slug}`}
            className="font-heading font-semibold text-base leading-tight hover:underline focus:outline-none"
            style={{ color: 'var(--color-text)' }}
          >
            {tournament.name}
          </Link>

          <div className="flex items-center gap-1 shrink-0">
            {/* Type badge */}
            <span
              className={cn('text-xs font-medium px-2 py-0.5 rounded-full', badge.className)}
              aria-label={t(badge.label)}
            >
              <span aria-hidden>{badge.icon}</span>{' '}
              {t(badge.label)}
            </span>
            {/* Status badge */}
            <span className={cn('text-xs font-medium px-2 py-0.5 rounded-full', statusClass)}>
              {t(`tournament.status_${tournament.status}`)}
            </span>
          </div>
        </div>

        {/* Game system badge */}
        {tournament.game_system && (
          <div
            className="self-start inline-flex items-center gap-1.5 text-xs font-medium px-2 py-1 rounded"
            style={{
              backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, transparent)',
              color: 'var(--color-primary)',
            }}
            aria-label={`${t('tournament_form.game_system_label')}: ${tournament.game_system}`}
          >
            <Swords className="h-3.5 w-3.5 shrink-0" aria-hidden />
            {tournament.game_system}
          </div>
        )}

        {/* Meta info */}
        <dl className="flex flex-wrap gap-x-4 gap-y-1 text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 65%, transparent)' }}>
          {tournament.event_date && (
            <div className="flex items-center gap-1">
              <CalendarDays className="h-3.5 w-3.5" aria-hidden />
              <dt className="sr-only">{t('tournament.event_date')}</dt>
              <dd>{new Date(tournament.event_date).toLocaleDateString()}</dd>
            </div>
          )}
          {tournament.location && (
            <div className="flex items-center gap-1">
              <MapPin className="h-3.5 w-3.5" aria-hidden />
              <dt className="sr-only">{t('tournament.location')}</dt>
              <dd>{tournament.location}</dd>
            </div>
          )}
          <div className="flex items-center gap-1">
            <Grid2x2 className="h-3.5 w-3.5" aria-hidden />
            <dt className="sr-only">{t('tournament.tables_count', { count: tournament.table_count })}</dt>
            <dd>{t('tournament.tables_count', { count: tournament.table_count })}</dd>
          </div>
        </dl>
      </div>

      {/* CTA — always at bottom */}
      <Link
        to={`/t/${tournament.slug}`}
        className="self-start text-sm font-medium hover:underline"
        style={{ color: 'var(--color-primary)' }}
        tabIndex={-1}
        aria-hidden
      >
        {t('tournament.view_tables')} →
      </Link>
    </article>
  );
}
