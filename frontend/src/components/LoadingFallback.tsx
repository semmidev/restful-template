import { useEffect } from 'react';
import LoadingPage from './LoadingPage';

export default function LoadingFallback() {
  useEffect(() => {
    window.dispatchEvent(new Event('route-loading-start'));
    return () => {
      window.dispatchEvent(new Event('route-loading-end'));
    };
  }, []);

  return <LoadingPage />;
}
