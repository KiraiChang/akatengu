import { authStore } from '../stores/auth.svelte';

interface LoginResponse {
  token: string;
}

interface ProblemDetail {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
}

export class AuthError extends Error {
  readonly status: number;
  readonly detail: string;

  constructor(problem: ProblemDetail) {
    super(problem.title);
    this.name = 'AuthError';
    this.status = problem.status;
    this.detail = problem.detail ?? problem.title;
  }
}

export const login = async (username: string, password: string): Promise<void> => {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) {
    const problem: ProblemDetail = await response.json();
    throw new AuthError(problem);
  }

  const data: LoginResponse = await response.json();
  authStore.setPendingToken(data.token);
};
