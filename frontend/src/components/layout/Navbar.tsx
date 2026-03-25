import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';

export function Navbar() {
  const { t, i18n } = useTranslation();
  const { isAuthenticated, logout, role } = useAuthStore();

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
  };

  return (
    <header className="border-b" style={{ backgroundColor: 'var(--color-surface)' }}>
      <nav
        className="container mx-auto flex items-center justify-between px-4 py-3"
        aria-label="Main navigation"
      >
        <Link to="/" className="font-heading text-xl font-bold" style={{ color: 'var(--color-primary)' }}>
          {t('app.name')}
        </Link>

        <div className="flex items-center gap-4">
          <Link to="/" className="text-sm hover:underline">{t('nav.home')}</Link>

          {isAuthenticated() && (role === 'organizer' || role === 'helper') && (
            <Link to="/dashboard" className="text-sm hover:underline">
              {t('nav.dashboard')}
            </Link>
          )}

          {isAuthenticated() ? (
            <button
              onClick={logout}
              className="text-sm hover:underline"
              aria-label={t('nav.logout')}
            >
              {t('nav.logout')}
            </button>
          ) : (
            <Link to="/login" className="text-sm hover:underline">{t('nav.login')}</Link>
          )}

          <div className="flex gap-1" role="group" aria-label="Select language">
            <button
              onClick={() => changeLanguage('en')}
              aria-pressed={i18n.language === 'en'}
              className={i18n.language === 'en' ? 'opacity-100' : 'opacity-50'}
              aria-label="English"
            >
              🇬🇧
            </button>
            <button
              onClick={() => changeLanguage('de')}
              aria-pressed={i18n.language === 'de'}
              className={i18n.language === 'de' ? 'opacity-100' : 'opacity-50'}
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
