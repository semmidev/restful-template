import { useEffect, useState } from 'react';

export default function RouteProgressBar() {
  const [width, setWidth] = useState(0);
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    let trickleInterval: any;
    let fadeTimeout: any;
    let resetTimeout: any;

    const handleStart = () => {
      clearTimeout(fadeTimeout);
      clearTimeout(resetTimeout);
      clearInterval(trickleInterval);

      setWidth(0);
      setVisible(true);

      // Start trickle animation
      setTimeout(() => setWidth(10), 10);

      // Trickle progress slowly up to 90%
      trickleInterval = setInterval(() => {
        setWidth((prev) => {
          if (prev >= 90) return prev;
          const remaining = 90 - prev;
          return prev + remaining * 0.1;
        });
      }, 300);
    };

    const handleEnd = () => {
      clearInterval(trickleInterval);
      setWidth(100);

      // Fade out after reaching 100%
      fadeTimeout = setTimeout(() => {
        setVisible(false);
      }, 400);

      // Reset width to 0% after fade animation completes
      resetTimeout = setTimeout(() => {
        setWidth(0);
      }, 700);
    };

    window.addEventListener('route-loading-start', handleStart);
    window.addEventListener('route-loading-end', handleEnd);

    return () => {
      window.removeEventListener('route-loading-start', handleStart);
      window.removeEventListener('route-loading-end', handleEnd);
      clearInterval(trickleInterval);
      clearTimeout(fadeTimeout);
      clearTimeout(resetTimeout);
    };
  }, []);

  if (!visible) return null;

  return (
    <div className="fixed top-0 left-0 w-full z-[99999] pointer-events-none">
      <div
        className="h-[3px] bg-primary shadow-[0_0_8px_hsl(var(--primary))] transition-all duration-300 ease-out"
        style={{ width: `${width}%` }}
      />
    </div>
  );
}
