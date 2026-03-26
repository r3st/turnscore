import { useEffect, useState } from 'react';
import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuthStore } from '@/stores/authStore';
import { useMe } from '@/api/hooks/useUser';
import { TurnScoreLogo } from '@/components/ui/TurnScoreLogo';

export function DashboardLayout() {
  const { t } = useTranslation();
  const { logout } = useAuthStore();
  const navigate = useNavigate();
  const { data: profile } = useMe();
  const [menuOpen, setMenuOpen] = useState(false);

  // Apply user's saved theme + color mode preferences
  useEffect(() => {
    if (!profile) return;
    const root = document.documentElement;
    root.setAttribute('data-theme', profile.theme);
    root.setAttribute('data-color-mode', profile.color_mode);
  }, [profile?.theme, profile?.color_mode]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const navItems = [
    { to: '/dashboard/my-tournaments', label: t('dashboard.my_tournaments') },
    { to: '/dashboard/helper',          label: t('dashboard.as_helper') },
    { to: '/dashboard/archived',        label: t('dashboard.archived') },
  ];

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    [
      'flex items-center gap-2 px-3 py-2 rounded text-sm font-medium transition-colors',
      isActive
        ? 'text-[var(--color-background)] bg-[var(--color-primary)]'
        : 'hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)]',
    ].join(' ');

  return (
    <div className="flex h-screen overflow-hidden" style={{ backgroundColor: 'var(--color-background)', color: 'var(--color-text)' }}>

      {/* ── Sidebar (desktop) ─────────────────────────── */}
      <aside
        className="hidden md:flex flex-col w-56 shrink-0 border-r"
        style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
      >
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 px-4 py-4 font-heading font-bold text-lg" style={{ color: 'var(--color-primary)' }}>
          <TurnScoreLogo height={28} />
          <span>{t('app.name')}</span>
        </Link>

        {/* Quick create */}
        <div className="px-3 mb-2">
          <Link
            to="/dashboard/tournaments/new"
            className="flex items-center justify-center gap-1 w-full px-3 py-2 rounded text-sm font-medium"
            style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
          >
            + {t('dashboard.create_tournament')}
          </Link>
        </div>

        {/* Nav items */}
        <nav className="flex-1 flex flex-col gap-1 px-3" aria-label="Dashboard navigation">
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} className={navLinkClass}>
              {item.label}
            </NavLink>
          ))}
        </nav>

        {/* Bottom: profile + logout */}
        <div className="px-3 py-4 flex flex-col gap-1 border-t" style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}>
          {profile && (
            <div className="flex items-center gap-2 px-3 py-2 text-sm truncate">
              {profile.avatar_url && (
                <img src={profile.avatar_url} alt="" className="h-6 w-6 rounded-full shrink-0" />
              )}
              <span className="truncate font-medium">{profile.name}</span>
            </div>
          )}
          <NavLink to="/dashboard/profile" className={navLinkClass}>
            {t('profile.title')}
          </NavLink>
          <button
            onClick={handleLogout}
            className="flex items-center gap-2 px-3 py-2 rounded text-sm font-medium text-left hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)] w-full"
            aria-label={t('nav.logout')}
          >
            {t('nav.logout')}
          </button>
        </div>
      </aside>

      {/* ── Mobile top bar ────────────────────────────── */}
      <div className="flex flex-col flex-1 min-w-0 overflow-auto">
        <header
          className="md:hidden flex items-center justify-between px-4 py-3 border-b shrink-0"
          style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
        >
          <Link to="/" className="font-heading font-bold" style={{ color: 'var(--color-primary)' }}>
            {t('app.name')}
          </Link>
          <button
            onClick={() => setMenuOpen((o) => !o)}
            aria-label="Toggle menu"
            className="px-3 py-2 text-sm rounded"
            style={{ border: '1px solid var(--color-primary)', color: 'var(--color-primary)' }}
          >
            ☰
          </button>
        </header>

        {/* Mobile dropdown menu */}
        {menuOpen && (
          <div
            className="md:hidden flex flex-col gap-1 px-4 py-3 border-b"
            style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
          >
            <Link
              to="/dashboard/tournaments/new"
              onClick={() => setMenuOpen(false)}
              className="px-3 py-2 rounded text-sm font-medium text-center"
              style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
            >
              + {t('dashboard.create_tournament')}
            </Link>
            {navItems.map((item) => (
              <NavLink key={item.to} to={item.to} className={navLinkClass} onClick={() => setMenuOpen(false)}>
                {item.label}
              </NavLink>
            ))}
            <NavLink to="/dashboard/profile" className={navLinkClass} onClick={() => setMenuOpen(false)}>
              {t('profile.title')}
            </NavLink>
            <button onClick={handleLogout} className="px-3 py-2 rounded text-sm font-medium text-left hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)]">
              {t('nav.logout')}
            </button>
          </div>
        )}

        {/* Page content */}
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
