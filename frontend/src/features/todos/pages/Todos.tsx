import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  LogOut,
  Plus,
  Search,
  Trash2,
  Edit2,
  CheckCircle,
  Clock,
  Play,
  ChevronLeft,
  ChevronRight,
  Upload,
  X,
  FileText
} from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import useAuthStore from '../../auth/store';
import useTodoStore, { Todo } from '../store';
import { todoSchema } from '../../../lib/schemas';
import TodoSkeleton from '../components/TodoSkeleton';

export default function Todos() {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);

  const {
    todos,
    total,
    page,
    perPage,
    status,
    keyword,
    sortBy,
    sortDir,
    loading,
    error,
    editingTodo,
    fetchTodos,
    createTodo,
    updateTodo,
    toggleTodoStatus,
    deleteTodo,
    setFilters,
    setPage,
    setEditingTodo,
    setError,
  } = useTodoStore();

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);

  // Form state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [coverFile, setCoverFile] = useState<File | null>(null);
  const [coverPreview, setCoverPreview] = useState('');

  // Form validation errors
  const [validationErrors, setValidationErrors] = useState<{ title?: string; description?: string }>({});

  // Local debounced keyword state
  const [searchKeyword, setSearchKeyword] = useState(keyword);

  useEffect(() => {
    fetchTodos();
  }, [page, status, sortBy, sortDir, keyword, fetchTodos]);

  // Sync initial/updated keyword from store to local input if needed
  useEffect(() => {
    setSearchKeyword(keyword);
  }, [keyword]);

  // Debounced update of the store's keyword
  useEffect(() => {
    const delayDebounceFn = setTimeout(() => {
      if (searchKeyword !== keyword) {
        setFilters({ keyword: searchKeyword });
      }
    }, 400);

    return () => clearTimeout(delayDebounceFn);
  }, [searchKeyword, keyword, setFilters]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (file.size > 5 * 1024 * 1024) {
        setError('Image must be under 5MB');
        return;
      }
      setCoverFile(file);
      setCoverPreview(URL.createObjectURL(file));
    }
  };

  const handleCreateTodo = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setValidationErrors({});
    setError(null);

    const validation = todoSchema.safeParse({ title, description });
    if (!validation.success) {
      const fieldErrors: { title?: string; description?: string } = {};
      validation.error.issues.forEach((issue) => {
        const path = issue.path[0] as 'title' | 'description';
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const success = await createTodo(title, description || '', coverFile);
    if (success) {
      setIsCreateOpen(false);
      resetForm();
    }
  };

  const handleEditClick = (todo: Todo) => {
    setEditingTodo(todo);
    setTitle(todo.title);
    setDescription(todo.description || '');
    setCoverPreview(todo.cover || '');
    setIsEditOpen(true);
  };

  const handleUpdateTodo = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingTodo) return;
    setValidationErrors({});
    setError(null);

    const validation = todoSchema.safeParse({ title, description });
    if (!validation.success) {
      const fieldErrors: { title?: string; description?: string } = {};
      validation.error.issues.forEach((issue) => {
        const path = issue.path[0] as 'title' | 'description';
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const success = await updateTodo(
      editingTodo.id,
      title,
      description || '',
      coverFile,
      coverPreview,
      editingTodo.status,
      editingTodo.updated_at
    );
    if (success) {
      setIsEditOpen(false);
      resetForm();
    }
  };

  const handleToggleStatus = async (todo: Todo, nextStatus: 'pending' | 'in_progress' | 'done') => {
    await toggleTodoStatus(todo, nextStatus);
  };

  const handleDeleteTodo = async (id: string) => {
    if (!window.confirm('Delete this todo?')) return;
    await deleteTodo(id);
  };

  const resetForm = () => {
    setTitle('');
    setDescription('');
    setCoverFile(null);
    setCoverPreview('');
    setEditingTodo(null);
    setValidationErrors({});
  };

  const totalPages = Math.max(1, Math.ceil(total / perPage));

  return (
    <div className="min-h-screen bg-brutal-bg p-6 pb-24 text-black">
      {/* Navigation Header */}
      <header className="max-w-6xl mx-auto mb-8 card-brutal bg-white flex flex-col sm:flex-row justify-between items-center gap-4 py-4 px-6">
        <div className="flex items-center gap-3">
          <span className="bg-brutal-yellow border-2 border-black p-2 rounded shadow-brutal-sm font-black rotate-[-3deg]">
            TA
          </span>
          <h1 className="text-3xl font-black tracking-tight">
            TODO<span className="text-brutal-blue">APP</span>
          </h1>
        </div>
        <div className="flex items-center gap-4">
          <Button
            onClick={() => setIsCreateOpen(true)}
            className="btn-brutal bg-brutal-yellow text-sm font-black flex items-center gap-2"
          >
            <Plus size={18} /> New Task
          </Button>
          <Button
            onClick={handleLogout}
            className="btn-brutal-secondary text-sm font-black flex items-center gap-2"
          >
            <LogOut size={16} /> Logout
          </Button>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="max-w-6xl mx-auto mb-6 bg-brutal-pink border-3 border-black p-4 rounded-xl font-bold shadow-brutal flex justify-between items-center">
          <span>⚠️ {error}</span>
          <Button onClick={() => setError(null)} className="font-black hover:scale-110 p-0 h-auto bg-transparent text-black border-none shadow-none">
            <X size={20} />
          </Button>
        </div>
      )}

      {/* Main Board Content */}
      <main className="max-w-6xl mx-auto">
        {/* Filters and Controls */}
        <section className="card-brutal bg-white mb-8 grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
          {/* Search */}
          <div className="relative md:col-span-2">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500 z-10" size={18} />
            <Input
              type="text"
              placeholder="Search tasks..."
              className="input-brutal w-full pl-10"
              value={searchKeyword}
              onChange={(e) => setSearchKeyword(e.target.value)}
            />
          </div>

          {/* Sort Column */}
          <select
            className="input-brutal w-full"
            value={sortBy}
            onChange={(e) => setFilters({ sortBy: e.target.value })}
          >
            <option value="created_at">Created Time</option>
            <option value="updated_at">Updated Time</option>
            <option value="title">Title Alphabetical</option>
            <option value="status">Status Priority</option>
          </select>

          {/* Sort Direction */}
          <Button
            onClick={() => setFilters({ sortDir: sortDir === 'asc' ? 'desc' : 'asc' })}
            className="btn-brutal-secondary w-full text-center font-bold"
          >
            Sort: {sortDir === 'asc' ? '⬆️ Ascending' : '⬇️ Descending'}
          </Button>
        </section>

        {/* Status Tabs (using Shadcn Tabs) */}
        <section className="mb-8">
          <Tabs value={status} onValueChange={(val) => { setFilters({ status: val }); }}>
            <TabsList className="bg-transparent border-none gap-2 flex flex-wrap h-auto p-0">
              {[
                { label: 'All Tasks', value: '' },
                { label: 'Pending', value: 'pending' },
                { label: 'In Progress', value: 'in_progress' },
                { label: 'Completed', value: 'done' },
              ].map((tab) => (
                <TabsTrigger
                  key={tab.value}
                  value={tab.value}
                  className={`px-4 py-2 border-3 border-black font-extrabold rounded-lg shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] hover:shadow-brutal transition-all ${
                    status === tab.value
                      ? 'bg-brutal-yellow text-black data-active:bg-brutal-yellow data-active:text-black'
                      : 'bg-white text-black data-active:bg-white data-active:text-black'
                  }`}
                >
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        </section>

        {/* Todo Cards Grid */}
        {loading && todos.length === 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: 6 }).map((_, index) => (
              <TodoSkeleton key={index} />
            ))}
          </div>
        ) : todos.length === 0 ? (
          <Card className="card-brutal bg-white text-center py-16">
            <FileText size={48} className="mx-auto mb-4 text-neutral-400" />
            <h3 className="text-2xl font-black mb-2">No Tasks Found</h3>
            <p className="text-neutral-600 font-bold mb-6">Create a new task to get started!</p>
            <Button
              onClick={() => setIsCreateOpen(true)}
              className="btn-brutal text-md"
            >
              <Plus size={18} /> New Task
            </Button>
          </Card>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {todos.map((todo) => (
              <Card
                key={todo.id}
                className="card-brutal bg-white flex flex-col justify-between hover:shadow-brutal-hover transition-all overflow-hidden"
              >
                <div>
                  {/* Task Cover Image */}
                  {todo.cover && (
                    <img
                      src={todo.cover}
                      alt="Todo Cover"
                      className="w-[calc(100%+3rem)] h-40 object-cover border-b-3 border-black -mt-6 -mx-6 mb-4"
                    />
                  )}

                  {/* Task Info */}
                  <div className="flex justify-between items-start gap-2 mb-3">
                    <Badge
                      className={`badge-brutal text-xs uppercase py-1 px-2.5 h-auto ${
                        todo.status === 'done'
                          ? 'bg-brutal-green text-white'
                          : todo.status === 'in_progress'
                          ? 'bg-brutal-blue text-white'
                          : 'bg-brutal-yellow text-black'
                      }`}
                    >
                      {todo.status === 'in_progress' ? 'in progress' : todo.status}
                    </Badge>
                    <div className="flex gap-1.5">
                      <Button
                        onClick={() => handleEditClick(todo)}
                        className="p-1.5 h-auto w-auto border-2 border-black rounded bg-white hover:bg-neutral-100 shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] text-black"
                      >
                        <Edit2 size={14} />
                      </Button>
                      <Button
                        onClick={() => handleDeleteTodo(todo.id)}
                        className="p-1.5 h-auto w-auto border-2 border-black rounded bg-brutal-pink hover:bg-red-400 shadow-brutal-sm hover:translate-x-[-1px] hover:translate-y-[-1px] text-black"
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  </div>

                  <h3 className="text-xl font-black mb-2 text-black line-clamp-1">
                    {todo.title}
                  </h3>
                  <p className="text-neutral-700 font-medium text-sm mb-6 line-clamp-3">
                    {todo.description || 'No description provided.'}
                  </p>
                </div>

                {/* Footer Controls */}
                <div className="border-t-2 border-black pt-4 mt-auto">
                  <div className="flex flex-wrap justify-between items-center gap-2">
                    <div className="flex gap-1.5">
                      {todo.status !== 'pending' && (
                        <Button
                          onClick={() => handleToggleStatus(todo, 'pending')}
                          className="btn-brutal-secondary py-1 px-2.5 h-auto text-xs font-black flex items-center gap-1 shadow-brutal-sm text-black"
                        >
                          <Clock size={12} /> Reopen
                        </Button>
                      )}
                      {todo.status !== 'in_progress' && todo.status !== 'done' && (
                        <Button
                          onClick={() => handleToggleStatus(todo, 'in_progress')}
                          className="btn-brutal bg-brutal-blue text-white py-1 px-2.5 h-auto text-xs font-black flex items-center gap-1 shadow-brutal-sm hover:bg-blue-600"
                        >
                          <Play size={12} /> Start
                        </Button>
                      )}
                      {todo.status !== 'done' && (
                        <Button
                          onClick={() => handleToggleStatus(todo, 'done')}
                          className="btn-brutal bg-brutal-green text-white py-1 px-2.5 h-auto text-xs font-black flex items-center gap-1 shadow-brutal-sm hover:bg-green-600"
                        >
                          <CheckCircle size={12} /> Finish
                        </Button>
                      )}
                    </div>
                    <span className="text-[10px] text-neutral-500 font-bold uppercase">
                      updated {new Date(todo.updated_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}

        {/* Pagination Controls */}
        {totalPages > 1 && (
          <section className="flex justify-center items-center gap-4 mt-12">
            <Button
              onClick={() => setPage(Math.max(1, page - 1))}
              disabled={page === 1}
              className="btn-brutal-secondary flex items-center justify-center p-2.5 h-10 w-10 disabled:opacity-50 text-black"
            >
              <ChevronLeft size={20} />
            </Button>
            <span className="font-black text-lg bg-white px-4 py-2 border-3 border-black rounded-lg shadow-brutal-sm">
              Page {page} of {totalPages}
            </span>
            <Button
              onClick={() => setPage(Math.min(totalPages, page + 1))}
              disabled={page === totalPages}
              className="btn-brutal-secondary flex items-center justify-center p-2.5 h-10 w-10 disabled:opacity-50 text-black"
            >
              <ChevronRight size={20} />
            </Button>
          </section>
        )}
      </main>

      {/* Creation Dialog (using Shadcn Dialog) */}
      <Dialog open={isCreateOpen} onOpenChange={(open) => { if (!open) { setIsCreateOpen(false); resetForm(); } }}>
        <DialogContent className="card-brutal bg-white w-full max-w-lg shadow-[8px_8px_0_0_#000] p-6 max-h-[90vh] overflow-y-auto" showCloseButton={false}>
          <DialogHeader className="border-b-3 border-black pb-3 mb-6 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-2xl font-black flex items-center gap-2">
              <Plus size={24} /> Create Task
            </DialogTitle>
            <Button
              onClick={() => {
                setIsCreateOpen(false);
                resetForm();
              }}
              className="p-1 h-auto w-auto border-2 border-black rounded hover:bg-brutal-pink transition shadow-brutal-sm bg-white text-black hover:translate-x-0 hover:translate-y-0"
            >
              <X size={18} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleCreateTodo} className="space-y-5">
            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Title</label>
              <Input
                type="text"
                required
                placeholder="What needs to be done?"
                className="input-brutal w-full"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Description</label>
              <textarea
                rows={3}
                placeholder="Task details..."
                className="input-brutal w-full"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Cover Image</label>
              <div className="flex items-center gap-4">
                <label className="btn-brutal-secondary py-2 px-4 text-sm font-black cursor-pointer flex items-center gap-2">
                  <Upload size={16} /> Choose File
                  <input
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleFileChange}
                  />
                </label>
                {coverPreview && (
                  <Button
                    type="button"
                    onClick={() => {
                      setCoverFile(null);
                      setCoverPreview('');
                    }}
                    className="btn-brutal-danger py-2 px-3 h-auto text-sm font-black flex items-center gap-1 hover:translate-x-0 hover:translate-y-0"
                  >
                    Remove
                  </Button>
                )}
              </div>
              {coverPreview && (
                <img
                  src={coverPreview}
                  alt="Cover preview"
                  className="w-full h-32 object-cover border-3 border-black rounded-lg mt-4 shadow-brutal-sm"
                />
              )}
            </div>

            <div className="flex justify-end gap-3 pt-4 border-t-2 border-black">
              <Button
                type="button"
                onClick={() => {
                  setIsCreateOpen(false);
                  resetForm();
                }}
                className="btn-brutal-secondary h-auto"
              >
                Cancel
              </Button>
              <Button type="submit" className="btn-brutal bg-brutal-yellow h-auto">
                Create Task
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Editing Dialog (using Shadcn Dialog) */}
      <Dialog open={isEditOpen} onOpenChange={(open) => { if (!open) { setIsEditOpen(false); resetForm(); } }}>
        <DialogContent className="card-brutal bg-white w-full max-w-lg shadow-[8px_8px_0_0_#000] p-6 max-h-[90vh] overflow-y-auto" showCloseButton={false}>
          <DialogHeader className="border-b-3 border-black pb-3 mb-6 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-2xl font-black flex items-center gap-2">
              <Edit2 size={22} /> Edit Task
            </DialogTitle>
            <Button
              onClick={() => {
                setIsEditOpen(false);
                resetForm();
              }}
              className="p-1 h-auto w-auto border-2 border-black rounded hover:bg-brutal-pink transition shadow-brutal-sm bg-white text-black hover:translate-x-0 hover:translate-y-0"
            >
              <X size={18} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleUpdateTodo} className="space-y-5">
            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Title</label>
              <Input
                type="text"
                required
                placeholder="Task title"
                className="input-brutal w-full"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Description</label>
              <textarea
                rows={3}
                placeholder="Task description"
                className="input-brutal w-full"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            <div className="flex flex-col">
              <label className="text-sm font-black mb-1.5 uppercase">Cover Image</label>
              <div className="flex items-center gap-4">
                <label className="btn-brutal-secondary py-2 px-4 text-sm font-black cursor-pointer flex items-center gap-2">
                  <Upload size={16} /> Change File
                  <input
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleFileChange}
                  />
                </label>
                {coverPreview && (
                  <Button
                    type="button"
                    onClick={() => {
                      setCoverFile(null);
                      setCoverPreview('');
                    }}
                    className="btn-brutal-danger py-2 px-3 h-auto text-sm font-black flex items-center gap-1 hover:translate-x-0 hover:translate-y-0"
                  >
                    Remove
                  </Button>
                )}
              </div>
              {coverPreview && (
                <img
                  src={coverPreview}
                  alt="Cover preview"
                  className="w-full h-32 object-cover border-3 border-black rounded-lg mt-4 shadow-brutal-sm"
                />
              )}
            </div>

            <div className="flex justify-end gap-3 pt-4 border-t-2 border-black">
              <Button
                type="button"
                onClick={() => {
                  setIsEditOpen(false);
                  resetForm();
                }}
                className="btn-brutal-secondary h-auto"
              >
                Cancel
              </Button>
              <Button type="submit" className="btn-brutal bg-brutal-yellow h-auto">
                Save Changes
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
