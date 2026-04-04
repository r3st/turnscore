import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

type LegalPage = 'imprint' | 'privacy';
type Status = 'loading' | 'ok' | 'error';

interface LegalContent {
  content: string;
  status: Status;
}

/**
 * Fetches the Markdown file for the given legal page in the active language.
 * Falls back to English if the language-specific file is not found.
 * Files are served from /legal/{page}.{lang}.md (mounted via Docker Compose).
 */
export function useLegalContent(page: LegalPage): LegalContent {
  const { i18n } = useTranslation();
  const [content, setContent] = useState('');
  const [status, setStatus] = useState<Status>('loading');

  useEffect(() => {
    const lang = i18n.language.startsWith('de') ? 'de' : 'en';

    async function load() {
      setStatus('loading');
      const res = await fetch(`/legal/${page}.${lang}.md`);
      if (res.ok) {
        setContent(await res.text());
        setStatus('ok');
        return;
      }
      // Fallback to English
      if (lang !== 'en') {
        const fallback = await fetch(`/legal/${page}.en.md`);
        if (fallback.ok) {
          setContent(await fallback.text());
          setStatus('ok');
          return;
        }
      }
      setStatus('error');
    }

    load();
  }, [page, i18n.language]);

  return { content, status };
}
