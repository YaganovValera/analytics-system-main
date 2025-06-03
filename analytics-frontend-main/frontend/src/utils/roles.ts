// src/utils/roles.ts
import { useAuth } from '@context/AuthContext';

export function useIsAdmin(): boolean {
  const { user } = useAuth();
  return user?.roles.includes('admin') ?? false;
}

export function hasRole(user: { roles: string[] }, role: string): boolean {
  return user.roles.includes(role);
}
