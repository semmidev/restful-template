import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTheme } from 'next-themes';
import {
  ChevronLeft,
  Calendar,
  Clock,
  Play,
  CheckCircle,
  CheckCircle2,
  CircleDot,
  Circle,
  Trash2,
  AlertCircle,
  Save,
  RotateCcw,
  Sun,
  Moon,
  Shield,
  FileText,
  X,
  Edit2,
  Upload
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import {
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import { TooltipProvider } from '@/components/ui/tooltip';
import { AppSidebar } from '@/components/app-sidebar';
import { Comark } from '@comark/react';
import { MarkdownToolbar } from '../components/MarkdownToolbar';

import useTodoStore, { Todo } from '../store';
import { formatDueAtDisplay, formatToLocalISO } from '@/lib/utils';
import { todoSchema } from '../../../lib/schemas';

export default function TodoDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();

  const {
    currentTodo,
    currentTodoLoading,
    currentTodoError,
    fetchTodoById,
    updateTodo,
    toggleTodoStatus,
    deleteTodo,
    setError
  } = useTodoStore();

  const [isEditingDesc, setIsEditingDesc] = useState(false);
  const [descText, setDescText] = useState('');
  const [descTab, setDescTab] = useState<'preview' | 'edit'>('preview');
  const [isSaving, setIsSaving] = useState(false);
  const [isArchiveOpen, setIsArchiveOpen] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);

  // Edit dialog state variables
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [coverFile, setCoverFile] = useState<File | null>(null);
  const [coverPreview, setCoverPreview] = useState('');
  const [editDueAt, setEditDueAt] = useState('');
  const [validationErrors, setValidationErrors] = useState<{ title?: string; description?: string }>({});
  const [editModalDescTab, setEditModalDescTab] = useState<'write' | 'preview'>('write');
  const editModalTextareaRef = useRef<HTMLTextAreaElement | null>(null);

  useEffect(() => {
    if (id) {
      fetchTodoById(id).then((todo) => {
        if (todo) {
          setDescText(todo.description || '');
        }
      });
    }
    return () => {
      // Clear current todo when leaving the page or changing id to prevent stale data flash
      useTodoStore.setState({ currentTodo: null, currentTodoError: null });
    };
  }, [id, fetchTodoById]);

  if (currentTodoLoading || (!currentTodo && !currentTodoError) || (currentTodo && currentTodo.id !== id)) {
    return (
      <TooltipProvider>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <div className="flex items-center justify-center min-h-screen bg-background text-muted-foreground">
              <div className="flex flex-col items-center gap-3">
                <div className="size-8 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
                <span className="text-sm font-semibold">Loading task details...</span>
              </div>
            </div>
          </SidebarInset>
        </SidebarProvider>
      </TooltipProvider>
    );
  }

  if (currentTodoError || !currentTodo) {
    return (
      <TooltipProvider>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <div className="flex items-center justify-center min-h-screen bg-background">
              <Card className="max-w-md w-full border border-destructive/20 bg-destructive/5 p-6 rounded-lg text-center">
                <AlertCircle className="size-10 text-destructive mx-auto mb-3" />
                <h3 className="text-lg font-bold text-foreground mb-1">Failed to Load Task</h3>
                <p className="text-muted-foreground text-sm mb-4">
                  {currentTodoError || "The requested task could not be found or you do not have permission to view it."}
                </p>
                <Button onClick={() => navigate('/todos')} className="w-full">
                  Back to Tasks
                </Button>
              </Card>
            </div>
          </SidebarInset>
        </SidebarProvider>
      </TooltipProvider>
    );
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'done':
        return <CheckCircle2 size={16} className="text-emerald-500 fill-emerald-500/10" />;
      case 'in_progress':
        return <CircleDot size={16} className="text-primary animate-pulse" />;
      default:
        return <Circle size={16} className="text-muted-foreground/50" />;
    }
  };

  const getPriorityText = (todo: Todo) => {
    if (todo.importance && todo.urgency) return 'Q1: Urgent & Important';
    if (todo.importance && !todo.urgency) return 'Q2: Important & Not Urgent';
    if (!todo.importance && todo.urgency) return 'Q3: Urgent & Not Important';
    return 'Q4: Not Urgent & Not Important';
  };

  const getPriorityColor = (todo: Todo) => {
    if (todo.importance && todo.urgency) return 'bg-red-500/10 border-red-500/20 text-red-500';
    if (todo.importance && !todo.urgency) return 'bg-indigo-500/10 border-indigo-500/20 text-indigo-500';
    if (!todo.importance && todo.urgency) return 'bg-amber-500/10 border-amber-500/20 text-amber-500';
    return 'bg-slate-500/10 border-slate-500/20 text-slate-500';
  };

  const handleStatusChange = async (nextStatus: 'pending' | 'in_progress' | 'done') => {
    await toggleTodoStatus(currentTodo, nextStatus);
    // Refresh to get updated ETag/Timestamp silently
    await fetchTodoById(currentTodo.id, true);
  };

  const handleSaveDescription = async () => {
    setIsSaving(true);
    const success = await updateTodo(
      currentTodo.id,
      currentTodo.title,
      descText,
      null,
      currentTodo.cover || '',
      currentTodo.status,
      currentTodo.updated_at,
      currentTodo.importance,
      currentTodo.urgency,
      currentTodo.due_at || null
    );
    setIsSaving(false);
    if (success) {
      setIsEditingDesc(false);
      setDescTab('preview');
      await fetchTodoById(currentTodo.id, true);
    }
  };

  const handleCancelEdit = () => {
    setDescText(currentTodo.description || '');
    setIsEditingDesc(false);
    setDescTab('preview');
  };

  const handleEditClick = () => {
    if (currentTodo) {
      setEditTitle(currentTodo.title);
      setEditDescription(currentTodo.description || '');
      setCoverPreview(currentTodo.cover || '');
      setEditDueAt(formatToLocalISO(currentTodo.due_at));
      setCoverFile(null);
      setValidationErrors({});
      setEditModalDescTab('write');
      setIsEditOpen(true);
    }
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

  const handleUpdateTodo = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!currentTodo) return;
    setValidationErrors({});
    setError(null);

    const validation = todoSchema.safeParse({ title: editTitle, description: editDescription });
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
      currentTodo.id,
      editTitle,
      editDescription || '',
      coverFile,
      coverPreview,
      currentTodo.status,
      currentTodo.updated_at,
      currentTodo.importance,
      currentTodo.urgency,
      editDueAt ? new Date(editDueAt).toISOString() : null
    );
    if (success) {
      setIsEditOpen(false);
      setDescText(editDescription || '');
    }
  };

  const handleDelete = () => {
    setIsArchiveOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (currentTodo) {
      await deleteTodo(currentTodo.id);
      setIsArchiveOpen(false);
      navigate('/todos');
    }
  };

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
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => navigate('/todos')}
                className="h-7 px-2 text-muted-foreground hover:text-foreground text-xs font-semibold gap-1 rounded-md"
              >
                <ChevronLeft size={14} /> Back to Tasks
              </Button>
            </div>

            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="h-7 w-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-all"
                title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
              </Button>
            </div>
          </header>

          <main className="flex-1 overflow-y-auto bg-background p-6 lg:p-8">
            <div className="max-w-4xl mx-auto space-y-6 relative">
              {/* Decorative Ambient Background Glow */}
              <div className="absolute -top-12 -right-12 w-[250px] h-[250px] bg-primary/5 rounded-full blur-[85px] pointer-events-none z-0" />

              {/* Cover Banner */}
              {currentTodo.cover && (
                <div className="w-full h-48 sm:h-64 rounded-xl overflow-hidden border border-border bg-muted/20 relative z-10 shadow-sm">
                  <img
                    src={currentTodo.cover}
                    alt="Task cover banner"
                    className="w-full h-full object-cover"
                  />
                </div>
              )}

              {/* Task Title & Primary Actions */}
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 relative z-10">
                <div className="space-y-1.5">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="shrink-0">{getStatusIcon(currentTodo.status)}</span>
                    <span className="text-xs font-extrabold uppercase tracking-wider text-muted-foreground capitalize">
                      {currentTodo.status === 'in_progress' ? 'in progress' : currentTodo.status}
                    </span>
                    <span className={`text-xs font-bold px-2 py-0.5 border rounded-full ${getPriorityColor(currentTodo)}`}>
                      {getPriorityText(currentTodo)}
                    </span>
                  </div>
                  <h1 className="text-2xl md:text-3xl font-black tracking-tight text-foreground leading-tight">
                    {currentTodo.title}
                  </h1>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  {currentTodo.status !== 'pending' && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => handleStatusChange('pending')}
                      className="h-8 text-xs font-semibold px-3 border-border/80 hover:bg-accent rounded-md"
                    >
                      <Clock size={12} className="mr-1.5" /> Reopen
                    </Button>
                  )}
                  {currentTodo.status !== 'in_progress' && currentTodo.status !== 'done' && (
                    <Button
                      type="button"
                      variant="default"
                      size="sm"
                      onClick={() => handleStatusChange('in_progress')}
                      className="h-8 text-xs font-semibold px-3 bg-primary text-primary-foreground hover:bg-primary/95 rounded-md"
                    >
                      <Play size={12} className="mr-1.5" /> Start
                    </Button>
                  )}
                  {currentTodo.status !== 'done' && (
                    <Button
                      type="button"
                      variant="default"
                      size="sm"
                      onClick={() => handleStatusChange('done')}
                      className="h-8 text-xs font-semibold px-3 bg-emerald-600 hover:bg-emerald-700 text-white border-none rounded-md"
                    >
                      <CheckCircle size={12} className="mr-1.5" /> Finish
                    </Button>
                  )}

                  <Separator orientation="vertical" className="h-6 mx-1 hidden sm:block" />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                      handleDelete();
                    }}
                    className="h-8 text-xs font-semibold px-3 text-destructive hover:bg-destructive/10 rounded-md"
                  >
                    <Trash2 size={12} className="mr-1.5" /> Archive
                  </Button>
                </div>
              </div>

              {/* Task Metadata & Attributes */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 relative z-10">
                <Card className="border border-border bg-card/25 shadow-none rounded-xl p-5 flex flex-col justify-between">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block mb-2">Due Date</span>
                  <div className="flex items-center gap-2">
                    <Calendar className="size-4 text-primary" />
                    <span className="text-sm font-semibold text-foreground">
                      {currentTodo.due_at ? formatDueAtDisplay(currentTodo.due_at) : 'No due date set'}
                    </span>
                  </div>
                </Card>

                <Card className="border border-border bg-card/25 shadow-none rounded-xl p-5 flex flex-col justify-between">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block mb-2">Created</span>
                  <div className="flex items-center gap-2">
                    <Clock className="size-4 text-muted-foreground" />
                    <span className="text-sm font-semibold text-foreground">
                      {new Date(currentTodo.created_at).toLocaleString()}
                    </span>
                  </div>
                </Card>

                <Card className="border border-border bg-card/25 shadow-none rounded-xl p-5 flex flex-col justify-between">
                  <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block mb-2">Last Updated</span>
                  <div className="flex items-center gap-2">
                    <Clock className="size-4 text-muted-foreground" />
                    <span className="text-sm font-semibold text-foreground">
                      {new Date(currentTodo.updated_at).toLocaleString()}
                    </span>
                  </div>
                </Card>
              </div>

              {/* Task Description with Markdown Render & Inline Editor */}
              <Card className="border border-border bg-card/20 shadow-none rounded-xl overflow-hidden relative z-10">
                <CardHeader className="border-b border-border/60 px-6 py-4 flex flex-row items-center justify-between gap-4">
                  <div className="flex items-center gap-2">
                    <FileText className="size-4 text-primary" />
                    <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">Task Description</h3>
                  </div>

                  <div className="flex items-center gap-2">
                    {isEditingDesc ? (
                      <div className="flex border border-border/60 rounded-md p-0.5 bg-muted/20">
                        <button
                          type="button"
                          onClick={() => setDescTab('edit')}
                          className={`px-2.5 py-0.5 text-xs font-semibold rounded-sm transition-all border-none ${descTab === 'edit' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                        >
                          Write
                        </button>
                        <button
                          type="button"
                          onClick={() => setDescTab('preview')}
                          className={`px-2.5 py-0.5 text-xs font-semibold rounded-sm transition-all border-none ${descTab === 'preview' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                        >
                          Preview
                        </button>
                      </div>
                    ) : (
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          setIsEditingDesc(true);
                          setDescTab('edit');
                        }}
                        className="h-7 text-xs font-semibold px-2.5 border-border/80 hover:bg-accent rounded-md"
                      >
                        Edit Description
                      </Button>
                    )}
                  </div>
                </CardHeader>

                <CardContent className="p-6">
                  {isEditingDesc ? (
                    <div className="space-y-4">
                      {descTab === 'edit' ? (
                        <div className="flex flex-col border border-border/80 rounded-lg overflow-hidden bg-background/50">
                          <MarkdownToolbar
                            textareaRef={textareaRef}
                            value={descText}
                            setValue={setDescText}
                          />
                          <textarea
                            ref={textareaRef}
                            rows={12}
                            value={descText}
                            onChange={(e) => setDescText(e.target.value)}
                            placeholder="Write task details (Markdown supported, e.g. headings, code, lists)..."
                            className="w-full p-4 bg-transparent text-foreground text-sm border-none focus:outline-none focus:ring-0 leading-relaxed resize-y font-normal"
                          />
                        </div>
                      ) : (
                        <div className="w-full min-h-[260px] p-4 border border-border/65 bg-muted/10 text-sm rounded-lg overflow-y-auto markdown-content">
                          {descText.trim() ? (
                            <Comark>{descText}</Comark>
                          ) : (
                            <span className="text-muted-foreground/60 italic text-xs">Nothing to preview. Write something first!</span>
                          )}
                        </div>
                      )}

                      <div className="flex justify-end gap-2 pt-2">
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={handleCancelEdit}
                          className="h-8 text-xs font-semibold px-3 rounded-md"
                          disabled={isSaving}
                        >
                          <RotateCcw className="size-3 mr-1" /> Discard
                        </Button>
                        <Button
                          type="button"
                          onClick={handleSaveDescription}
                          className="h-8 text-xs font-semibold px-3 bg-primary text-primary-foreground hover:bg-primary/95 rounded-md"
                          disabled={isSaving}
                        >
                          <Save className="size-3 mr-1" /> {isSaving ? 'Saving...' : 'Save Changes'}
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="markdown-content min-h-[120px]">
                      {currentTodo.description?.trim() ? (
                        <Comark>{currentTodo.description}</Comark>
                      ) : (
                        <p className="text-muted-foreground/60 italic text-sm text-center py-8">
                          No description provided for this task. Click Edit Description to add details.
                        </p>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </main>
        </SidebarInset>
      </SidebarProvider>

      {/* Archive Confirmation Dialog */}
      <Dialog open={isArchiveOpen} onOpenChange={setIsArchiveOpen}>
        <DialogContent className="bg-card w-full max-w-[92vw] sm:max-w-md border border-border p-6 rounded-lg shadow-lg gap-3" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-base font-bold uppercase tracking-wider flex items-center gap-2 text-foreground">
              <Trash2 size={16} className="text-destructive" /> Archive Task
            </DialogTitle>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setIsArchiveOpen(false)}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <div className="space-y-4">
            <p className="text-muted-foreground text-sm leading-relaxed">
              Are you sure you want to archive <span className="font-bold text-foreground">"{currentTodo?.title}"</span>? You can view and restore it from the Archived tab later.
            </p>

            <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsArchiveOpen(false)}
                className="h-8 text-sm font-semibold"
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="destructive"
                onClick={handleDeleteConfirm}
                className="h-8 text-sm font-semibold px-4 rounded-md"
              >
                Archive
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Editing Dialog */}
      <Dialog open={isEditOpen} onOpenChange={(open) => { if (!open) setIsEditOpen(false); }}>
        <DialogContent className="bg-card w-full max-w-[92vw] sm:max-w-lg border border-border p-6 rounded-lg shadow-lg overflow-y-auto max-h-[90vh]" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-sm font-bold uppercase tracking-wider flex items-center gap-2 text-muted-foreground">
              <Edit2 size={13} className="text-primary" /> Edit Task Properties
            </DialogTitle>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              onClick={() => setIsEditOpen(false)}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <form onSubmit={handleUpdateTodo} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Title</label>
              <Input
                type="text"
                required
                placeholder="Task title..."
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-sm text-foreground px-3"
                value={editTitle}
                onChange={(e) => {
                  setEditTitle(e.target.value);
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
              <div className="flex items-center justify-between">
                <label className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Description</label>
                <div className="flex border border-border/60 rounded-md p-0.5 bg-muted/20">
                  <button
                    type="button"
                    onClick={() => setEditModalDescTab('write')}
                    className={`px-2 py-0.5 text-xs font-semibold rounded-sm transition-all border-none ${editModalDescTab === 'write' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                  >
                    Write
                  </button>
                  <button
                    type="button"
                    onClick={() => setEditModalDescTab('preview')}
                    className={`px-2 py-0.5 text-xs font-semibold rounded-sm transition-all border-none ${editModalDescTab === 'preview' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
                  >
                    Preview
                  </button>
                </div>
              </div>

              {editModalDescTab === 'write' ? (
                <div className="flex flex-col border border-border/80 rounded-md overflow-hidden bg-transparent">
                  <MarkdownToolbar
                    textareaRef={editModalTextareaRef}
                    value={editDescription}
                    setValue={setEditDescription}
                  />
                  <textarea
                    ref={editModalTextareaRef}
                    rows={4}
                    placeholder="Add task details & description (Markdown supported)..."
                    className="w-full p-3 bg-transparent text-sm text-foreground border-none focus:outline-none focus:ring-0 resize-none leading-relaxed"
                    value={editDescription}
                    onChange={(e) => {
                      setEditDescription(e.target.value);
                      if (validationErrors.description) {
                        setValidationErrors((prev) => ({ ...prev, description: undefined }));
                      }
                    }}
                  ></textarea>
                </div>
              ) : (
                <div className="w-full h-[106px] p-3 rounded-md border border-border/80 bg-muted/10 text-sm text-foreground resize-none leading-relaxed overflow-y-auto markdown-content">
                  {editDescription.trim() ? (
                    <Comark>{editDescription}</Comark>
                  ) : (
                    <span className="text-muted-foreground/60 italic text-xs">Nothing to preview. Write something first!</span>
                  )}
                </div>
              )}
              {validationErrors.description && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.description}
                </p>
              )}
            </div>

            {/* Metadata Attributes Section */}
            <div className="border-t border-border/50 pt-4 space-y-3">
              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Due Date</span>
                <div className="sm:col-span-2 relative">
                  <Calendar size={12} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60 pointer-events-none" />
                  <Input
                    type="datetime-local"
                    className="w-full h-8 pl-8 pr-2 border border-border/60 rounded-md bg-muted/10 text-sm focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 [color-scheme:dark] text-foreground custom-datetime-input cursor-pointer"
                    value={editDueAt}
                    onChange={(e) => setEditDueAt(e.target.value)}
                  />
                </div>
              </div>

              <div className="flex flex-col sm:grid sm:grid-cols-3 sm:items-center gap-1.5 sm:gap-2">
                <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider">Cover Image</span>
                <div className="sm:col-span-2 flex items-center gap-2">
                  <label className="h-7 px-2.5 text-xs font-semibold border border-border/80 rounded-md hover:bg-accent cursor-pointer flex items-center gap-1 select-none text-foreground bg-muted/10">
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
                      className="h-7 px-2 text-xs font-semibold text-destructive hover:bg-destructive/10 border-none"
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
                onClick={() => setIsEditOpen(false)}
                className="h-8 text-sm font-semibold"
              >
                Cancel
              </Button>
              <Button type="submit" className="h-8 text-sm font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-md">
                Save Changes
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}
