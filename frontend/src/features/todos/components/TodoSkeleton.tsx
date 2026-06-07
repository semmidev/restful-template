import React from 'react';
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

export default function TodoSkeleton() {
  return (
    <Card className="border border-border bg-card shadow-sm rounded-xl overflow-hidden flex flex-col justify-between p-0 gap-0">
      <div>
        {/* Cover image placeholder */}
        <div className="w-full h-40 bg-slate-100 animate-pulse border-b border-border" />
        
        {/* Content container */}
        <div className="p-6 space-y-4">
          {/* Badge & Buttons placeholder */}
          <div className="flex justify-between items-center gap-2">
            <Skeleton className="h-5 w-20 rounded-full" />
            <div className="flex gap-1.5">
              <Skeleton className="h-7 w-7 rounded-md" />
              <Skeleton className="h-7 w-7 rounded-md" />
            </div>
          </div>

          {/* Title placeholder */}
          <Skeleton className="h-6 w-3/4 rounded-md" />
          
          {/* Description placeholder */}
          <div className="space-y-2">
            <Skeleton className="h-4 w-full rounded-md" />
            <Skeleton className="h-4 w-5/6 rounded-md" />
          </div>
        </div>
      </div>

      {/* Footer controls placeholder */}
      <div className="border-t border-border py-4 px-6 flex justify-between items-center gap-4 mt-auto">
        <div className="flex gap-1.5">
          <Skeleton className="h-7 w-16 rounded-md" />
          <Skeleton className="h-7 w-20 rounded-md" />
        </div>
        <Skeleton className="h-3.5 w-16 rounded-md" />
      </div>
    </Card>
  );
}
