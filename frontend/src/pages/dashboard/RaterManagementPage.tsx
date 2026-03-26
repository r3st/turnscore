import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useTournament } from '@/api/hooks/useTournaments';
import { useListRaters, useCreateRater, useExportRatersPdf } from '@/api/hooks/useRaters';

export function RaterManagementPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { data: tournament } = useTournament(slug);
  const { data: raters, isLoading } = useListRaters(slug);
  const createRater = useCreateRater(slug);
  const exportPdf = useExportRatersPdf(slug);

  const [nickname, setNickname] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleAdd = async () => {
    if (!nickname.trim()) return;
    setError(null);
    try {
      await createRater.mutateAsync({
        nickname: nickname.trim(),
        code: code.trim() || undefined,
      });
      setNickname('');
      setCode('');
    } catch {
      setError(t('rater.error_add'));
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <Link
            to={`/dashboard/tournaments/${slug}`}
            className="text-xs hover:underline mb-1 block"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            ← {t('table_mgmt.back_to_tournament')}
          </Link>
          <h1 className="font-heading text-2xl font-bold">
            {t('rater.section_title')} — {tournament?.name ?? slug}
          </h1>
        </div>
        {raters && raters.length > 0 && (
          <button
            onClick={() => exportPdf.mutate()}
            disabled={exportPdf.isPending}
            className="text-sm px-4 py-2 rounded border font-medium"
            style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
          >
            {exportPdf.isPending ? t('common.loading') : t('rater.export_pdf')}
          </button>
        )}
      </div>

      {/* Add rater form */}
      <div className="flex flex-wrap gap-3 items-end">
        <div>
          <label className="block text-xs font-medium mb-1">{t('rater.nickname_label')}</label>
          <input
            type="text"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
            placeholder={t('rater.nickname_placeholder')}
            className="px-3 py-1.5 rounded text-sm border"
            style={{
              backgroundColor: 'var(--color-surface)',
              borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
              color: 'var(--color-text)',
            }}
          />
        </div>
        <div>
          <label className="block text-xs font-medium mb-1">{t('rater.code_label')}</label>
          <input
            type="text"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder={t('rater.code_placeholder')}
            maxLength={6}
            className="px-3 py-1.5 rounded text-sm border w-36"
            style={{
              backgroundColor: 'var(--color-surface)',
              borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
              color: 'var(--color-text)',
            }}
          />
        </div>
        <button
          onClick={handleAdd}
          disabled={createRater.isPending || !nickname.trim()}
          className="inline-flex items-center justify-center px-4 py-1.5 rounded text-sm font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
        >
          {createRater.isPending ? t('common.loading') : t('rater.add_button')}
        </button>
        {error && <p className="text-xs self-end" style={{ color: '#dc2626' }}>{error}</p>}
      </div>

      {/* Rater list */}
      {isLoading ? (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('common.loading')}
        </p>
      ) : raters && raters.length > 0 ? (
        <div
          className="rounded border overflow-hidden"
          style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr style={{ backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, transparent)' }}>
                <th className="text-left px-4 py-2 font-medium">{t('rater.nickname_label')}</th>
                <th className="text-left px-4 py-2 font-medium">{t('rater.code_label').split(' ')[0]}</th>
              </tr>
            </thead>
            <tbody>
              {raters.map((rater, i) => (
                <tr
                  key={rater.id}
                  style={{ backgroundColor: i % 2 === 0 ? 'transparent' : 'color-mix(in srgb, var(--color-primary) 4%, transparent)' }}
                >
                  <td className="px-4 py-2">{rater.nickname}</td>
                  <td className="px-4 py-2 font-mono text-xs">{rater.code}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('rater.no_raters')}
        </p>
      )}
    </div>
  );
}
