'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/lib/auth';

interface ProtectedRouteProps {
  children: React.ReactNode;
  requireAdmin?: boolean;
}

export default function ProtectedRoute({ children, requireAdmin = false }: ProtectedRouteProps) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    console.log('ProtectedRoute effect', { user, loading, requireAdmin });
    if (!loading && !user) {
      console.log('Redirecting to login');
      router.push('/login');
    }

    if (!loading && user && requireAdmin && user.role !== 'admin') {
      console.log('Redirecting to dashboard (non-admin)');
      router.push('/dashboard');
    }
  }, [user, loading, router, requireAdmin]);

  console.log('ProtectedRoute render', { loading, user });
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
          <p className="mt-4 text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  if (requireAdmin && user.role !== 'admin') {
    return null;
  }

  return <>{children}</>;
}