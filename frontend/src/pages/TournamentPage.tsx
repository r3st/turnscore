import { useParams } from 'react-router-dom';
import { AppLayout } from '@/components/layout/AppLayout';

export function TournamentPage() {
  const { slug } = useParams<{ slug: string }>();

  return (
    <AppLayout>
      <main className="container mx-auto px-4 py-8">
        <p className="text-sm text-muted-foreground">Tournament: {slug}</p>
      </main>
    </AppLayout>
  );
}
