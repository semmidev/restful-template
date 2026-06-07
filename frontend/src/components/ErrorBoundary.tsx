import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from '@/components/ui/card';
import { AlertOctagon } from 'lucide-react';

interface Props {
  children?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Uncaught error in boundary:", error, errorInfo);
  }

  private handleReload = () => {
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-slate-50/50 flex flex-col justify-center items-center p-6">
          <div className="w-full max-w-md">
            <Card className="border border-border bg-card shadow-lg rounded-xl">
              <CardHeader className="space-y-1">
                <CardTitle className="text-2xl font-bold text-destructive flex items-center gap-2">
                  <AlertOctagon size={24} /> 
                  Oops! Crash Detected
                </CardTitle>
                <CardDescription>
                  Something went wrong in the rendering tree of the workspace
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-5">
                <p className="text-sm font-medium text-slate-600 leading-relaxed">
                  Our task synchronizer ran into an unexpected runtime script error. You can try reloading the page or reset your credentials.
                </p>
                
                {this.state.error && (
                  <div className="bg-slate-100 border border-slate-200 p-4 rounded-lg font-mono text-[11px] overflow-x-auto select-all max-h-40 text-slate-800">
                    <span className="font-bold text-destructive">Error: </span>
                    {this.state.error.message}
                    {this.state.error.stack && (
                      <div className="mt-2 text-slate-500 whitespace-pre">
                        {this.state.error.stack}
                      </div>
                    )}
                  </div>
                )}

                <div className="flex flex-col sm:flex-row gap-3 pt-2">
                  <Button
                    onClick={this.handleReload}
                    className="w-full h-10 text-sm font-semibold bg-primary text-primary-foreground hover:bg-primary/95 transition-all shadow-sm rounded-md"
                  >
                    Reload Page
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      localStorage.clear();
                      window.location.href = '/login';
                    }}
                    className="w-full h-10 text-sm font-semibold rounded-md"
                  >
                    Reset & Sign In
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
