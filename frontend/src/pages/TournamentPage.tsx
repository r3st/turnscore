import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { useTournament } from '@/api/hooks/useTournaments';
import { useTheme } from '@/hooks/useTheme';
import { useAuthStore } from '@/stores/authStore';
import { useEventRatingStatus, useSubmitEventRating } from '@/api/hooks/useEventRating';
import type { TableSummary } from '@/api/hooks/useTables';
import type { components } from '@/api/generated/api';

type CriteriaKey = components['schemas']['CriteriaKey'];

const OPTIONAL_CRITERIA: CriteriaKey[] = ['catering', 'venue', 'organization'];
const GRADES = [1, 2, 3, 4, 5, 6] as const;

export function TournamentPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data: tournament, isLoading, isError } = useTournament(slug);
  const { applyTheme, clearTheme } = useTheme();
  const { isAuthenticated, role } = useAuthStore();
  const isRater = isAuthenticated() && role === 'rater';

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

            {/* Event Rating — shown above tables when optional criteria exist and rater is logged in */}
            {isRater && tournament.active_criteria?.some((c) => OPTIONAL_CRITERIA.includes(c as CriteriaKey)) && (
              <EventRatingSection slug={slug} activeCriteria={(tournament.active_criteria as CriteriaKey[]).filter((c) => OPTIONAL_CRITERIA.includes(c))} />
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

// ── EventRatingSection ────────────────────────────────────────────────────────

function EventRatingSection({ slug, activeCriteria }: { slug: string; activeCriteria: CriteriaKey[] }) {
  const { t } = useTranslation();
  const { data: status } = useEventRatingStatus(slug, true);
  const submitEventRating = useSubmitEventRating(slug);

  const [scores, setScores] = useState<Record<string, number>>({});
  const [comment, setComment] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  if (status?.submitted || submitted) {
    return (
      <div
        className="p-4 rounded border space-y-1"
        style={{
          backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, transparent)',
          borderColor: 'color-mix(in srgb, var(--color-primary) 25%, transparent)',
        }}
      >
        <p className="text-sm font-medium" style={{ color: 'var(--color-primary)' }}>
          ✓ {t('rating.event_already_rated')}
        </p>
      </div>
    );
  }

  const allScored = activeCriteria.every((c) => scores[c] !== undefined);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!allScored) return;
    setSubmitError(null);
    try {
      await submitEventRating.mutateAsync({
        criteria_scores: scores,
        comment: comment.trim() || null,
      });
      setSubmitted(true);
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        setSubmitted(true);
      } else {
        setSubmitError(t('errors.generic'));
      }
    }
  };

  const inputStyle = {
    backgroundColor: 'var(--color-background)',
    borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
    color: 'var(--color-text)',
  };

  return (
    <div
      className="p-4 rounded border space-y-4"
      style={{
        backgroundColor: 'var(--color-surface)',
        borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)',
      }}
    >
      <div>
        <h2 className="font-heading text-lg font-bold">{t('rating.event_title')}</h2>
        <p className="text-xs mt-0.5" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('rating.event_subtitle')}
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {activeCriteria.map((criterion) => (
          <div key={criterion} className="space-y-1.5">
            <span className="text-sm font-medium">{t(`criteria.${criterion}`)}</span>
            <div className="flex gap-1">
              {GRADES.map((g) => (
                <button
                  key={g}
                  type="button"
                  onClick={() => setScores((prev) => ({ ...prev, [criterion]: g }))}
                  className="flex-1 h-9 rounded font-mono font-bold text-sm border"
                  style={{
                    borderColor: scores[criterion] === g ? 'var(--color-primary)' : 'color-mix(in srgb, var(--color-text) 25%, transparent)',
                    backgroundColor: scores[criterion] === g ? 'var(--color-primary)' : 'transparent',
                    color: scores[criterion] === g ? 'var(--color-background)' : 'var(--color-text)',
                  }}
                >
                  {g}
                </button>
              ))}
            </div>
          </div>
        ))}

        <textarea
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          placeholder={t('rating.comment_placeholder')}
          maxLength={1000}
          rows={2}
          className="w-full px-3 py-2 rounded border text-sm resize-none"
          style={inputStyle}
        />

        {submitError && (
          <p className="text-sm" style={{ color: '#dc2626' }}>{submitError}</p>
        )}

        <button
          type="submit"
          disabled={submitEventRating.isPending || !allScored}
          className="w-full py-2 rounded font-medium text-sm"
          style={{
            backgroundColor: 'var(--color-primary)',
            color: 'var(--color-background)',
            opacity: submitEventRating.isPending || !allScored ? 0.6 : 1,
          }}
        >
          {submitEventRating.isPending ? t('common.loading') : t('rating.event_submit')}
        </button>
      </form>
    </div>
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
