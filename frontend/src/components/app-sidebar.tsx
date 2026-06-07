import * as React from "react"
import { useNavigate } from "react-router-dom"
import {
  LayoutDashboard,
  Clock,
  Activity,
  CheckCircle,
  LogOut,
  CheckSquare,
  Sparkles
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
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import useTodoStore from "@/features/todos/store"
import useAuthStore from "@/features/auth/store"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const navigate = useNavigate()
  const { todos, status, setFilters } = useTodoStore()
  const logout = useAuthStore((state) => state.logout)

  // Extract email from access token
  const email = React.useMemo(() => {
    try {
      const token = localStorage.getItem("access_token")
      if (!token) return "user@example.com"
      const payload = JSON.parse(atob(token.split(".")[1]))
      return payload.email || "user@example.com"
    } catch {
      return "user@example.com"
    }
  }, [])

  const username = React.useMemo(() => {
    return email.split("@")[0] || "User"
  }, [email])

  // Count items for badges
  const counts = React.useMemo(() => {
    const total = todos.length
    const pending = todos.filter((t) => t.status === "pending").length
    const inProgress = todos.filter((t) => t.status === "in_progress").length
    const done = todos.filter((t) => t.status === "done").length
    return { total, pending, inProgress, done }
  }, [todos])

  const handleLogout = () => {
    logout()
    navigate("/login")
  }

  const navItems = [
    {
      title: "All Tasks",
      value: "",
      icon: LayoutDashboard,
      count: counts.total,
    },
    {
      title: "Pending",
      value: "pending",
      icon: Clock,
      count: counts.pending,
      color: "text-amber-500",
    },
    {
      title: "In Progress",
      value: "in_progress",
      icon: Activity,
      count: counts.inProgress,
      color: "text-primary",
    },
    {
      title: "Completed",
      value: "done",
      icon: CheckCircle,
      count: counts.done,
      color: "text-emerald-500",
    },
  ]

  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader className="border-b border-sidebar-border px-4 py-3">
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
        <SidebarGroup>
          <SidebarGroupLabel className="px-2 font-bold uppercase tracking-wider text-[10px]">
            Workspace
          </SidebarGroupLabel>
          <SidebarMenu className="gap-0.5">
            {navItems.map((item) => {
              const isActive = status === item.value
              const Icon = item.icon
              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    isActive={isActive}
                    onClick={() => setFilters({ status: item.value })}
                    className="h-9 font-medium transition-all group-data-[collapsible=icon]:p-2!"
                    tooltip={item.title}
                  >
                    <Icon className={`size-4 ${item.color || ""}`} />
                    <span className="text-sm font-medium">{item.title}</span>
                    <span className="ml-auto rounded-full bg-slate-100 dark:bg-slate-800 text-[10px] px-2 py-0.5 font-bold tabular-nums text-slate-500 group-data-[collapsible=icon]:hidden">
                      {item.count}
                    </span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )
            })}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter className="border-t border-sidebar-border p-2">
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
                    </div>
                  </div>
                </DropdownMenuLabel>
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
  )
}
