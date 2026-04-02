import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Globe, EyeOff } from 'lucide-react';
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

  return (
    <div className="space-y-8 max-w-4xl">
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

      <ResultsView
        data={data}
        onToggleApproval={isMember
          ? (commentId, approved) => setCommentApproved.mutate({ commentId, approved })
          : undefined}
        approvalPending={setCommentApproved.isPending}
      />
    </div>
  );
}
