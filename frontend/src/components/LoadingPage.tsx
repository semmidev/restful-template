import React from 'react';
import { Shield } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';

export default function LoadingPage() {
  return (
    <div className="min-h-screen bg-background relative flex flex-col justify-center items-center p-4 overflow-hidden">
      {/* CSS Animation injection for the sliding progress bar */}
      <style dangerouslySetInnerHTML={{ __html: `
        @keyframes progress-slide {
          0% { left: -60%; width: 60%; }
          50% { width: 40%; }
          100% { left: 100%; width: 60%; }
        }
        .animate-progress-slide {
          animation: progress-slide 1.8s infinite ease-in-out;
        }
      `}} />

      {/* Premium ambient glow backdrops */}
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[380px] h-[380px] bg-primary/10 rounded-full blur-[90px] pointer-events-none" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[220px] h-[220px] bg-indigo-500/10 rounded-full blur-[70px] pointer-events-none" />

      <Card className="border border-border/50 bg-card/60 backdrop-blur-xl p-8 max-w-sm w-full text-center flex flex-col items-center gap-6 shadow-2xl rounded-2xl relative z-10 transition-all duration-300">
        <CardContent className="flex flex-col items-center gap-5 p-0 w-full">
          {/* Glowing Animated Spinner & Icon */}
          <div className="relative flex items-center justify-center w-16 h-16">
            {/* Outer spinning gradient border */}
            <div className="absolute inset-0 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
            
            {/* Middle pulsing background */}
            <div className="absolute inset-2 rounded-full bg-primary/5 animate-ping opacity-60" />
            
            {/* Center icon container */}
            <div className="absolute inset-2 rounded-full bg-primary/10 flex items-center justify-center text-primary shadow-[inset_0_0_12px_rgba(var(--primary),0.15)]">
              <Shield size={20} className="animate-pulse" />
            </div>
          </div>

          {/* Secure loading text */}
          <div className="space-y-2">
            <h2 className="text-sm font-bold uppercase tracking-widest text-foreground/90">
              Syncing Board
            </h2>
            <p className="text-muted-foreground text-xs leading-relaxed max-w-[280px]">
              Establishing a secure session and decrypting your task vault. Please hold on...
            </p>
          </div>

          {/* Infinite sliding progress bar */}
          <div className="w-full max-w-[180px] bg-muted h-1 rounded-full overflow-hidden relative">
            <div className="bg-primary h-full rounded-full animate-progress-slide absolute left-0 top-0 shadow-[0_0_8px_rgba(var(--primary),0.5)]" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

