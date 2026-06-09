import axios, { AxiosResponse } from 'axios';
import { useAuthStore } from '../features/auth/store';

interface FailedRequest {
  resolve: () => void;
  reject: (err: any) => void;
}

const client = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

let isRefreshing = false;
let failedQueue: FailedRequest[] = [];

const processQueue = (error: any) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve();
    }
  });
  failedQueue = [];
};

client.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: any) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      const url = originalRequest.url || '';
      const isAuthRequest =
        url.endsWith('/auth/login') ||
        url.endsWith('/auth/register') ||
        url.endsWith('/auth/refresh') ||
        url.endsWith('/auth/logout') ||
        url.includes('auth/login') ||
        url.includes('auth/register') ||
        url.includes('auth/refresh') ||
        url.includes('auth/logout');

      if (isAuthRequest) {
        if (url.includes('auth/refresh')) {
          useAuthStore.getState().logout();
        }
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise<void>((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => client(originalRequest))
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        await axios.post('/api/v1/auth/refresh', {}, { withCredentials: true });

        processQueue(null);
        isRefreshing = false;
        return client(originalRequest);
      } catch (refreshErr) {
        processQueue(refreshErr);
        isRefreshing = false;
        useAuthStore.getState().logout();
        return Promise.reject(refreshErr);
      }
    }

    return Promise.reject(error);
  }
);

export default client;
