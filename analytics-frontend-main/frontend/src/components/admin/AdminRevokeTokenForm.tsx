// src/components/admin/AdminRevokeTokenForm.tsx
import { useState } from 'react';
import { adminRevokeToken } from '@api/admin';

function AdminRevokeTokenForm() {
  const [token, setToken] = useState('');
  const [status, setStatus] = useState<'idle' | 'success' | 'error'>('idle');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await adminRevokeToken(token.trim());
      setToken('');
      setStatus('success');
    } catch {
      setStatus('error');
    }
  };

  return (
    <form onSubmit={submit} className="admin-revoke-form">
      <h4>Отозвать refresh токен</h4>
      <input value={token} onChange={(e) => setToken(e.target.value)} placeholder="refresh_token" />
      <button type="submit" disabled={!token}>Отозвать</button>
      {status === 'success' && <p style={{ color: 'green' }}>Успешно</p>}
      {status === 'error' && <p style={{ color: 'red' }}>Ошибка</p>}
    </form>
  );
}

export default AdminRevokeTokenForm;
