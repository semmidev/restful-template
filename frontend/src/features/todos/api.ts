import client from '../../lib/client';

export interface FetchTodosParams {
  page: number;
  per_page: number;
  status?: string;
  q?: string;
  sort_by: string;
  sort_dir: string;
  archived?: boolean;
}

export async function fetchTodosRequest(params: FetchTodosParams) {
  return client.get('/todos', { params });
}

export async function createTodoRequest(formData: FormData) {
  return client.post('/todos', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

export async function updateTodoRequest(
  id: string,
  formData: FormData,
  updateMask: string,
  etag: string
) {
  return client.patch(`/todos/${id}`, formData, {
    params: {
      update_mask: updateMask,
    },
    headers: {
      'Content-Type': 'multipart/form-data',
      'If-Match': etag,
    },
  });
}

export async function deleteTodoRequest(id: string) {
  return client.delete(`/todos/${id}`);
}

export async function restoreTodoRequest(id: string) {
  return client.post(`/todos/${id}/restore`);
}

export async function fetchTodoStatsRequest() {
  return client.get('/todos/stats');
}

export async function fetchTodoByIdRequest(id: string) {
  return client.get(`/todos/${id}`);
}

