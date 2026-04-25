const TOKEN_KEY = 'auth_token';

function createAuthStore() {
  let token = $state<string | null>(localStorage.getItem(TOKEN_KEY));

  return {
    get token() { return token; },
    setToken(newToken: string): void {
      token = newToken;
      localStorage.setItem(TOKEN_KEY, newToken);
    },
    clearToken(): void {
      token = null;
      localStorage.removeItem(TOKEN_KEY);
    },
  };
}

export const authStore = createAuthStore();
