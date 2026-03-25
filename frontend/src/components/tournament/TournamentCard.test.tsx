import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { MemoryRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n';
import { TournamentCard } from './TournamentCard';
import type { TournamentSummary } from '@/api/hooks/useTournaments';

expect.extend(toHaveNoViolations);

const mockTournament: TournamentSummary = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  slug: 'essen-open-2025',
  name: 'Essen Open 2025',
  type: 'fantasy',
  status: 'active',
  event_date: '2025-10-15',
  location: 'Essen, Germany',
  table_count: 20,
};

function Wrapper({ children }: { children: React.ReactNode }) {
  return (
    <I18nextProvider i18n={i18n}>
      <MemoryRouter>{children}</MemoryRouter>
    </I18nextProvider>
  );
}

describe('TournamentCard', () => {
  it('renders tournament name as a link', () => {
    render(<TournamentCard tournament={mockTournament} />, { wrapper: Wrapper });
    const link = screen.getByRole('link', { name: /Essen Open 2025/i });
    expect(link).toHaveAttribute('href', '/t/essen-open-2025');
  });

  it('shows Fantasy badge for fantasy type', () => {
    render(<TournamentCard tournament={mockTournament} />, { wrapper: Wrapper });
    expect(screen.getByText(/Fantasy/i)).toBeInTheDocument();
  });

  it('shows Sci-Fi badge for scifi type', () => {
    const scifi: TournamentSummary = { ...mockTournament, type: 'scifi' };
    render(<TournamentCard tournament={scifi} />, { wrapper: Wrapper });
    expect(screen.getByText(/Sci-Fi/i)).toBeInTheDocument();
  });

  it('shows event date', () => {
    render(<TournamentCard tournament={mockTournament} />, { wrapper: Wrapper });
    expect(screen.getByText(/10\/15\/2025|15\.10\.2025/)).toBeInTheDocument();
  });

  it('shows location', () => {
    render(<TournamentCard tournament={mockTournament} />, { wrapper: Wrapper });
    expect(screen.getByText('Essen, Germany')).toBeInTheDocument();
  });

  it('shows table count', () => {
    render(<TournamentCard tournament={mockTournament} />, { wrapper: Wrapper });
    // Both sr-only <dt> and visible <dd> contain the text
    const matches = screen.getAllByText(/20 tables/i);
    expect(matches.length).toBeGreaterThanOrEqual(1);
  });

  it('does not render date row when event_date is null', () => {
    const noDate: TournamentSummary = { ...mockTournament, event_date: null };
    render(<TournamentCard tournament={noDate} />, { wrapper: Wrapper });
    expect(screen.queryByText(/10\/15\/2025/)).not.toBeInTheDocument();
  });

  it('has no accessibility violations', async () => {
    const { container } = render(<TournamentCard tournament={mockTournament} />, {
      wrapper: Wrapper,
    });
    expect(await axe(container)).toHaveNoViolations();
  });
});
