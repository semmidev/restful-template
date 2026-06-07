import client from '../../lib/client';
import { LoginInput, RegisterInput } from '../../lib/schemas';

export async function loginRequest(input: LoginInput) {
  return client.post('/auth/login', input);
}

export async function registerRequest(input: Omit<RegisterInput, 'confirmPassword'>) {
  return client.post('/auth/register', input);
}

export async function googleLoginRequest(code: string, codeVerifier: string) {
  return client.post('/auth/google', { code, code_verifier: codeVerifier });
}

export async function getGoogleConfig() {
  return client.get('/auth/google/config');
}

export async function logoutRequest() {
  return client.post('/auth/logout');
}

