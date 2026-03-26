import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { TurnScoreLogo } from '@/components/ui/TurnScoreLogo';

export function Navbar() {
  const { t, i18n } = useTranslation();
  const { isAuthenticated, logout, role } = useAuthStore();
  const navigate = useNavigate();

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const linkClass = 'flex items-center h-9 px-3 rounded text-sm font-medium hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)] transition-colors';
  const logoutClass = 'flex items-center h-9 px-3 rounded text-sm font-medium border transition-colors hover:opacity-80';

  return (
    <header className="border-b shrink-0" style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}>
      <nav
        className="container mx-auto flex items-center justify-between px-4 h-14"
        aria-label="Main navigation"
      >
        <Link
          to="/"
          className="flex items-center gap-2 font-heading text-xl font-bold"
          style={{ color: 'var(--color-primary)' }}
          aria-label={t('app.name')}
        >
          <TurnScoreLogo height={32} />
          <span>{t('app.name')}</span>
        </Link>

        <div className="flex items-center gap-1">
          <Link to="/" className={linkClass}>{t('nav.home')}</Link>

          {isAuthenticated() && role === 'user' && (
            <Link to="/dashboard" className={linkClass}>
              {t('nav.dashboard')}
            </Link>
          )}

          {isAuthenticated() ? (
            <button
              onClick={handleLogout}
              className={logoutClass}
              style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
              aria-label={t('nav.logout')}
            >
              {t('nav.logout')}
            </button>
          ) : (
            <Link to="/login" className={linkClass}>{t('nav.login')}</Link>
          )}

          <div className="flex gap-1 ml-2" role="group" aria-label="Select language">
            <button
              onClick={() => changeLanguage('en')}
              aria-pressed={i18n.language === 'en'}
              className={`flex items-center justify-center h-9 w-9 rounded text-base ${i18n.language === 'en' ? 'opacity-100' : 'opacity-40'}`}
              aria-label="English"
            >
              🇬🇧
            </button>
            <button
              onClick={() => changeLanguage('de')}
              aria-pressed={i18n.language === 'de'}
              className={`flex items-center justify-center h-9 w-9 rounded text-base ${i18n.language === 'de' ? 'opacity-100' : 'opacity-40'}`}
              aria-label="Deutsch"
            >
              🇩🇪
            </button>
          </div>
        </div>
      </nav>
    </header>
  );
}
