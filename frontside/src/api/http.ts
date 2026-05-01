import { authStore } from '../stores/auth.svelte';

export async function apiFetch(input: string, init: RequestInit = {}): Promise<Response> {
  const response = await fetch(input, {
    ...init,
    headers: {
      Authorization: `Bearer ${authStore.token}`,
      ...init.headers,
    },
  });

  if (response.status === 401) {
    authStore.clearToken();
    window.location.hash = '#/';
    throw new Error('未授權，請重新登入');
  }

  return response;
}
