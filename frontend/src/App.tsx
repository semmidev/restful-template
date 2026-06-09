import { Suspense, lazy } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from 'next-themes';
import useAuthStore from './features/auth/store';
import LoadingFallback from './components/LoadingFallback';
import RouteProgressBar from './components/RouteProgressBar';
import ErrorBoundary from './components/ErrorBoundary';

// Lazy load page bundles for code-splitting
const Login = lazy(() => import('./features/auth/pages/Login'));
const Register = lazy(() => import('./features/auth/pages/Register'));
const GoogleCallback = lazy(() => import('./features/auth/pages/GoogleCallback'));
const Todos = lazy(() => import('./features/todos/pages/Todos'));
const Dashboard = lazy(() => import('./features/todos/pages/Dashboard'));
const EisenhowerMatrix = lazy(() => import('./features/todos/pages/EisenhowerMatrix'));
const Users = lazy(() => import('./features/adm/users/pages/Users'));

interface RouteProps {
  children: React.ReactNode;
}

const PrivateRoute: React.FC<RouteProps> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />;
};

const PublicRoute: React.FC<RouteProps> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  return isAuthenticated ? <Navigate to="/" replace /> : <>{children}</>;
};

function App() {
  return (
    <ThemeProvider attribute="class" defaultTheme="light" enableSystem={false}>
      <RouteProgressBar />
      <ErrorBoundary>
        <Router>
          <Suspense fallback={<LoadingFallback />}>
            <Routes>
              <Route
                path="/login"
                element={
                  <PublicRoute>
                    <Login />
                  </PublicRoute>
                }
              />
              <Route
                path="/login/google/callback"
                element={
                  <PublicRoute>
                    <GoogleCallback />
                  </PublicRoute>
                }
              />
              <Route
                path="/register"
                element={
                  <PublicRoute>
                    <Register />
                  </PublicRoute>
                }
              />
              <Route
                path="/"
                element={
                  <PrivateRoute>
                    <Dashboard />
                  </PrivateRoute>
                }
              />
              <Route
                path="/todos"
                element={
                  <PrivateRoute>
                    <Todos />
                  </PrivateRoute>
                }
              />
              <Route
                path="/matrix"
                element={
                  <PrivateRoute>
                    <EisenhowerMatrix />
                  </PrivateRoute>
                }
              />
              <Route
                path="/users"
                element={
                  <PrivateRoute>
                    <Users />
                  </PrivateRoute>
                }
              />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </Router>
      </ErrorBoundary>
    </ThemeProvider>
  );
}

export default App;
