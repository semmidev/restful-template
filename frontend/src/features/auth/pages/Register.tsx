import React, { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { UserPlus, Mail, Lock, Eye, EyeOff } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import useAuthStore from '../store';
import { registerSchema } from '../../../lib/schemas';

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{ email?: string; password?: string; confirmPassword?: string }>({});
  const navigate = useNavigate();

  const { register, isLoading, error: authError, clearError } = useAuthStore();

  useEffect(() => {
    clearError();
  }, [clearError]);

  const handleRegister = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setValidationErrors({});
    clearError();

    const result = registerSchema.safeParse({ email, password, confirmPassword });
    if (!result.success) {
      const fieldErrors: { email?: string; password?: string; confirmPassword?: string } = {};
      result.error.issues.forEach((issue) => {
        const path = issue.path[0] as 'email' | 'password' | 'confirmPassword';
        if (path) {
          fieldErrors[path] = issue.message;
        }
      });
      setValidationErrors(fieldErrors);
      return;
    }

    const res = await register({ email, password, confirmPassword });
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
              <UserPlus size={28} /> Register
            </CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            {authError && (
              <div className="bg-brutal-pink border-3 border-black p-3 mb-6 rounded-lg font-bold text-sm shadow-brutal-sm">
                ⚠️ {authError}
              </div>
            )}

            <form onSubmit={handleRegister} className="space-y-6">
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
                  Password (min 8 chars)
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

              <div className="flex flex-col">
                <label className="text-sm font-black text-black mb-1.5 uppercase tracking-wide">
                  Confirm Password
                </label>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500 z-10" size={18} />
                  <Input
                    type={showConfirmPassword ? "text" : "password"}
                    required
                    placeholder="••••••••"
                    className="w-full pl-11 pr-12"
                    value={confirmPassword}
                    onChange={(e) => {
                      setConfirmPassword(e.target.value);
                      if (validationErrors.confirmPassword) {
                        setValidationErrors((prev) => ({ ...prev, confirmPassword: undefined }));
                      }
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-neutral-500 hover:text-black z-10 transition-colors"
                  >
                    {showConfirmPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                  </button>
                </div>
                {validationErrors.confirmPassword && (
                  <p className="text-red-600 text-xs font-black mt-1 uppercase tracking-wide">
                    ⚠️ {validationErrors.confirmPassword}
                  </p>
                )}
              </div>

              <Button
                type="submit"
                disabled={isLoading}
                className="btn-brutal w-full justify-center text-lg mt-2 h-12"
              >
                {isLoading ? 'Creating Account...' : 'Register'}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Link to Login */}
        <div className="mt-6 text-center">
          <p className="font-bold text-neutral-800">
            Already have an account?{' '}
            <Link
              to="/login"
              className="text-brutal-blue underline decoration-3 hover:text-black font-black transition-all"
            >
              Login Here
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
