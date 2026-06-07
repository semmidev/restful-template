import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { Loader2, AlertCircle, ArrowLeft } from 'lucide-react';
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardDescription } from "@/components/ui/card";
import useAuthStore from '../store';

export default function GoogleCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { loginWithGoogle, isLoading } = useAuthStore();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');

    if (!code) {
      setError('Authorization code was not returned by Google.');
      return;
    }

    // Retrieve state and verifier from sessionStorage to prevent CSRF
    const storedState = sessionStorage.getItem('oauth_state');
    const storedVerifier = sessionStorage.getItem('oauth_code_verifier');

    if (!storedState || !storedVerifier) {
      setError('OAuth session expired. Please try logging in again.');
      return;
    }

    if (state !== storedState) {
      setError('Security validation failed: state mismatch. Access denied.');
      return;
    }

    // Clear session storage once used
    sessionStorage.removeItem('oauth_state');
    sessionStorage.removeItem('oauth_code_verifier');

    let isMounted = true;

    async function exchangeCode() {
      const result = await loginWithGoogle(code, storedVerifier!);
      if (isMounted) {
        if (result.success) {
          navigate('/');
        } else {
          setError(result.error || 'Failed to complete login with Google.');
        }
      }
    }

    exchangeCode();

    return () => {
      isMounted = false;
    };
  }, [searchParams, loginWithGoogle, navigate]);

  return (
    <div className="flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10 animate-fade-in">
      <div className="w-full max-w-sm">
        <Card className="shadow-lg border-border/80">
          <CardHeader className="text-center">
            <CardTitle className="text-xl">Google Authentication</CardTitle>
            <CardDescription>
              {error ? 'An error occurred during authentication' : 'Completing sign-in process...'}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center justify-center p-6 text-center">
            {error ? (
              <div className="space-y-4">
                <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10 text-destructive">
                  <AlertCircle className="size-6" />
                </div>
                <p className="text-sm font-medium text-destructive">{error}</p>
                <Button asChild variant="outline" className="w-full border-border/80">
                  <Link to="/login" className="flex items-center justify-center gap-2">
                    <ArrowLeft className="size-4" />
                    Back to Login
                  </Link>
                </Button>
              </div>
            ) : (
              <div className="space-y-4">
                <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto" />
                <p className="text-sm text-muted-foreground">
                  Validating credentials and establishing secure session...
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
