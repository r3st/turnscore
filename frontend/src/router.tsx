import { createBrowserRouter } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';
import { ProtectedRoute } from '@/components/layout/ProtectedRoute';
import { HomePage } from '@/pages/HomePage';
import { TournamentPage } from '@/pages/TournamentPage';
import { RatePage } from '@/pages/RatePage';
import { LoginPage } from '@/pages/LoginPage';
import { OAuthCallbackPage } from '@/pages/OAuthCallbackPage';
import { RaterLoginPage } from '@/pages/RaterLoginPage';
import { ImprintPage } from '@/pages/ImprintPage';
import { PrivacyPage } from '@/pages/PrivacyPage';
import { DashboardPage } from '@/pages/dashboard/DashboardPage';
import { TournamentEditPage } from '@/pages/dashboard/TournamentEditPage';
import { TableManagementPage } from '@/pages/dashboard/TableManagementPage';
import { ResultsPage } from '@/pages/dashboard/ResultsPage';

export const router = createBrowserRouter([
  { path: '/', element: <HomePage /> },
  { path: '/t/:slug', element: <TournamentPage /> },
  { path: '/rate/:slug/:tableNum', element: <RatePage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/auth/callback', element: <OAuthCallbackPage /> },
  { path: '/rate-login/:slug?', element: <RaterLoginPage /> },
  { path: '/imprint', element: <ImprintPage /> },
  { path: '/privacy', element: <PrivacyPage /> },
  {
    path: '/dashboard',
    element: (
      <ProtectedRoute roles={['organizer', 'helper']}>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'tournaments/new', element: <TournamentEditPage /> },
      { path: 'tournaments/:slug', element: <TournamentEditPage /> },
      { path: 'tournaments/:slug/tables', element: <TableManagementPage /> },
      {
        path: 'tournaments/:slug/results',
        element: (
          <ProtectedRoute roles={['organizer']}>
            <ResultsPage />
          </ProtectedRoute>
        ),
      },
    ],
  },
]);
