import React from 'react';
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

export default function TodoSkeleton() {
  return (
    <Card className="card-brutal bg-white flex flex-col justify-between h-[360px] overflow-hidden">
      <div>
        {/* Cover image placeholder */}
        <div className="w-full h-40 bg-neutral-200 animate-pulse border-b-3 border-black -mt-6 -mx-6 mb-4 rounded-t-lg max-w-[calc(100%+3rem)]" />
        
        {/* Badge & Buttons placeholder */}
        <div className="flex justify-between items-start gap-2 mb-3">
          <Skeleton className="h-6 w-20 rounded" />
          <div className="flex gap-1.5">
            <Skeleton className="h-8 w-8 rounded" />
            <Skeleton className="h-8 w-8 rounded" />
          </div>
        </div>

        {/* Title placeholder */}
        <Skeleton className="h-6 w-3/4 rounded mb-2.5" />
        
        {/* Description placeholder */}
        <div className="space-y-2 mb-6">
          <Skeleton className="h-4 w-full rounded" />
          <Skeleton className="h-4 w-5/6 rounded" />
        </div>
      </div>

      {/* Footer controls placeholder */}
      <div className="border-t-2 border-black pt-4 mt-auto">
        <div className="flex justify-between items-center gap-2">
          <div className="flex gap-1.5">
            <Skeleton className="h-7 w-16 rounded" />
            <Skeleton className="h-7 w-20 rounded" />
          </div>
          <Skeleton className="h-3.5 w-24 rounded" />
        </div>
      </div>
    </Card>
  );
}
