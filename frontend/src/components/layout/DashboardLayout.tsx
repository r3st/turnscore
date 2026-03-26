import { useEffect, useState } from 'react';
import { Link, NavLink, Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useMe } from '@/api/hooks/useUser';
import { Navbar } from '@/components/layout/Navbar';

export function DashboardLayout() {
  const { t } = useTranslation();
  const { data: profile } = useMe();
  const [menuOpen, setMenuOpen] = useState(false);

  // Apply user's saved theme + color mode as soon as profile loads
  useEffect(() => {
    if (!profile) return;
    document.documentElement.setAttribute('data-theme', profile.theme);
    document.documentElement.setAttribute('data-color-mode', profile.color_mode);
  }, [profile?.theme, profile?.color_mode]);

  const navItems = [
    { to: '/dashboard/my-tournaments', label: t('dashboard.my_tournaments') },
    { to: '/dashboard/helper',          label: t('dashboard.as_helper') },
    { to: '/dashboard/archived',        label: t('dashboard.archived') },
  ];

  const navLinkClass = ({ isActive }: { isActive: boolean }) =>
    [
      'flex items-center px-3 py-2 rounded text-sm font-medium transition-colors',
      isActive
        ? 'bg-[var(--color-primary)] text-[var(--color-background)]'
        : 'hover:bg-[color-mix(in_srgb,var(--color-primary)_12%,transparent)]',
    ].join(' ');

  return (
    <div className="flex flex-col h-screen overflow-hidden" style={{ backgroundColor: 'var(--color-background)', color: 'var(--color-text)' }}>

      {/* ── Persistent top navbar ─────────────────────── */}
      <Navbar />

      {/* ── Sidebar + Content ─────────────────────────── */}
      <div className="flex flex-1 overflow-hidden">

        {/* Sidebar (desktop) */}
        <aside
          className="hidden md:flex flex-col w-56 shrink-0 border-r overflow-y-auto"
          style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
        >
          {/* Quick create */}
          <div className="px-3 pt-4 pb-2">
            <Link
              to="/dashboard/tournaments/new"
              className="flex items-center justify-center gap-1 w-full px-3 py-2 rounded text-sm font-medium"
              style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
            >
              + {t('dashboard.create_tournament')}
            </Link>
          </div>

          {/* Nav items */}
          <nav className="flex-1 flex flex-col gap-1 px-3 py-2" aria-label="Dashboard navigation">
            {navItems.map((item) => (
              <NavLink key={item.to} to={item.to} className={navLinkClass}>
                {item.label}
              </NavLink>
            ))}
          </nav>

          {/* Bottom: profile name + profile link */}
          <div className="px-3 py-4 flex flex-col gap-1 border-t" style={{ borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}>
            {profile && (
              <div className="flex items-center gap-2 px-3 py-1 text-xs truncate" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
                {profile.avatar_url && (
                  <img src={profile.avatar_url} alt="" className="h-5 w-5 rounded-full shrink-0" />
                )}
                <span className="truncate">{profile.name}</span>
              </div>
            )}
            <NavLink to="/dashboard/profile" className={navLinkClass}>
              {t('profile.title')}
            </NavLink>
          </div>
        </aside>

        {/* Mobile sidebar toggle (below navbar) */}
        <div className="md:hidden flex flex-col flex-1 min-w-0 overflow-auto">
          <div
            className="flex items-center gap-2 px-4 py-2 border-b"
            style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
          >
            <button
              onClick={() => setMenuOpen((o) => !o)}
              aria-label="Toggle dashboard menu"
              aria-expanded={menuOpen}
              className="flex items-center gap-2 px-3 py-1.5 rounded text-sm font-medium border"
              style={{ borderColor: 'var(--color-primary)', color: 'var(--color-primary)' }}
            >
              ☰ {t('dashboard.title')}
            </button>
            <Link
              to="/dashboard/tournaments/new"
              className="flex items-center gap-1 px-3 py-1.5 rounded text-sm font-medium"
              style={{ backgroundColor: 'var(--color-primary)', color: 'var(--color-background)' }}
            >
              + {t('dashboard.create_tournament')}
            </Link>
          </div>

          {menuOpen && (
            <div
              className="flex flex-col gap-1 px-4 py-3 border-b"
              style={{ backgroundColor: 'var(--color-surface)', borderColor: 'color-mix(in srgb, var(--color-primary) 20%, transparent)' }}
            >
              {navItems.map((item) => (
                <NavLink key={item.to} to={item.to} className={navLinkClass} onClick={() => setMenuOpen(false)}>
                  {item.label}
                </NavLink>
              ))}
              <NavLink to="/dashboard/profile" className={navLinkClass} onClick={() => setMenuOpen(false)}>
                {t('profile.title')}
              </NavLink>
            </div>
          )}

          <main className="flex-1 overflow-auto p-6">
            <Outlet />
          </main>
        </div>

        {/* Desktop content */}
        <main className="hidden md:block flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
