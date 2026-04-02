import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import {
  useTournament,
  useCreateTournament,
  useUpdateTournament,
  useUpdateResultConfig,
  type CriteriaKey,
} from '@/api/hooks/useTournaments';

// ── Constants ─────────────────────────────────────────────────────────────────

const CORE_CRITERIA: CriteriaKey[] = [
  'balance', 'aesthetics', 'terrain_density', 'labeling', 'overall', 'zone_a', 'zone_b',
];
const OPTIONAL_CRITERIA: CriteriaKey[] = ['catering', 'venue', 'organization'];
const ALL_CRITERIA = [...CORE_CRITERIA, ...OPTIONAL_CRITERIA];
const STATUSES = ['draft', 'active', 'archived'] as const;

// ── Page ──────────────────────────────────────────────────────────────────────

export function TournamentEditPage() {
  const { slug } = useParams<{ slug?: string }>();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: me } = useMe();
  const isNew = !slug;
  const isOrganizer = me?.memberships.some(
    (m) => m.tournament_slug === slug && m.role === 'organizer'
  ) ?? false;

  const { data: tournament, isLoading } = useTournament(slug ?? '');
  const createTournament = useCreateTournament();
  const updateTournament = useUpdateTournament(slug ?? '');
  const updateResultConfig = useUpdateResultConfig(slug ?? '');

  // ── Form state ────────────────────────────────────────────────────────────

  const [name, setName] = useState('');
  const [type, setType] = useState<'fantasy' | 'scifi'>('fantasy');
  const [gameSystem, setGameSystem] = useState('');
  const [tableCount, setTableCount] = useState(10);
  const [description, setDescription] = useState('');
  const [location, setLocation] = useState('');
  const [eventDate, setEventDate] = useState('');
  const [votingStart, setVotingStart] = useState('');
  const [votingEnd, setVotingEnd] = useState('');
  const [status, setStatus] = useState('draft');
  const [links, setLinks] = useState<{ url: string; label: string }[]>([]);
  const [activeCriteria, setActiveCriteria] = useState<CriteriaKey[]>(CORE_CRITERIA);
  const [formError, setFormError] = useState('');
  const [formSuccess, setFormSuccess] = useState('');

  // ── Result config state ───────────────────────────────────────────────────

  const [showComments, setShowComments] = useState(false);
  const [visibleCriteria, setVisibleCriteria] = useState<CriteriaKey[]>([]);
  const [rcSuccess, setRcSuccess] = useState('');
  const [rcError, setRcError] = useState('');

  // ── Helper invite state ───────────────────────────────────────────────────


  // ── Populate form from loaded tournament ──────────────────────────────────

  useEffect(() => {
    if (!tournament) return;
    setName(tournament.name);
    setType(tournament.type as 'fantasy' | 'scifi');
    setGameSystem(tournament.game_system ?? '');
    setTableCount(tournament.table_count);
    setDescription(tournament.description ?? '');
    setLocation(tournament.location ?? '');
    setEventDate(tournament.event_date ?? '');
    setVotingStart(toDatetimeLocal(tournament.voting_start ?? ''));
    setVotingEnd(toDatetimeLocal(tournament.voting_end ?? ''));
    setStatus(tournament.status);
    setLinks(tournament.links ?? []);
    setActiveCriteria(
      (tournament.active_criteria ?? CORE_CRITERIA) as CriteriaKey[],
    );
  }, [tournament]);

  // ── Submit handlers ───────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormError('');
    setFormSuccess('');
    if (!name.trim()) {
      setFormError(t('tournament_form.error_name_required'));
      return;
    }
    try {
      if (isNew) {
        const created = await createTournament.mutateAsync({
          name,
          type,
          game_system: gameSystem || null,
          table_count: tableCount,
          description: description || null,
          location: location || null,
          event_date: eventDate || null,
          voting_start: votingStart ? toISO(votingStart) : null,
          voting_end: votingEnd ? toISO(votingEnd) : null,
          links,
          active_criteria: activeCriteria,
        });
        navigate(`/dashboard/tournaments/${created.slug}`);
      } else {
        await updateTournament.mutateAsync({
          name,
          type,
          game_system: gameSystem || null,
          description: description || null,
          location: location || null,
          event_date: eventDate || null,
          status: status as typeof STATUSES[number],
          voting_start: votingStart ? toISO(votingStart) : null,
          voting_end: votingEnd ? toISO(votingEnd) : null,
          links,
          active_criteria: activeCriteria,
        });
        setFormSuccess(t('tournament_form.save_success'));
      }
    } catch {
      setFormError(isNew ? t('tournament_form.error_create') : t('tournament_form.error_save'));
    }
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

  const toggleOptionalCriteria = (key: CriteriaKey) => {
    setActiveCriteria((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key],
    );
  };

  const toggleVisibleCriteria = (key: CriteriaKey) => {
    setVisibleCriteria((prev) =>
      prev.includes(key) ? prev.filter((k) => k !== key) : [...prev, key],
    );
  };

  // ── Early returns ─────────────────────────────────────────────────────────

  if (!isNew && isLoading) {
    return (
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
        {t('common.loading')}
      </p>
    );
  }

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div className="space-y-10 max-w-2xl">

      {/* ── Tournament Form ────────────────────────────────────── */}
      <section>
        <h1 className="font-heading text-2xl font-bold mb-6">
          {isNew ? t('tournament_form.create_title') : t('tournament_form.edit_title')}
        </h1>

        <form onSubmit={handleSubmit} className="space-y-5" noValidate>

          {/* Name */}
          <Field label={t('tournament_form.name_label')} htmlFor="t-name" required>
            <input
              id="t-name"
              type="text"
              required
              maxLength={200}
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('tournament_form.name_placeholder')}
              className="w-full px-3 py-2 rounded border text-sm"
              style={inputStyle}
            />
          </Field>

          {/* Type */}
          <Field label={t('tournament_form.type_label')} htmlFor="t-type">
            <div className="flex gap-4" role="radiogroup" aria-label={t('tournament_form.type_label')}>
              {(['fantasy', 'scifi'] as const).map((v) => (
                <label key={v} className="flex items-center gap-2 text-sm cursor-pointer">
                  <input
                    type="radio"
                    name="type"
                    value={v}
                    checked={type === v}
                    onChange={() => setType(v)}
                  />
                  {t(`tournament_form.type_${v}`)}
                </label>
              ))}
            </div>
          </Field>

          {/* Game System */}
          <Field label={t('tournament_form.game_system_label')} htmlFor="t-game-system">
            <input
              id="t-game-system"
              type="text"
              maxLength={100}
              value={gameSystem}
              onChange={(e) => setGameSystem(e.target.value)}
              placeholder={t('tournament_form.game_system_placeholder')}
              className="w-full px-3 py-2 rounded border text-sm"
              style={inputStyle}
            />
          </Field>

          {/* Table Count — new only */}
          {isNew && (
            <Field
              label={t('tournament_form.table_count_label')}
              htmlFor="t-table-count"
              hint={t('tournament_form.table_count_hint')}
              required
            >
              <input
                id="t-table-count"
                type="number"
                required
                min={1}
                max={100}
                value={tableCount}
                onChange={(e) => setTableCount(Number(e.target.value))}
                className="w-28 px-3 py-2 rounded border text-sm"
                style={inputStyle}
              />
            </Field>
          )}

          {/* Status — edit only */}
          {!isNew && (
            <Field label={t('tournament_form.status_label')} htmlFor="t-status">
              <select
                id="t-status"
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                className="px-3 py-2 rounded border text-sm"
                style={inputStyle}
              >
                {STATUSES.map((s) => (
                  <option key={s} value={s}>{t(`tournament.status_${s}`)}</option>
                ))}
              </select>
            </Field>
          )}

          {/* Location */}
          <Field label={t('tournament_form.location_label')} htmlFor="t-location">
            <input
              id="t-location"
              type="text"
              maxLength={200}
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              placeholder={t('tournament_form.location_placeholder')}
              className="w-full px-3 py-2 rounded border text-sm"
              style={inputStyle}
            />
          </Field>

          {/* Event Date */}
          <Field label={t('tournament_form.event_date_label')} htmlFor="t-event-date">
            <input
              id="t-event-date"
              type="date"
              value={eventDate}
              onChange={(e) => setEventDate(e.target.value)}
              className="px-3 py-2 rounded border text-sm"
              style={inputStyle}
            />
          </Field>

          {/* Voting Period */}
          <div className="grid grid-cols-2 gap-4">
            <Field label={t('tournament_form.voting_start_label')} htmlFor="t-voting-start">
              <input
                id="t-voting-start"
                type="datetime-local"
                value={votingStart}
                onChange={(e) => setVotingStart(e.target.value)}
                className="w-full px-3 py-2 rounded border text-sm"
                style={inputStyle}
              />
            </Field>
            <Field label={t('tournament_form.voting_end_label')} htmlFor="t-voting-end">
              <input
                id="t-voting-end"
                type="datetime-local"
                value={votingEnd}
                onChange={(e) => setVotingEnd(e.target.value)}
                className="w-full px-3 py-2 rounded border text-sm"
                style={inputStyle}
              />
            </Field>
          </div>

          {/* Description */}
          <Field label={t('tournament_form.description_label')} htmlFor="t-description">
            <textarea
              id="t-description"
              rows={4}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('tournament_form.description_placeholder')}
              className="w-full px-3 py-2 rounded border text-sm resize-y"
              style={inputStyle}
            />
            <p className="text-xs mt-1" style={{ color: 'color-mix(in srgb, var(--color-text) 50%, transparent)' }}>
              {t('tournament_form.description_markdown_hint')}{' '}
              <a
                href="https://www.markdownguide.org/basic-syntax/"
                target="_blank"
                rel="noopener noreferrer"
                style={{ color: 'var(--color-primary)' }}
              >
                Markdown
              </a>
            </p>
          </Field>

          {/* Links */}
          <fieldset>
            <legend className="block text-sm font-medium mb-1">
              {t('tournament_form.links_label')}
            </legend>
            <p className="text-xs mb-2" style={{ color: 'color-mix(in srgb, var(--color-text) 55%, transparent)' }}>
              {t('tournament_form.links_hint')}
            </p>
            <div className="space-y-2">
              {links.map((link, i) => (
                <div key={i} className="flex gap-2">
                  <input
                    type="url"
                    value={link.url}
                    onChange={(e) => {
                      const next = [...links];
                      next[i] = { ...next[i], url: e.target.value };
                      setLinks(next);
                    }}
                    placeholder={t('tournament_form.link_url_placeholder')}
                    aria-label={`Link ${i + 1} URL`}
                    className="flex-1 px-3 py-2 rounded border text-sm"
                    style={inputStyle}
                  />
                  <input
                    type="text"
                    value={link.label}
                    onChange={(e) => {
                      const next = [...links];
                      next[i] = { ...next[i], label: e.target.value };
                      setLinks(next);
                    }}
                    placeholder={t('tournament_form.link_label_placeholder')}
                    aria-label={`Link ${i + 1} label`}
                    className="flex-1 px-3 py-2 rounded border text-sm"
                    style={inputStyle}
                  />
                  <button
                    type="button"
                    onClick={() => setLinks(links.filter((_, j) => j !== i))}
                    className="text-xs px-3 py-2 rounded border"
                    style={{ borderColor: '#dc2626', color: '#dc2626' }}
                    aria-label={`${t('tournament_form.remove_link')} ${i + 1}`}
                  >
                    {t('tournament_form.remove_link')}
                  </button>
                </div>
              ))}
              <button
                type="button"
                onClick={() => setLinks([...links, { url: '', label: '' }])}
                className="text-xs px-3 py-2 rounded border"
                style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
              >
                + {t('tournament_form.add_link')}
              </button>
            </div>
          </fieldset>

          {/* Active Criteria */}
          <fieldset>
            <legend className="block text-sm font-medium mb-2">
              {t('tournament_form.criteria_label')}
            </legend>
            <p className="text-xs mb-1 font-medium" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {t('tournament_form.criteria_core')}
            </p>
            <div className="flex flex-wrap gap-3 mb-3">
              {CORE_CRITERIA.map((key) => (
                <label key={key} className="flex items-center gap-1 text-sm">
                  <input type="checkbox" checked disabled readOnly />
                  {t(`criteria.${key}`)}
                </label>
              ))}
            </div>
            <p className="text-xs mb-1 font-medium" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {t('tournament_form.criteria_optional')}
            </p>
            <div className="flex flex-wrap gap-3">
              {OPTIONAL_CRITERIA.map((key) => (
                <label key={key} className="flex items-center gap-1 text-sm cursor-pointer">
                  <input
                    type="checkbox"
                    checked={activeCriteria.includes(key)}
                    onChange={() => toggleOptionalCriteria(key)}
                  />
                  {t(`criteria.${key}`)}
                </label>
              ))}
            </div>
          </fieldset>

          {/* Feedback */}
          {formError && (
            <p role="alert" className="text-sm" style={{ color: '#dc2626' }}>{formError}</p>
          )}
          {formSuccess && (
            <p role="status" className="text-sm" style={{ color: '#16a34a' }}>{formSuccess}</p>
          )}

          <div className="flex gap-3">
            <button
              type="submit"
              disabled={createTournament.isPending || updateTournament.isPending}
              className="px-5 py-2 rounded font-medium text-sm"
              style={{ backgroundColor: 'var(--color-primary)', color: 'white' }}
            >
              {t('common.save')}
            </button>
            <button
              type="button"
              onClick={() => navigate('/dashboard')}
              className="px-5 py-2 rounded font-medium text-sm border"
              style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
            >
              {t('common.cancel')}
            </button>
          </div>
        </form>
      </section>


      {/* ── Result Config (edit + organizer only) ─────────────── */}
      {!isNew && isOrganizer && (
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

// ── Shared helpers ────────────────────────────────────────────────────────────

const inputStyle: React.CSSProperties = {
  borderColor: 'color-mix(in srgb, var(--color-text) 20%, transparent)',
  backgroundColor: 'var(--color-surface)',
};

function Field({
  label, htmlFor, hint, required, children,
}: {
  label: string;
  htmlFor: string;
  hint?: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1">
      <label htmlFor={htmlFor} className="block text-sm font-medium">
        {label}{required && <span aria-hidden> *</span>}
      </label>
      {hint && (
        <p className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 55%, transparent)' }}>
          {hint}
        </p>
      )}
      {children}
    </div>
  );
}

function toDatetimeLocal(iso: string): string {
  if (!iso) return '';
  return iso.slice(0, 16);
}

function toISO(local: string): string {
  if (!local) return '';
  return new Date(local).toISOString();
}
