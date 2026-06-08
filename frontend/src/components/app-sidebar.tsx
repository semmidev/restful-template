import * as React from "react"
import { useNavigate, useLocation, Link } from "react-router-dom"
import {
  LayoutDashboard,
  LayoutGrid,
  LogOut,
  CheckSquare,
  Shield,
  Check,
  User,
  CheckCircle2,
} from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarGroup,
  SidebarGroupLabel,
} from "@/components/ui/sidebar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuSub,
  DropdownMenuSubTrigger,
  DropdownMenuSubContent,
  DropdownMenuPortal,
} from "@/components/ui/dropdown-menu"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import useAuthStore from "@/features/auth/store"
import { usePermission } from "@/hooks/usePermission"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const navigate = useNavigate()
  const logout = useAuthStore((state) => state.logout)
  const switchRole = useAuthStore((state) => state.switchRole)

  const userEmail = useAuthStore((state) => state.userEmail)
  const email = userEmail || "user@example.com"
  const activeRole = useAuthStore((state) => state.activeRole) || "user"
  const roles = useAuthStore((state) => state.roles) || ["user"]
  const canListUsers = usePermission("user:list")

  const [roleSwitchDialog, setRoleSwitchDialog] = React.useState<{
    open: boolean
    newRole: string
  }>({ open: false, newRole: "" })
  const [isSwitching, setIsSwitching] = React.useState(false)

  const username = React.useMemo(() => {
    return email.split("@")[0] || "User"
  }, [email])

  const handleLogout = () => {
    logout()
    navigate("/login")
  }

  const handleSwitchRole = async (role: string) => {
    if (role === activeRole) return
    setIsSwitching(true)
    const result = await switchRole(role)
    setIsSwitching(false)
    if (result.success) {
      setRoleSwitchDialog({ open: true, newRole: role })
    }
  }

  const location = useLocation()
  const isDashboardActive = location.pathname === "/"
  const isTasksActive = location.pathname === "/todos"
  const isMatrixActive = location.pathname === "/matrix"
  const isUsersActive = location.pathname === "/users"

  return (
    <>
      <Sidebar collapsible="icon" {...props}>
        <SidebarHeader className="border-b border-sidebar-border px-4 py-3 group-data-[state=collapsed]:hidden">
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton size="lg" asChild>
                <div className="flex items-center gap-2.5">
                  <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
                    <CheckSquare className="size-4" />
                  </div>
                  <div className="flex flex-col gap-0.5 text-left leading-none">
                    <span className="text-sm font-extrabold tracking-tight text-sidebar-foreground">
                      TodoApp
                    </span>
                    <span className="text-[10px] text-muted-foreground font-semibold">
                      v1.0.0
                    </span>
                  </div>
                </div>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
          {/* General workspace navigation */}
          <SidebarGroup>
            <SidebarGroupLabel className="px-2 font-bold uppercase tracking-wider text-[10px]">
              Workspace
            </SidebarGroupLabel>
            <SidebarMenu className="gap-1">
              <SidebarMenuItem>
                <SidebarMenuButton
                  isActive={isDashboardActive}
                  className="h-8.5 font-medium transition-all group-data-[collapsible=icon]:p-2! data-[active=true]:bg-accent/80 data-[active=true]:text-foreground rounded-lg"
                  tooltip="Dashboard"
                  asChild
                >
                  <Link to="/" className="flex items-center gap-2.5 px-2">
                    <LayoutDashboard className={`size-4 transition-colors ${isDashboardActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'}`} />
                    <span className={`text-xs font-semibold ${isDashboardActive ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}>Dashboard</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton
                  isActive={isTasksActive}
                  className="h-8.5 font-medium transition-all group-data-[collapsible=icon]:p-2! data-[active=true]:bg-accent/80 data-[active=true]:text-foreground rounded-lg"
                  tooltip="All Tasks"
                  asChild
                >
                  <Link to="/todos" className="flex items-center gap-2.5 px-2">
                    <CheckSquare className={`size-4 transition-colors ${isTasksActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'}`} />
                    <span className={`text-xs font-semibold ${isTasksActive ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}>All Tasks</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton
                  isActive={isMatrixActive}
                  className="h-8.5 font-medium transition-all group-data-[collapsible=icon]:p-2! data-[active=true]:bg-accent/80 data-[active=true]:text-foreground rounded-lg"
                  tooltip="Eisenhower Matrix"
                  asChild
                >
                  <Link to="/matrix" className="flex items-center gap-2.5 px-2">
                    <LayoutGrid className={`size-4 transition-colors ${isMatrixActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'}`} />
                    <span className={`text-xs font-semibold ${isMatrixActive ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}>Eisenhower Matrix</span>
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroup>

          {/* Administration section — only visible to admin users */}
          {canListUsers && (
            <SidebarGroup>
              <SidebarGroupLabel className="px-2 font-bold uppercase tracking-wider text-[10px]">
                Administration
              </SidebarGroupLabel>
              <SidebarMenu className="gap-1">
                <SidebarMenuItem>
                  <SidebarMenuButton
                    isActive={isUsersActive}
                    className="h-8.5 font-medium transition-all group-data-[collapsible=icon]:p-2! data-[active=true]:bg-accent/80 data-[active=true]:text-foreground rounded-lg"
                    tooltip="User Management"
                    asChild
                  >
                    <Link to="/users" className="flex items-center gap-2.5 px-2">
                      <User className={`size-4 transition-colors ${isUsersActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'}`} />
                      <span className={`text-xs font-semibold ${isUsersActive ? 'text-foreground' : 'text-muted-foreground group-hover:text-foreground'}`}>User Management</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroup>
          )}
        </SidebarContent>

        <SidebarFooter className="border-t border-sidebar-border p-2 group-data-[state=collapsed]:hidden">
          <SidebarMenu>
            <SidebarMenuItem>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  >
                    <Avatar className="h-8 w-8 rounded-lg">
                      <AvatarFallback className="rounded-lg bg-primary/10 text-primary font-bold text-xs uppercase">
                        {username.slice(0, 2)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                      <span className="truncate font-semibold text-slate-800 dark:text-slate-200">
                        {username}
                      </span>
                      <span className="truncate text-xs text-muted-foreground font-medium">
                        {email}
                      </span>
                      <span className="truncate text-[9px] text-primary/90 font-extrabold uppercase tracking-wider mt-0.5">
                        {activeRole}
                      </span>
                    </div>
                  </SidebarMenuButton>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  className="w-[var(--radix-dropdown-menu-trigger-width)] min-w-56 rounded-lg"
                  side="right"
                  align="end"
                  sideOffset={4}
                >
                  <DropdownMenuLabel className="p-0 font-normal">
                    <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                      <Avatar className="h-8 w-8 rounded-lg">
                        <AvatarFallback className="rounded-lg bg-primary/10 text-primary font-bold text-xs uppercase">
                          {username.slice(0, 2)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="grid flex-1 text-left text-sm leading-tight">
                        <span className="truncate font-semibold">{username}</span>
                        <span className="truncate text-xs text-muted-foreground font-medium">
                          {email}
                        </span>
                        <span className="truncate text-[9px] text-primary/90 font-extrabold uppercase tracking-wider mt-0.5">
                          {activeRole}
                        </span>
                      </div>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  {/* Switch Role submenu — only shown when user has multiple roles */}
                  {roles.length > 1 && (
                    <DropdownMenuSub>
                      <DropdownMenuSubTrigger className="cursor-pointer font-semibold text-xs py-1.5">
                        <Shield className="size-3.5 mr-2 text-muted-foreground" />
                        <span>Switch Role</span>
                      </DropdownMenuSubTrigger>
                      <DropdownMenuPortal>
                        <DropdownMenuSubContent className="w-48">
                          {roles.map((r) => (
                            <DropdownMenuItem
                              key={r}
                              onClick={() => handleSwitchRole(r)}
                              disabled={r === activeRole || isSwitching}
                              className="flex items-center justify-between font-medium cursor-pointer"
                            >
                              <div className="flex items-center gap-2">
                                {r === "admin" ? (
                                  <Shield className="size-3.5 text-indigo-500" />
                                ) : (
                                  <User className="size-3.5 text-slate-500" />
                                )}
                                <span className="capitalize text-xs">{r}</span>
                              </div>
                              {r === activeRole && <Check className="size-3.5 text-primary" />}
                            </DropdownMenuItem>
                          ))}
                        </DropdownMenuSubContent>
                      </DropdownMenuPortal>
                    </DropdownMenuSub>
                  )}
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={handleLogout} className="text-destructive font-semibold">
                    <LogOut className="size-4 mr-2" />
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </Sidebar>

      {/* Role switch confirmation dialog */}
      <Dialog
        open={roleSwitchDialog.open}
        onOpenChange={(open) => setRoleSwitchDialog((prev) => ({ ...prev, open }))}
      >
        <DialogContent showCloseButton={false} className="sm:max-w-xs">
          <div className="flex flex-col items-center gap-4 py-2">
            {/* Icon badge */}
            <div className={`flex h-14 w-14 items-center justify-center rounded-2xl ${
              roleSwitchDialog.newRole === "admin"
                ? "bg-indigo-500/10 ring-1 ring-indigo-500/20"
                : "bg-primary/10 ring-1 ring-primary/20"
            }`}>
              {roleSwitchDialog.newRole === "admin" ? (
                <Shield className="size-7 text-indigo-500" />
              ) : (
                <User className="size-7 text-primary" />
              )}
            </div>

            {/* Text */}
            <DialogHeader className="items-center text-center gap-1 w-full">
              <DialogTitle className="text-base font-semibold tracking-tight">
                Role Switched
              </DialogTitle>
              <DialogDescription className="text-sm text-muted-foreground leading-relaxed">
                You are now signed in as{" "}
                <span className="font-semibold text-foreground capitalize">
                  {roleSwitchDialog.newRole}
                </span>
                . Your permissions have been updated.
              </DialogDescription>
            </DialogHeader>

            {/* CTA button — inline, not in a footer bar */}
            <Button
              className="w-full mt-1"
              onClick={() => setRoleSwitchDialog({ open: false, newRole: "" })}
            >
              <CheckCircle2 className="size-4 mr-1.5" />
              Got it
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
