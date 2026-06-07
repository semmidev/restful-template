import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
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
        <div className="min-h-screen bg-brutal-bg flex flex-col justify-center items-center p-6">
          <div className="w-full max-w-lg">
            <Card className="border-3 border-black bg-white shadow-[8px_8px_0px_#000] p-8">
              <CardHeader className="border-b-3 border-black pb-4 mb-6">
                <CardTitle className="text-3xl font-black text-black flex items-center gap-2">
                  <AlertOctagon size={32} className="text-brutal-pink" /> 
                  Oops! Crash Detected
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0 space-y-6">
                <p className="font-bold text-neutral-800">
                  Something went wrong in the rendering tree. Our task synchronizer ran into an unexpected error.
                </p>
                
                {this.state.error && (
                  <div className="bg-neutral-100 border-2 border-black p-4 rounded-lg font-mono text-xs overflow-x-auto select-all max-h-40">
                    <span className="font-black text-red-600">Error: </span>
                    {this.state.error.message}
                    {this.state.error.stack && (
                      <div className="mt-2 text-neutral-600 whitespace-pre">
                        {this.state.error.stack}
                      </div>
                    )}
                  </div>
                )}

                <div className="flex flex-col sm:flex-row gap-4 pt-2">
                  <Button
                    onClick={this.handleReload}
                    className="btn-brutal bg-brutal-violet w-full h-12 text-md justify-center font-black"
                  >
                    Reload Page
                  </Button>
                  <Button
                    onClick={() => {
                      localStorage.clear();
                      window.location.href = '/login';
                    }}
                    className="btn-brutal-secondary w-full h-12 text-md justify-center font-black"
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
