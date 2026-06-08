import client from '../../lib/client';

export interface FetchUsersParams {
  page: number;
  per_page: number;
  q?: string;
  sort_by: string;
  sort_dir: string;
}

export interface CreateUserInput {
  email: string;
  password?: string;
  active_role: string;
  roles: string[];
}

export interface UpdateUserInput {
  email?: string;
  password?: string;
  active_role?: string;
  roles?: string[];
}

export async function fetchUsersRequest(params: FetchUsersParams) {
  return client.get('/admin/users', { params });
}

export async function createUserRequest(input: CreateUserInput) {
  return client.post('/admin/users', input);
}

export async function updateUserRequest(id: string, input: UpdateUserInput) {
  return client.patch(`/admin/users/${id}`, input);
}

export async function deleteUserRequest(id: string) {
  return client.delete(`/admin/users/${id}`);
}
