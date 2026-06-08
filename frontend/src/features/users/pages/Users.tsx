import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTheme } from 'next-themes';
import {
  Search,
  Plus,
  Edit2,
  Trash2,
  ChevronLeft,
  ChevronRight,
  X,
  Sun,
  Moon,
  Shield,
  UserCheck,
  UserX,
  ArrowUp,
  ArrowDown,
  User as UserIcon,
  ChevronRight as ChevronRightIcon,
  AlertTriangle,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import {
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table';
import { TooltipProvider } from '@/components/ui/tooltip';
import { AppSidebar } from '@/components/app-sidebar';
import useAuthStore from '../../auth/store';
import useUsersStore, { User } from '../store';
import { createUserSchema, updateUserSchema } from '../../../lib/schemas';
import { usePermission } from '@/hooks/usePermission';

export default function Users() {
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();
  const logout = useAuthStore((state) => state.logout);
  const currentUserEmail = useAuthStore((state) => state.userEmail);
  const activeRole = useAuthStore((state) => state.activeRole);
  const canListUsers = usePermission('user:list');

  const {
    users,
    total,
    page,
    perPage,
    keyword,
    sortBy,
    sortDir,
    loading,
    error,
    editingUser,
    fetchUsers,
    createUser,
    updateUser,
    deleteUser,
    setFilters,
    setPage,
    setEditingUser,
    setError,
  } = useUsersStore();

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [userToDelete, setUserToDelete] = useState<User | null>(null);

  // Form states
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [roles, setRoles] = useState<string[]>(['user']);
  const [userActiveRole, setUserActiveRole] = useState('user');

  // Form validation errors
  const [validationErrors, setValidationErrors] = useState<{
    email?: string;
    password?: string;
    roles?: string;
    activeRole?: string;
  }>({});

  // Local debounced keyword state
  const [searchKeyword, setSearchKeyword] = useState(keyword);

  // Redirect if not authorized
  useEffect(() => {
    if (!canListUsers) {
      navigate('/');
    }
  }, [canListUsers, navigate]);

  useEffect(() => {
    if (canListUsers) {
      fetchUsers();
    }
  }, [page, keyword, sortBy, sortDir, fetchUsers, canListUsers]);

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

  const resetForm = () => {
    setEmail('');
    setPassword('');
    setRoles(['user']);
    setUserActiveRole('user');
    setEditingUser(null);
    setValidationErrors({});
  };

  const handleRoleCheckboxChange = (role: string, checked: boolean) => {
    if (checked) {
      if (!roles.includes(role)) {
        setRoles([...roles, role]);
      }
    } else {
      setRoles(roles.filter((r) => r !== role));
    }
  };

  const handleCreateUser = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setValidationErrors({});
    setError(null);

    const validation = createUserSchema.safeParse({
      email,
      password,
      activeRole: userActiveRole,
      roles,
    });

    if (!validation.success) {
      const fieldErrors: typeof validationErrors = {};
      validation.error.issues.forEach((issue) => {
        const path = issue.path[0] as keyof typeof validationErrors;
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const success = await createUser({
      email,
      password,
      active_role: userActiveRole,
      roles,
    });

    if (success) {
      setIsCreateOpen(false);
      resetForm();
    }
  };

  const handleEditClick = (user: User) => {
    setEditingUser(user);
    setEmail(user.email);
    setPassword('');
    setRoles(user.roles);
    setUserActiveRole(user.active_role);
    setIsEditOpen(true);
  };

  const handleUpdateUser = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingUser) return;
    setValidationErrors({});
    setError(null);

    const validation = updateUserSchema.safeParse({
      email,
      password: password || undefined,
      activeRole: userActiveRole,
      roles,
    });

    if (!validation.success) {
      const fieldErrors: typeof validationErrors = {};
      validation.error.issues.forEach((issue) => {
        const path = issue.path[0] as keyof typeof validationErrors;
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const payload: any = {
      email,
      active_role: userActiveRole,
      roles,
    };
    if (password) {
      payload.password = password;
    }

    const success = await updateUser(editingUser.id, payload);
    if (success) {
      setIsEditOpen(false);
      resetForm();
    }
  };

  const handleDeleteClick = (user: User) => {
    if (user.email === currentUserEmail) {
      return; // UI guard
    }
    setUserToDelete(user);
  };

  const handleDeleteConfirm = async () => {
    if (userToDelete) {
      const success = await deleteUser(userToDelete.id);
      if (success) {
        setUserToDelete(null);
      }
    }
  };

  const totalPages = Math.max(1, Math.ceil(total / perPage));

  if (!canListUsers) {
    return null;
  }

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          {/* Header */}
          <header className="flex h-14 shrink-0 items-center justify-between gap-4 border-b bg-background/50 backdrop-blur px-4 lg:px-6">
            <div className="flex items-center gap-2">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <span
                  className="font-medium hover:text-foreground transition-colors cursor-pointer"
                  onClick={() => navigate('/')}
                >
                  Workspace
                </span>
                <ChevronRightIcon size={12} className="text-muted-foreground/60" />
                <span className="font-semibold text-foreground">User Management</span>
              </div>
            </div>

            {/* Header Actions */}
            <div className="flex items-center gap-2">
              <div className="relative w-40 sm:w-60">
                <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60 size-3.5" />
                <Input
                  type="text"
                  placeholder="Search email..."
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
                className="h-7 px-2.5 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 transition-all rounded-md flex items-center gap-1"
              >
                <Plus size={14} /> New User
              </Button>
            </div>
          </header>

          {/* Main Content */}
          <main className="flex flex-col gap-6 p-6 lg:p-8 bg-background min-h-[calc(100vh-3.5rem)]">
            {/* Error Banner */}
            {error && (
              <div className="bg-destructive/10 border border-destructive/20 text-destructive text-xs font-semibold p-4 rounded-lg flex justify-between items-center shadow-sm">
                <span>⚠️ {error}</span>
                <Button
                  variant="ghost"
                  onClick={() => setError(null)}
                  className="h-6 w-6 p-0 text-destructive hover:bg-destructive/10"
                >
                  <X size={14} />
                </Button>
              </div>
            )}

            {/* Toolbar Filters / Stats */}
            <section className="flex flex-col md:flex-row gap-4 justify-between items-stretch md:items-center bg-card/25 border border-border p-3 rounded-lg">
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold text-muted-foreground uppercase tracking-wider">
                  Filters & Sorting
                </span>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <select
                  className="h-8 px-2.5 rounded-md border border-border/80 bg-background text-xs font-semibold focus:outline-none focus:ring-1 focus:ring-primary text-foreground"
                  value={sortBy}
                  onChange={(e) => setFilters({ sortBy: e.target.value })}
                >
                  <option value="created_at">Created Time</option>
                  <option value="email">Email</option>
                  <option value="active_role">Active Role</option>
                </select>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setFilters({ sortDir: sortDir === 'asc' ? 'desc' : 'asc' })}
                  className="h-8 text-xs font-semibold border-border/80 flex items-center gap-1 hover:bg-accent"
                >
                  Sort: {sortDir === 'asc' ? 'Ascending' : 'Descending'}
                  {sortDir === 'asc' ? <ArrowUp className="size-3.5" /> : <ArrowDown className="size-3.5" />}
                </Button>
              </div>
            </section>

            {/* Users Table */}
            <Card className="border border-border bg-card/25 shadow-none rounded-lg overflow-hidden">
              <CardContent className="p-0">
                {loading && users.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-20 gap-3">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                    <span className="text-xs text-muted-foreground">Loading users...</span>
                  </div>
                ) : users.length === 0 ? (
                  <div className="text-center py-16">
                    <UserIcon className="mx-auto size-8 text-muted-foreground/60 mb-2" />
                    <h3 className="text-sm font-bold text-foreground">No users found</h3>
                    <p className="text-xs text-muted-foreground mt-1">
                      No registered users matched the active filters.
                    </p>
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow className="border-b hover:bg-transparent">
                        <TableHead className="font-bold text-[10px] uppercase tracking-wider py-3 px-5 text-muted-foreground">
                          Email
                        </TableHead>
                        <TableHead className="font-bold text-[10px] uppercase tracking-wider py-3 px-5 text-muted-foreground">
                          Active Role
                        </TableHead>
                        <TableHead className="font-bold text-[10px] uppercase tracking-wider py-3 px-5 text-muted-foreground">
                          Assigned Roles
                        </TableHead>
                        <TableHead className="font-bold text-[10px] uppercase tracking-wider py-3 px-5 text-muted-foreground text-right">
                          Actions
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {users.map((user) => {
                        const isSelf = user.email === currentUserEmail;
                        return (
                          <TableRow
                            key={user.id}
                            className="border-b transition-colors hover:bg-muted/30"
                          >
                            <TableCell className="py-3 px-5 font-medium text-xs text-foreground">
                              <div className="flex items-center gap-2">
                                <span>{user.email}</span>
                                {isSelf && (
                                  <Badge className="bg-primary/10 border-primary/20 border text-primary text-[9px] font-extrabold uppercase shadow-none py-0.5 px-1.5 rounded">
                                    You
                                  </Badge>
                                )}
                              </div>
                            </TableCell>
                            <TableCell className="py-3 px-5 text-xs">
                              <Badge
                                className={`text-[9px] uppercase tracking-wider font-extrabold py-0.5 px-2 rounded border shadow-none ${
                                  user.active_role === 'admin'
                                    ? 'bg-indigo-500/5 border-indigo-500/20 text-indigo-500'
                                    : 'bg-muted border-border/80 text-muted-foreground'
                                }`}
                              >
                                {user.active_role}
                              </Badge>
                            </TableCell>
                            <TableCell className="py-3 px-5 text-xs">
                              <div className="flex flex-wrap gap-1">
                                {user.roles.map((role) => (
                                  <Badge
                                    key={role}
                                    variant="outline"
                                    className="text-[9px] font-semibold py-0.5 px-1.5 rounded border border-border/60 text-muted-foreground"
                                  >
                                    {role}
                                  </Badge>
                                ))}
                              </div>
                            </TableCell>
                            <TableCell className="py-3 px-5 text-right">
                              <div className="flex gap-1.5 justify-end">
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleEditClick(user)}
                                  className="h-7 w-7 rounded-md text-muted-foreground hover:text-foreground hover:bg-accent border-none"
                                >
                                  <Edit2 size={12} />
                                </Button>
                                <Button
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => handleDeleteClick(user)}
                                  disabled={isSelf}
                                  className={`h-7 w-7 rounded-md border-none ${
                                    isSelf
                                      ? 'text-muted-foreground/30 cursor-not-allowed'
                                      : 'text-muted-foreground hover:text-destructive hover:bg-destructive/10'
                                  }`}
                                  title={isSelf ? 'You cannot delete your own account' : 'Delete user'}
                                >
                                  <Trash2 size={12} />
                                </Button>
                              </div>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <section className="flex justify-center items-center gap-2 mt-4">
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

      {/* Create User Dialog */}
      <Dialog
        open={isCreateOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsCreateOpen(false);
            resetForm();
          }
        }}
      >
        <DialogContent
          className="bg-card w-full max-w-lg border border-border p-6 rounded-lg shadow-lg overflow-y-auto max-h-[90vh]"
          showCloseButton={false}
        >
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-sm font-bold uppercase tracking-wider flex items-center gap-2">
              <Plus size={16} className="text-primary" /> Create User
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

          <form onSubmit={handleCreateUser} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Email Address
              </label>
              <Input
                type="email"
                required
                placeholder="user@example.com"
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  if (validationErrors.email) {
                    setValidationErrors((prev) => ({ ...prev, email: undefined }));
                  }
                }}
              />
              {validationErrors.email && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.email}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Password
              </label>
              <Input
                type="password"
                required
                placeholder="••••••••"
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  if (validationErrors.password) {
                    setValidationErrors((prev) => ({ ...prev, password: undefined }));
                  }
                }}
              />
              {validationErrors.password && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.password}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Assigned Roles
              </label>
              <div className="flex gap-4 p-2.5 rounded-md border border-border/80 bg-background/20">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="role-user"
                    checked={roles.includes('user')}
                    onCheckedChange={(checked) => handleRoleCheckboxChange('user', !!checked)}
                  />
                  <label htmlFor="role-user" className="text-xs font-medium cursor-pointer">
                    User
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="role-admin"
                    checked={roles.includes('admin')}
                    onCheckedChange={(checked) => handleRoleCheckboxChange('admin', !!checked)}
                  />
                  <label htmlFor="role-admin" className="text-xs font-medium cursor-pointer">
                    Admin
                  </label>
                </div>
              </div>
              {validationErrors.roles && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.roles}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Active Role
              </label>
              <Select
                value={userActiveRole}
                onValueChange={(val) => {
                  setUserActiveRole(val);
                  if (validationErrors.activeRole) {
                    setValidationErrors((prev) => ({ ...prev, activeRole: undefined }));
                  }
                }}
              >
                <SelectTrigger className="w-full h-9 text-xs rounded-md bg-transparent border-border/80 text-foreground">
                  <SelectValue placeholder="Select active role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
              {validationErrors.activeRole && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.activeRole}
                </p>
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
              <Button
                type="submit"
                className="h-8 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-md"
              >
                Create User
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit User Dialog */}
      <Dialog
        open={isEditOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsEditOpen(false);
            resetForm();
          }
        }}
      >
        <DialogContent
          className="bg-card w-full max-w-lg border border-border p-6 rounded-lg shadow-lg overflow-y-auto max-h-[90vh]"
          showCloseButton={false}
        >
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-sm font-bold uppercase tracking-wider flex items-center gap-2">
              <Edit2 size={14} className="text-primary" /> Edit User
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

          <form onSubmit={handleUpdateUser} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Email Address
              </label>
              <Input
                type="email"
                required
                placeholder="user@example.com"
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  if (validationErrors.email) {
                    setValidationErrors((prev) => ({ ...prev, email: undefined }));
                  }
                }}
              />
              {validationErrors.email && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.email}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Password <span className="text-[9px] text-muted-foreground lowercase font-normal">(Leave blank to keep unchanged)</span>
              </label>
              <Input
                type="password"
                placeholder="••••••••"
                className="w-full h-9 rounded-md border border-border/80 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0 bg-transparent text-xs"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  if (validationErrors.password) {
                    setValidationErrors((prev) => ({ ...prev, password: undefined }));
                  }
                }}
              />
              {validationErrors.password && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.password}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Assigned Roles
              </label>
              <div className="flex gap-4 p-2.5 rounded-md border border-border/80 bg-background/20">
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="edit-role-user"
                    checked={roles.includes('user')}
                    onCheckedChange={(checked) => handleRoleCheckboxChange('user', !!checked)}
                  />
                  <label htmlFor="edit-role-user" className="text-xs font-medium cursor-pointer">
                    User
                  </label>
                </div>
                <div className="flex items-center gap-2">
                  <Checkbox
                    id="edit-role-admin"
                    checked={roles.includes('admin')}
                    onCheckedChange={(checked) => handleRoleCheckboxChange('admin', !!checked)}
                  />
                  <label htmlFor="edit-role-admin" className="text-xs font-medium cursor-pointer">
                    Admin
                  </label>
                </div>
              </div>
              {validationErrors.roles && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.roles}
                </p>
              )}
            </div>

            <div className="space-y-1.5">
              <label className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
                Active Role
              </label>
              <Select
                value={userActiveRole}
                onValueChange={(val) => {
                  setUserActiveRole(val);
                  if (validationErrors.activeRole) {
                    setValidationErrors((prev) => ({ ...prev, activeRole: undefined }));
                  }
                }}
              >
                <SelectTrigger className="w-full h-9 text-xs rounded-md bg-transparent border-border/80 text-foreground">
                  <SelectValue placeholder="Select active role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="admin">Admin</SelectItem>
                </SelectContent>
              </Select>
              {validationErrors.activeRole && (
                <p className="text-destructive text-xs font-semibold mt-1">
                  ⚠️ {validationErrors.activeRole}
                </p>
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
              <Button
                type="submit"
                className="h-8 text-xs font-semibold bg-primary text-primary-foreground hover:bg-primary/90 rounded-md"
              >
                Save Changes
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete User Dialog */}
      <Dialog
        open={!!userToDelete}
        onOpenChange={(open) => {
          if (!open) setUserToDelete(null);
        }}
      >
        <DialogContent className="bg-card w-full max-w-md border border-border p-6 rounded-lg shadow-lg" showCloseButton={false}>
          <DialogHeader className="border-b border-border/60 pb-3 mb-5 flex flex-row justify-between items-center gap-2">
            <DialogTitle className="text-sm font-bold uppercase tracking-wider flex items-center gap-2 text-foreground">
              <UserX size={16} className="text-destructive" /> Delete User
            </DialogTitle>
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setUserToDelete(null)}
              className="h-7 w-7 rounded-md hover:bg-accent border-none text-muted-foreground hover:text-foreground"
            >
              <X size={14} />
            </Button>
          </DialogHeader>

          <div className="space-y-4">
            <div className="flex gap-3 p-3 rounded-md bg-destructive/10 border border-destructive/20 text-destructive text-xs leading-relaxed font-semibold">
              <AlertTriangle className="size-4 shrink-0" />
              <span>
                Warning: This will permanently delete the account of{' '}
                <span className="font-extrabold">{userToDelete?.email}</span>. This action is irreversible.
              </span>
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t border-border/60">
              <Button
                type="button"
                variant="outline"
                onClick={() => setUserToDelete(null)}
                className="h-8 text-xs font-semibold"
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                onClick={handleDeleteConfirm}
                className="h-8 text-xs font-semibold px-4 rounded-md"
              >
                Delete User
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </TooltipProvider>
  );
}
