import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { TurnScoreLogo } from '@/components/ui/TurnScoreLogo';

export function Navbar() {
  const { t, i18n } = useTranslation();
  const { isAuthenticated, logout, role } = useAuthStore();
  const navigate = useNavigate();
  const [menuOpen, setMenuOpen] = useState(false);

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
    setMenuOpen(false);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
    setMenuOpen(false);
  };

  const linkClass = 'flex items-center h-9 px-3 rounded text-sm font-medium hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)] transition-colors';
  const logoutClass = 'flex items-center h-9 px-3 rounded text-sm font-medium border transition-colors hover:opacity-80';
  const mobileLinkClass = 'flex items-center h-11 px-4 rounded text-sm font-medium hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)] transition-colors';

  return (
    <header className="border-b shrink-0 relative" style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}>
      <nav
        className="container mx-auto flex items-center justify-between px-4 h-14"
        aria-label="Main navigation"
      >
        {/* Logo */}
        <Link
          to="/"
          className="flex items-center gap-2 font-heading text-xl font-bold shrink-0"
          style={{ color: 'var(--color-primary)' }}
          aria-label={t('app.name')}
        >
          <TurnScoreLogo height={32} />
          <span>{t('app.name')}</span>
        </Link>

        {/* Desktop nav */}
        <div className="hidden sm:flex items-center gap-1">
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

        {/* Mobile hamburger button */}
        <button
          className="sm:hidden flex items-center justify-center h-9 w-9 rounded"
          onClick={() => setMenuOpen((o) => !o)}
          aria-expanded={menuOpen}
          aria-label="Toggle menu"
          style={{ color: 'var(--color-text)' }}
        >
          {menuOpen ? (
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="4" y1="4" x2="16" y2="16" />
              <line x1="16" y1="4" x2="4" y2="16" />
            </svg>
          ) : (
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="3" y1="6" x2="17" y2="6" />
              <line x1="3" y1="10" x2="17" y2="10" />
              <line x1="3" y1="14" x2="17" y2="14" />
            </svg>
          )}
        </button>
      </nav>

      {/* Mobile dropdown menu */}
      {menuOpen && (
        <div
          className="sm:hidden absolute top-14 left-0 right-0 z-50 border-b px-2 py-2 flex flex-col gap-1"
          style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
        >
          <Link to="/" className={mobileLinkClass} onClick={() => setMenuOpen(false)}>
            {t('nav.home')}
          </Link>

          {isAuthenticated() && role === 'user' && (
            <Link to="/dashboard" className={mobileLinkClass} onClick={() => setMenuOpen(false)}>
              {t('nav.dashboard')}
            </Link>
          )}

          {isAuthenticated() ? (
            <button
              onClick={handleLogout}
              className="flex items-center h-11 px-4 rounded text-sm font-medium border transition-colors hover:opacity-80"
              style={{ borderColor: 'var(--color-accent)', color: 'var(--color-accent)' }}
            >
              {t('nav.logout')}
            </button>
          ) : (
            <Link to="/login" className={mobileLinkClass} onClick={() => setMenuOpen(false)}>
              {t('nav.login')}
            </Link>
          )}

          <div className="flex gap-1 px-1 pt-1 border-t mt-1" style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 15%, transparent)' }} role="group" aria-label="Select language">
            <button
              onClick={() => changeLanguage('en')}
              aria-pressed={i18n.language === 'en'}
              className={`flex items-center justify-center h-10 w-10 rounded text-base ${i18n.language === 'en' ? 'opacity-100' : 'opacity-40'}`}
              aria-label="English"
            >
              🇬🇧
            </button>
            <button
              onClick={() => changeLanguage('de')}
              aria-pressed={i18n.language === 'de'}
              className={`flex items-center justify-center h-10 w-10 rounded text-base ${i18n.language === 'de' ? 'opacity-100' : 'opacity-40'}`}
              aria-label="Deutsch"
            >
              🇩🇪
            </button>
          </div>
        </div>
      )}
    </header>
  );
}
