import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import { useDeleteTournament } from '@/api/hooks/useTournaments';
import { useLeaveTournament } from '@/api/hooks/useMembers';
import type { TournamentMembership } from '@/api/hooks/useUser';

export function DashboardPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: profile, isLoading } = useMe();
  const deleteTournament = useDeleteTournament();
  const [copiedCode, setCopiedCode] = useState(false);

  const organizerMemberships =
    profile?.memberships.filter((m) => m.role === 'organizer') ?? [];
  const helperMemberships =
    profile?.memberships.filter((m) => m.role === 'helper') ?? [];

  const handleCopyCode = async () => {
    if (!profile?.helper_invite_code) return;
    await navigator.clipboard.writeText(profile.helper_invite_code);
    setCopiedCode(true);
    setTimeout(() => setCopiedCode(false), 2000);
  };

  const handleDelete = async (m: TournamentMembership) => {
    if (!window.confirm(t('dashboard.delete_confirm'))) return;
    await deleteTournament.mutateAsync(m.tournament_slug);
  };

  if (isLoading) {
    return (
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('common.loading')}
      </p>
    );
  }

  return (
    <div className="space-y-10">

      {/* ── Profile / invite code strip ───────────────────────── */}
      {profile && (
        <section
          className="flex flex-wrap items-center gap-4 p-4 rounded"
          style={{ backgroundColor: 'var(--color-surface)' }}
        >
          {profile.avatar_url && (
            <img src={profile.avatar_url} alt={profile.name} className="h-10 w-10 rounded-full" />
          )}
          <div className="flex-1 min-w-0">
            <p className="font-medium text-sm truncate">{profile.name}</p>
            <p className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {t('dashboard.invite_code_hint')}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <p className="sr-only" id="invite-code-label">{t('dashboard.invite_code_label')}</p>
            <code
              aria-labelledby="invite-code-label"
              className="font-mono text-sm px-2 py-1 rounded"
              style={{ backgroundColor: 'color-mix(in srgb, var(--color-primary) 12%, transparent)' }}
            >
              {profile.helper_invite_code}
            </code>
            <button
              onClick={handleCopyCode}
              className="text-xs px-3 py-1 rounded font-medium"
              style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
              aria-label={t('dashboard.invite_code_label')}
            >
              {copiedCode ? t('dashboard.invite_code_copied') : 'Copy'}
            </button>
          </div>
        </section>
      )}

      {/* ── My Tournaments (organizer) ─────────────────────────── */}
      <section aria-labelledby="org-heading">
        <div className="flex items-center justify-between mb-4">
          <h2 id="org-heading" className="font-heading text-xl font-bold">
            {t('dashboard.my_tournaments')}
          </h2>
          <Link
            to="/dashboard/tournaments/new"
            className="px-4 py-2 rounded text-sm font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
          >
            {t('dashboard.create_tournament')}
          </Link>
        </div>

        {organizerMemberships.length === 0 ? (
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('dashboard.no_tournaments')}
          </p>
        ) : (
          <ul className="space-y-2" role="list">
            {organizerMemberships.map((m) => (
              <li
                key={m.tournament_id}
                className="flex flex-wrap items-center justify-between gap-2 p-3 rounded"
                style={{ backgroundColor: 'var(--color-surface)' }}
              >
                <span className="font-medium text-sm">{m.tournament_name}</span>
                <div className="flex gap-2 flex-wrap">
                  <button
                    onClick={() => navigate(`/dashboard/tournaments/${m.tournament_slug}`)}
                    className="text-xs px-3 py-1 rounded border"
                    style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                  >
                    {t('dashboard.manage')}
                  </button>
                  <Link
                    to={`/dashboard/tournaments/${m.tournament_slug}/tables`}
                    className="text-xs px-3 py-1 rounded border"
                    style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
                  >
                    {t('dashboard.tables')}
                  </Link>
                  <Link
                    to={`/dashboard/tournaments/${m.tournament_slug}/results`}
                    className="text-xs px-3 py-1 rounded border"
                    style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
                  >
                    {t('results.title')}
                  </Link>
                  <button
                    onClick={() => handleDelete(m)}
                    disabled={deleteTournament.isPending}
                    className="text-xs px-3 py-1 rounded border"
                    style={{ borderColor: '#dc2626', color: '#dc2626' }}
                    aria-label={`${t('dashboard.delete_tournament')} ${m.tournament_name}`}
                  >
                    {t('dashboard.delete_tournament')}
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      {/* ── As Helper ─────────────────────────────────────────── */}
      <section aria-labelledby="helper-heading">
        <h2 id="helper-heading" className="font-heading text-xl font-bold mb-4">
          {t('dashboard.as_helper')}
        </h2>

        {helperMemberships.length === 0 ? (
          <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
            {t('dashboard.no_helper_tournaments')}
          </p>
        ) : (
          <ul className="space-y-2" role="list">
            {helperMemberships.map((m) => (
              <HelperRow key={m.tournament_id} membership={m} />
            ))}
          </ul>
        )}
      </section>

    </div>
  );
}

function HelperRow({ membership: m }: { membership: TournamentMembership }) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const leave = useLeaveTournament(m.tournament_slug);

  const handleLeave = async () => {
    if (!window.confirm(t('dashboard.leave_confirm'))) return;
    await leave.mutateAsync();
  };

  return (
    <li
      className="flex flex-wrap items-center justify-between gap-2 p-3 rounded"
      style={{ backgroundColor: 'var(--color-surface)' }}
    >
      <span className="font-medium text-sm">{m.tournament_name}</span>
      <div className="flex gap-2">
        <button
          onClick={() => navigate(`/dashboard/tournaments/${m.tournament_slug}`)}
          className="text-xs px-3 py-1 rounded border"
          style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
        >
          {t('dashboard.manage')}
        </button>
        <button
          onClick={handleLeave}
          disabled={leave.isPending}
          className="text-xs px-3 py-1 rounded border"
          style={{ borderColor: '#dc2626', color: '#dc2626' }}
          aria-label={`${t('dashboard.leave_tournament')} ${m.tournament_name}`}
        >
          {t('dashboard.leave_tournament')}
        </button>
      </div>
    </li>
  );
}
