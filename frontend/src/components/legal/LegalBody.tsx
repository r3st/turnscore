import Markdown from 'react-markdown';
import { TFunction } from 'i18next';

interface Props {
  status: 'loading' | 'ok' | 'error';
  content: string;
  t: TFunction;
}

/**
 * Renders the body of a legal page: loading skeleton, error message, or Markdown content.
 */
export function LegalBody({ status, content, t }: Props) {
  if (status === 'loading') {
    return (
      <div className="space-y-3 animate-pulse">
        {[100, 90, 95, 70].map((w, i) => (
          <div
            key={i}
            className="h-4 rounded"
            style={{
              width: `${w}%`,
              backgroundColor: 'color-mix(in srgb, var(--color-text) 12%, transparent)',
            }}
          />
        ))}
      </div>
    );
  }

  if (status === 'error') {
    return (
      <p className="text-sm" style={{ color: 'color-mix(in srgb, var(--color-text) 55%, transparent)' }}>
        {t('legal.not_available')}
      </p>
    );
  }

  return (
    <div
      className="prose prose-sm max-w-none"
      style={{ color: 'var(--color-text)' }}
    >
      <Markdown>{content}</Markdown>
    </div>
  );
}
