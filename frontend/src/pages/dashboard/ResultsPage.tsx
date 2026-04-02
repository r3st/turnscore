import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Globe, EyeOff, Check, X } from 'lucide-react';
import { useTournamentResults, usePublishResults, useSetCommentApproved } from '@/api/hooks/useResults';
import { useTournament } from '@/api/hooks/useTournaments';
import { useMe } from '@/api/hooks/useUser';
import { ResultsView } from '@/components/results/ResultsView';

export function ResultsPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data, isLoading, error } = useTournamentResults(slug ?? '');
  const { data: tournament } = useTournament(slug ?? '');
  const { data: me } = useMe();
  const publishResults = usePublishResults(slug ?? '');
  const setCommentApproved = useSetCommentApproved(slug ?? '');

  const isMember = me?.memberships.some((m) => m.tournament_slug === slug) ?? false;

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  if (error || !data) {
    return <p className="text-sm" style={{ color: '#dc2626' }}>{t('errors.generic')}</p>;
  }

  const isPublished = tournament?.results_published ?? false;

  // Collect all comments across all tables for moderation panel.
  // Backend always returns comments for managers regardless of show_comments setting.
  const allComments = data.ranking.flatMap((row) =>
    (row.comments ?? []).map((c) => ({
      ...c,
      table_number: row.table_number,
      table_name: row.table_name,
    }))
  );

  return (
    <div className="space-y-8 max-w-4xl">
      {/* Header with publish toggle */}
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="font-heading text-2xl font-bold">{t('results.title')}</h1>
          <p className="text-sm font-medium mt-1" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {data.tournament_name}
          </p>
        </div>
        {isMember && (
          <div className="flex items-center gap-3">
            {isPublished && (
              <span
                className="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium"
                style={{ backgroundColor: 'color-mix(in srgb, #16a34a 15%, transparent)', color: '#16a34a' }}
              >
                <Globe size={12} />
                {t('results.published_badge')}
              </span>
            )}
            <button
              onClick={() => publishResults.mutate({ published: !isPublished })}
              disabled={publishResults.isPending}
              className="inline-flex items-center gap-2 text-sm px-4 py-2 rounded font-medium"
              style={{
                backgroundColor: isPublished ? 'color-mix(in srgb, var(--color-text) 10%, transparent)' : 'var(--color-primary)',
                color: isPublished ? 'var(--color-text)' : 'var(--color-background)',
                opacity: publishResults.isPending ? 0.6 : 1,
              }}
            >
              {isPublished ? <EyeOff size={15} /> : <Globe size={15} />}
              {publishResults.isPending
                ? t('common.loading')
                : isPublished
                ? t('results.unpublish')
                : t('results.publish')}
            </button>
          </div>
        )}
      </div>

      {/* Ranking */}
      <ResultsView data={data} />

      {/* Comment moderation — always visible for members, regardless of show_comments setting */}
      {isMember && (
        <section aria-labelledby="moderation-heading">
          <h2 id="moderation-heading" className="font-heading text-lg font-semibold mb-1">
            {t('results.moderation_heading')}
          </h2>
          <p className="text-sm mb-4" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('results.moderation_hint')}
          </p>

          {allComments.length === 0 ? (
            <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
              {t('results.no_comments')}
            </p>
          ) : (
            <div className="space-y-2">
              {allComments.map((c) => (
                <div
                  key={c.id}
                  className="flex items-start gap-3 px-4 py-3 rounded"
                  style={{
                    backgroundColor: 'var(--color-surface)',
                    border: '1px solid color-mix(in srgb, var(--color-text) 15%, transparent)',
                    opacity: setCommentApproved.isPending ? 0.7 : 1,
                  }}
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-xs mb-1" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
                      {c.table_name
                        ? `Tisch ${c.table_number} — ${c.table_name}`
                        : `Tisch ${c.table_number}`}
                    </p>
                    <p className="text-sm">{c.comment}</p>
                  </div>
                  <div className="shrink-0 flex items-center gap-2">
                    {c.approved && (
                      <span
                        className="text-xs px-2 py-0.5 rounded-full font-medium"
                        style={{ backgroundColor: 'color-mix(in srgb, #16a34a 15%, transparent)', color: '#16a34a' }}
                      >
                        {t('results.approved_badge')}
                      </span>
                    )}
                    <button
                      onClick={() => setCommentApproved.mutate({ commentId: c.id, approved: !c.approved })}
                      disabled={setCommentApproved.isPending}
                      className="inline-flex items-center gap-1 text-xs px-3 py-1.5 rounded font-medium"
                      style={{
                        backgroundColor: c.approved
                          ? 'color-mix(in srgb, var(--color-text) 10%, transparent)'
                          : 'var(--color-primary)',
                        color: c.approved ? 'var(--color-text)' : 'var(--color-background)',
                      }}
                    >
                      {c.approved ? <X size={12} /> : <Check size={12} />}
                      {c.approved ? t('results.revoke') : t('results.approve')}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}
    </div>
  );
}
