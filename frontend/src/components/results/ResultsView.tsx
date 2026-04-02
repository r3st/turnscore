import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, ChevronUp, Check, X } from 'lucide-react';
import type { TournamentResults, TableResult } from '@/api/hooks/useResults';

const EVENT_CRITERIA = new Set(['catering', 'venue', 'organization']);

interface ResultsViewProps {
  data: TournamentResults;
  onToggleApproval?: (commentId: string, approved: boolean) => void;
  approvalPending?: boolean;
}

export function ResultsView({ data, onToggleApproval, approvalPending }: ResultsViewProps) {
  const { t } = useTranslation();
  const tableCriteria = data.active_criteria.filter((k) => !EVENT_CRITERIA.has(k));

  if (data.ranking.length === 0) {
    return (
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('results.no_results')}
      </p>
    );
  }

  return (
    <div className="space-y-8">
      {/* Table Results */}
      <section aria-labelledby="table-results-heading">
        <h2 id="table-results-heading" className="font-heading text-lg font-semibold mb-3">
          {t('results.table_results')}
        </h2>
        <div className="space-y-3">
          {data.ranking.map((row, index) => (
            <TableResultRow
              key={row.table_number}
              rank={index + 1}
              result={row}
              activeCriteria={tableCriteria}
              onToggleApproval={onToggleApproval}
              approvalPending={approvalPending}
            />
          ))}
        </div>
      </section>

      {/* Event Rating — only when event averages exist */}
      {data.event_averages && Object.keys(data.event_averages).length > 0 && (
        <section aria-labelledby="event-rating-heading">
          <h2 id="event-rating-heading" className="font-heading text-lg font-semibold mb-1">
            {t('results.event_rating')}
          </h2>
          <p className="text-sm mb-3" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('results.event_rating_subtitle')}
          </p>
          <div
            className="rounded p-4 space-y-2"
            style={{
              backgroundColor: 'var(--color-surface)',
              border: '1px solid color-mix(in srgb, var(--color-text) 15%, transparent)',
            }}
          >
            {Object.entries(data.event_averages).map(([key, avg]) => (
              <div key={key} className="flex items-center justify-between gap-4">
                <span className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 80%, transparent)' }}>
                  {t(`criteria.${key}`, { defaultValue: key })}
                </span>
                <span className="font-heading text-lg font-bold tabular-nums" style={{ color: 'var(--color-primary)' }}>
                  {(avg as number).toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

interface TableResultRowProps {
  rank: number;
  result: TableResult;
  activeCriteria: string[];
  onToggleApproval?: (commentId: string, approved: boolean) => void;
  approvalPending?: boolean;
}

function TableResultRow({ rank, result, activeCriteria, onToggleApproval, approvalPending }: TableResultRowProps) {
  const { t } = useTranslation();
  const [expanded, setExpanded] = useState(false);

  const overallAvg = result.average_scores['overall'];
  const hasComments = result.comments != null && result.comments.length > 0;
  const hasDetails =
    activeCriteria.filter((k) => k !== 'overall').length > 0 ||
    result.zone_distribution.zone_a > 0 ||
    result.zone_distribution.zone_b > 0 ||
    hasComments;

  return (
    <div
      className="rounded overflow-hidden"
      style={{
        backgroundColor: 'var(--color-surface)',
        border: '1px solid color-mix(in srgb, var(--color-text) 15%, transparent)',
      }}
    >
      <div className="flex items-center gap-4 px-4 py-3">
        <span
          className="font-heading text-lg font-bold w-8 text-center shrink-0"
          style={{ color: rank <= 3 ? 'var(--color-primary)' : 'color-mix(in srgb, var(--color-text) 45%, transparent)' }}
          aria-label={`${t('results.rank')} ${rank}`}
        >
          {rank}
        </span>

        <div className="flex-1 min-w-0">
          <span className="font-medium text-sm">
            {result.table_name
              ? `${t('results.table')} ${result.table_number} — ${result.table_name}`
              : `${t('results.table')} ${result.table_number}`}
          </span>
          <p className="text-xs mt-0.5" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
            {t('results.ratings_count', { count: result.rating_count })}
          </p>
        </div>

        <div className="text-right shrink-0">
          <span className="font-heading text-xl font-bold" style={{ color: 'var(--color-primary)' }}>
            {overallAvg !== undefined ? overallAvg.toFixed(2) : '—'}
          </span>
          <p className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
            {t('results.average')}
          </p>
        </div>

        {hasDetails && (
          <button
            onClick={() => setExpanded((v) => !v)}
            aria-expanded={expanded}
            aria-label={expanded ? t('results.hide_criteria') : t('results.show_criteria')}
            className="shrink-0 p-1 rounded hover:opacity-70 transition-opacity"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            {expanded ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
          </button>
        )}
      </div>

      {expanded && (
        <div
          className="px-4 pb-4 space-y-4 pt-3"
          style={{ borderTop: '1px solid color-mix(in srgb, var(--color-text) 10%, transparent)' }}
        >
          {activeCriteria.filter((k) => k !== 'overall').length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-semibold uppercase tracking-wide" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
                {t('results.criteria_breakdown')}
              </p>
              <div className="grid grid-cols-2 gap-x-6 gap-y-1.5">
                {activeCriteria.filter((k) => k !== 'overall').map((key) => {
                  const avg = result.average_scores[key];
                  return (
                    <div key={key} className="flex items-center justify-between gap-2">
                      <span className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}>
                        {t(`criteria.${key}`, { defaultValue: key })}
                      </span>
                      <span className="text-xs font-semibold tabular-nums">
                        {avg !== undefined ? avg.toFixed(2) : '—'}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          <div className="space-y-1.5">
            <p className="text-xs font-semibold uppercase tracking-wide" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
              {t('results.zone_distribution')}
            </p>
            <div className="flex gap-6 text-xs">
              <ZoneBadge label={t('results.zone_a')} count={result.zone_distribution.zone_a} />
              <ZoneBadge label={t('results.zone_b')} count={result.zone_distribution.zone_b} />
              <ZoneBadge label={t('results.zone_none')} count={result.zone_distribution.none} muted />
            </div>
          </div>

          {hasComments && (
            <div className="space-y-2">
              <p className="text-xs font-semibold uppercase tracking-wide" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
                {t('results.comments')}
              </p>
              <ul className="space-y-2">
                {result.comments!.map((c) => (
                  <li
                    key={c.id}
                    className="flex items-start gap-3 px-3 py-2 rounded"
                    style={{
                      backgroundColor: 'color-mix(in srgb, var(--color-text) 5%, transparent)',
                      opacity: onToggleApproval && !c.approved ? 0.5 : 1,
                    }}
                  >
                    <span className="flex-1 text-sm">{c.comment}</span>
                    {onToggleApproval && (
                      <button
                        onClick={() => onToggleApproval(c.id, !c.approved)}
                        disabled={approvalPending}
                        className="shrink-0 inline-flex items-center gap-1 text-xs px-2.5 py-1 rounded font-medium"
                        style={{
                          backgroundColor: c.approved
                            ? 'color-mix(in srgb, var(--color-text) 10%, transparent)'
                            : 'var(--color-primary)',
                          color: c.approved ? 'var(--color-text)' : 'var(--color-background)',
                          opacity: approvalPending ? 0.6 : 1,
                        }}
                      >
                        {c.approved ? <X size={11} /> : <Check size={11} />}
                        {c.approved ? t('results.revoke') : t('results.approve')}
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function ZoneBadge({ label, count, muted }: { label: string; count: number; muted?: boolean }) {
  return (
    <span
      className="flex items-center gap-1"
      style={{ color: muted ? 'color-mix(in srgb, var(--color-text) 45%, transparent)' : 'var(--color-text)' }}
    >
      <span className="font-semibold tabular-nums">{count}</span>
      <span>{label}</span>
    </span>
  );
}
