import { useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { useTournament } from '@/api/hooks/useTournaments';
import { useTheme } from '@/hooks/useTheme';
import type { TableSummary } from '@/api/hooks/useTables';

export function TournamentPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data: tournament, isLoading, isError } = useTournament(slug);
  const { applyTheme, clearTheme } = useTheme();

  useEffect(() => {
    if (tournament?.type) {
      applyTheme(tournament.type as 'fantasy' | 'scifi');
    }
    return () => clearTheme();
  }, [tournament?.type]);

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8 max-w-3xl">
        {isLoading && (
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('common.loading')}
          </p>
        )}

        {isError && (
          <p className="text-sm" style={{ color: '#dc2626' }}>{t('errors.not_found')}</p>
        )}

        {tournament && (
          <div className="space-y-8">

            {/* Header */}
            <div className="space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <span
                  className="text-xs px-2 py-0.5 rounded font-medium uppercase tracking-wide"
                  style={{ backgroundColor: 'color-mix(in srgb, var(--color-primary) 15%, transparent)', color: 'var(--color-primary)' }}
                >
                  {t(`tournament.type_${tournament.type}`)}
                </span>
                <span
                  className="text-xs px-2 py-0.5 rounded font-medium"
                  style={{ backgroundColor: 'color-mix(in srgb, var(--color-text) 10%, transparent)', color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}
                >
                  {t(`tournament.status_${tournament.status}`)}
                </span>
              </div>
              <h1 className="font-heading text-3xl font-bold">{tournament.name}</h1>
            </div>

            {/* Meta info */}
            {(tournament.event_date || tournament.location) && (
              <div className="flex flex-wrap gap-4 text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}>
                {tournament.event_date && (
                  <span>
                    <span className="font-medium">{t('tournament.event_date')}:</span>{' '}
                    {new Date(tournament.event_date).toLocaleDateString()}
                  </span>
                )}
                {tournament.location && (
                  <span>
                    <span className="font-medium">{t('tournament.location')}:</span>{' '}
                    {tournament.location}
                  </span>
                )}
              </div>
            )}

            {/* Description */}
            {tournament.description && (
              <div
                className="p-4 rounded text-sm leading-relaxed whitespace-pre-wrap"
                style={{ backgroundColor: 'var(--color-surface)' }}
              >
                {tournament.description}
              </div>
            )}

            {/* External links */}
            {tournament.links && tournament.links.length > 0 && (
              <div className="space-y-2">
                <div className="flex flex-wrap gap-2">
                  {tournament.links.map((link, i) => (
                    <a
                      key={i}
                      href={link.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-sm px-3 py-1.5 rounded border"
                      style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                    >
                      {link.label || link.url}
                      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5">
                        <path d="M5 2H2a1 1 0 00-1 1v7a1 1 0 001 1h7a1 1 0 001-1V7M7 1h4m0 0v4m0-4L5 7" />
                      </svg>
                    </a>
                  ))}
                </div>
                <p className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
                  {t('tournament.links_disclaimer')}
                </p>
              </div>
            )}

            {/* Tables */}
            <div className="space-y-3">
              <h2 className="font-heading text-xl font-bold">
                {t('tournament.tables_count', { count: tournament.table_count })}
              </h2>

              {(!tournament.tables || tournament.tables.length === 0) ? (
                <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
                  {t('tournament.no_tables')}
                </p>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {tournament.tables.map((table) => (
                    <TableCard key={table.id} slug={slug} table={table} rateLabel={t('tournament.rate_table')} tableLabel={t('table.number', { number: table.number })} />
                  ))}
                </div>
              )}
            </div>

          </div>
        )}
      </main>
    </AppLayout>
  );
}

function TableCard({ slug, table, rateLabel, tableLabel }: {
  slug: string;
  table: TableSummary;
  rateLabel: string;
  tableLabel: string;
}) {
  const firstPhoto = table.photos?.[0];
  const thumbUrl = firstPhoto?.thumbnail_url ?? firstPhoto?.url;

  return (
    <div
      className="rounded border overflow-hidden"
      style={{
        backgroundColor: 'var(--color-surface)',
        borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)',
      }}
    >
      {thumbUrl && (
        <img
          src={thumbUrl}
          alt=""
          className="w-full h-32 object-cover"
        />
      )}
      <div className="p-3 flex items-center justify-between gap-2">
        <div className="min-w-0">
          <p className="font-medium text-sm truncate">{tableLabel}</p>
          {table.name && (
            <p className="text-xs truncate" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {table.name}
            </p>
          )}
        </div>
        <Link
          to={`/rate/${slug}/${table.number}`}
          className="shrink-0 inline-flex items-center justify-center text-xs px-3 py-1.5 rounded font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
        >
          {rateLabel}
        </Link>
      </div>
    </div>
  );
}
