// src/components/admin/UserTable.tsx
import './UserTable.css';

import { useState, useMemo } from 'react';
import type { User } from '../../types/admin';
import { updateUserRoles } from '@api/admin';
import UserRoleEditor from './UserRoleEditor';

interface Props {
  users: User[];
  onRolesUpdated: () => void;
}

function UserTable({ users, onRolesUpdated }: Props) {
  const [editingUserId, setEditingUserId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState('');
  const [sortKey, setSortKey] = useState<'username' | 'roles' | 'created_at'>('username');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');

  const handleRoleUpdate = async (userId: string, roles: string[]) => {
    try {
      await updateUserRoles(userId, roles);
      setEditingUserId(null);
      onRolesUpdated();
    } catch {
      setError('Не удалось обновить роли');
    }
  };

  const filteredAndSorted = useMemo(() => {
    const filtered = users.filter(
      (user) =>
        user.username.toLowerCase().startsWith(filter.toLowerCase())
    );

    return filtered.sort((a, b) => {
      let aValue: string | number = '';
      let bValue: string | number = '';

      if (sortKey === 'username') {
        aValue = a.username.toLowerCase();
        bValue = b.username.toLowerCase();
      } else if (sortKey === 'roles') {
        aValue = a.roles.join(',').toLowerCase();
        bValue = b.roles.join(',').toLowerCase();
      } else if (sortKey === 'created_at') {
        aValue = new Date(a.created_at).getTime();
        bValue = new Date(b.created_at).getTime();
      }

      if (aValue < bValue) return sortOrder === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });
  }, [users, filter, sortKey, sortOrder]);

  const toggleSort = (key: 'username' | 'roles' | 'created_at') => {
    if (sortKey === key) {
      setSortOrder((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortOrder('asc');
    }
  };

  return (
    <div className="user-table">
      <h3>Список пользователей</h3>
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <input
        type="text"
        placeholder="Фильтр по имени..."
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        style={{ marginBottom: '0.5rem', width: '100%' }}
      />

      <table>
        <thead>
          <tr>
            <th onClick={() => toggleSort('username')}>
              Имя пользователя {sortKey === 'username' ? (sortOrder === 'asc' ? '▲' : '▼') : ''}
            </th>
            <th onClick={() => toggleSort('roles')}>
              Роли {sortKey === 'roles' ? (sortOrder === 'asc' ? '▲' : '▼') : ''}
            </th>
            <th onClick={() => toggleSort('created_at')}>
              Создан {sortKey === 'created_at' ? (sortOrder === 'asc' ? '▲' : '▼') : ''}
            </th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody>
          {filteredAndSorted.map((user) => (
            <tr key={user.id}>
              <td>{user.username}</td>
              <td>{user.roles?.join(', ') ?? '-'}</td>
              <td>{new Date(user.created_at).toLocaleString()}</td>
              <td>
                {editingUserId === user.id ? (
                  <UserRoleEditor
                    initialRoles={user.roles}
                    onSave={(newRoles) => handleRoleUpdate(user.id, newRoles)}
                    onCancel={() => setEditingUserId(null)}
                  />
                ) : (
                  <button onClick={() => setEditingUserId(user.id)}>Изменить роли</button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default UserTable;
