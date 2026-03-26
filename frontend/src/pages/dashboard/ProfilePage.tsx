import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useMe, useUpdatePreferences } from '@/api/hooks/useUser';

export function ProfilePage() {
  const { t } = useTranslation();
  const { data: profile, isLoading } = useMe();
  const updatePreferences = useUpdatePreferences();

  const [theme, setTheme] = useState<'scifi' | 'fantasy'>('scifi');
  const [colorMode, setColorMode] = useState<'dark' | 'light'>('dark');
  const [saved, setSaved] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  useEffect(() => {
    if (profile) {
      setTheme(profile.theme as 'scifi' | 'fantasy');
      setColorMode(profile.color_mode as 'dark' | 'light');
    }
  }, [profile?.theme, profile?.color_mode]);

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
            <p className="text-xs mt-0.5" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
              {t('dashboard.invite_code_label')}: <code className="font-mono">{profile.helper_invite_code}</code>
            </p>
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
    </section>
  );
}
