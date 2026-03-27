import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useTournament } from '@/api/hooks/useTournaments';
import { useTournamentMembers, useAddMember, useRemoveMember } from '@/api/hooks/useMembers';
import { useAuthStore } from '@/stores/authStore';

export function HelperManagementPage() {
  const { slug = '' } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const { userId } = useAuthStore();
  const { data: tournament } = useTournament(slug);
  const { data: members, isLoading } = useTournamentMembers(slug);
  const addMember = useAddMember(slug);
  const removeMember = useRemoveMember(slug);

  const [inviteCode, setInviteCode] = useState('');
  const [addError, setAddError] = useState<string | null>(null);
  const [addSuccess, setAddSuccess] = useState<string | null>(null);

  const handleAdd = async () => {
    if (!inviteCode.trim()) return;
    setAddError(null);
    setAddSuccess(null);
    try {
      const preview = await addMember.mutateAsync(inviteCode.trim());
      setAddSuccess(t('helper.added_success', { name: preview.name }));
      setInviteCode('');
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 404) setAddError(t('helper.not_found'));
      else if (status === 409) setAddError(t('helper.already_member'));
      else setAddError(t('helper.error_add'));
    }
  };

  const helpers = members?.filter((m) => m.id !== userId) ?? [];

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Header */}
      <div>
        <Link
          to={`/dashboard/tournaments/${slug}`}
          className="text-xs hover:underline mb-1 block"
          style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
        >
          ← {t('table_mgmt.back_to_tournament')}
        </Link>
        <h1 className="font-heading text-2xl font-bold">
          {t('helper.section_title')} — {tournament?.name ?? slug}
        </h1>
      </div>

      {/* Add helper form */}
      <div className="flex flex-wrap gap-3 items-end">
        <div>
          <label className="block text-xs font-medium mb-1">{t('helper.add_label')}</label>
          <input
            type="text"
            value={inviteCode}
            onChange={(e) => setInviteCode(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
            placeholder={t('helper.add_placeholder')}
            className="px-3 py-1.5 rounded text-sm border font-mono"
            style={{
              backgroundColor: 'var(--color-surface)',
              borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
              color: 'var(--color-text)',
            }}
          />
        </div>
        <button
          onClick={handleAdd}
          disabled={addMember.isPending || !inviteCode.trim()}
          className="inline-flex items-center justify-center px-4 py-1.5 rounded text-sm font-medium"
          style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
        >
          {addMember.isPending ? t('common.loading') : t('helper.add_button')}
        </button>
        {addError && <p className="text-xs self-end" style={{ color: '#dc2626' }}>{addError}</p>}
        {addSuccess && <p className="text-xs self-end" style={{ color: '#16a34a' }}>{addSuccess}</p>}
      </div>

      {/* Helper list */}
      {isLoading ? (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('common.loading')}
        </p>
      ) : helpers.length > 0 ? (
        <div
          className="rounded border overflow-hidden"
          style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr style={{ backgroundColor: 'color-mix(in srgb, var(--color-primary) 8%, transparent)' }}>
                <th className="text-left px-4 py-2 font-medium">{t('rater.nickname_label')}</th>
                <th className="px-4 py-2" />
              </tr>
            </thead>
            <tbody>
              {helpers.map((m, i) => (
                <tr
                  key={m.id}
                  style={{ backgroundColor: i % 2 === 0 ? 'transparent' : 'color-mix(in srgb, var(--color-primary) 4%, transparent)' }}
                >
                  <td className="px-4 py-2">{m.name}</td>
                  <td className="px-4 py-2 text-right">
                    <button
                      onClick={async () => {
                        if (!window.confirm(t('helper.remove_confirm'))) return;
                        await removeMember.mutateAsync(m.id);
                      }}
                      disabled={removeMember.isPending}
                      className="text-xs px-2 py-1 rounded border"
                      style={{ color: '#dc2626', borderColor: '#dc2626' }}
                    >
                      {t('helper.remove_button')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
          {t('helper.no_helpers')}
        </p>
      )}
    </div>
  );
}
