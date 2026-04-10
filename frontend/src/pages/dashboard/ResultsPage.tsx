import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Globe, EyeOff } from 'lucide-react';
import { useTournamentResults, usePublishResults, useSetCommentApproved } from '@/api/hooks/useResults';
import { useTournament, useUpdateResultConfig, type CriteriaKey, type ResultConfig } from '@/api/hooks/useTournaments';
import { useMe } from '@/api/hooks/useUser';
import { ResultsView } from '@/components/results/ResultsView';

const ALL_CRITERIA: CriteriaKey[] = [
  'balance', 'aesthetics', 'terrain_density', 'labeling', 'overall', 'zone_a', 'zone_b',
  'catering', 'venue', 'organization',
];

export function ResultsPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data, isLoading, error } = useTournamentResults(slug ?? '');
  const { data: tournament } = useTournament(slug ?? '');
  const { data: me } = useMe();
  const publishResults = usePublishResults(slug ?? '');
  const setCommentApproved = useSetCommentApproved(slug ?? '');
  const updateResultConfig = useUpdateResultConfig(slug ?? '');

  const isMember = me?.memberships.some((m) => m.tournament_slug === slug) ?? false;
  const isOrganizer = me?.memberships.some(
    (m) => m.tournament_slug === slug && m.role === 'organizer'
  ) ?? false;

  // ── Result config form state ──────────────────────────────────────────────

  const [showComments, setShowComments] = useState(false);
  const [visibleCriteria, setVisibleCriteria] = useState<CriteriaKey[]>([]);
  const [rcSuccess, setRcSuccess] = useState('');
  const [rcError, setRcError] = useState('');

  useEffect(() => {
    if (!tournament?.result_config) return;
    const rc = tournament.result_config as ResultConfig;
    setShowComments(rc.show_comments);
    setVisibleCriteria((rc.visible_comment_criteria ?? []) as CriteriaKey[]);
  }, [tournament]);

  const toggleVisibleCriteria = (key: CriteriaKey) => {
    setVisibleCriteria((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key],
    );
  };

  const handleResultConfig = async (e: React.FormEvent) => {
    e.preventDefault();
    setRcError('');
    setRcSuccess('');
    try {
      await updateResultConfig.mutateAsync({
        show_comments: showComments,
        visible_comment_criteria: visibleCriteria,
      });
      setRcSuccess(t('result_config.save_success'));
    } catch {
      setRcError(t('result_config.error_save'));
    }
  };

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

      {/* ── Result Config (organizer only) ──────────────────────────────── */}
      {isOrganizer && (
        <section aria-labelledby="rc-heading">
          <h2 id="rc-heading" className="font-heading text-xl font-bold mb-4">
            {t('result_config.section_title')}
          </h2>
          <form onSubmit={handleResultConfig} className="space-y-4">
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input
                type="checkbox"
                checked={showComments}
                onChange={(e) => setShowComments(e.target.checked)}
              />
              {t('result_config.show_comments_label')}
            </label>

            {showComments && (
              <fieldset>
                <legend className="block text-sm font-medium mb-2">
                  {t('result_config.visible_criteria_label')}
                </legend>
                <div className="flex flex-wrap gap-3">
                  {ALL_CRITERIA.map((key) => (
                    <label key={key} className="flex items-center gap-1 text-sm cursor-pointer">
                      <input
                        type="checkbox"
                        checked={visibleCriteria.includes(key)}
                        onChange={() => toggleVisibleCriteria(key)}
                      />
                      {t(`criteria.${key}`)}
                    </label>
                  ))}
                </div>
              </fieldset>
            )}

            {rcError && (
              <p role="alert" className="text-sm" style={{ color: '#dc2626' }}>{rcError}</p>
            )}
            {rcSuccess && (
              <p role="status" className="text-sm" style={{ color: '#16a34a' }}>{rcSuccess}</p>
            )}

            <button
              type="submit"
              disabled={updateResultConfig.isPending}
              className="px-5 py-2 rounded font-medium text-sm"
              style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
            >
              {t('common.save')}
            </button>
          </form>
        </section>
      )}
    </div>
  );
}
