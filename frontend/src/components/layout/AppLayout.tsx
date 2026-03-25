import { type ReactNode } from 'react';
import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';
import { Footer } from './Footer';

interface AppLayoutProps {
  children?: ReactNode;
}

/**
 * AppLayout can be used in two ways:
 * 1. As a React Router layout route (children via <Outlet />)
 * 2. As a wrapper component (children prop) for standalone pages
 */
export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main className="flex-1 container mx-auto px-4 py-8">
        {children ?? <Outlet />}
      </main>
      <Footer />
    </div>
  );
}
