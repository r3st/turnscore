import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { useGetTable, useSubmitRating } from '@/api/hooks/useTables';
import { useAuthStore } from '@/stores/authStore';
import type { components } from '@/api/generated/api';

type CriteriaKey = components['schemas']['CriteriaKey'];
type PlayedZone = components['schemas']['PlayedZone'];
type Photo = components['schemas']['Photo'];

const GRADES = [1, 2, 3, 4, 5, 6] as const;

export function RatePage() {
  const { slug = '', tableNum = '' } = useParams<{ slug: string; tableNum: string }>();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { isAuthenticated, hasRole } = useAuthStore();

  const { data: table, isLoading } = useGetTable(slug, tableNum);
  const submitRating = useSubmitRating(slug, tableNum);

  const [scores, setScores] = useState<Record<string, number>>({});
  const [playedZone, setPlayedZone] = useState<PlayedZone>('none');
  const [comment, setComment] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [alreadyRated, setAlreadyRated] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Redirect unauthenticated raters to login
  if (!isAuthenticated() || !hasRole('rater')) {
    return (
      <AppLayout>
        <main className="container mx-auto px-4 py-12 max-w-xl text-center space-y-4">
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('auth.rater_login_title')}
          </p>
          <Link
            to={`/rate-login/${slug}`}
            className="inline-flex items-center justify-center px-6 py-2 rounded font-medium text-sm"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            {t('auth.login_button')}
          </Link>
        </main>
      </AppLayout>
    );
  }

  if (isLoading) {
    return (
      <AppLayout>
        <main className="container mx-auto px-4 py-8">
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('common.loading')}
          </p>
        </main>
      </AppLayout>
    );
  }

  if (!table) {
    return (
      <AppLayout>
        <main className="container mx-auto px-4 py-8">
          <p className="text-sm" style={{ color: '#dc2626' }}>{t('errors.not_found')}</p>
        </main>
      </AppLayout>
    );
  }

  const activeCriteria = table.active_criteria as CriteriaKey[];
  const zonePhotos = {
    zone_a: table.photos.filter((p) => p.category === 'zone_a'),
    zone_b: table.photos.filter((p) => p.category === 'zone_b'),
    general: table.photos.filter((p) => p.category === 'general'),
  };

  const allScored = activeCriteria.every((c) => scores[c] !== undefined);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!allScored) return;
    setError(null);
    try {
      await submitRating.mutateAsync({
        criteria_scores: scores,
        played_zone: playedZone,
        comment: comment.trim() || null,
      });
      setSubmitted(true);
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        setAlreadyRated(true);
      } else {
        setError(t('errors.generic'));
      }
    }
  };

  if (submitted || alreadyRated) {
    return (
      <AppLayout>
        <main className="container mx-auto px-4 py-12 max-w-xl text-center space-y-4">
          <p className="font-heading text-2xl font-bold">
            {alreadyRated ? t('rating.already_rated') : t('rating.success')}
          </p>
          <Link
            to={`/t/${slug}`}
            className="inline-flex items-center justify-center px-6 py-2 rounded font-medium text-sm"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            {t('common.back')}
          </Link>
        </main>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8 max-w-xl space-y-8">
        {/* Header */}
        <div>
          <Link
            to={`/t/${slug}`}
            className="text-xs hover:underline mb-1 block"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            ← {table.tournament_name}
          </Link>
          <h1 className="font-heading text-2xl font-bold">
            {t('table.number', { number: table.number })}
            {table.name && <span className="ml-2 text-base font-normal">— {table.name}</span>}
          </h1>
          {table.description && (
            <p
              className="mt-1 text-sm"
              style={{ color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}
            >
              {table.description}
            </p>
          )}
        </div>

        {/* General photos */}
        {zonePhotos.general.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {zonePhotos.general.map((p) => (
              <img
                key={p.id}
                src={p.url}
                alt=""
                className="h-28 w-28 object-cover rounded"
              />
            ))}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-8">
          {/* Criteria ratings */}
          <div className="space-y-6">
            {activeCriteria.map((criterion) => (
              <CriterionRow
                key={criterion}
                criterion={criterion}
                value={scores[criterion]}
                onChange={(v) => setScores((prev) => ({ ...prev, [criterion]: v }))}
                photos={
                  criterion === 'zone_a'
                    ? zonePhotos.zone_a
                    : criterion === 'zone_b'
                    ? zonePhotos.zone_b
                    : []
                }
              />
            ))}
          </div>

          {/* Played zone */}
          <div>
            <p className="text-sm font-medium mb-2">{t('rating.played_zone')}</p>
            <div className="flex flex-wrap gap-2">
              {(['none', 'zone_a', 'zone_b'] as PlayedZone[]).map((zone) => (
                <button
                  key={zone}
                  type="button"
                  onClick={() => setPlayedZone(zone)}
                  className="px-4 py-1.5 rounded border text-sm font-medium"
                  style={{
                    borderColor:
                      playedZone === zone
                        ? 'var(--color-primary)'
                        : 'color-mix(in srgb, var(--color-text) 30%, transparent)',
                    backgroundColor:
                      playedZone === zone
                        ? 'color-mix(in srgb, var(--color-primary) 15%, transparent)'
                        : 'transparent',
                    color:
                      playedZone === zone
                        ? 'var(--color-primary)'
                        : 'var(--color-text)',
                  }}
                >
                  {zone === 'none'
                    ? t('rating.played_none')
                    : zone === 'zone_a'
                    ? t('rating.played_zone_a')
                    : t('rating.played_zone_b')}
                </button>
              ))}
            </div>
          </div>

          {/* Comment */}
          <div>
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder={t('rating.comment_placeholder')}
              maxLength={1000}
              rows={3}
              className="w-full px-3 py-2 rounded border text-sm resize-none"
              style={{
                backgroundColor: 'var(--color-surface)',
                borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
                color: 'var(--color-text)',
              }}
            />
            <p
              className="text-xs mt-1 text-right"
              style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}
            >
              {comment.length} / 1000
            </p>
          </div>

          {error && <p className="text-sm" style={{ color: '#dc2626' }}>{error}</p>}

          <button
            type="submit"
            disabled={submitRating.isPending || !allScored}
            className="w-full py-3 rounded font-medium text-sm"
            style={{
              backgroundColor: 'var(--color-primary)',
              color: 'var(--color-background)',
              opacity: submitRating.isPending || !allScored ? 0.6 : 1,
            }}
          >
            {submitRating.isPending ? t('rating.submitting') : t('rating.submit')}
          </button>
        </form>
      </main>
    </AppLayout>
  );
}

// ── CriterionRow ──────────────────────────────────────────────────────────────

function CriterionRow({
  criterion,
  value,
  onChange,
  photos,
}: {
  criterion: CriteriaKey;
  value: number | undefined;
  onChange: (v: number) => void;
  photos: Photo[];
}) {
  const { t } = useTranslation();
  const label = t(`criteria.${criterion}`);

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{label}</span>
        {value !== undefined && (
          <span
            className="text-xs"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            {t(`rating.grade_${value}`)}
          </span>
        )}
      </div>

      {/* Zone photos inline */}
      {photos.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {photos.map((p) => (
            <img
              key={p.id}
              src={p.thumbnail_url ?? p.url}
              alt=""
              className="h-20 w-20 object-cover rounded"
            />
          ))}
        </div>
      )}

      {/* 1–6 grade buttons */}
      <div className="flex gap-1">
        <span
          className="text-xs self-center mr-1 w-14 text-right"
          style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}
        >
          {t('rating.grade_1_short')}
        </span>
        {GRADES.map((g) => (
          <button
            key={g}
            type="button"
            onClick={() => onChange(g)}
            className="flex-1 h-10 rounded font-mono font-bold text-sm border"
            style={{
              borderColor:
                value === g
                  ? 'var(--color-primary)'
                  : 'color-mix(in srgb, var(--color-text) 25%, transparent)',
              backgroundColor:
                value === g
                  ? 'var(--color-primary)'
                  : 'transparent',
              color: value === g ? 'var(--color-background)' : 'var(--color-text)',
            }}
          >
            {g}
          </button>
        ))}
        <span
          className="text-xs self-center ml-1 w-14"
          style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}
        >
          {t('rating.grade_6_short')}
        </span>
      </div>
    </div>
  );
}
