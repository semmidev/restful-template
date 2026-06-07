import React from 'react';
import { Loader2 } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';

export default function LoadingPage() {
  return (
    <div className="min-h-screen bg-slate-50/50 flex flex-col justify-center items-center p-4">
      <Card className="border border-border bg-card p-8 max-w-sm w-full text-center flex flex-col items-center gap-6 shadow-lg rounded-xl">
        <CardContent className="flex flex-col items-center gap-5 p-0">
          <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center text-primary animate-spin">
            <Loader2 size={24} />
          </div>
          <div className="space-y-1.5">
            <h2 className="text-xl font-bold tracking-tight text-slate-900">
              Syncing Board...
            </h2>
            <p className="text-muted-foreground text-xs font-semibold">
              Fetching your secure tasks. Please hold on.
            </p>
          </div>
          {/* Progress bar placeholder */}
          <div className="w-48 bg-slate-100 h-1 rounded-full overflow-hidden relative">
            <div className="bg-primary h-full w-2/3 animate-[pulse_1.5s_infinite] rounded-full"></div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
