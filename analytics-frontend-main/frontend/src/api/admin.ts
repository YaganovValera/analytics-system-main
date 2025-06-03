import api from './axios';
import type { User, UserListResponse } from '../types/admin';

export const listUsers = async (
  pageSize: number = 50,
  pageToken?: string,
  query?: string
): Promise<UserListResponse> => {
  const res = await api.get('/admin/users', {
    params: {
      page_size: pageSize,
      page_token: pageToken,
      query,
    },
  });
  return res.data;
};

export const getUser = async (id: string): Promise<User> => {
  const res = await api.get(`/admin/users/${id}`);
  return res.data;
};

export const updateUserRoles = async (id: string, roles: string[]): Promise<void> => {
  await api.put(`/admin/users/${id}/roles`, { roles });
};

export const registerUser = async (
  username: string,
  password: string,
  roles: string[]
): Promise<User> => {
  const res = await api.post('/register', { username, password, roles });
  return res.data;
};

export const adminRevokeToken = async (token: string): Promise<void> => {
  await api.post('/admin/revoke', { token });
};
