import { authStore } from '../stores/auth.svelte';

export async function apiFetch(input: string, init: RequestInit = {}, token: string | null = null): Promise<Response> {
  let isFromAuth = false;
  let header_token = token;
  if (header_token == null) {
      header_token = authStore.token;
      isFromAuth = true;
  }

  const response = await fetch(input, {
    ...init,
    headers: {
      Authorization: `Bearer ${header_token}`,
      ...init.headers,
    },
  });

  if (response.status === 401) {
    if (isFromAuth) {
      authStore.clearToken();
    }
    window.location.hash = '#/';
    throw new Error('未授權，請重新登入');
  }

  return response;
}
