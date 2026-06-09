import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTheme } from 'next-themes';
import {
  AlertTriangle,
  Calendar,
  ChevronRight,
  Clock,
  GripVertical,
  Layers,
  Moon,
  Plus,
  Sun,
  Trash2,
  X,
  Zap,
  CheckCircle2,
  CircleDot,
  Circle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import {
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import { TooltipProvider } from '@/components/ui/tooltip';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { AppSidebar } from '@/components/app-sidebar';
import useTodoStore, { Todo } from '../store';
import { todoSchema } from '../../../lib/schemas';

// ─── Quadrant Config ─────────────────────────────────────────────────────────

interface Quadrant {
  id: string;
  label: string;
  sublabel: string;
  importance: boolean;
  urgency: boolean;
  icon: React.ReactNode;
  accent: string;
  accentBg: string;
  accentBorder: string;
  accentText: string;
  dotColor: string;
}

const QUADRANTS: Quadrant[] = [
  {
    id: 'do-first',
    label: 'Do First',
    sublabel: 'Urgent & Important',
    importance: true,
    urgency: true,
    icon: <Zap size={14} />,
    accent: 'text-rose-500',
    accentBg: 'bg-rose-500/5',
    accentBorder: 'border-rose-500/20',
    accentText: 'text-rose-500',
    dotColor: '#f43f5e',
  },
  {
    id: 'schedule',
    label: 'Schedule',
    sublabel: 'Not Urgent & Important',
    importance: true,
    urgency: false,
    icon: <Calendar size={14} />,
    accent: 'text-primary',
    accentBg: 'bg-primary/5',
    accentBorder: 'border-primary/20',
    accentText: 'text-primary',
    dotColor: 'hsl(250 56% 60%)',
  },
  {
    id: 'delegate',
    label: 'Delegate',
    sublabel: 'Urgent & Not Important',
    importance: false,
    urgency: true,
    icon: <Clock size={14} />,
    accent: 'text-amber-500',
    accentBg: 'bg-amber-500/5',
    accentBorder: 'border-amber-500/20',
    accentText: 'text-amber-500',
    dotColor: '#f59e0b',
  },
  {
    id: 'eliminate',
    label: 'Eliminate',
    sublabel: 'Not Urgent & Not Important',
    importance: false,
    urgency: false,
    icon: <Trash2 size={14} />,
    accent: 'text-muted-foreground',
    accentBg: 'bg-muted/30',
    accentBorder: 'border-border/60',
    accentText: 'text-muted-foreground',
    dotColor: 'hsl(240 4% 46%)',
  },
];

// ─── Todo Card ────────────────────────────────────────────────────────────────

interface TodoCardProps {
  todo: Todo;
  quadrant: Quadrant;
  onDragStart: (e: React.DragEvent, todo: Todo) => void;
  onDelete: (id: string) => void;
}

function TodoCard({ todo, quadrant, onDragStart, onDelete }: TodoCardProps) {
  return (
    <div
      draggable
      onDragStart={(e) => onDragStart(e, todo)}
      className="group flex items-start gap-2.5 px-3 py-2.5 rounded-lg border border-border/60 bg-card/40 hover:bg-card/80 hover:border-border/80 cursor-grab active:cursor-grabbing transition-all duration-150 select-none"
    >
      <GripVertical
        size={12}
        className="mt-0.5 shrink-0 text-muted-foreground/30 group-hover:text-muted-foreground/60 transition-colors"
      />
      <div className="flex-1 min-w-0">
        <p className="text-xs font-semibold text-foreground leading-snug line-clamp-2">
          {todo.title}
        </p>
        {todo.description && (
          <p className="text-[10px] text-muted-foreground mt-0.5 line-clamp-1">
            {todo.description}
          </p>
        )}
        <div className="flex items-center gap-1.5 mt-1.5">
          {todo.status === 'done' ? (
            <CheckCircle2 size={11} className="text-emerald-500 shrink-0" />
          ) : todo.status === 'in_progress' ? (
            <CircleDot size={11} className="text-primary shrink-0 animate-pulse" />
          ) : (
            <Circle size={11} className="text-muted-foreground/50 shrink-0" />
          )}
          <span className="text-[10px] text-muted-foreground/80 font-semibold capitalize">
            {todo.status === 'in_progress' ? 'in progress' : todo.status}
          </span>
          <span className="text-muted-foreground/30 text-[9px]">•</span>
          <span className="text-[9px] text-muted-foreground/60 font-medium">
            {todo.due_at ? `Due: ${new Date(todo.due_at).toLocaleDateString()}` : new Date(todo.updated_at).toLocaleDateString()}
          </span>
        </div>
      </div>
      <button
        onClick={(e) => {
          e.stopPropagation();
          onDelete(todo.id);
        }}
        className="opacity-0 group-hover:opacity-100 shrink-0 p-1 rounded text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-all"
      >
        <X size={10} />
      </button>
    </div>
  );
}

// ─── Quadrant Panel ──────────────────────────────────────────────────────────

interface QuadrantPanelProps {
  quadrant: Quadrant;
  todos: Todo[];
  isDragOver: boolean;
  onDragOver: (e: React.DragEvent) => void;
  onDragLeave: () => void;
  onDrop: (e: React.DragEvent, q: Quadrant) => void;
  onDragStart: (e: React.DragEvent, todo: Todo) => void;
  onDelete: (id: string) => void;
  onQuickAdd: (q: Quadrant) => void;
}

function QuadrantPanel({
  quadrant,
  todos,
  isDragOver,
  onDragOver,
  onDragLeave,
  onDrop,
  onDragStart,
  onDelete,
  onQuickAdd,
}: QuadrantPanelProps) {
  return (
    <div
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={(e) => onDrop(e, quadrant)}
      className={`flex flex-col min-h-0 rounded-xl border transition-all duration-200 ${
        isDragOver
          ? `${quadrant.accentBg} ${quadrant.accentBorder} shadow-sm`
          : 'border-border/60 bg-card/15'
      }`}
    >
      {/* Quadrant Header */}
      <div className={`flex items-center justify-between px-4 py-3 border-b ${isDragOver ? quadrant.accentBorder : 'border-border/40'}`}>
        <div className="flex items-center gap-2">
          <div
            className="flex items-center justify-center w-5 h-5 rounded"
            style={{ backgroundColor: `${quadrant.dotColor}20`, color: quadrant.dotColor }}
          >
            {quadrant.icon}
          </div>
          <div>
            <p className={`text-xs font-bold ${quadrant.accentText}`}>
              {quadrant.label}
            </p>
            <p className="text-[9px] text-muted-foreground font-medium">
              {quadrant.sublabel}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="text-[10px] font-bold text-muted-foreground bg-muted/60 px-1.5 py-0.5 rounded-md">
            {todos.length}
          </span>
          <button
            onClick={() => onQuickAdd(quadrant)}
            className={`flex items-center justify-center w-5 h-5 rounded border transition-all ${quadrant.accentBorder} ${quadrant.accentBg} ${quadrant.accentText} hover:opacity-80`}
            title={`Add task to ${quadrant.label}`}
          >
            <Plus size={10} />
          </button>
        </div>
      </div>

      {/* Quadrant Body */}
      <div className="flex-1 overflow-y-auto p-3 space-y-2 min-h-[160px] max-h-[calc(50vh-80px)]">
        {todos.length === 0 ? (
          <div
            className={`flex flex-col items-center justify-center h-full min-h-[120px] rounded-lg border border-dashed transition-colors ${
              isDragOver ? quadrant.accentBorder : 'border-border/30'
            }`}
          >
            <div
              className="w-6 h-6 rounded-full flex items-center justify-center mb-2"
              style={{ backgroundColor: `${quadrant.dotColor}15` }}
            >
              <Layers size={12} style={{ color: quadrant.dotColor }} />
            </div>
            <p className="text-[10px] text-muted-foreground font-semibold">
              {isDragOver ? 'Drop here' : 'No tasks'}
            </p>
          </div>
        ) : (
          todos.map((todo) => (
            <TodoCard
              key={todo.id}
              todo={todo}
              quadrant={quadrant}
              onDragStart={onDragStart}
              onDelete={onDelete}
            />
          ))
        )}
      </div>
    </div>
  );
}

// ─── Quick Add Dialog ─────────────────────────────────────────────────────────

interface QuickAddDialogProps {
  quadrant: Quadrant | null;
  open: boolean;
  onClose: () => void;
  onSubmit: (title: string, description: string, importance: boolean, urgency: boolean) => Promise<boolean>;
}

function QuickAddDialog({ quadrant, open, onClose, onSubmit }: QuickAddDialogProps) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [titleError, setTitleError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const titleRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setTitle('');
      setDescription('');
      setTitleError('');
      setTimeout(() => titleRef.current?.focus(), 50);
    }
  }, [open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setTitleError('');
    const result = todoSchema.safeParse({ title, description });
    if (!result.success) {
      const titleIssue = result.error.issues.find((i) => i.path[0] === 'title');
      if (titleIssue) setTitleError(titleIssue.message);
      return;
    }
    if (!quadrant) return;
    setSubmitting(true);
    const ok = await onSubmit(title, description, quadrant.importance, quadrant.urgency);
    setSubmitting(false);
    if (ok) onClose();
  };

  if (!quadrant) return null;

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent
        className="bg-card w-full max-w-md border border-border p-6 rounded-xl shadow-xl"
        showCloseButton={false}
      >
        <DialogHeader className="border-b border-border/60 pb-3 mb-4 flex flex-row justify-between items-center gap-2">
          <DialogTitle className="text-sm font-bold flex items-center gap-2">
            <div
              className="flex items-center justify-center w-6 h-6 rounded-md"
              style={{ backgroundColor: `${quadrant.dotColor}20`, color: quadrant.dotColor }}
            >
              {quadrant.icon}
            </div>
            <span>Add to <span style={{ color: quadrant.dotColor }}>{quadrant.label}</span></span>
          </DialogTitle>
          <button
            onClick={onClose}
            className="w-6 h-6 flex items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent transition-all"
          >
            <X size={14} />
          </button>
        </DialogHeader>

        <div className={`flex items-center gap-2 px-3 py-2 rounded-lg border text-[10px] font-semibold mb-4 ${quadrant.accentBg} ${quadrant.accentBorder} ${quadrant.accentText}`}>
          <AlertTriangle size={11} />
          {quadrant.sublabel}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              Title <span className="text-destructive">*</span>
            </label>
            <Input
              ref={titleRef}
              type="text"
              placeholder="What needs to be done?"
              className="h-9 text-xs border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent"
              value={title}
              onChange={(e) => {
                setTitle(e.target.value);
                if (titleError) setTitleError('');
              }}
            />
            {titleError && (
              <p className="text-destructive text-xs font-semibold">⚠️ {titleError}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              Description
            </label>
            <textarea
              rows={3}
              placeholder="Optional details..."
              className="w-full p-2.5 rounded-md border border-border/80 focus:outline-none focus:ring-1 focus:ring-primary bg-transparent text-xs resize-none"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="flex justify-end gap-2 pt-2 border-t border-border/60">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              className="h-8 text-xs font-semibold"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={submitting}
              className="h-8 text-xs font-semibold"
              style={{ backgroundColor: quadrant.dotColor }}
            >
              {submitting ? 'Adding...' : 'Add Task'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// ─── Main Matrix Page ─────────────────────────────────────────────────────────

export default function EisenhowerMatrix() {
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();

  const {
    todos,
    loading,
    error,
    fetchTodos,
    createTodo,
    deleteTodo,
    updateTodoPriority,
    setError,
  } = useTodoStore();

  // On mount, fetch all todos (no per-page limit for the matrix view)
  useEffect(() => {
    fetchTodos();
  }, [fetchTodos]);

  // Drag state
  const [draggingTodo, setDraggingTodo] = useState<Todo | null>(null);
  const [dragOverQuadrant, setDragOverQuadrant] = useState<string | null>(null);

  // Quick add dialog
  const [quickAddQuadrant, setQuickAddQuadrant] = useState<Quadrant | null>(null);

  // Map todos to quadrants
  const getQuadrantTodos = (q: Quadrant): Todo[] =>
    todos.filter(
      (t) => t.importance === q.importance && t.urgency === q.urgency && !t.deleted_at
    );

  // ── Drag handlers ──

  const handleDragStart = (e: React.DragEvent, todo: Todo) => {
    setDraggingTodo(todo);
    e.dataTransfer.effectAllowed = 'move';
    // Add semi-transparent ghost image
    const ghost = e.currentTarget.cloneNode(true) as HTMLElement;
    ghost.style.opacity = '0.6';
    ghost.style.position = 'absolute';
    ghost.style.top = '-1000px';
    document.body.appendChild(ghost);
    e.dataTransfer.setDragImage(ghost, 0, 0);
    setTimeout(() => document.body.removeChild(ghost), 0);
  };

  const handleDragOver = (e: React.DragEvent, quadrantId: string) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverQuadrant(quadrantId);
  };

  const handleDragLeave = () => {
    setDragOverQuadrant(null);
  };

  const handleDrop = async (e: React.DragEvent, quadrant: Quadrant) => {
    e.preventDefault();
    setDragOverQuadrant(null);
    if (!draggingTodo) return;

    // Skip if same quadrant
    if (
      draggingTodo.importance === quadrant.importance &&
      draggingTodo.urgency === quadrant.urgency
    ) {
      setDraggingTodo(null);
      return;
    }

    await updateTodoPriority(draggingTodo, quadrant.importance, quadrant.urgency);
    setDraggingTodo(null);
  };

  const handleQuickAdd = async (
    title: string,
    description: string,
    importance: boolean,
    urgency: boolean
  ): Promise<boolean> => {
    return createTodo(title, description, null, importance, urgency);
  };

  const handleDelete = async (id: string) => {
    await deleteTodo(id);
  };

  const totalTasks = todos.length;
  const doFirstCount = getQuadrantTodos(QUADRANTS[0]).length;

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          {/* Header */}
          <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-background/50 backdrop-blur px-4 lg:px-6">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />

            <div className="flex-1 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span
                className="font-medium hover:text-foreground transition-colors cursor-pointer"
                onClick={() => navigate('/')}
              >
                Workspace
              </span>
              <ChevronRight size={12} className="text-muted-foreground/60" />
              <span className="font-semibold text-foreground">Eisenhower Matrix</span>
            </div>

            <div className="flex items-center gap-2">
              {/* Summary pills */}
              <div className="hidden sm:flex items-center gap-1.5">
                <span className="text-[10px] font-bold bg-muted/60 border border-border/60 px-2 py-0.5 rounded-md text-muted-foreground">
                  {totalTasks} tasks
                </span>
                {doFirstCount > 0 && (
                  <span className="text-[10px] font-bold bg-rose-500/10 border border-rose-500/20 text-rose-500 px-2 py-0.5 rounded-md">
                    {doFirstCount} urgent
                  </span>
                )}
              </div>

              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="h-7 w-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-all"
              >
                {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
              </Button>
            </div>
          </header>

          {/* Main content */}
          <main className="flex flex-col gap-0 p-0 bg-background min-h-[calc(100vh-3.5rem)] overflow-hidden">
            {/* Axis labels */}
            <div className="flex items-center justify-center pt-4 pb-2 px-6">
              <div className="flex items-center gap-2 text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                <span>↑ Important</span>
                <span className="text-border/80">·</span>
                <span>→ Urgent</span>
              </div>
            </div>

            {/* Error */}
            {error && (
              <div className="mx-6 mb-2 bg-destructive/10 border border-destructive/20 text-destructive text-xs font-semibold p-3 rounded-lg flex justify-between items-center">
                <span>⚠️ {error}</span>
                <button onClick={() => setError(null)}>
                  <X size={12} />
                </button>
              </div>
            )}

            {/* Matrix Grid */}
            <div className="flex-1 grid grid-cols-2 grid-rows-2 gap-px bg-border/40 mx-4 mb-4 rounded-xl overflow-hidden border border-border/60 shadow-sm">
              {/* Row labels */}
              {QUADRANTS.map((q) => (
                <QuadrantPanel
                  key={q.id}
                  quadrant={q}
                  todos={loading ? [] : getQuadrantTodos(q)}
                  isDragOver={dragOverQuadrant === q.id}
                  onDragOver={(e) => handleDragOver(e, q.id)}
                  onDragLeave={handleDragLeave}
                  onDrop={handleDrop}
                  onDragStart={handleDragStart}
                  onDelete={handleDelete}
                  onQuickAdd={(quadrant) => setQuickAddQuadrant(quadrant)}
                />
              ))}
            </div>

            {/* Legend Footer */}
            <div className="flex items-center justify-center gap-6 pb-4 flex-wrap px-6">
              {QUADRANTS.map((q) => (
                <div key={q.id} className="flex items-center gap-1.5">
                  <div
                    className="w-2 h-2 rounded-full"
                    style={{ backgroundColor: q.dotColor }}
                  />
                  <span className="text-[10px] font-semibold text-muted-foreground">
                    {q.label}
                  </span>
                </div>
              ))}
            </div>
          </main>
        </SidebarInset>
      </SidebarProvider>

      {/* Quick Add Dialog */}
      <QuickAddDialog
        quadrant={quickAddQuadrant}
        open={!!quickAddQuadrant}
        onClose={() => setQuickAddQuadrant(null)}
        onSubmit={handleQuickAdd}
      />
    </TooltipProvider>
  );
}
