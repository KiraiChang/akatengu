<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
  <link
    href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,500;1,300&family=JetBrains+Mono:wght@300;400;500&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<script lang="ts">
  import { login, AuthError } from "../api/auth";

  let username = $state('')
  let password = $state('')
  let showPassword = $state(false)
  let isLoading = $state(false)
  let errorMessage = $state('')

  let isValid = $derived(username.trim().length > 0 && password.length > 0)

  async function handleSubmit(e: SubmitEvent): Promise<void> {
    e.preventDefault()
    if (!isValid || isLoading) return

    isLoading = true
    errorMessage = ''

    try {
      await login(username, password)
      window.location.hash = '#/home'
    } catch (err) {
      if (err instanceof AuthError) {
        errorMessage = err.detail
      } else {
        errorMessage = '登入失敗，請稍後再試。'
      }
    } finally {
      isLoading = false
    }
  }
</script>

<div class="login-page">
  <main class="login-card">
    <header class="login-brand">
      <div class="login-brand-mark" aria-hidden="true">赤</div>
      <div class="login-brand-text">
        <p class="login-brand-name">AKATENGU</p>
        <p class="login-brand-sub">會計管理系統</p>
      </div>
    </header>

    <hr class="login-sep" />

    <form onsubmit={handleSubmit} novalidate>
      <div class="login-field">
        <label class="login-field-label" for="username">使用者名稱</label>
        <input
          id="username"
          type="text"
          class="login-field-input"
          bind:value={username}
          autocomplete="username"
          spellcheck={false}
          placeholder="請輸入使用者名稱"
          disabled={isLoading}
        />
      </div>

      <div class="login-field">
        <label class="login-field-label" for="password">密碼</label>
        <div class="login-pw-wrap">
          <input
            id="password"
            type={showPassword ? 'text' : 'password'}
            class="login-field-input"
            bind:value={password}
            autocomplete="current-password"
            placeholder="請輸入密碼"
            disabled={isLoading}
          />
          <button
            type="button"
            class="login-pw-toggle"
            onclick={() => (showPassword = !showPassword)}
            aria-label={showPassword ? '隱藏密碼' : '顯示密碼'}
            tabindex="-1"
          >
            {#if showPassword}
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
            {:else}
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            {/if}
          </button>
        </div>
      </div>

      {#if errorMessage}
        <p class="login-error" role="alert">{errorMessage}</p>
      {/if}

      <button type="submit" class="login-submit" disabled={!isValid || isLoading}>
        {#if isLoading}
          <span class="login-spinner" aria-hidden="true"></span>
          <span>驗證中...</span>
        {:else}
          <span>登 入</span>
        {/if}
      </button>
    </form>

    <footer class="login-footer">
      <span>© 2026 AKATENGU</span>
    </footer>
  </main>
</div>
