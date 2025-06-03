import React, { createContext, useContext, useEffect, useState } from 'react';
import api, { setAccessToken } from '@api/axios';
import { clearAccessToken } from '@api/axios';

type AuthContextType = {
  isAuthenticated: boolean;
  initialized: boolean;
  user: { user_id: string; roles: string[] } | null;
  logout: () => void;
  setUser: (user: AuthContextType['user']) => void;
};

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  initialized: false,
  user: null,
  logout: () => {},
  setUser: () => {},
});

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<AuthContextType['user']>(null);
    const [initialized, setInitialized] = useState(false);

    useEffect(() => {
    const init = async () => {
        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) {
        setInitialized(true);
        return;
        }

        try {
        const res = await api.post('/refresh', { refresh_token: refreshToken });
        localStorage.setItem('refresh_token', res.data.refresh_token);
        setAccessToken(res.data.access_token);

        const userRes = await api.get('/me');
        setUser({
            user_id: userRes.data.user_id,
            roles: userRes.data.roles,
        });
        } catch {
        localStorage.removeItem('refresh_token');
        setUser(null);
        } finally {
        setInitialized(true); 
        }
    };

    init();
    }, []);


    const logout = async () => {
      try {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          await api.post('/logout', { refresh_token: refreshToken });
        }
      } catch (err) {
        console.warn('Logout API failed (ignored)', err);
      } finally {
        clearAccessToken();
        localStorage.removeItem('refresh_token');
        window.location.href = '/login';
      }
    };
    


  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: !!user,
        initialized,
        user,
        logout,
        setUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
