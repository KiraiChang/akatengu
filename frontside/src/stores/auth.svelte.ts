const TOKEN_KEY = 'auth_token';
const PENDING_TOKEN_KEY = 'pending_token';

function createAuthStore() {
  let token = $state<string | null>(localStorage.getItem(TOKEN_KEY));
  let pendingToken = $state<string | null>(sessionStorage.getItem(PENDING_TOKEN_KEY));

  return {
    get token() { return token; },
    get pendingToken() { return pendingToken; },
    setToken(newToken: string): void {
      token = newToken;
      localStorage.setItem(TOKEN_KEY, newToken);
    },
    clearToken(): void {
      token = null;
      pendingToken = null;
      localStorage.removeItem(TOKEN_KEY);
      sessionStorage.removeItem(PENDING_TOKEN_KEY);
    },
    setPendingToken(newToken: string): void {
      pendingToken = newToken;
      sessionStorage.setItem(PENDING_TOKEN_KEY, newToken);
    },
    clearPendingToken(): void {
      pendingToken = null;
      sessionStorage.removeItem(PENDING_TOKEN_KEY);
    },
  };
}

export const authStore = createAuthStore();
