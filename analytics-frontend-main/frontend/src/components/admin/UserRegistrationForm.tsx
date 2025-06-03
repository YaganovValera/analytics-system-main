import { useState } from 'react';
import { registerUser } from '@api/admin';
import axios from 'axios';
import './UserRegistrationForm.css';

interface Props {
  onSuccess: () => void;
}

const ALL_ROLES = ['user', 'admin'];

function validateUsername(username: string): string | null {
  const trimmed = username.trim();
  if (trimmed.length < 3 || trimmed.length > 64) return 'Логин от 3 до 64 символов.';
  if (!/^[a-zA-Z0-9_]+$/.test(trimmed)) return 'Только латиница, цифры, подчёркивания.';
  if (/^_/.test(trimmed) || /_$/.test(trimmed)) return 'Нельзя начинать или заканчивать подчёркиванием.';
  return null;
}

function validatePassword(password: string): string | null {
  if (password.length < 8 || password.length > 128) return 'Пароль от 8 до 128 символов.';
  if (/\s/.test(password)) return 'Без пробелов.';
  if (!/[A-Za-z]/.test(password)) return 'Хотя бы одна буква.';
  if (!/\d/.test(password)) return 'Хотя бы одна цифра.';
  return null;
}

function UserRegistrationForm({ onSuccess }: Props) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [roles, setRoles] = useState<string[]>(['user']);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleToggleRole = (role: string) => {
    setRoles((prev) =>
      prev.includes(role) ? prev.filter((r) => r !== role) : [...prev, role]
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const uErr = validateUsername(username);
    const pErr = validatePassword(password);
    if (uErr) return setError(uErr);
    if (pErr) return setError(pErr);
    if (roles.length === 0) return setError('Назначьте хотя бы одну роль');

    setError(null);
    setLoading(true);
    try {
      await registerUser(username.trim(), password, roles);
      onSuccess();
      setUsername('');
      setPassword('');
      setRoles(['user']);
    } catch (err: any) {
      if (axios.isAxiosError(err)) {
        const status = err.response?.status;
        if (status === 409) setError('Пользователь уже существует.');
        else setError(`Ошибка регистрации: ${status}`);
      } else {
        setError('Неизвестная ошибка');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="user-registration-form">
      <h3>Регистрация нового пользователя</h3>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <div className="form-group">
        <label>Логин:</label>
        <input value={username} onChange={(e) => setUsername(e.target.value)} required />
      </div>

      <div className="form-group">
        <label>Пароль:</label>
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required />
      </div>

      <div className="form-group">
        <label>Роли:</label>
        <div className="roles">
          {ALL_ROLES.map((role) => (
            <label key={role}>
              <input
                type="checkbox"
                checked={roles.includes(role)}
                onChange={() => handleToggleRole(role)}
              />
              {role}
            </label>
          ))}
        </div>
      </div>

      <button type="submit" disabled={loading}>
        {loading ? 'Создание...' : 'Создать'}
      </button>
    </form>
  );
}

export default UserRegistrationForm;
