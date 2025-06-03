import React, { useState } from 'react';
import axios from 'axios';
import './AuthForm.css';
import { Link } from 'react-router-dom';

type AuthFormProps = {
  mode: 'login' | 'register';
  onSubmit: (username: string, password: string) => Promise<void>;
};

function validateUsername(username: string): string | null {
  const trimmed = username.trim();
  if (trimmed.length < 3 || trimmed.length > 64) {
    return 'Логин должен содержать от 3 до 64 символов.';
  }
  if (!/^[a-zA-Z0-9_]+$/.test(trimmed)) {
    return 'Логин может содержать только латиницу, цифры и подчёркивания.';
  }
  if (/^_/.test(trimmed) || /_$/.test(trimmed)) {
    return 'Логин не может начинаться или заканчиваться подчёркиванием.';
  }
  return null;
}

function validatePassword(password: string): string | null {
  if (password.length < 8 || password.length > 128) {
    return 'Пароль должен содержать от 8 до 128 символов.';
  }
  if (/\s/.test(password)) {
    return 'Пароль не должен содержать пробелов.';
  }
  if (!/[A-Za-z]/.test(password)) {
    return 'Пароль должен содержать хотя бы одну букву.';
  }
  if (!/\d/.test(password)) {
    return 'Пароль должен содержать хотя бы одну цифру.';
  }
  return null;
}

function AuthForm({ mode, onSubmit }: AuthFormProps) {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const uErr = validateUsername(username);
    const pErr = validatePassword(password);
    if (uErr) return setError(uErr);
    if (pErr) return setError(pErr);

    try {
      await onSubmit(username.trim(), password);
    } catch (err: any) {
      if (axios.isAxiosError(err)) {
        const status = err.response?.status;
        if (mode === 'login') {
          if (status === 401) setError('Неверный логин или пароль.');
          else setError(`Ошибка входа: ${status}`);
        } else {
          if (status === 409) setError('Пользователь уже существует.');
          else setError(`Ошибка регистрации: ${status}`);
        }
      } else {
        setError('Неизвестная ошибка');
      }
    }
  };

  return (
    <form onSubmit={handleSubmit} className="auth-form">
      <h2>{mode === 'login' ? 'Вход в систему' : 'Регистрация'}</h2>
      {error && <p className="error">{error}</p>}
      <div className="form-group">
        <label htmlFor="username">Логин:</label>
        <input
          id="username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          required
        />
      </div>
      <div className="form-group">
        <label htmlFor="password">Пароль:</label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
      </div>
      <button type="submit">{mode === 'login' ? 'Войти' : 'Зарегистрироваться'}</button>

      <p className="auth-switch">
        {mode === 'login' ? (
          <>
            Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
          </>
        ) : (
          <>
            Уже есть аккаунт? <Link to="/login">Войти</Link>
          </>
        )}
      </p>
    
    </form>
  );
}

export default AuthForm;
