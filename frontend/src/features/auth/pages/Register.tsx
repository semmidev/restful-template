import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Mail, Lock, Eye, EyeOff, CheckSquare, Loader2 } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Field, FieldDescription, FieldGroup, FieldLabel, FieldError } from "@/components/ui/field";
import useAuthStore from '../store';
import { registerSchema } from '../../../lib/schemas';
import { generateCodeVerifier, generateCodeChallenge, generateState } from '../pkce';
import { getGoogleConfig } from '../api';

const GoogleIcon = () => (
  <svg className="mr-2 h-4 w-4" aria-hidden="true" focusable="false" role="img" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 488 512">
    <path fill="currentColor" d="M488 261.8C488 403.3 391.1 504 248 504 110.8 504 0 393.2 0 256S110.8 8 248 8c66.8 0 123 24.5 166.3 64.9l-67.5 64.9C258.5 52.6 94.3 116.6 94.3 256c0 86.5 69.1 156.6 153.7 156.6 98.2 0 135-70.4 140.8-106.9H248v-85.3h236.1c2.3 12.7 3.9 24.9 3.9 41.4z"></path>
  </svg>
);

export default function Register() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState<{ email?: string; password?: string; confirmPassword?: string }>({});
  const navigate = useNavigate();

  const { register, isLoading, error: authError, clearError } = useAuthStore();
  const [googleLoading, setGoogleLoading] = useState(false);
  const [googleError, setGoogleError] = useState<string | null>(null);

  const handleGoogleLogin = async () => {
    setGoogleLoading(true);
    setGoogleError(null);
    try {
      const configRes = await getGoogleConfig();
      const { client_id, redirect_uri } = configRes.data;

      if (!client_id || !redirect_uri) {
        throw new Error('Google OAuth is not configured on the backend.');
      }

      const verifier = generateCodeVerifier();
      const challenge = await generateCodeChallenge(verifier);
      const state = generateState();

      sessionStorage.setItem('oauth_code_verifier', verifier);
      sessionStorage.setItem('oauth_state', state);

      const authUrl = `https://accounts.google.com/o/oauth2/v2/auth` +
        `?client_id=${encodeURIComponent(client_id)}` +
        `&redirect_uri=${encodeURIComponent(redirect_uri)}` +
        `&response_type=code` +
        `&scope=openid%20email%20profile` +
        `&state=${encodeURIComponent(state)}` +
        `&code_challenge=${encodeURIComponent(challenge)}` +
        `&code_challenge_method=S256`;

      window.location.href = authUrl;
    } catch (err: any) {
      console.error(err);
      setGoogleError(err.message || 'Failed to initialize Google login.');
      setGoogleLoading(false);
    }
  };

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
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-muted p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        {/* Brand Logo Header */}
        <Link to="/" className="flex items-center gap-2 self-center font-bold text-lg">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
            <CheckSquare className="size-4" />
          </div>
          TodoApp
        </Link>

        {/* Register Form Card */}
        <Card className="shadow-lg">
          <CardHeader className="text-center">
            <CardTitle className="text-xl">Create your account</CardTitle>
            <CardDescription>
              Enter your details below to create your secure account
            </CardDescription>
          </CardHeader>
          <CardContent>
            {(authError || googleError) && (
              <div className="bg-destructive/10 border border-destructive/20 text-destructive text-sm font-semibold p-3 mb-4 rounded-lg text-center animate-shake">
                ⚠️ {authError || googleError}
              </div>
            )}

            <Button
              type="button"
              variant="outline"
              disabled={isLoading || googleLoading}
              onClick={handleGoogleLogin}
              className="w-full flex items-center justify-center border-border hover:bg-muted mb-4 h-10 font-semibold"
            >
              {googleLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin text-muted-foreground" />
              ) : (
                <GoogleIcon />
              )}
              {googleLoading ? 'Redirecting to Google...' : 'Continue with Google'}
            </Button>

            <div className="relative my-4">
              <div className="absolute inset-0 flex items-center">
                <span className="w-full border-t border-border" />
              </div>
              <div className="relative flex justify-center text-xs uppercase">
                <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
              </div>
            </div>

            <form onSubmit={handleRegister}>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="email">Email Address</FieldLabel>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground z-10 size-4" />
                    <Input
                      id="email"
                      type="email"
                      required
                      placeholder="m@example.com"
                      className="pl-9 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0"
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
                    <FieldError>{validationErrors.email}</FieldError>
                  )}
                </Field>

                <Field>
                  <FieldLabel htmlFor="password">Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground z-10 size-4" />
                    <Input
                      id="password"
                      type={showPassword ? "text" : "password"}
                      required
                      placeholder="••••••••"
                      className="pl-9 pr-9 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0"
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
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground z-10 transition-colors"
                    >
                      {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                    </button>
                  </div>
                  <FieldDescription className="text-[10px] text-muted-foreground">
                    Must be at least 8 characters long.
                  </FieldDescription>
                  {validationErrors.password && (
                    <FieldError>{validationErrors.password}</FieldError>
                  )}
                </Field>

                <Field>
                  <FieldLabel htmlFor="confirmPassword">Confirm Password</FieldLabel>
                  <div className="relative">
                    <Lock className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground z-10 size-4" />
                    <Input
                      id="confirmPassword"
                      type={showConfirmPassword ? "text" : "password"}
                      required
                      placeholder="••••••••"
                      className="pl-9 pr-9 focus-visible:ring-1 focus-visible:ring-primary focus-visible:ring-offset-0"
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
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground z-10 transition-colors"
                    >
                      {showConfirmPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                    </button>
                  </div>
                  {validationErrors.confirmPassword && (
                    <FieldError>{validationErrors.confirmPassword}</FieldError>
                  )}
                </Field>

                <Field>
                  <Button type="submit" disabled={isLoading} className="w-full h-10 font-semibold bg-primary text-primary-foreground hover:bg-primary/95 transition-all shadow-sm">
                    {isLoading ? 'Creating Account...' : 'Create Account'}
                  </Button>
                  <FieldDescription className="text-center mt-2 text-xs">
                    Already have an account?{' '}
                    <Link to="/login" className="underline font-bold text-primary hover:text-primary/90">
                      Sign In
                    </Link>
                  </FieldDescription>
                </Field>
              </FieldGroup>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
