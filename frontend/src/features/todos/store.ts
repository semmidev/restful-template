import { create } from 'zustand';
import {
  fetchTodosRequest,
  createTodoRequest,
  updateTodoRequest,
  deleteTodoRequest,
  restoreTodoRequest,
  fetchTodoStatsRequest,
  fetchTodoByIdRequest
} from './api';

export interface DailyStat {
  date: string;
  created: number;
  completed: number;
}

export interface TodoStats {
  total: number;
  pending: number;
  in_progress: number;
  completed: number;
  completion_rate: number;
  daily_stats: DailyStat[];
}

export interface Todo {
  id: string;
  user_id: string;
  title: string;
  description: string;
  cover: string | null;
  status: 'pending' | 'in_progress' | 'done';
  importance: boolean;
  urgency: boolean;
  due_at: string | null;
  deleted_at: string | null;
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
  archived: boolean;
  loading: boolean;
  error: string | null;
  editingTodo: Todo | null;
  stats: TodoStats | null;
  statsLoading: boolean;
  statsError: string | null;
  currentTodo: Todo | null;
  currentTodoLoading: boolean;
  currentTodoError: string | null;

  fetchTodos: () => Promise<void>;
  fetchTodoById: (id: string, silent?: boolean) => Promise<Todo | null>;
  createTodo: (title: string, description: string, coverFile: File | null, importance?: boolean, urgency?: boolean, dueAt?: string | null) => Promise<boolean>;
  updateTodo: (id: string, title: string, description: string, coverFile: File | null, coverPreview: string, currentStatus: 'pending' | 'in_progress' | 'done', originalUpdatedAt: string, importance?: boolean, urgency?: boolean, dueAt?: string | null) => Promise<boolean>;
  updateTodoPriority: (todo: Todo, importance: boolean, urgency: boolean) => Promise<void>;
  toggleTodoStatus: (todo: Todo, nextStatus: 'pending' | 'in_progress' | 'done') => Promise<void>;
  deleteTodo: (id: string) => Promise<void>;
  restoreTodo: (id: string) => Promise<void>;
  setFilters: (filters: Partial<Pick<TodoState, 'status' | 'keyword' | 'sortBy' | 'sortDir' | 'archived'>>) => void;
  setPage: (page: number) => void;
  setEditingTodo: (todo: Todo | null) => void;
  setError: (error: string | null) => void;
  fetchStats: () => Promise<void>;
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
  archived: false,
  loading: false,
  error: null,
  editingTodo: null,
  currentTodo: null,
  currentTodoLoading: false,
  currentTodoError: null,

  fetchTodos: async () => {
    set({ loading: true, error: null });
    const { page, perPage, status, keyword, sortBy, sortDir, archived } = get();
    try {
      const res = await fetchTodosRequest({
        page,
        per_page: perPage,
        status: status || undefined,
        q: keyword || undefined,
        sort_by: sortBy,
        sort_dir: sortDir,
        archived: archived || undefined,
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

  fetchTodoById: async (id, silent = false) => {
    if (!silent) {
      set({ currentTodoLoading: true });
    }
    set({ currentTodoError: null });
    try {
      const res = await fetchTodoByIdRequest(id);
      const todo = res.data.data;
      set({ currentTodo: todo, currentTodoLoading: false });
      return todo;
    } catch (err: any) {
      console.error(err);
      set({ currentTodoError: 'Failed to fetch todo details.', currentTodoLoading: false });
      return null;
    }
  },

  createTodo: async (title, description, coverFile, importance = true, urgency = false, dueAt = null) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    formData.append('importance', String(importance));
    formData.append('urgency', String(urgency));
    if (dueAt) {
      formData.append('due_at', dueAt);
    }
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

  updateTodo: async (id, title, description, coverFile, coverPreview, currentStatus, originalUpdatedAt, importance, urgency, dueAt = undefined) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('title', title);
    formData.append('description', description);
    formData.append('status', currentStatus);

    const maskFields = ['title', 'description', 'status'];

    if (importance !== undefined) {
      formData.append('importance', String(importance));
      maskFields.push('importance');
    }
    if (urgency !== undefined) {
      formData.append('urgency', String(urgency));
      maskFields.push('urgency');
    }

    if (dueAt !== undefined) {
      formData.append('due_at', dueAt || '');
      maskFields.push('due_at');
    }

    if (coverFile) {
      formData.append('cover', coverFile);
      maskFields.push('cover');
    } else if (coverPreview === '' && get().editingTodo?.cover) {
      formData.append('cover', '');
      maskFields.push('cover');
    }

    try {
      const etag = `"${originalUpdatedAt}"`;
      const res = await updateTodoRequest(id, formData, maskFields.join(','), etag);
      const updatedTodo = res.data.data;

      const currentTodo = get().currentTodo;
      if (currentTodo && currentTodo.id === id) {
        set({ currentTodo: updatedTodo });
      }

      await get().fetchTodos();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to update todo. It might have been modified elsewhere.' });
      return false;
    }
  },

  updateTodoPriority: async (todo, importance, urgency) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('importance', String(importance));
    formData.append('urgency', String(urgency));

    try {
      const etag = `"${todo.updated_at}"`;
      const res = await updateTodoRequest(todo.id, formData, 'importance,urgency', etag);
      const updatedTodo = res.data.data;

      const currentTodo = get().currentTodo;
      if (currentTodo && currentTodo.id === todo.id) {
        set({ currentTodo: updatedTodo });
      }

      // Update local state with the latest values from server
      set((state) => ({
        todos: state.todos.map((t) =>
          t.id === todo.id ? updatedTodo : t
        ),
      }));
    } catch (err: any) {
      console.error(err);
      set({ error: 'Concurrency conflict. Refreshing...' });
      await get().fetchTodos();
    }
  },

  toggleTodoStatus: async (todo, nextStatus) => {
    set({ error: null });
    const formData = new FormData();
    formData.append('status', nextStatus);

    try {
      const etag = `"${todo.updated_at}"`;
      const res = await updateTodoRequest(todo.id, formData, 'status', etag);
      const updatedTodo = res.data.data;

      const currentTodo = get().currentTodo;
      if (currentTodo && currentTodo.id === todo.id) {
        set({ currentTodo: updatedTodo });
      }

      // Update local list state immediately with the updatedTodo returned by server
      set((state) => ({
        todos: state.todos.map((t) =>
          t.id === todo.id ? updatedTodo : t
        ),
      }));
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

  restoreTodo: async (id) => {
    set({ error: null });
    try {
      await restoreTodoRequest(id);
      await get().fetchTodos();
    } catch (err: any) {
      console.error(err);
      set({ error: 'Failed to restore todo' });
    }
  },

  setFilters: (filters) => {
    set((state) => ({ ...state, ...filters, page: 1 }));
  },

  stats: null,
  statsLoading: false,
  statsError: null,

  fetchStats: async () => {
    set({ statsLoading: true, statsError: null });
    try {
      const res = await fetchTodoStatsRequest();
      set({
        stats: res.data.data || null,
        statsLoading: false,
      });
    } catch (err: any) {
      console.error(err);
      set({ statsError: 'Failed to fetch statistics.', statsLoading: false });
    }
  },

  setPage: (page) => set({ page }),
  setEditingTodo: (editingTodo) => set({ editingTodo }),
  setError: (error) => set({ error }),
}));
export default useTodoStore;
