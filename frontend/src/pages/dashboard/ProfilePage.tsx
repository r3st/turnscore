import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { Copy, Check, Share2 } from 'lucide-react';
import { useMe, useUpdatePreferences, useDeleteAccount } from '@/api/hooks/useUser';
import { useAuthStore } from '@/stores/authStore';

export function ProfilePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: profile, isLoading } = useMe();
  const updatePreferences = useUpdatePreferences();
  const deleteAccount = useDeleteAccount();
  const logout = useAuthStore((s) => s.logout);

  const [theme, setTheme] = useState<'scifi' | 'fantasy'>('scifi');
  const [colorMode, setColorMode] = useState<'dark' | 'light'>('dark');
  const [saved, setSaved] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [codeCopied, setCodeCopied] = useState(false);

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [confirmEmail, setConfirmEmail] = useState('');
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const confirmInputRef = useRef<HTMLInputElement>(null);

  const handleCopyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      // Fallback for non-secure contexts (HTTP dev)
      const el = document.createElement('textarea');
      el.value = code;
      el.style.position = 'fixed';
      el.style.opacity = '0';
      document.body.appendChild(el);
      el.select();
      document.execCommand('copy');
      document.body.removeChild(el);
    }
    setCodeCopied(true);
    setTimeout(() => setCodeCopied(false), 2000);
  };

  const handleShareCode = async (code: string) => {
    if (typeof navigator.share === 'function') {
      try {
        await navigator.share({ title: t('dashboard.invite_code_label'), text: code });
        return;
      } catch {
        // AbortError = user cancelled; other errors fall through to clipboard
      }
    }
    handleCopyCode(code);
  };

  useEffect(() => {
    if (profile) {
      setTheme(profile.theme as 'scifi' | 'fantasy');
      setColorMode(profile.color_mode as 'dark' | 'light');
    }
  }, [profile?.theme, profile?.color_mode]);

  const handleDeleteAccount = async () => {
    setDeleteError(null);
    try {
      await deleteAccount.mutateAsync({ confirm_email: confirmEmail });
      logout();
      navigate('/');
    } catch (err: unknown) {
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 400) {
        setDeleteError(t('profile.delete_account_email_mismatch'));
      } else {
        setDeleteError(t('profile.delete_account_error'));
      }
    }
  };

  const openDeleteDialog = () => {
    setConfirmEmail('');
    setDeleteError(null);
    setShowDeleteDialog(true);
    setTimeout(() => confirmInputRef.current?.focus(), 50);
  };

  const handleSave = async () => {
    setSaveError(null);
    try {
      await updatePreferences.mutateAsync({ theme, color_mode: colorMode });
      document.documentElement.setAttribute('data-theme', theme);
      document.documentElement.setAttribute('data-color-mode', colorMode);
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (e) {
      setSaveError(t('errors.generic'));
    }
  };

  if (isLoading) {
    return <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>{t('common.loading')}</p>;
  }

  return (
    <section aria-labelledby="profile-heading" className="max-w-lg space-y-8">
      <h1 id="profile-heading" className="font-heading text-2xl font-bold">
        {t('profile.title')}
      </h1>

      {/* User info */}
      {profile && (
        <div
          className="flex items-center gap-4 p-4 rounded"
          style={{ backgroundColor: 'var(--color-surface)' }}
        >
          {profile.avatar_url && (
            <img src={profile.avatar_url} alt="" className="h-12 w-12 rounded-full" />
          )}
          <div>
            <p className="font-medium">{profile.name}</p>
            <div className="flex items-center gap-1.5 mt-0.5">
              <span className="text-xs" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
                {t('dashboard.invite_code_label')}:
              </span>
              <code className="font-mono text-xs">{profile.helper_invite_code}</code>
              <button
                onClick={() => handleCopyCode(profile.helper_invite_code!)}
                title={codeCopied ? t('dashboard.invite_code_copied') : t('dashboard.invite_code_label')}
                className="p-0.5 rounded hover:opacity-70 transition-opacity"
                style={{ color: codeCopied ? '#16a34a' : 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
              >
                {codeCopied ? <Check size={13} /> : <Copy size={13} />}
              </button>
              <button
                onClick={() => handleShareCode(profile.helper_invite_code!)}
                title={t('dashboard.invite_code_share')}
                className="p-0.5 rounded hover:opacity-70 transition-opacity sm:hidden"
                style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
              >
                <Share2 size={13} />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Theme preference */}
      <div className="space-y-4">
        <h2 className="font-heading text-lg font-semibold">{t('profile.appearance')}</h2>

        {/* Theme selector */}
        <fieldset>
          <legend className="text-sm font-medium mb-2">{t('profile.theme_label')}</legend>
          <div className="flex gap-3">
            {(['scifi', 'fantasy'] as const).map((v) => (
              <label
                key={v}
                className="flex items-center gap-2 px-4 py-2 rounded cursor-pointer border"
                style={{
                  borderColor: theme === v ? 'var(--color-primary)' : 'color-mix(in srgb, var(--color-text) 20%, transparent)',
                  backgroundColor: theme === v ? 'color-mix(in srgb, var(--color-primary) 12%, transparent)' : 'transparent',
                }}
              >
                <input
                  type="radio"
                  name="theme"
                  value={v}
                  checked={theme === v}
                  onChange={() => setTheme(v)}
                  className="sr-only"
                />
                <span className="text-sm font-medium">{t(`profile.theme_${v}`)}</span>
              </label>
            ))}
          </div>
        </fieldset>

        {/* Color mode selector */}
        <fieldset>
          <legend className="text-sm font-medium mb-2">{t('profile.color_mode_label')}</legend>
          <div className="flex gap-3">
            {(['dark', 'light'] as const).map((v) => (
              <label
                key={v}
                className="flex items-center gap-2 px-4 py-2 rounded cursor-pointer border"
                style={{
                  borderColor: colorMode === v ? 'var(--color-primary)' : 'color-mix(in srgb, var(--color-text) 20%, transparent)',
                  backgroundColor: colorMode === v ? 'color-mix(in srgb, var(--color-primary) 12%, transparent)' : 'transparent',
                }}
              >
                <input
                  type="radio"
                  name="color_mode"
                  value={v}
                  checked={colorMode === v}
                  onChange={() => setColorMode(v)}
                  className="sr-only"
                />
                <span className="text-sm font-medium">{t(`profile.color_mode_${v}`)}</span>
              </label>
            ))}
          </div>
        </fieldset>

        <div className="flex items-center gap-4">
          <button
            onClick={handleSave}
            disabled={updatePreferences.isPending}
            className="px-6 py-2 rounded text-sm font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            {updatePreferences.isPending ? t('common.loading') : saved ? t('profile.save_success') : t('common.save')}
          </button>
          {saveError && <p className="text-sm" style={{ color: '#dc2626' }}>{saveError}</p>}
        </div>
      </div>
      {/* Danger Zone */}
      <div className="space-y-4 pt-4" style={{ borderTop: '1px solid color-mix(in srgb, #dc2626 30%, transparent)' }}>
        <h2 className="font-heading text-lg font-semibold" style={{ color: '#dc2626' }}>
          {t('profile.danger_zone')}
        </h2>
        <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}>
          {t('profile.delete_account_warning')}
        </p>
        {!showDeleteDialog ? (
          <button
            onClick={openDeleteDialog}
            className="px-4 py-2 rounded text-sm font-medium border"
            style={{ borderColor: '#dc2626', color: '#dc2626' }}
          >
            {t('profile.delete_account')}
          </button>
        ) : (
          <div className="space-y-3 p-4 rounded" style={{ backgroundColor: 'color-mix(in srgb, #dc2626 8%, transparent)', border: '1px solid color-mix(in srgb, #dc2626 30%, transparent)' }}>
            <label htmlFor="confirm-email" className="block text-sm font-medium">
              {t('profile.delete_account_confirm_label')}
            </label>
            <input
              ref={confirmInputRef}
              id="confirm-email"
              type="email"
              value={confirmEmail}
              onChange={(e) => setConfirmEmail(e.target.value)}
              placeholder={t('profile.delete_account_confirm_placeholder')}
              className="w-full px-3 py-2 rounded text-sm border"
              style={{ backgroundColor: 'var(--color-background)', color: 'var(--color-text)', borderColor: 'color-mix(in srgb, var(--color-text) 30%, transparent)' }}
              onKeyDown={(e) => e.key === 'Enter' && confirmEmail && handleDeleteAccount()}
            />
            {deleteError && (
              <p role="alert" className="text-sm" style={{ color: '#dc2626' }}>{deleteError}</p>
            )}
            <div className="flex gap-3">
              <button
                onClick={handleDeleteAccount}
                disabled={!confirmEmail || deleteAccount.isPending}
                className="px-4 py-2 rounded text-sm font-medium"
                style={{ backgroundColor: '#dc2626', color: '#fff', opacity: (!confirmEmail || deleteAccount.isPending) ? 0.5 : 1 }}
              >
                {deleteAccount.isPending ? t('common.loading') : t('profile.delete_account_confirm_button')}
              </button>
              <button
                onClick={() => setShowDeleteDialog(false)}
                className="px-4 py-2 rounded text-sm"
                style={{ color: 'color-mix(in srgb, var(--color-text) 70%, transparent)' }}
              >
                {t('common.cancel')}
              </button>
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
