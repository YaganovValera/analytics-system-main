import { useEffect, useState } from 'react';
import api from '@api/axios';
import './MePage.css';

type UserInfo = {
  user_id: string;
  roles: string[];
  expires_at: string;
};

function MePage() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .get('/me')
      .then((res) => setUser(res.data))
      .catch(() => setError('Не удалось получить информацию о пользователе'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p>Загрузка...</p>;
  if (error) return <p className="error">{error}</p>;
  if (!user) return null;

  return (
    <div className="profile-container">
      <h2>Профиль</h2>
      <p><strong>ID:</strong> {user.user_id}</p>
      <p><strong>Роли:</strong> {user.roles.join(', ')}</p>
      <p><strong>Токен истекает:</strong> {new Date(user.expires_at).toLocaleString()}</p>
    </div>
  );
}

export default MePage;
