import { create } from 'zustand';
import {
  fetchTodosRequest,
  createTodoRequest,
  updateTodoRequest,
  deleteTodoRequest
} from './api';

export interface Todo {
  id: string;
  user_id: string;
  title: string;
  description: string;
  cover: string | null;
  status: 'pending' | 'in_progress' | 'done';
  created_at: string;
  updated_at: string;
}

interface TodoState {
  todos: Todo[];
  total: number;
  page: number;
  perPage: number;
  status: string;
  keyword: string;
  sortBy: string;
  sortDir: string;
  loading: boolean;
  error: string | null;
  editingTodo: Todo | null;

  fetchTodos: () => Promise<void>;
  createTodo: (title: string, description: string, coverFile: File | null) => Promise<boolean>;
  updateTodo: (id: string, title: string, description: string, coverFile: File | null, coverPreview: string, currentStatus: 'pending' | 'in_progress' | 'done', originalUpdatedAt: string) => Promise<boolean>;
  toggleTodoStatus: (todo: Todo, nextStatus: 'pending' | 'in_progress' | 'done') => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
  setFilters: (filters: Partial<Pick<TodoState, 'status' | 'keyword' | 'sortBy' | 'sortDir'>>) => void;
  setPage: (page: number) => void;
  setEditingTodo: (todo: Todo | null) => void;
  setError: (error: string | null) => void;
}

export const useTodoStore = create<TodoState>((set, get) => ({
  todos: [],
  total: 0,
  page: 1,
  perPage: 6,
  status: '',
  keyword: '',
  sortBy: 'created_at',
  sortDir: 'desc',
  loading: false,
  error: null,
  editingTodo: null,

  fetchTodos: async () => {
    set({ loading: true, error: null });
    const { page, perPage, status, keyword, sortBy, sortDir } = get();
    try {
      const res = await fetchTodosRequest({
        page,
        per_page: perPage,
        status: status || undefined,
        q: keyword || undefined,
        sort_by: sortBy,
        sort_dir: sortDir,
      });
      set({
        todos: res.data.data.items || [],
        total: res.data.data.total || 0,
        loading: false,
      });
    } catch (err: any) {
      console.error(err);
      set({ error: 'Failed to fetch todos. Please try again.', loading: false });
    }
  },

  createTodo: async (title, description, coverFile) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    if (coverFile) {
      formData.append('cover', coverFile);
    }

    try {
      await createTodoRequest(formData);
      get().setPage(1);
      await get().fetchTodos();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to create todo' });
      return false;
    }
  },

  updateTodo: async (id, title, description, coverFile, coverPreview, currentStatus, originalUpdatedAt) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    formData.append('status', currentStatus);
    
    const maskFields = ['title', 'description', 'status'];
    if (coverFile) {
      formData.append('cover', coverFile);
      maskFields.push('cover');
    } else if (coverPreview === '' && get().editingTodo?.cover) {
      formData.append('cover', '');
      maskFields.push('cover');
    }

    try {
      const etag = `"${originalUpdatedAt}"`;
      await updateTodoRequest(id, formData, maskFields.join(','), etag);
      await get().fetchTodos();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to update todo. It might have been modified elsewhere.' });
      return false;
    }
  },

  toggleTodoStatus: async (todo, nextStatus) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('status', nextStatus);

    try {
      const etag = `"${todo.updated_at}"`;
      await updateTodoRequest(todo.id, formData, 'status', etag);
      await get().fetchTodos();
    } catch (err: any) {
      console.error(err);
      set({ error: 'Concurrency conflict. Refreshing todos...' });
      await get().fetchTodos();
    }
  },

  deleteTodo: async (id) => {
    set({ error: null });
    try {
      await deleteTodoRequest(id);
      await get().fetchTodos();
    } catch (err: any) {
      console.error(err);
      set({ error: 'Failed to delete todo' });
    }
  },

  setFilters: (filters) => {
    set((state) => ({ ...state, ...filters, page: 1 }));
  },

  setPage: (page) => set({ page }),
  setEditingTodo: (editingTodo) => set({ editingTodo }),
  setError: (error) => set({ error }),
}));
export default useTodoStore;
