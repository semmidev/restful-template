import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTheme } from 'next-themes';
import {
  CheckCircle,
  Clock,
  Activity,
  FileText,
  Sun,
  Moon,
  TrendingUp,
  BarChart2,
  PieChart as PieChartIcon,
  ChevronRight
} from 'lucide-react';
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  BarChart,
  Bar
} from 'recharts';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import {
  SidebarProvider,
  SidebarTrigger,
  SidebarInset,
} from '@/components/ui/sidebar';
import { TooltipProvider } from '@/components/ui/tooltip';
import { AppSidebar } from '@/components/app-sidebar';

import useTodoStore from '../store';

// Custom Linear-style Tooltip for Recharts
const ChartTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    return (
      <div className="bg-card/95 border border-border/80 px-3 py-2 rounded-lg shadow-md text-xs backdrop-blur-sm">
        <p className="font-semibold text-muted-foreground mb-1">{label}</p>
        <div className="flex flex-col gap-1">
          {payload.map((item: any, index: number) => (
            <div key={index} className="flex items-center gap-2 font-medium">
              <span className="size-2 rounded-full" style={{ backgroundColor: item.color || item.fill }}></span>
              <span className="text-foreground">{item.name}:</span>
              <span className="text-foreground font-bold">{item.value}</span>
            </div>
          ))}
        </div>
      </div>
    );
  }
  return null;
};

