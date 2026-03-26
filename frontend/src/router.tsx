import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';
import { DashboardLayout } from '@/components/layout/DashboardLayout';
import { ProtectedRoute } from '@/components/layout/ProtectedRoute';
import { HomePage } from '@/pages/HomePage';
import { TournamentPage } from '@/pages/TournamentPage';
import { RatePage } from '@/pages/RatePage';
import { LoginPage } from '@/pages/LoginPage';
import { OAuthCallbackPage } from '@/pages/OAuthCallbackPage';
import { RaterLoginPage } from '@/pages/RaterLoginPage';
import { ImprintPage } from '@/pages/ImprintPage';
import { PrivacyPage } from '@/pages/PrivacyPage';
import { MyTournamentsPage } from '@/pages/dashboard/MyTournamentsPage';
import { HelperTournamentsPage } from '@/pages/dashboard/HelperTournamentsPage';
import { ArchivedTournamentsPage } from '@/pages/dashboard/ArchivedTournamentsPage';
import { ProfilePage } from '@/pages/dashboard/ProfilePage';
import { TournamentEditPage } from '@/pages/dashboard/TournamentEditPage';
import { TableManagementPage } from '@/pages/dashboard/TableManagementPage';
import { RaterManagementPage } from '@/pages/dashboard/RaterManagementPage';
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
      <ProtectedRoute roles={['user']}>
        <DashboardLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard/my-tournaments" replace /> },
      { path: 'my-tournaments', element: <MyTournamentsPage /> },
      { path: 'helper', element: <HelperTournamentsPage /> },
      { path: 'archived', element: <ArchivedTournamentsPage /> },
      { path: 'profile', element: <ProfilePage /> },
      { path: 'tournaments/new', element: <TournamentEditPage /> },
      { path: 'tournaments/:slug', element: <TournamentEditPage /> },
      { path: 'tournaments/:slug/tables', element: <TableManagementPage /> },
      { path: 'tournaments/:slug/raters', element: <RaterManagementPage /> },
      {
        path: 'tournaments/:slug/results',
        element: (
          <ProtectedRoute roles={['user']}>
            <ResultsPage />
          </ProtectedRoute>
        ),
      },
    ],
  },
]);
