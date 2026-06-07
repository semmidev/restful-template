import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useTheme } from 'next-themes';
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  Calendar,
  Clock,
  Play,
  CheckCircle,
  FileText,
  ChevronLeft,
  ChevronRight,
  Activity,
  Upload,
  X,
  Sun,
  Moon
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Separator } from '@/components/ui/separator';
import {
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import { TooltipProvider } from '@/components/ui/tooltip';
import { AppSidebar } from '@/components/app-sidebar';
import useAuthStore from '../../auth/store';
import useTodoStore, { Todo } from '../store';
import { todoSchema } from '../../../lib/schemas';
import TodoSkeleton from '../components/TodoSkeleton';

export default function Todos() {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.logout);
  const { theme, setTheme } = useTheme();

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
  const [todoToDelete, setTodoToDelete] = useState<Todo | null>(null);

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

  const handleDeleteClick = (todo: Todo) => {
    setTodoToDelete(todo);
  };

  const handleDeleteConfirm = async () => {
    if (todoToDelete) {
      await deleteTodo(todoToDelete.id);
      setTodoToDelete(null);
    }
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

  // Compute metrics from todos in local context/store
  const pendingCount = todos.filter(t => t.status === 'pending').length;
  const inProgressCount = todos.filter(t => t.status === 'in_progress').length;
  const completedCount = todos.filter(t => t.status === 'done').length;
  const completionRate = total > 0 ? Math.round((completedCount / total) * 100) : 0;

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          {/* Site Header scaffolded after dashboard-01 */}
          <header className="flex h-16 shrink-0 items-center gap-2 border-b bg-background/95 backdrop-blur px-4 lg:px-6">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
            
            {/* Breadcrumb section */}
            <div className="flex-1 flex items-center gap-1 text-sm text-muted-foreground">
              <span className="font-semibold text-slate-800 dark:text-slate-200">Workspace</span>
              <span>/</span>
              <span className="font-medium">Tasks</span>
            </div>

            {/* Quick Header Search & Actions */}
            <div className="flex items-center gap-2">
              <div className="relative w-40 sm:w-60 md:w-80">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground size-4" />
                <Input
                  type="text"
                  placeholder="Search tasks..."
                  className="w-full pl-8 h-8 rounded-md bg-muted/40 border border-input focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 text-xs"
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="h-8 w-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-all"
                title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {theme === 'dark' ? <Sun size={15} /> : <Moon size={15} />}
              </Button>
              <Button
                onClick={() => setIsCreateOpen(true)}
                className="h-8 px-3 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/95 transition-all shadow-sm rounded-md flex items-center gap-1"
              >
                <Plus size={14} /> New Task
              </Button>
            </div>
          </header>

          {/* Main content grid */}
          <main className="flex flex-col gap-6 p-6 lg:p-8">
            {/* Error Banner */}
            {error && (
              <div className="bg-destructive/10 border border-destructive/20 text-destructive text-sm font-semibold p-4 rounded-xl flex justify-between items-center shadow-sm">
                <span>⚠️ {error}</span>
                <Button variant="ghost" onClick={() => setError(null)} className="h-8 w-8 p-0 text-destructive hover:bg-destructive/10">
                  <X size={16} />
                </Button>
              </div>
            )}

            {/* Dashboard Metrics (dashboard-01 style blocks) */}
            <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <Card className="border border-border bg-card shadow-sm rounded-xl">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Total Tasks</CardTitle>
                  <FileText size={18} className="text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-extrabold">{total}</div>
                  <p className="text-[10px] text-muted-foreground mt-1">Overall database records</p>
                </CardContent>
              </Card>

              <Card className="border border-border bg-card shadow-sm rounded-xl">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Pending</CardTitle>
                  <Clock size={18} className="text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-extrabold text-amber-600">{pendingCount}</div>
                  <p className="text-[10px] text-muted-foreground mt-1">Awaiting workspace start</p>
                </CardContent>
              </Card>

              <Card className="border border-border bg-card shadow-sm rounded-xl">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-bold uppercase tracking-wider text-muted-foreground">In Progress</CardTitle>
                  <Activity size={18} className="text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-extrabold text-primary">{inProgressCount}</div>
                  <p className="text-[10px] text-muted-foreground mt-1">Currently being processed</p>
                </CardContent>
              </Card>

              <Card className="border border-border bg-card shadow-sm rounded-xl">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Completion Rate</CardTitle>
                  <CheckCircle size={18} className="text-muted-foreground" />
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="text-2xl font-extrabold text-emerald-600">{completionRate}%</div>
                  <div className="w-full bg-slate-100 dark:bg-slate-800 h-1.5 rounded-full overflow-hidden">
                    <div 
                      className="bg-emerald-500 h-full rounded-full transition-all duration-500" 
                      style={{ width: `${completionRate}%` }}
                    ></div>
                  </div>
                </CardContent>
              </Card>
            </section>

            {/* Filters and Controls */}
            <section className="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center bg-card border border-border p-4 rounded-xl shadow-sm">
              {/* Tabs for fast status switching on mobile/tablet */}
              <Tabs value={status} onValueChange={(val) => { setFilters({ status: val }); }}>
                <TabsList className="bg-muted p-1 rounded-lg gap-1">
                  {[
                    { label: 'All Tasks', value: '' },
                    { label: 'Pending', value: 'pending' },
                    { label: 'In Progress', value: 'in_progress' },
                    { label: 'Completed', value: 'done' },
                  ].map((tab) => (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className="px-4 py-1.5 text-xs font-semibold rounded-md transition-all data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm"
                    >
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>

              <div className="flex flex-wrap items-center gap-3">
                {/* Sort Column */}
                <select
                  className="h-9 px-3 rounded-md border border-input bg-background text-sm font-medium focus:outline-none focus:ring-1 focus:ring-primary"
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
                  variant="outline"
                  size="sm"
                  onClick={() => setFilters({ sortDir: sortDir === 'asc' ? 'desc' : 'asc' })}
                  className="h-9 font-semibold text-xs flex items-center gap-1"
                >
                  Sort: {sortDir === 'asc' ? 'Ascending ⬆️' : 'Descending ⬇️'}
                </Button>
              </div>
            </section>

            {/* Todo Cards Grid */}
            {loading && todos.length === 0 ? (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {Array.from({ length: 6 }).map((_, index) => (
                  <TodoSkeleton key={index} />
                ))}
              </div>
            ) : todos.length === 0 ? (
              <Card className="border border-border border-dashed bg-card text-center py-16 rounded-xl shadow-sm">
                <CardContent className="flex flex-col items-center justify-center p-0">
                  <div className="w-12 h-12 bg-slate-100 rounded-full flex items-center justify-center text-slate-400 mb-4">
                    <FileText size={24} />
                  </div>
                  <h3 className="text-xl font-bold mb-1">No Tasks Found</h3>
                  <p className="text-muted-foreground text-sm max-w-xs mb-6">Create a new task and specify details to start tracking milestones.</p>
                  <Button
                    onClick={() => setIsCreateOpen(true)}
                    className="h-9 font-semibold text-xs"
                  >
                    <Plus size={14} className="mr-1" /> New Task
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {todos.map((todo) => (
                  <Card
                    key={todo.id}
                    className="border border-border bg-card shadow-sm rounded-xl overflow-hidden flex flex-col justify-between hover:shadow-md transition-all duration-200 p-0 gap-0"
                  >
                    <div>
                      {/* Task Cover Image */}
                      {todo.cover ? (
                        <div className="w-full h-40 overflow-hidden border-b border-border bg-slate-50 flex items-center justify-center">
                          <img
                            src={todo.cover}
                            alt="Todo Cover"
                            className="w-full h-full object-cover"
                          />
                        </div>
                      ) : (
                        <div className="w-full h-2 bg-primary"></div>
                      )}

                      {/* Task Header & Badge */}
                      <CardHeader className="pb-2 space-y-1.5 pt-5 px-6">
                        <div className="flex justify-between items-start gap-2">
                          <Badge
                            variant="secondary"
                            className={`text-[10px] uppercase font-bold py-0.5 px-2 rounded-full border shadow-none ${
                              todo.status === 'done'
                                ? 'bg-emerald-50 border-emerald-200 text-emerald-700 dark:bg-emerald-950/20'
                                : todo.status === 'in_progress'
                                ? 'bg-indigo-50 border-indigo-200 text-indigo-700 dark:bg-indigo-950/20'
                                : 'bg-amber-50 border-amber-200 text-amber-700 dark:bg-amber-950/20'
                            }`}
                          >
                            {todo.status === 'in_progress' ? 'in progress' : todo.status}
                          </Badge>
                          <div className="flex gap-1.5">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleEditClick(todo)}
                              className="h-7 w-7 rounded-md border text-muted-foreground hover:text-foreground"
                            >
                              <Edit2 size={12} />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDeleteClick(todo)}
                              className="h-7 w-7 rounded-md border text-muted-foreground hover:text-destructive hover:bg-destructive/10"
                            >
                              <Trash2 size={12} />
                            </Button>
                          </div>
                        </div>
                        <CardTitle className="text-lg font-bold text-slate-900 line-clamp-1">
                          {todo.title}
                        </CardTitle>
                      </CardHeader>

                      <CardContent className="pb-6 px-6">
                        <p className="text-slate-600 text-sm leading-relaxed line-clamp-3">
                          {todo.description || 'No description provided.'}
                        </p>
                      </CardContent>
                    </div>

                    {/* Footer Controls */}
                    <CardFooter className="border-t border-border pt-4 pb-4 px-6 flex items-center justify-between gap-4 mt-auto">
                      <div className="flex gap-1.5">
                        {todo.status !== 'pending' && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleToggleStatus(todo, 'pending')}
                            className="h-7 text-[10px] font-bold px-2 rounded-md flex items-center gap-1 text-slate-700"
                          >
                            <Clock size={10} /> Reopen
                          </Button>
                        )}
                        {todo.status !== 'in_progress' && todo.status !== 'done' && (
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => handleToggleStatus(todo, 'in_progress')}
                            className="h-7 text-[10px] font-bold px-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/95 flex items-center gap-1 shadow-none"
                          >
                            <Play size={10} /> Start
                          </Button>
                        )}
                        {todo.status !== 'done' && (
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => handleToggleStatus(todo, 'done')}
                            className="h-7 text-[10px] font-bold px-2 rounded-md bg-emerald-600 text-white hover:bg-emerald-700 flex items-center gap-1 shadow-none"
                          >
                            <CheckCircle size={10} /> Finish
                          </Button>
                        )}
                      </div>
                      <div className="flex items-center gap-1 text-[10px] text-muted-foreground font-semibold">
                        <Calendar size={10} />
                        <span>{new Date(todo.updated_at).toLocaleDateString()}</span>
                      </div>
                    </CardFooter>
                  </Card>
                ))}
              </div>
            )}

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <section className="flex justify-center items-center gap-3 mt-12">
                <Button
                  variant="outline"
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="h-9 w-9 p-0 flex items-center justify-center rounded-lg"
                >
                  <ChevronLeft size={16} />
                </Button>
                <span className="text-sm font-semibold bg-background border border-border px-4 py-1.5 rounded-lg shadow-sm">
                  Page {page} of {totalPages}
                </span>
                <Button
                  variant="outline"
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages}
                  className="h-9 w-9 p-0 flex items-center justify-center rounded-lg"
                >
                  <ChevronRight size={16} />
                </Button>
              </section>
            )}
          </main>
        </SidebarInset>
      </SidebarProvider>

      {/* Creation Dialog (using Shadcn Dialog) */}
      <Dialog open={isCreateOpen} onOpenChange={(open) => { if (!open) { setIsCreateOpen(false); resetForm(); } }}>
        <DialogContent className="bg-card w-full max-w-lg border border-border p-6 rounded-xl shadow-lg overflow-y-auto max-h-[90vh]" showCloseButton={false}>
          <DialogHeader className="border-b border-border pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-xl font-bold flex items-center gap-2">
              <Plus size={20} className="text-primary" /> Create Task
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                setIsCreateOpen(false);
                resetForm();
              }}
              className="h-8 w-8 rounded-md hover:bg-slate-100"
            >
              <X size={16} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleCreateTodo} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Title</label>
              <Input
                type="text"
                required
                placeholder="What needs to be done?"
                className="w-full h-10 rounded-md border border-input focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-sm"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Description</label>
              <textarea
                rows={3}
                placeholder="Task details..."
                className="w-full p-3 rounded-md border border-input focus:outline-none focus:ring-1 focus:ring-primary bg-transparent text-sm"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Cover Image</label>
              <div className="flex items-center gap-3">
                <label className="h-9 px-4 text-xs font-semibold border border-input rounded-md hover:bg-slate-50 cursor-pointer flex items-center gap-1.5 select-none">
                  <Upload size={14} /> Choose File
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
                    variant="destructive"
                    onClick={() => {
                      setCoverFile(null);
                      setCoverPreview('');
                    }}
                    className="h-9 px-3 text-xs font-semibold"
                  >
                    Remove
                  </Button>
                )}
              </div>
              {coverPreview && (
                <div className="w-full h-32 overflow-hidden border border-border rounded-lg mt-3">
                  <img
                    src={coverPreview}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                  />
                </div>
              )}
            </div>

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsCreateOpen(false);
                  resetForm();
                }}
                className="h-9 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button type="submit" className="h-9 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/95 shadow-sm rounded-md">
                Create Task
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Editing Dialog (using Shadcn Dialog) */}
      <Dialog open={isEditOpen} onOpenChange={(open) => { if (!open) { setIsEditOpen(false); resetForm(); } }}>
        <DialogContent className="bg-card w-full max-w-lg border border-border p-6 rounded-xl shadow-lg overflow-y-auto max-h-[90vh]" showCloseButton={false}>
          <DialogHeader className="border-b border-border pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-xl font-bold flex items-center gap-2">
              <Edit2 size={18} className="text-primary" /> Edit Task
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                setIsEditOpen(false);
                resetForm();
              }}
              className="h-8 w-8 rounded-md hover:bg-slate-100"
            >
              <X size={16} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleUpdateTodo} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Title</label>
              <Input
                type="text"
                required
                placeholder="Task title"
                className="w-full h-10 rounded-md border border-input focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-sm"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Description</label>
              <textarea
                rows={3}
                placeholder="Task description"
                className="w-full p-3 rounded-md border border-input focus:outline-none focus:ring-1 focus:ring-primary bg-transparent text-sm"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-bold text-slate-700 uppercase tracking-wider">Cover Image</label>
              <div className="flex items-center gap-3">
                <label className="h-9 px-4 text-xs font-semibold border border-input rounded-md hover:bg-slate-50 cursor-pointer flex items-center gap-1.5 select-none">
                  <Upload size={14} /> Change File
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
                    variant="destructive"
                    onClick={() => {
                      setCoverFile(null);
                      setCoverPreview('');
                    }}
                    className="h-9 px-3 text-xs font-semibold"
                  >
                    Remove
                  </Button>
                )}
              </div>
              {coverPreview && (
                <div className="w-full h-32 overflow-hidden border border-border rounded-lg mt-3">
                  <img
                    src={coverPreview}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                  />
                </div>
              )}
            </div>

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsEditOpen(false);
                  resetForm();
                }}
                className="h-9 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button type="submit" className="h-9 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/95 shadow-sm rounded-md">
                Save Changes
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog (using Shadcn Dialog) */}
      <Dialog open={!!todoToDelete} onOpenChange={(open) => { if (!open) setTodoToDelete(null); }}>
        <DialogContent className="bg-card w-full max-w-md border border-border p-6 rounded-xl shadow-lg" showCloseButton={false}>
          <DialogHeader className="border-b border-border pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-xl font-bold flex items-center gap-2 text-slate-900">
              <Trash2 size={20} className="text-destructive" /> Delete Task
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setTodoToDelete(null)}
              className="h-8 w-8 rounded-md hover:bg-slate-100"
            >
              <X size={16} />
            </Button>
          </DialogHeader>

          <div className="space-y-5">
            <p className="text-slate-600 text-sm leading-relaxed">
              Are you sure you want to delete the task <span className="font-bold text-slate-900">"{todoToDelete?.title}"</span>? This action is permanent and cannot be undone.
            </p>

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <Button
                type="button"
                variant="outline"
                onClick={() => setTodoToDelete(null)}
                className="h-9 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button 
                variant="destructive"
                onClick={handleDeleteConfirm} 
                className="h-9 text-xs font-semibold px-4 shadow-sm rounded-md"
              >
                Delete
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}