export default function Dashboard() {
  const navigate = useNavigate();
  const { theme, setTheme } = useTheme();

  const {
    stats,
    statsLoading,
    statsError,
    fetchStats
  } = useTodoStore();

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  // Color mapping based on Linear UI Palette
  const pieData = stats
    ? [
        { name: 'Pending', value: stats.pending, color: '#f59e0b' },      // Amber
        { name: 'In Progress', value: stats.in_progress, color: 'hsl(var(--primary))' }, // Linear Purple
        { name: 'Completed', value: stats.completed, color: '#10b981' },    // Emerald
      ].filter((d) => d.value > 0)
    : [];

  const hasData = stats && stats.total > 0;

  return (
    <TooltipProvider>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          {/* Header */}
          <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-background/50 backdrop-blur px-4 lg:px-6">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />

            <div className="flex-1 flex items-center gap-1.5 text-sm text-muted-foreground">
              <span className="font-medium hover:text-foreground transition-colors cursor-pointer hidden sm:inline" onClick={() => navigate('/')}>Workspace</span>
              <ChevronRight size={14} className="text-muted-foreground/60 hidden sm:inline" />
              <span className="font-semibold text-foreground">Dashboard</span>
            </div>

            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
                className="h-8 w-8 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/60 transition-all"
                title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
              >
                {theme === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate('/todos')}
                className="h-7 px-2.5 text-xs font-semibold hover:bg-accent hover:text-accent-foreground border-border/80 transition-all rounded-md"
              >
                Manage Tasks
              </Button>
            </div>
          </header>

          {/* Main Dashboard Content */}
          <main className="flex flex-col gap-6 p-6 lg:p-8 bg-background min-h-[calc(100vh-3.5rem)] relative overflow-hidden">
            {/* Ambient decorative glow */}
            <div className="absolute top-0 right-0 w-[300px] h-[300px] bg-primary/5 rounded-full blur-[80px] pointer-events-none z-0" />

            {statsError && (
              <div className="bg-destructive/10 border border-destructive/20 text-destructive text-xs font-semibold p-4 rounded-lg shadow-sm relative z-10">
                ⚠️ {statsError}
              </div>
            )}

            {/* Metrics cards */}
            <section className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 relative z-10">
              <Card className="border border-border bg-card/25 shadow-none rounded-lg p-5 hover:border-border/80 transition-colors">
                <div className="flex items-center justify-between text-muted-foreground mb-3">
                  <span className="text-xs font-bold uppercase tracking-wider">Total Tasks</span>
                  <FileText size={15} />
                </div>
                <div className="text-2xl font-bold tracking-tight">
                  {statsLoading ? '...' : stats?.total ?? 0}
                </div>
                <div className="text-xs text-muted-foreground mt-1">Accumulated tasks in workspace</div>
              </Card>

              <Card className="border border-border bg-card/25 shadow-none rounded-lg p-5 hover:border-border/80 transition-colors">
                <div className="flex items-center justify-between text-muted-foreground mb-3">
                  <span className="text-xs font-bold uppercase tracking-wider">Pending</span>
                  <Clock size={15} />
                </div>
                <div className="text-2xl font-bold tracking-tight text-amber-500">
                  {statsLoading ? '...' : stats?.pending ?? 0}
                </div>
                <div className="text-xs text-muted-foreground mt-1">Task items yet to be started</div>
              </Card>

              <Card className="border border-border bg-card/25 shadow-none rounded-lg p-5 hover:border-border/80 transition-colors">
                <div className="flex items-center justify-between text-muted-foreground mb-3">
                  <span className="text-xs font-bold uppercase tracking-wider">In Progress</span>
                  <Activity size={15} />
                </div>
                <div className="text-2xl font-bold tracking-tight text-primary">
                  {statsLoading ? '...' : stats?.in_progress ?? 0}
                </div>
                <div className="text-xs text-muted-foreground mt-1">Items currently being processed</div>
              </Card>

              <Card className="border border-border bg-card/25 shadow-none rounded-lg p-5 hover:border-border/80 transition-colors">
                <div className="flex items-center justify-between text-muted-foreground mb-3">
                  <span className="text-xs font-bold uppercase tracking-wider">Completion Rate</span>
                  <CheckCircle size={15} />
                </div>
                <div className="flex items-baseline gap-2">
                  <div className="text-2xl font-bold tracking-tight text-emerald-500">
                    {statsLoading ? '...' : `${stats?.completion_rate ?? 0}%`}
                  </div>
                  <div className="flex-1 bg-muted/80 h-1 rounded-full overflow-hidden self-center max-w-[100px]">
                    <div
                      className="bg-emerald-500 h-full rounded-full transition-all duration-500"
                      style={{ width: `${stats?.completion_rate ?? 0}%` }}
                    ></div>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground mt-1">Ratio of completed tasks</div>
              </Card>
            </section>

            {/* Charts Section */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 relative z-10">
              {/* Pie Chart: Status Breakdown */}
              <Card className="lg:col-span-1 border border-border bg-card/25 shadow-none rounded-lg p-5 flex flex-col justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <PieChartIcon className="size-4 text-primary" />
                    <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">Status Distribution</h3>
                  </div>
                  <p className="text-xs text-muted-foreground">Representation of task volume by state</p>
                </div>

                <div className="flex flex-col items-center justify-center my-6 min-h-[220px]">
                  {statsLoading ? (
                    <div className="text-muted-foreground text-xs font-semibold animate-pulse">Loading charts...</div>
                  ) : !hasData ? (
                    <div className="text-center py-6">
                      <div className="size-16 mx-auto flex items-center justify-center rounded-full border border-dashed border-border mb-3">
                        <PieChartIcon size={20} className="text-muted-foreground" />
                      </div>
                      <p className="text-xs font-bold text-muted-foreground">No tasks to display</p>
                    </div>
                  ) : (
                    <div className="w-full h-[200px] relative">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={pieData}
                            cx="50%"
                            cy="50%"
                            innerRadius={55}
                            outerRadius={75}
                            paddingAngle={3}
                            dataKey="value"
                          >
                            {pieData.map((entry, index) => (
                              <Cell key={`cell-${index}`} fill={entry.color} />
                            ))}
                          </Pie>
                          <Tooltip content={<ChartTooltip />} />
                        </PieChart>
                      </ResponsiveContainer>
                      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-center">
                        <span className="text-xs font-bold uppercase tracking-wider text-muted-foreground">Total</span>
                        <div className="text-lg font-black text-foreground leading-none">{stats?.total}</div>
                      </div>
                    </div>
                  )}
                </div>

                {hasData && (
                  <div className="flex flex-wrap gap-x-4 gap-y-1 justify-center border-t border-border/40 pt-4">
                    {pieData.map((item, idx) => (
                      <div key={idx} className="flex items-center gap-1.5 text-xs font-medium">
                        <span className="size-2 rounded-full" style={{ backgroundColor: item.color }}></span>
                        <span className="text-muted-foreground">{item.name}</span>
                        <span className="text-foreground font-bold">{stats ? (stats as any)[item.name.toLowerCase().replace(' ', '_')] : 0}</span>
                      </div>
                    ))}
                  </div>
                )}
              </Card>

              {/* Area Chart: Task Trend Timeline */}
              <Card className="lg:col-span-2 border border-border bg-card/25 shadow-none rounded-lg p-5 flex flex-col justify-between">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <TrendingUp className="size-4 text-primary" />
                    <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">Task Velocity</h3>
                  </div>
                  <p className="text-xs text-muted-foreground">Daily creations vs completions over the last 7 days</p>
                </div>

                <div className="my-6 min-h-[220px] flex items-center justify-center">
                  {statsLoading ? (
                    <div className="text-muted-foreground text-xs font-semibold animate-pulse">Loading timeline data...</div>
                  ) : !hasData ? (
                    <div className="text-center py-6">
                      <div className="size-10 rounded-full border border-border flex items-center justify-center text-muted-foreground mb-3 mx-auto">
                        <TrendingUp size={16} />
                      </div>
                      <p className="text-xs font-bold text-muted-foreground">No recent activity detected</p>
                    </div>
                  ) : (
                    <div className="w-full h-[220px]">
                      <ResponsiveContainer width="100%" height="100%">
                        <AreaChart
                          data={stats?.daily_stats || []}
                          margin={{ top: 10, right: 5, left: -25, bottom: 0 }}
                        >
                          <defs>
                            <linearGradient id="colorCreated" x1="0" y1="0" x2="0" y2="1">
                              <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.15}/>
                              <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0.0}/>
                            </linearGradient>
                            <linearGradient id="colorCompleted" x1="0" y1="0" x2="0" y2="1">
                              <stop offset="5%" stopColor="#10b981" stopOpacity={0.15}/>
                              <stop offset="95%" stopColor="#10b981" stopOpacity={0.0}/>
                            </linearGradient>
                          </defs>
                          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border)/60)" />
                          <XAxis
                            dataKey="date"
                            stroke="hsl(var(--muted-foreground))"
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            tickFormatter={(str) => {
                              try {
                                const parts = str.split('-');
                                if (parts.length === 3) {
                                  return `${parts[1]}/${parts[2]}`;
                                }
                              } catch {}
                              return str;
                            }}
                          />
                          <YAxis
                            stroke="hsl(var(--muted-foreground))"
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            allowDecimals={false}
                          />
                          <Tooltip content={<ChartTooltip />} />
                          <Legend wrapperStyle={{ fontSize: '12px', paddingTop: '12px' }} />
                          <Area
                            name="Created"
                            type="monotone"
                            dataKey="created"
                            stroke="hsl(var(--primary))"
                            strokeWidth={2}
                            fillOpacity={1}
                            fill="url(#colorCreated)"
                          />
                          <Area
                            name="Completed"
                            type="monotone"
                            dataKey="completed"
                            stroke="#10b981"
                            strokeWidth={2}
                            fillOpacity={1}
                            fill="url(#colorCompleted)"
                          />
                        </AreaChart>
                      </ResponsiveContainer>
                    </div>
                  )}
                </div>
              </Card>
            </div>

            {/* Sub-row: Productivity Comparison */}
            <div className="grid grid-cols-1 gap-6 relative z-10">
              <Card className="border border-border bg-card/25 shadow-none rounded-lg p-5">
                <div className="mb-4">
                  <div className="flex items-center gap-2 mb-1">
                    <BarChart2 className="size-4 text-primary" />
                    <h3 className="text-sm font-bold uppercase tracking-wider text-foreground">Task Productivity Balance</h3>
                  </div>
                  <p className="text-xs text-muted-foreground">Side-by-side volume comparisons per day</p>
                </div>

                <div className="min-h-[200px] flex items-center justify-center">
                  {statsLoading ? (
                    <div className="text-muted-foreground text-xs font-semibold animate-pulse">Loading comparative stats...</div>
                  ) : !hasData ? (
                    <div className="text-center py-6">
                      <div className="size-10 rounded-full border border-border flex items-center justify-center text-muted-foreground mb-3 mx-auto">
                        <BarChart2 size={16} />
                      </div>
                      <p className="text-xs font-bold text-muted-foreground">No historical data found</p>
                    </div>
                  ) : (
                    <div className="w-full h-[200px]">
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart
                          data={stats?.daily_stats || []}
                          margin={{ top: 10, right: 5, left: -25, bottom: 0 }}
                          barSize={12}
                        >
                          <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border)/60)" />
                          <XAxis
                            dataKey="date"
                            stroke="hsl(var(--muted-foreground))"
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            tickFormatter={(str) => {
                              try {
                                const parts = str.split('-');
                                if (parts.length === 3) {
                                  return `${parts[1]}/${parts[2]}`;
                                }
                              } catch {}
                              return str;
                            }}
                          />
                          <YAxis
                            stroke="hsl(var(--muted-foreground))"
                            fontSize={10}
                            tickLine={false}
                            axisLine={false}
                            allowDecimals={false}
                          />
                          <Tooltip content={<ChartTooltip />} />
                          <Legend wrapperStyle={{ fontSize: '12px', paddingTop: '12px' }} />
                          <Bar
                            name="Created"
                            dataKey="created"
                            fill="hsl(var(--primary))"
                            radius={[3, 3, 0, 0]}
                          />
                          <Bar
                            name="Completed"
                            dataKey="completed"
                            fill="#10b981"
                            radius={[3, 3, 0, 0]}
                          />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  )}
                </div>
              </Card>
            </div>
          </main>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}
