import { useIsAdmin } from '@utils/roles';
import UserRegistrationForm from '@components/admin/UserRegistrationForm';
import UserTable from '@components/admin/UserTable';
import './AdminDashboardPage.css';
import { useState, useEffect } from 'react';
import { listUsers } from '@api/admin';
import type { User } from '../../types/admin';

function AdminDashboardPage() {
  const isAdmin = useIsAdmin();
  const [users, setUsers] = useState<User[]>([]);
  const [nextPageToken, setNextPageToken] = useState<string | undefined>(undefined);
  const [prevTokens, setPrevTokens] = useState<string[]>(['']);
  const [query] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [pageIndex, setPageIndex] = useState(0);

  const fetchPage = async (token: string | undefined) => {
    try {
      setError(null);
      const res = await listUsers(50, token, query);
      setUsers(res.users);
      setNextPageToken(res.next_page_token);
    } catch {
      setError('Не удалось загрузить пользователей');
    }
  };

  useEffect(() => {
    fetchPage(undefined);
    setPrevTokens(['']);
    setPageIndex(0);
  }, [query]);

  const handleNext = async () => {
    if (!nextPageToken) return;
    const newPrev = [...prevTokens, nextPageToken];
    setPrevTokens(newPrev);
    setPageIndex(newPrev.length - 1);
    await fetchPage(nextPageToken);
  };

  const handlePrev = async () => {
    if (pageIndex === 0) return;
    const prevToken = prevTokens[pageIndex - 1];
    setPageIndex(pageIndex - 1);
    await fetchPage(prevToken);
  };

  const refreshUsers = async () => {
    await fetchPage(prevTokens[pageIndex]);
  };

  if (!isAdmin) return <p>У вас нет доступа к админке</p>;

  return (
    <div className="admin-dashboard">
      <h2>Администрирование пользователей</h2>
      <UserRegistrationForm onSuccess={refreshUsers} />
      <UserTable users={users} onRolesUpdated={refreshUsers} />
      <div className="pagination-controls">
        <button onClick={handlePrev} disabled={pageIndex === 0}>← Назад</button>
        <button onClick={handleNext} disabled={!nextPageToken}>Вперёд →</button>
      </div>

      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  );
}

export default AdminDashboardPage;
