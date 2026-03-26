import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { TurnScoreLogo } from '@/components/ui/TurnScoreLogo';
import { useRaterLogin } from '@/api/hooks/useAuth';
import { useAuthStore } from '@/stores/authStore';

export function RaterLoginPage() {
  const { slug: slugFromUrl } = useParams<{ slug?: string }>();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const raterLogin = useRaterLogin();

  const [slug, setSlug] = useState(slugFromUrl ?? '');
  const [nickname, setNickname] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState<string | null>(null);

  const isSlugFromUrl = Boolean(slugFromUrl);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!nickname.trim() || !code.trim() || !slug.trim()) return;
    setError(null);
    try {
      const data = await raterLogin.mutateAsync({
        tournament_slug: slug.trim(),
        nickname: nickname.trim(),
        code: code.trim(),
      });
      login(data.access_token, data.refresh_token ?? null, nickname.trim());
      navigate(`/t/${slug.trim()}`);
    } catch {
      setError(t('auth.error_invalid_credentials'));
    }
  };

  const inputStyle = {
    backgroundColor: 'var(--color-background)',
    borderColor: 'color-mix(in srgb, var(--color-primary) 30%, transparent)',
    color: 'var(--color-text)',
  };

  const readonlyStyle = {
    backgroundColor: 'color-mix(in srgb, var(--color-surface) 80%, transparent)',
    borderColor: 'color-mix(in srgb, var(--color-text) 15%, transparent)',
    color: 'color-mix(in srgb, var(--color-text) 60%, transparent)',
  };

  return (
    <AppLayout>
      <main className="flex min-h-[60vh] items-center justify-center">
        <div
          className="w-full max-w-sm space-y-6 p-8 rounded"
          style={{ backgroundColor: 'var(--color-surface)' }}
        >
          <div className="flex justify-center">
            <TurnScoreLogo height={72} />
          </div>
          <h1 className="font-heading text-2xl font-bold text-center">
            {t('auth.rater_login_title')}
          </h1>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">
                {t('auth.slug_label')}
              </label>
              <input
                type="text"
                value={slug}
                onChange={isSlugFromUrl ? undefined : (e) => setSlug(e.target.value)}
                readOnly={isSlugFromUrl}
                placeholder={isSlugFromUrl ? undefined : t('auth.slug_placeholder')}
                className="w-full px-3 py-2 rounded border text-sm font-mono"
                style={isSlugFromUrl ? readonlyStyle : inputStyle}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                {t('auth.nickname_label')}
              </label>
              <input
                type="text"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
                placeholder={t('auth.nickname_placeholder')}
                autoComplete="nickname"
                className="w-full px-3 py-2 rounded border text-sm"
                style={inputStyle}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                {t('auth.code_label')}
              </label>
              <input
                type="text"
                inputMode="numeric"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={t('auth.code_placeholder')}
                maxLength={6}
                className="w-full px-3 py-2 rounded border text-sm font-mono"
                style={inputStyle}
              />
            </div>

            {error && (
              <p className="text-sm" style={{ color: '#dc2626' }}>
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={raterLogin.isPending || !nickname.trim() || !code.trim() || !slug.trim()}
              className="w-full py-2 px-4 rounded font-medium text-sm"
              style={{
                backgroundColor: 'var(--color-primary)',
                color: 'var(--color-background)',
                opacity: raterLogin.isPending || !nickname.trim() || !code.trim() || !slug.trim() ? 0.6 : 1,
              }}
            >
              {raterLogin.isPending ? t('common.loading') : t('auth.login_button')}
            </button>
          </form>
        </div>
      </main>
    </AppLayout>
  );
}
