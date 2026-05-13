import { apiRequest } from './http/client';

export interface AuthCredentials {
  email: string;
  password: string;
}

export interface AuthResult {
  user_id: string;
  email: string;
  role: string;
  access_token: string;
  token_type: string;
}

export interface LogoutResult {
  revoked: boolean;
}

export const login = (body: AuthCredentials) => {
  return apiRequest<AuthResult>('/api/v1/auth/login', {
    method: 'POST',
    data: body,
  });
};

export const register = (body: AuthCredentials) => {
  return apiRequest<AuthResult>('/api/v1/auth/register', {
    method: 'POST',
    data: body,
  });
};

export const logout = () => {
  return apiRequest<LogoutResult>('/api/v1/auth/logout', {
    method: 'POST',
  });
};
