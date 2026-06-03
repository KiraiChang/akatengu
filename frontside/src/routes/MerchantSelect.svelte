<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
  <link
    href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,300;0,500;1,300&family=JetBrains+Mono:wght@300;400;500&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth.svelte';
  import { list, select } from '../api/merchant';
  import type { Merchant } from '../types/merchant';

  let merchants = $state<Merchant[]>([]);
  let isLoading = $state(true);
  let errorMessage = $state('');
  let selectingId = $state<number | null>(null);

  onMount(async () => {
    if (authStore.token) {
      window.location.hash = '#/home';
      return;
    }
    if (!authStore.pendingToken) {
      window.location.hash = '#/';
      return;
    }
    await loadMerchants();
  });

  async function loadMerchants(): Promise<void> {
    try {
      const all = await list(authStore.pendingToken);
      merchants = all.filter((m) => m.status === 'ACTIVE');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : '載入商戶失敗';
    } finally {
      isLoading = false;
    }
  }

  async function handleSelect(merchant: Merchant): Promise<void> {
    if (selectingId !== null) return;
    selectingId = merchant.merchant_id;
    errorMessage = '';

    try {
      const result = await select(merchant.merchant_id, authStore.pendingToken);
      authStore.setToken(result.token);
      authStore.clearPendingToken();
      window.location.hash = '#/home';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : '選擇商戶失敗';
      selectingId = null;
    }
  }

  function handleBack(): void {
    authStore.clearPendingToken();
    window.location.hash = '#/';
  }
</script>

<div class="ms-page">
  <main class="ms-card">
    <header class="ms-brand">
      <div class="ms-brand-mark" aria-hidden="true">赤</div>
      <div class="ms-brand-text">
        <p class="ms-brand-name">AKATENGU</p>
        <p class="ms-brand-sub">選擇商戶</p>
      </div>
    </header>

    <hr class="ms-sep" />

    {#if isLoading}
      <div class="ms-loading" aria-live="polite">
        <span class="ms-spinner" aria-hidden="true"></span>
        <span>載入中…</span>
      </div>
    {:else if merchants.length === 0 && !errorMessage}
      <div class="ms-empty">
        <p>目前沒有可用的商戶</p>
      </div>
    {:else}
      {#if errorMessage}
        <p class="ms-error" role="alert">{errorMessage}</p>
      {/if}
      <ul class="ms-list" role="list">
        {#each merchants as merchant (merchant.merchant_id)}
          <li>
            <button
              class="ms-item"
              class:is-selecting={selectingId === merchant.merchant_id}
              disabled={selectingId !== null}
              onclick={() => handleSelect(merchant)}
            >
              <span class="ms-item-info">
                <span class="ms-item-name">{merchant.display_name}</span>
                <span class="ms-item-sub">{merchant.name}</span>
              </span>
              <span class="ms-item-right">
                <span class="ms-item-currency">{merchant.currency}</span>
                {#if selectingId === merchant.merchant_id}
                  <span class="ms-item-spinner" aria-hidden="true"></span>
                {/if}
              </span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}

    <footer class="ms-footer">
      <button type="button" class="ms-back" onclick={handleBack}>
        返回登入
      </button>
    </footer>
  </main>
</div>
