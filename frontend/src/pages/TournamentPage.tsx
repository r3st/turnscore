import React, { useEffect, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Swords, Camera, X } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { useTournament } from '@/api/hooks/useTournaments';
import { useTheme } from '@/hooks/useTheme';
import { useAuthStore } from '@/stores/authStore';
import { useEventRatingStatus, useSubmitEventRating, useMyRatings } from '@/api/hooks/useEventRating';
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
  const isGuest = !isAuthenticated() || role !== 'rater';
  const { data: myRatingsData } = useMyRatings(slug, isRater);
  const ratedTables = myRatingsData?.rated_table_numbers ?? [];

  const isVotingActive = (() => {
    if (!tournament || tournament.status !== 'active') return false;
    const now = new Date();
    if (tournament.voting_start && now < new Date(tournament.voting_start)) return false;
    if (tournament.voting_end && now > new Date(tournament.voting_end)) return false;
    return true;
  })();

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

            {/* Game system */}
            {tournament.game_system && (
              <div
                className="self-start inline-flex items-center gap-2 text-sm font-medium px-3 py-1.5 rounded"
                style={{
                  backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, transparent)',
                  color: 'var(--color-primary)',
                }}
              >
                <Swords className="h-4 w-4 shrink-0" aria-hidden />
                {tournament.game_system}
              </div>
            )}

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
                className="p-4 rounded text-sm leading-relaxed"
                style={{ backgroundColor: 'var(--color-surface)', color: 'var(--color-text)' }}
              >
                <ReactMarkdown
                  components={{
                    h1: ({ children }) => <h1 className="text-xl font-bold font-heading mb-2 mt-4 first:mt-0">{children}</h1>,
                    h2: ({ children }) => <h2 className="text-lg font-bold font-heading mb-2 mt-4 first:mt-0">{children}</h2>,
                    h3: ({ children }) => <h3 className="text-base font-bold mb-1 mt-3 first:mt-0">{children}</h3>,
                    p: ({ children }) => <p className="mb-3 last:mb-0">{children}</p>,
                    ul: ({ children }) => <ul className="list-disc list-outside pl-5 mb-3 space-y-1">{children}</ul>,
                    ol: ({ children }) => <ol className="list-decimal list-outside pl-5 mb-3 space-y-1">{children}</ol>,
                    li: ({ children }) => <li>{children}</li>,
                    blockquote: ({ children }) => (
                      <blockquote
                        className="pl-3 mb-3 italic"
                        style={{ borderLeft: '3px solid var(--color-primary)', color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}
                      >
                        {children}
                      </blockquote>
                    ),
                    code: ({ children, className }) => {
                      const isBlock = !!className;
                      return isBlock ? (
                        <code className="block p-3 rounded text-xs font-mono mb-3 overflow-x-auto whitespace-pre" style={{ backgroundColor: 'color-mix(in srgb, var(--color-text) 8%, transparent)' }}>{children}</code>
                      ) : (
                        <code className="px-1 py-0.5 rounded text-xs font-mono" style={{ backgroundColor: 'color-mix(in srgb, var(--color-text) 8%, transparent)' }}>{children}</code>
                      );
                    },
                    pre: ({ children }) => <>{children}</>,
                    a: ({ href, children }) => (
                      <a href={href} target="_blank" rel="noopener noreferrer" style={{ color: 'var(--color-primary)' }}>
                        {children}
                      </a>
                    ),
                  }}
                >
                  {tournament.description}
                </ReactMarkdown>
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

            {/* CTA for guests when voting is active */}
            {isVotingActive && isGuest && (
              <div
                className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 p-4 rounded border"
                style={{
                  backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, transparent)',
                  borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
                }}
              >
                <div>
                  <p className="font-medium text-sm" style={{ color: 'var(--color-primary)' }}>
                    {t('tournament.voting_open_cta_title')}
                  </p>
                  <p className="text-xs mt-0.5" style={{ color: 'color-mix(in srgb, var(--color-text) 65%, transparent)' }}>
                    {t('tournament.voting_open_cta_text')}
                  </p>
                </div>
                <Link
                  to={`/rate-login/${slug}`}
                  className="shrink-0 inline-flex items-center justify-center px-4 py-2 rounded font-medium text-sm"
                  style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
                >
                  {t('tournament.voting_open_cta_button')}
                </Link>
              </div>
            )}

            {/* Event Rating — shown above tables when optional criteria exist, rater is logged in, and voting is open */}
            {isRater && isVotingActive && tournament.active_criteria?.some((c) => OPTIONAL_CRITERIA.includes(c as CriteriaKey)) && (
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
                    <TableCard
                      key={table.id}
                      slug={slug}
                      table={table}
                      showRateButton={isRater && isVotingActive}
                      isRated={ratedTables.includes(table.number)}
                      rateLabel={t('tournament.rate_table')}
                      ratedLabel={t('tournament.table_rated')}
                      tableLabel={t('table.number', { number: table.number })}
                    />
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

function TableCard({ slug, table, showRateButton, isRated, rateLabel, ratedLabel, tableLabel }: {
  slug: string;
  table: TableSummary;
  showRateButton: boolean;
  isRated: boolean;
  rateLabel: string;
  ratedLabel: string;
  tableLabel: string;
}) {
  const { t } = useTranslation();
  const [galleryIndex, setGalleryIndex] = useState<number | null>(null);
  const photos = table.photos ?? [];
  const photoCount = photos.length;

  const openGallery = (i: number) => setGalleryIndex(i);
  const closeGallery = () => setGalleryIndex(null);

  return (
    <>
      <div
        className="rounded border overflow-hidden"
        style={{
          backgroundColor: 'var(--color-surface)',
          borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)',
        }}
      >
        {/* Photo gallery preview */}
        {photoCount === 1 && (
          <button type="button" className="w-full block" onClick={() => openGallery(0)}>
            <img
              src={photos[0].thumbnail_url ?? photos[0].url}
              alt=""
              className="w-full h-36 object-cover cursor-zoom-in"
            />
          </button>
        )}
        {photoCount > 1 && (
          <div className="flex gap-1 p-1">
            {/* Main photo */}
            <button
              type="button"
              className="relative flex-1 overflow-hidden rounded"
              style={{ height: 128 }}
              onClick={() => openGallery(0)}
            >
              <img
                src={photos[0].thumbnail_url ?? photos[0].url}
                alt=""
                className="w-full h-full object-cover cursor-zoom-in"
              />
              <CategoryBadge category={photos[0].category} />
            </button>
            {/* Thumbnail column — max 2 visible */}
            <div className="flex flex-col gap-1">
              {photos.slice(1, 3).map((p, i) => {
                const isLast = i === 1;
                const remaining = photoCount - 3;
                return (
                  <button
                    key={p.id}
                    type="button"
                    className="relative overflow-hidden rounded"
                    style={{ width: 60, height: 62 }}
                    onClick={() => openGallery(i + 1)}
                  >
                    <img
                      src={p.thumbnail_url ?? p.url}
                      alt=""
                      className="w-full h-full object-cover cursor-zoom-in"
                    />
                    <CategoryBadge category={p.category} />
                    {isLast && remaining > 0 && (
                      <div
                        className="absolute inset-0 flex items-center justify-center text-sm font-bold text-white"
                        style={{ backgroundColor: 'rgba(0,0,0,0.55)' }}
                      >
                        +{remaining}
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        )}

        <div className="p-3 flex items-center justify-between gap-2">
          <div className="min-w-0">
            <p className="font-medium text-sm truncate">{tableLabel}</p>
            {table.name && (
              <p className="text-xs truncate" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
                {table.name}
              </p>
            )}
            {photoCount > 0 && (
              <p className="flex items-center gap-1 text-xs mt-0.5" style={{ color: 'color-mix(in srgb, var(--color-text) 55%, transparent)' }}>
                <Camera className="h-3 w-3" aria-hidden />
                {photoCount} {t('table.photos')}
              </p>
            )}
          </div>
          {showRateButton && !isRated && (
            <Link
              to={`/rate/${slug}/${table.number}`}
              className="shrink-0 inline-flex items-center justify-center text-xs px-3 py-1.5 rounded font-medium"
              style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
            >
              {rateLabel}
            </Link>
          )}
          {showRateButton && isRated && (
            <span
              className="shrink-0 inline-flex items-center gap-1 text-xs px-3 py-1.5 rounded font-medium"
              style={{
                backgroundColor: 'color-mix(in srgb, var(--color-primary) 15%, transparent)',
                color: 'var(--color-primary)',
              }}
            >
              ✓ {ratedLabel}
            </span>
          )}
        </div>
      </div>

      {galleryIndex !== null && (
        <GalleryLightbox photos={photos} initialIndex={galleryIndex} onClose={closeGallery} />
      )}
    </>
  );
}

function CategoryBadge({ category }: { category?: string | null }) {
  if (!category || category === 'general') return null;
  return (
    <span
      className="absolute bottom-1 right-1 text-xs font-bold px-1.5 py-0.5 rounded leading-none"
      style={{ backgroundColor: 'rgba(0,0,0,0.6)', color: 'white' }}
    >
      {category === 'zone_a' ? 'A' : 'B'}
    </span>
  );
}

// ── GalleryLightbox ───────────────────────────────────────────────────────────

function GalleryLightbox({ photos, initialIndex, onClose }: {
  photos: TableSummary['photos'];
  initialIndex: number;
  onClose: () => void;
}) {
  const [index, setIndex] = useState(initialIndex);
  const total = photos.length;
  const prev = () => setIndex((i) => (i - 1 + total) % total);
  const next = () => setIndex((i) => (i + 1) % total);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowLeft') setIndex((i) => (i - 1 + total) % total);
      if (e.key === 'ArrowRight') setIndex((i) => (i + 1) % total);
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose, total]);

  // Touch swipe
  const touchStartX = useRef<number | null>(null);
  const handleTouchStart = (e: React.TouchEvent) => { touchStartX.current = e.touches[0].clientX; };
  const handleTouchEnd = (e: React.TouchEvent) => {
    if (touchStartX.current === null) return;
    const delta = e.changedTouches[0].clientX - touchStartX.current;
    if (Math.abs(delta) > 50) { delta < 0 ? next() : prev(); }
    touchStartX.current = null;
  };

  const current = photos[index];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{ backgroundColor: 'rgba(0,0,0,0.90)' }}
      onClick={onClose}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
      role="dialog"
      aria-modal="true"
    >
      {/* Close */}
      <button
        type="button"
        className="absolute top-4 right-4 p-2 rounded-full"
        style={{ backgroundColor: 'rgba(255,255,255,0.15)', color: 'white' }}
        onClick={onClose}
        aria-label="Close"
      >
        <X className="h-5 w-5" />
      </button>

      {/* Counter + category */}
      <div
        className="absolute top-4 left-4 flex items-center gap-2 text-sm text-white"
        style={{ textShadow: '0 1px 3px rgba(0,0,0,0.8)' }}
      >
        <span>{index + 1} / {total}</span>
        {current.category && current.category !== 'general' && (
          <span
            className="px-2 py-0.5 rounded text-xs font-bold"
            style={{ backgroundColor: 'rgba(255,255,255,0.2)' }}
          >
            Zone {current.category === 'zone_a' ? 'A' : 'B'}
          </span>
        )}
      </div>

      {/* Image */}
      <img
        src={current.url}
        alt=""
        className="max-h-[85vh] max-w-[90vw] rounded object-contain"
        onClick={(e) => e.stopPropagation()}
      />

      {/* Prev / Next */}
      {total > 1 && (
        <>
          <button
            type="button"
            className="absolute left-3 top-1/2 -translate-y-1/2 p-3 rounded-full"
            style={{ backgroundColor: 'rgba(255,255,255,0.15)', color: 'white' }}
            onClick={(e) => { e.stopPropagation(); prev(); }}
            aria-label="Previous photo"
          >
            ‹
          </button>
          <button
            type="button"
            className="absolute right-3 top-1/2 -translate-y-1/2 p-3 rounded-full"
            style={{ backgroundColor: 'rgba(255,255,255,0.15)', color: 'white' }}
            onClick={(e) => { e.stopPropagation(); next(); }}
            aria-label="Next photo"
          >
            ›
          </button>
        </>
      )}
    </div>
  );
}
