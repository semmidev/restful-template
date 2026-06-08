import { create } from 'zustand';
import {
  fetchUsersRequest,
  createUserRequest,
  updateUserRequest,
  deleteUserRequest,
  CreateUserInput,
  UpdateUserInput,
} from './api';

export interface User {
  id: string;
  email: string;
  active_role: string;
  roles: string[];
}

interface UsersState {
  users: User[];
  total: number;
  page: number;
  perPage: number;
  keyword: string;
  sortBy: string;
  sortDir: string;
  loading: boolean;
  error: string | null;
  editingUser: User | null;

  fetchUsers: () => Promise<void>;
  createUser: (input: CreateUserInput) => Promise<boolean>;
  updateUser: (id: string, input: UpdateUserInput) => Promise<boolean>;
  deleteUser: (id: string) => Promise<boolean>;
  setFilters: (filters: Partial<Pick<UsersState, 'keyword' | 'sortBy' | 'sortDir'>>) => void;
  setPage: (page: number) => void;
  setEditingUser: (user: User | null) => void;
  setError: (error: string | null) => void;
}

export const useUsersStore = create<UsersState>((set, get) => ({
  users: [],
  total: 0,
  page: 1,
  perPage: 10,
  keyword: '',
  sortBy: 'created_at',
  sortDir: 'desc',
  loading: false,
  error: null,
  editingUser: null,

  fetchUsers: async () => {
    set({ loading: true, error: null });
    const { page, perPage, keyword, sortBy, sortDir } = get();
    try {
      const res = await fetchUsersRequest({
        page,
        per_page: perPage,
        q: keyword || undefined,
        sort_by: sortBy,
        sort_dir: sortDir,
      });
      set({
        users: res.data.data.items || [],
        total: res.data.data.total || 0,
        loading: false,
      });
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to fetch users.', loading: false });
    }
  },

  createUser: async (input) => {
    set({ error: null });
    try {
      await createUserRequest(input);
      set({ page: 1 });
      await get().fetchUsers();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to create user.' });
      return false;
    }
  },

  updateUser: async (id, input) => {
    set({ error: null });
    try {
      await updateUserRequest(id, input);
      await get().fetchUsers();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to update user.' });
      return false;
    }
  },

  deleteUser: async (id) => {
    set({ error: null });
    try {
      await deleteUserRequest(id);
      await get().fetchUsers();
      return true;
    } catch (err: any) {
      console.error(err);
      set({ error: err.response?.data?.detail || 'Failed to delete user.' });
      return false;
    }
  },

  setFilters: (filters) => {
    set((state) => ({ ...state, ...filters, page: 1 }));
  },

  setPage: (page) => set({ page }),
  setEditingUser: (editingUser) => set({ editingUser }),
  setError: (error) => set({ error }),
}));

export default useUsersStore;
