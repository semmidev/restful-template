import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
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
  CheckCircle2,
  FileText,
  ChevronLeft,
  ChevronRight,
  Activity,
  Upload,
  X,
  Sun,
  Moon,
  ArrowUp,
  ArrowDown,
  Archive,
  CircleDot,
  Circle,
  List,
  LayoutGrid
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
import { formatToLocalISO, formatDueAtDisplay } from '@/lib/utils';
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
    archived,
    loading,
    error,
    editingTodo,
    fetchTodos,
    createTodo,
    updateTodo,
    toggleTodoStatus,
    deleteTodo,
    restoreTodo,
    setFilters,
    setPage,
    setEditingTodo,
    setError,
  } = useTodoStore();

  // View mode preference
  const [viewMode, setViewMode] = useState<'list' | 'card'>((localStorage.getItem('todos_view_mode') as 'list' | 'card') || 'list');

  useEffect(() => {
    localStorage.setItem('todos_view_mode', viewMode);
  }, [viewMode]);

  // Helper for status icons
  const getStatusIcon = (item: Todo) => {
    if (item.deleted_at) {
      return <Archive size={13} className="text-muted-foreground/60" />;
    }
    switch (item.status) {
      case 'done':
        return <CheckCircle2 size={13} className="text-emerald-500 fill-emerald-500/10" />;
      case 'in_progress':
        return <CircleDot size={13} className="text-primary animate-pulse" />;
      default:
        return <Circle size={13} className="text-muted-foreground/50" />;
    }
  };

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [todoToDelete, setTodoToDelete] = useState<Todo | null>(null);

  // Form state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [coverFile, setCoverFile] = useState<File | null>(null);
  const [coverPreview, setCoverPreview] = useState('');
  const [dueAt, setDueAt] = useState('');

  // Form validation errors
  const [validationErrors, setValidationErrors] = useState<{ title?: string; description?: string }>({});

  // Local debounced keyword state
  const [searchKeyword, setSearchKeyword] = useState(keyword);

  useEffect(() => {
    fetchTodos();
  }, [page, status, sortBy, sortDir, keyword, archived, fetchTodos]);

  useEffect(() => {
    setSearchKeyword(keyword);
  }, [keyword]);

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

    const success = await createTodo(
      title,
      description || '',
      coverFile,
      undefined,
      undefined,
      dueAt ? new Date(dueAt).toISOString() : null
    );
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
    setDueAt(formatToLocalISO(todo.due_at));
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
      editingTodo.updated_at,
      undefined,
      undefined,
      dueAt ? new Date(dueAt).toISOString() : null
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
    setDueAt('');
    setEditingTodo(null);
    setValidationErrors({});
  };

  const totalPages = Math.max(1, Math.ceil(total / perPage));

  return (
    <TooltipProvider>
      {/* CSS custom styles for datetime input */}
      <style dangerouslySetInnerHTML={{ __html: `
        .custom-datetime-input::-webkit-calendar-picker-indicator {
          background: transparent;
          bottom: 0;
          color: transparent;
          cursor: pointer;
          height: auto;
          left: 0;
          position: absolute;
          right: 0;
          top: 0;
          width: auto;
        }
      `}} />
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          {/* Header */}
          <header className="flex h-14 shrink-0 items-center justify-between gap-4 border-b bg-background/50 backdrop-blur px-4 lg:px-6">
            <div className="flex items-center gap-2">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <span className="font-medium hover:text-foreground transition-colors cursor-pointer hidden sm:inline" onClick={() => navigate('/')}>Workspace</span>
                <ChevronRight size={12} className="text-muted-foreground/60 hidden sm:inline" />
                <span className="font-semibold text-foreground">Tasks</span>
              </div>
            </div>

            {/* Header Actions */}
            <div className="flex items-center gap-2">
              <div className="relative w-24 sm:w-48 md:w-60">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60 size-3.5" />
                <Input
                  type="text"
                  placeholder="Search tasks..."
                  className="w-full pl-8 h-7 rounded-md bg-muted/40 border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 text-xs"
                  value={searchKeyword}
                  onChange={(e) => setSearchKeyword(e.target.value)}
                />
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="h-7 w-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-all"
                title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
              </Button>
              <Button
                onClick={() => setIsCreateOpen(true)}
                className="h-7 px-2 sm:px-2.5 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition-all rounded-md flex items-center gap-1"
              >
                <Plus size={14} />
                <span className="hidden sm:inline">New Task</span>
              </Button>
            </div>
          </header>

          {/* Main Content Area */}
          <main className="flex flex-col gap-6 p-6 lg:p-8 bg-background min-h-[calc(100vh-3.5rem)]">
            {/* Error Banner */}
            {error && (
              <div className="bg-destructive/10 border border-destructive/20 text-destructive text-xs font-semibold p-4 rounded-lg flex justify-between items-center shadow-sm">
                <span>⚠️ {error}</span>
                <Button variant="ghost" onClick={() => setError(null)} className="h-6 w-6 p-0 text-destructive hover:bg-destructive/10">
                  <X size={14} />
                </Button>
              </div>
            )}

            {/* Filters Toolbar */}
            <section className="flex flex-col md:flex-row gap-3.5 justify-between items-stretch md:items-center py-2 border-b border-border/50">
              <Tabs value={archived ? 'archived' : status} onValueChange={(val) => {
                if (val === 'archived') {
                  setFilters({ archived: true, status: '' });
                } else {
                  setFilters({ archived: false, status: val });
                }
              }}>
                <TabsList className="bg-transparent p-0 gap-1 h-8 border-b border-transparent overflow-x-auto max-w-full flex-nowrap justify-start [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
                  {[
                    { label: 'All Tasks', value: '' },
                    { label: 'Pending', value: 'pending' },
                    { label: 'In Progress', value: 'in_progress' },
                    { label: 'Completed', value: 'done' },
                    { label: 'Archived', value: 'archived' },
                  ].map((tab) => (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className="px-3 h-7 text-xs font-semibold text-muted-foreground hover:text-foreground transition-all border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:bg-transparent rounded-none shadow-none shrink-0"
                    >
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>

              <div className="flex flex-wrap items-center gap-2">
                {/* View Switcher */}
                <div className="flex items-center border border-border/60 rounded-md p-0.5 bg-muted/20">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setViewMode('list')}
                    className={`h-6 w-6 rounded-sm p-0 transition-all border-none ${viewMode === 'list' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                    title="List View"
                  >
                    <List size={13} />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setViewMode('card')}
                    className={`h-6 w-6 rounded-sm p-0 transition-all border-none ${viewMode === 'card' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                    title="Card/Grid View"
                  >
                    <LayoutGrid size={13} />
                  </Button>
                </div>

                <Separator orientation="vertical" className="h-4 hidden sm:block" />

                <select
                  className="h-7 px-2 rounded-md border border-border/80 bg-background text-[11px] font-semibold focus:outline-none focus:ring-1 focus:ring-primary text-foreground"
                  value={sortBy}
                  onChange={(e) => setFilters({ sortBy: e.target.value })}
                >
                  <option value="created_at">Sort by Created</option>
                  <option value="updated_at">Sort by Updated</option>
                  <option value="title">Sort by Title</option>
                  <option value="status">Sort by Status</option>
                </select>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFilters({ sortDir: sortDir === 'asc' ? 'desc' : 'asc' })}
                  className="h-7 text-[11px] font-semibold border-border/80 flex items-center gap-1 hover:bg-accent px-2 rounded-md"
                >
                  {sortDir === 'asc' ? <ArrowUp className="size-3" /> : <ArrowDown className="size-3" />}
                  {sortDir === 'asc' ? 'Asc' : 'Desc'}
                </Button>
              </div>
            </section>

            {/* Tasks Section */}
            {loading && todos.length === 0 ? (
              viewMode === 'list' ? (
                <div className="flex flex-col border border-border/60 rounded-lg bg-card/10 overflow-hidden divide-y divide-border/40">
                  {Array.from({ length: 6 }).map((_, index) => (
                    <div key={index} className="h-10 w-full animate-pulse bg-muted/20" />
                  ))}
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {Array.from({ length: 6 }).map((_, index) => (
                    <TodoSkeleton key={index} />
                  ))}
                </div>
              )
            ) : todos.length === 0 ? (
              <Card className="border border-border border-dashed bg-transparent text-center py-16 rounded-lg">
                <CardContent className="flex flex-col items-center justify-center p-0">
                  <div className="w-10 h-10 border border-border rounded-full flex items-center justify-center text-muted-foreground mb-3">
                    <FileText size={18} />
                  </div>
                  <h3 className="text-sm font-bold mb-1">No tasks in this view</h3>
                  <p className="text-muted-foreground text-xs max-w-xs mb-4">Create a task to populate this category.</p>
                  <Button
                    onClick={() => setIsCreateOpen(true)}
                    className="h-8 text-xs font-semibold"
                  >
                    <Plus size={14} className="mr-1" /> New Task
                  </Button>
                </CardContent>
              </Card>
            ) : viewMode === 'list' ? (
              <div className="flex flex-col border border-border/60 rounded-lg bg-card/10 overflow-hidden divide-y divide-border/40">
                {todos.map((todo) => (
                  <div key={todo.id} className="group flex flex-col sm:flex-row sm:items-center justify-between px-4 py-3 hover:bg-muted/30 transition-colors gap-3 text-xs">
                    <div className="flex items-center gap-3 min-w-0 flex-1">
                      <span className="shrink-0">{getStatusIcon(todo)}</span>
                      <div className="flex flex-col sm:flex-row sm:items-baseline gap-1 sm:gap-2.5 min-w-0 flex-1">
                        <span className={`font-semibold text-foreground truncate ${todo.status === 'done' ? 'line-through text-muted-foreground/60' : ''}`}>
                          {todo.title}
                        </span>
                        {todo.description && (
                          <span className="text-[11px] text-muted-foreground truncate max-w-md">
                            — {todo.description}
                          </span>
                        )}
                      </div>
                    </div>
                    
                    <div className="flex flex-wrap items-center gap-4 shrink-0 justify-between sm:justify-end">
                      {/* Due date column */}
                      <div className="w-32 flex items-center gap-1.5 text-[10px] text-muted-foreground font-medium shrink-0">
                        {todo.due_at ? (
                          <>
                            <Calendar size={11} className="text-muted-foreground/60 shrink-0" />
                            <span>{formatDueAtDisplay(todo.due_at)}</span>
                          </>
                        ) : (
                          <span className="text-muted-foreground/30">—</span>
                        )}
                      </div>
                      
                      {/* Cover Image Indicator */}
                      <div className="w-14 shrink-0 flex justify-start">
                        {todo.cover ? (
                          <div className="text-[9px] font-extrabold uppercase tracking-wider bg-primary/10 border border-primary/20 text-primary px-1.5 py-0.5 rounded">
                            Image
                          </div>
                        ) : (
                          <span className="text-muted-foreground/30">—</span>
                        )}
                      </div>
                      
                      {/* Actions Controls (aligned like Card View) */}
                      <div className="flex items-center gap-1.5 min-w-[150px] justify-end">
                        {todo.deleted_at ? (
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => restoreTodo(todo.id)}
                            className="h-6 text-[10px] font-semibold px-2 rounded-md bg-emerald-600 hover:bg-emerald-700 text-white transition-colors border-none shadow-none"
                          >
                            Restore
                          </Button>
                        ) : (
                          <>
                            {todo.status !== 'pending' && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'pending')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md border border-border/80 hover:bg-accent text-foreground transition-colors"
                              >
                                <Clock size={10} className="sm:mr-1 inline-block" />
                                <span className="hidden sm:inline"> Reopen</span>
                              </Button>
                            )}
                            {todo.status !== 'in_progress' && todo.status !== 'done' && (
                              <Button
                                variant="default"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'in_progress')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md bg-primary hover:bg-primary/90 text-primary-foreground transition-colors border-none shadow-none"
                              >
                                <Play size={10} className="sm:mr-1 inline-block" />
                                <span className="hidden sm:inline"> Start</span>
                              </Button>
                            )}
                            {todo.status !== 'done' && (
                              <Button
                                variant="default"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'done')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md bg-emerald-600 hover:bg-emerald-700 text-white transition-colors border-none shadow-none"
                              >
                                <CheckCircle size={10} className="sm:mr-1 inline-block" />
                                <span className="hidden sm:inline"> Finish</span>
                              </Button>
                            )}

                            <Separator orientation="vertical" className="h-4 mx-1" />

                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleEditClick(todo)}
                              className="h-6 w-6 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent border-none"
                              title="Edit task"
                            >
                              <Edit2 size={11} />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDeleteClick(todo)}
                              className="h-6 w-6 rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 border-none"
                              title="Archive task"
                            >
                              <Trash2 size={11} />
                            </Button>
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {todos.map((todo) => (
                  <Card
                    key={todo.id}
                    className="border border-border bg-card/25 hover:border-border/80 transition-all duration-200 rounded-lg overflow-hidden flex flex-col justify-between p-0 shadow-none hover:shadow-sm"
                  >
                    <div>
                      {/* Cover Image */}
                      {todo.cover && (
                        <div className="w-full h-32 overflow-hidden border-b border-border bg-muted/30 flex items-center justify-center">
                          <img
                            src={todo.cover}
                            alt="Todo Cover"
                            className="w-full h-full object-cover"
                          />
                        </div>
                      )}

                      <CardHeader className="pt-4 pb-2 px-5 space-y-2">
                        <div className="flex justify-between items-center gap-2">
                          <div className="flex items-center gap-1.5">
                            {getStatusIcon(todo)}
                            <span className="text-[10px] font-bold text-muted-foreground capitalize">
                              {todo.status === 'in_progress' ? 'in progress' : todo.status}
                            </span>
                          </div>
                          <div className="flex gap-1">
                            {!todo.deleted_at && (
                              <>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleEditClick(todo)}
                                  className="h-6 w-6 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent border-none"
                                >
                                  <Edit2 size={12} />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleDeleteClick(todo)}
                                  className="h-6 w-6 rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 border-none"
                                >
                                  <Trash2 size={12} />
                                </Button>
                              </>
                            )}
                          </div>
                        </div>
                        <CardTitle className="text-sm font-bold text-foreground line-clamp-1 leading-snug">
                          {todo.title}
                        </CardTitle>
                      </CardHeader>

                      <CardContent className="pb-4 px-5">
                        <p className="text-muted-foreground text-xs leading-relaxed line-clamp-3">
                          {todo.description || 'No description provided.'}
                        </p>
                      </CardContent>
                    </div>

                    <CardFooter className="border-t border-border/60 pt-3 pb-3 px-5 flex items-center justify-between gap-4 mt-auto">
                      <div className="flex gap-1.5">
                        {todo.deleted_at ? (
                          <Button
                            variant="default"
                            size="sm"
                            onClick={() => restoreTodo(todo.id)}
                            className="h-6 text-[10px] font-semibold px-2 rounded-md bg-emerald-600 hover:bg-emerald-700 text-white transition-colors shadow-none"
                          >
                            Restore
                          </Button>
                        ) : (
                          <>
                            {todo.status !== 'pending' && (
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'pending')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md border border-border/80 hover:bg-accent text-foreground transition-colors"
                              >
                                <Clock size={10} className="mr-1 inline-block" /> Reopen
                              </Button>
                            )}
                            {todo.status !== 'in_progress' && todo.status !== 'done' && (
                              <Button
                                variant="default"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'in_progress')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md bg-primary hover:bg-primary/90 text-primary-foreground transition-colors shadow-none"
                              >
                                <Play size={10} className="mr-1 inline-block" /> Start
                              </Button>
                            )}
                            {todo.status !== 'done' && (
                              <Button
                                variant="default"
                                size="sm"
                                onClick={() => handleToggleStatus(todo, 'done')}
                                className="h-6 text-[10px] font-semibold px-2 rounded-md bg-emerald-600 hover:bg-emerald-700 text-white transition-colors shadow-none"
                              >
                                <CheckCircle size={10} className="mr-1 inline-block" /> Finish
                              </Button>
                            )}
                          </>
                        )}
                      </div>
                      <div className="flex items-center gap-1 text-[9px] text-muted-foreground font-medium">
                        <Calendar size={10} />
                        <span>{todo.due_at ? `Due: ${formatDueAtDisplay(todo.due_at)}` : new Date(todo.updated_at).toLocaleDateString()}</span>
                      </div>
                    </CardFooter>
                  </Card>
                ))}
              </div>
            )}

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <section className="flex justify-center items-center gap-2 mt-8">
                <Button
                  variant="outline"
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page === 1}
                  className="h-8 w-8 p-0 flex items-center justify-center rounded-md border-border/80 hover:bg-accent"
                >
                  <ChevronLeft size={14} />
                </Button>
                <span className="text-xs font-semibold bg-card/25 border border-border px-3 py-1.5 rounded-md">
                  Page {page} of {totalPages}
                </span>
                <Button
                  variant="outline"
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page === totalPages}
                  className="h-8 w-8 p-0 flex items-center justify-center rounded-md border-border/80 hover:bg-accent"
                >
                  <ChevronRight size={14} />
                </Button>
              </section>
            )}
          </main>
        </SidebarInset>
      </SidebarProvider>

      {/* Creation Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={(open) => { if (!open) { setIsCreateOpen(false); resetForm(); } }}>
        <DialogContent className="bg-card w-full max-w-[92vw] sm:max-w-lg border border-border p-6 rounded-lg shadow-lg overflow-y-auto max-h-[90vh]" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-xs font-bold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
              <Plus size={14} className="text-primary" /> Create New Task
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                setIsCreateOpen(false);
                resetForm();
              }}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleCreateTodo} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Title</label>
              <Input
                type="text"
                required
                placeholder="Task title..."
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs text-foreground px-3"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-destructive text-[11px] font-semibold mt-1">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Description</label>
              <textarea
                rows={4}
                placeholder="Add task details & description..."
                className="w-full p-3 rounded-md border border-border/80 focus:outline-none focus:ring-1 focus:ring-primary bg-transparent text-xs text-foreground resize-none leading-relaxed"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-destructive text-[11px] font-semibold mt-1">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            {/* Metadata Attributes Section */}
            <div className="border-t border-border/50 pt-4 space-y-3">
              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Due Date</span>
                <div className="sm:col-span-2 relative">
                  <Calendar size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60 pointer-events-none" />
                  <Input
                    type="datetime-local"
                    className="w-full h-8 pl-8 pr-2 border border-border/60 rounded-md bg-muted/10 text-xs focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 [color-scheme:dark] text-foreground custom-datetime-input cursor-pointer"
                    value={dueAt}
                    onChange={(e) => setDueAt(e.target.value)}
                  />
                </div>
              </div>

              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Cover Image</span>
                <div className="sm:col-span-2 flex items-center gap-2">
                  <label className="h-7 px-2.5 text-[11px] font-semibold border border-border/80 rounded-md hover:bg-accent cursor-pointer flex items-center gap-1 select-none text-foreground bg-muted/10">
                    <Upload size={12} /> Choose Image
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
                      variant="ghost"
                      onClick={() => {
                        setCoverFile(null);
                        setCoverPreview('');
                      }}
                      className="h-7 px-2 text-[10px] font-semibold text-destructive hover:bg-destructive/10 border-none"
                    >
                      Remove
                    </Button>
                  )}
                </div>
              </div>

              {coverPreview && (
                <div className="w-full h-24 overflow-hidden border border-border/65 rounded-md mt-1">
                  <img
                    src={coverPreview}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                  />
                </div>
              )}
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsCreateOpen(false);
                  resetForm();
                }}
                className="h-8 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button type="submit" className="h-8 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-md">
                Create Task
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Editing Dialog */}
      <Dialog open={isEditOpen} onOpenChange={(open) => { if (!open) { setIsEditOpen(false); resetForm(); } }}>
        <DialogContent className="bg-card w-full max-w-[92vw] sm:max-w-lg border border-border p-6 rounded-lg shadow-lg overflow-y-auto max-h-[90vh]" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-xs font-bold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
              <Edit2 size={13} className="text-primary" /> Edit Task Properties
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => {
                setIsEditOpen(false);
                resetForm();
              }}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleUpdateTodo} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Title</label>
              <Input
                type="text"
                required
                placeholder="Task title..."
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs text-foreground px-3"
                value={title}
                onChange={(e) => {
                  setTitle(e.target.value);
                  if (validationErrors.title) {
                    setValidationErrors((prev) => ({ ...prev, title: undefined }));
                  }
                }}
              />
              {validationErrors.title && (
                <p className="text-destructive text-[11px] font-semibold mt-1">
                  ⚠️ {validationErrors.title}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Description</label>
              <textarea
                rows={4}
                placeholder="Add task details & description..."
                className="w-full p-3 rounded-md border border-border/80 focus:outline-none focus:ring-1 focus:ring-primary bg-transparent text-xs text-foreground resize-none leading-relaxed"
                value={description}
                onChange={(e) => {
                  setDescription(e.target.value);
                  if (validationErrors.description) {
                    setValidationErrors((prev) => ({ ...prev, description: undefined }));
                  }
                }}
              ></textarea>
              {validationErrors.description && (
                <p className="text-destructive text-[11px] font-semibold mt-1">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            {/* Metadata Attributes Section */}
            <div className="border-t border-border/50 pt-4 space-y-3">
              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Due Date</span>
                <div className="sm:col-span-2 relative">
                  <Calendar size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60 pointer-events-none" />
                  <Input
                    type="datetime-local"
                    className="w-full h-8 pl-8 pr-2 border border-border/60 rounded-md bg-muted/10 text-xs focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 [color-scheme:dark] text-foreground custom-datetime-input cursor-pointer"
                    value={dueAt}
                    onChange={(e) => setDueAt(e.target.value)}
                  />
                </div>
              </div>

              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">Cover Image</span>
                <div className="sm:col-span-2 flex items-center gap-2">
                  <label className="h-7 px-2.5 text-[11px] font-semibold border border-border/80 rounded-md hover:bg-accent cursor-pointer flex items-center gap-1 select-none text-foreground bg-muted/10">
                    <Upload size={12} /> Choose Image
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
                      variant="ghost"
                      onClick={() => {
                        setCoverFile(null);
                        setCoverPreview('');
                      }}
                      className="h-7 px-2 text-[10px] font-semibold text-destructive hover:bg-destructive/10 border-none"
                    >
                      Remove
                    </Button>
                  )}
                </div>
              </div>

              {coverPreview && (
                <div className="w-full h-24 overflow-hidden border border-border/65 rounded-md mt-1">
                  <img
                    src={coverPreview}
                    alt="Cover preview"
                    className="w-full h-full object-cover"
                  />
                </div>
              )}
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsEditOpen(false);
                  resetForm();
                }}
                className="h-8 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button type="submit" className="h-8 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-md">
                Save Changes
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={!!todoToDelete} onOpenChange={(open) => { if (!open) setTodoToDelete(null); }}>
        <DialogContent className="bg-card w-full max-w-[92vw] sm:max-w-md border border-border p-6 rounded-lg shadow-lg gap-3" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-sm font-bold uppercase tracking-wider flex items-center gap-2 text-foreground">
              <Trash2 size={16} className="text-destructive" /> Archive Task
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setTodoToDelete(null)}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <div className="space-y-4">
            <p className="text-muted-foreground text-xs leading-relaxed">
              Are you sure you want to archive <span className="font-bold text-foreground">"{todoToDelete?.title}"</span>? You can view and restore it from the Archived tab later.
            </p>

            <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                onClick={() => setTodoToDelete(null)}
                className="h-8 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteConfirm}
                className="h-8 text-xs font-semibold px-4 rounded-md"
              >
                Archive
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}
