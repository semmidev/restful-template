import React from 'react';

export default function LoadingPage() {
  return (
    <div className="min-h-screen bg-brutal-bg flex flex-col justify-center items-center p-4">
      <div className="card-brutal bg-white p-8 max-w-sm w-full text-center flex flex-col items-center gap-6">
        {/* Soft Brutalist Loading Animation: A bouncing bold square */}
        <div className="w-16 h-16 bg-brutal-yellow border-4 border-black shadow-[4px_4px_0px_#000] animate-bounce duration-1000 rotate-12 flex items-center justify-center font-black text-2xl select-none">
          ST
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-black uppercase tracking-tight text-black">
            Syncing Board...
          </h2>
          <p className="text-neutral-600 font-bold text-sm">
            Fetching your secure tasks. Please hold on.
          </p>
        </div>
        {/* Progress bar placeholder */}
        <div className="w-full bg-neutral-100 border-3 border-black h-4 rounded-full overflow-hidden relative shadow-brutal-sm">
          <div className="bg-brutal-blue h-full w-2/3 animate-[pulse_1.5s_infinite] border-r-3 border-black"></div>
        </div>
      </div>
    </div>
  );
}
