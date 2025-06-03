export interface User {
  id: string;
  username: string;
  roles: string[];
  created_at: string;
}

export interface UserListResponse {
  users: User[];
  next_page_token?: string;
}
