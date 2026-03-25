import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

export function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="border-t py-4 text-center text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}>
      <nav aria-label="Footer navigation" className="flex justify-center gap-4">
        <Link to="/imprint" className="hover:underline">{t('nav.imprint')}</Link>
        <Link to="/privacy" className="hover:underline">{t('nav.privacy')}</Link>
      </nav>
    </footer>
  );
}
