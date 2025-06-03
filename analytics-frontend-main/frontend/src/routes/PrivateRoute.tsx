import { Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import { useAuth } from '@context/AuthContext';

type PrivateRouteProps = {
  children: ReactNode;
};

function PrivateRoute({ children }: PrivateRouteProps) {
  const { isAuthenticated, initialized } = useAuth();

  if (!initialized) return <p>Загрузка...</p>;
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return children;
}

export default PrivateRoute;
