import { useTranslation } from 'react-i18next';
import { AppLayout } from '@/components/layout/AppLayout';
import { TournamentCard } from '@/components/tournament/TournamentCard';
import { useTournaments } from '@/api/hooks/useTournaments';

function TournamentSection({
  titleKey,
  tournaments,
  isLoading,
}: {
  titleKey: string;
  tournaments: ReturnType<typeof useTournaments>['data'] extends { upcoming: infer U } ? U : never;
  isLoading: boolean;
}) {
  const { t } = useTranslation();

  return (
    <section aria-labelledby={`${titleKey}-heading`} className="space-y-4">
      <h2
        id={`${titleKey}-heading`}
        className="font-heading text-xl font-semibold"
        style={{ color: 'var(--color-text)' }}
      >
        {t(titleKey)}
      </h2>

      {isLoading ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3" aria-busy="true">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="h-32 rounded animate-pulse"
              style={{ backgroundColor: 'var(--color-surface)' }}
              aria-hidden
            />
          ))}
        </div>
      ) : !tournaments || tournaments.length === 0 ? (
        <p
          className="text-sm"
          style={{ color: 'color-mix(in srgb, var(--color-text) 55%, transparent)' }}
        >
          {t('home.no_tournaments')}
        </p>
      ) : (
        <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 list-none p-0">
          {tournaments.map((tournament) => (
            <li key={tournament.id}>
              <TournamentCard tournament={tournament} />
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export function HomePage() {
  const { t } = useTranslation();
  const { data, isLoading, isError } = useTournaments();

  return (
    <AppLayout>
      <div className="space-y-10">
        <header>
          <h1 className="font-heading text-3xl font-bold" style={{ color: 'var(--color-text)' }}>
            {t('app.name')}
          </h1>
          <p
            className="mt-1 text-sm"
            style={{ color: 'color-mix(in srgb, var(--color-text) 60%, transparent)' }}
          >
            {t('app.tagline')}
          </p>
        </header>

        {isError && (
          <div role="alert" className="rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {t('errors.generic')}
          </div>
        )}

        <TournamentSection
          titleKey="home.upcoming_tournaments"
          tournaments={data?.upcoming ?? []}
          isLoading={isLoading}
        />

        <TournamentSection
          titleKey="home.past_tournaments"
          tournaments={data?.past ?? []}
          isLoading={isLoading}
        />
      </div>
    </AppLayout>
  );
}
