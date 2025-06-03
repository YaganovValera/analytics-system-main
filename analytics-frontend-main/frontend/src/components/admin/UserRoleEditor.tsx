// src/components/admin/UserRoleEditor.tsx
import { useState } from 'react';

interface Props {
  initialRoles: string[];
  onSave: (roles: string[]) => void;
  onCancel: () => void;
}

const ALL_ROLES = ['user', 'admin'];

function UserRoleEditor({ initialRoles, onSave, onCancel }: Props) {
  const [roles, setRoles] = useState<string[]>(initialRoles);

  const toggle = (role: string) => {
    setRoles(prev => prev.includes(role) ? prev.filter(r => r !== role) : [...prev, role]);
  };

  return (
    <form onSubmit={(e) => { e.preventDefault(); onSave(roles); }}>
      {ALL_ROLES.map(role => (
        <label key={role} style={{ marginRight: '1rem' }}>
          <input type="checkbox" checked={roles.includes(role)} onChange={() => toggle(role)} />
          {role}
        </label>
      ))}
      <div>
        <button type="submit">Сохранить</button>
        <button type="button" onClick={onCancel}>Отмена</button>
      </div>
    </form>
  );
}

export default UserRoleEditor;
