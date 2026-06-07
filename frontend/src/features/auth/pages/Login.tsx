import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { LogIn, Mail, Lock, Eye, EyeOff } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import useAuthStore from '../store';
import { loginSchema } from '../../../lib/schemas';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{ email?: string; password?: string }>({});
  const navigate = useNavigate();

  const { login, isLoading, error: authError, clearError } = useAuthStore();

  useEffect(() => {
    clearError();
  }, [clearError]);

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setValidationErrors({});
    clearError();

    const result = loginSchema.safeParse({ email, password });
    if (!result.success) {
      const fieldErrors: { email?: string; password?: string } = {};
      result.error.issues.forEach((issue) => {
        const path = issue.path[0] as 'email' | 'password';
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const res = await login({ email, password });
    if (res.success) {
      navigate('/');
    }
  };

  return (
    <div className="min-h-screen bg-brutal-bg flex flex-col justify-center items-center p-4">
      <div className="w-full max-w-lg">
        {/* Soft Brutalist Logo/Header */}
        <div className="mb-8 text-center">
          <span className="inline-block bg-brutal-pink text-black border-3 border-black font-black px-4 py-1 text-sm uppercase tracking-wider shadow-brutal-sm rotate-[-2deg] mb-3">
            Task Organizer v1.0
          </span>
          <h1 className="text-5xl font-black tracking-tight text-black mt-2">
            STARK<span className="text-brutal-blue">TODO</span>
          </h1>
        </div>

        {/* Form Card */}
        <Card className="border-3 border-black bg-white shadow-[8px_8px_0px_0px_#000000] p-8">
          <CardHeader className="border-b-3 border-black pb-4 mb-6">
            <CardTitle className="text-3xl font-black text-black flex items-center gap-2">
              <LogIn size={28} /> Log In
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {authError && (
              <div className="bg-brutal-pink border-3 border-black p-3 mb-6 rounded-lg font-bold text-sm shadow-brutal-sm">
                ⚠️ {authError}
              </div>
            )}

            <form onSubmit={handleLogin} className="space-y-6">
              <div className="flex flex-col">
                <label className="text-sm font-black text-black mb-1.5 uppercase tracking-wide">
                  Email Address
                </label>
                <div className="relative">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500 z-10" size={18} />
                  <Input
                    type="email"
                    required
                    placeholder="name@example.com"
                    className="w-full pl-11"
                    value={email}
                    onChange={(e) => {
                      setEmail(e.target.value);
                      if (validationErrors.email) {
                        setValidationErrors((prev) => ({ ...prev, email: undefined }));
                      }
                    }}
                  />
                </div>
                {validationErrors.email && (
                  <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                    ⚠️ {validationErrors.email}
                  </p>
                )}
              </div>

              <div className="flex flex-col">
                <label className="text-sm font-black text-black mb-1.5 uppercase tracking-wide">
                  Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500 z-10" size={18} />
                  <Input
                    type={showPassword ? "text" : "password"}
                    required
                    placeholder="••••••••"
                    className="w-full pl-11 pr-12"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (validationErrors.password) {
                        setValidationErrors((prev) => ({ ...prev, password: undefined }));
                      }
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-neutral-500 hover:text-black z-10 transition-colors"
                  >
                    {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                  </button>
                </div>
                {validationErrors.password && (
                  <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                    ⚠️ {validationErrors.password}
                  </p>
                )}
              </div>

              <Button
                type="submit"
                disabled={isLoading}
                className="btn-brutal w-full justify-center text-lg mt-2 h-12"
              >
                {isLoading ? 'Logging in...' : 'Sign In'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Link to Register */}
        <div className="mt-6 text-center">
          <p className="font-bold text-neutral-800">
            Don't have an account?{' '}
            <Link
              to="/register"
              className="text-brutal-blue underline decoration-3 hover:text-black font-black transition-all"
            >
              Create Account
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
