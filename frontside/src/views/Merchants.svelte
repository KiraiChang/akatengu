<script lang="ts">
  import { onMount } from 'svelte';
  import { authStore } from '../stores/auth.svelte';
  import { list, update, deactivate } from '../api/merchant';
  import type { Merchant, UpdateMerchantRequest } from '../types/merchant';

  let merchants = $state<Merchant[]>([]);
  let isLoading = $state(true);
  let pageError = $state('');

  let editTarget = $state<Merchant | null>(null);
  let editForm = $state<UpdateMerchantRequest>({ name: '', display_name: '', currency: '' });
  let isSaving = $state(false);
  let editError = $state('');

  let confirmId = $state<number | null>(null);
  let isDeactivating = $state(false);

  let isFormValid = $derived(
    editForm.name.trim().length > 0 &&
    editForm.display_name.trim().length > 0 &&
    editForm.currency.trim().length > 0
  );

  onMount(async () => {
    await loadMerchants();
  });

  async function loadMerchants(): Promise<void> {
    isLoading = true;
    pageError = '';
    try {
      merchants = await list(authStore.token);
    } catch (err) {
      pageError = err instanceof Error ? err.message : '載入商戶失敗';
    } finally {
      isLoading = false;
    }
  }

  function openEdit(merchant: Merchant): void {
    editTarget = merchant;
    editForm = { name: merchant.name, display_name: merchant.display_name, currency: merchant.currency };
    editError = '';
  }

  function closeEdit(): void {
    if (isSaving) return;
    editTarget = null;
    editError = '';
  }

  async function handleUpdate(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    if (!editTarget || !isFormValid || isSaving) return;
    isSaving = true;
    editError = '';
    try {
      await update(editTarget.merchant_id, editForm, authStore.token);
      const id = editTarget.merchant_id;
      const saved = { ...editForm };
      merchants = merchants.map(m => m.merchant_id === id ? { ...m, ...saved } : m);
      closeEdit();
    } catch (err) {
      editError = err instanceof Error ? err.message : '更新商戶失敗';
    } finally {
      isSaving = false;
    }
  }

  async function handleDeactivate(): Promise<void> {
    if (confirmId === null || isDeactivating) return;
    const id = confirmId;
    isDeactivating = true;
    try {
      await deactivate(id, authStore.token);
      merchants = merchants.map(m => m.merchant_id === id ? { ...m, status: 'INACTIVE' } : m);
      confirmId = null;
    } catch (err) {
      pageError = err instanceof Error ? err.message : '關閉商戶失敗';
      confirmId = null;
    } finally {
      isDeactivating = false;
    }
  }
</script>

<div>
  <div class="content-header">
    <h1 class="content-title">商戶管理</h1>
  </div>

  {#if isLoading}
    <div class="query-loading">
      <span class="spinner" aria-hidden="true"></span>
      <span>載入中…</span>
    </div>
  {:else}
    {#if pageError}
      <p class="query-error" role="alert">{pageError}</p>
    {/if}

    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>顯示名稱</th>
            <th>識別名稱</th>
            <th>幣別</th>
            <th>狀態</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {#if merchants.length === 0}
            <tr>
              <td colspan="5" class="table-empty">尚無商戶資料</td>
            </tr>
          {:else}
            {#each merchants as merchant (merchant.merchant_id)}
              <tr class:merchants-row-inactive={merchant.status !== 'ACTIVE'}>
                <td>{merchant.display_name}</td>
                <td class="mono">{merchant.name}</td>
                <td class="mono">{merchant.currency}</td>
                <td>
                  {#if merchant.status === 'ACTIVE'}
                    <span class="badge approved">啟用</span>
                  {:else}
                    <span class="badge-plain">停用</span>
                  {/if}
                </td>
                <td class="merchants-actions">
                  <button class="merchants-btn-edit" onclick={() => openEdit(merchant)}>
                    編輯
                  </button>
                  {#if merchant.status === 'ACTIVE'}
                    {#if confirmId === merchant.merchant_id}
                      <span class="merchants-confirm">
                        <span>確認關閉？</span>
                        <button
                          class="merchants-btn-danger"
                          disabled={isDeactivating}
                          onclick={handleDeactivate}
                        >{isDeactivating ? '處理中…' : '確認'}</button>
                        <button
                          class="merchants-btn-muted"
                          onclick={() => { confirmId = null; }}
                        >取消</button>
                      </span>
                    {:else}
                      <button
                        class="merchants-btn-deactivate"
                        onclick={() => { confirmId = merchant.merchant_id; }}
                      >關閉</button>
                    {/if}
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  {/if}
</div>

{#if editTarget}
  <div class="merchants-modal-backdrop" role="presentation" onclick={closeEdit}>
    <div
      class="merchants-modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="merchants-modal-title"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <header class="merchants-modal-header">
        <h2 id="merchants-modal-title" class="merchants-modal-title">編輯商戶</h2>
        <button class="merchants-modal-close" onclick={closeEdit} aria-label="關閉">✕</button>
      </header>

      <form onsubmit={handleUpdate}>
        <div class="merchants-field">
          <label for="edit-display-name">顯示名稱</label>
          <input
            id="edit-display-name"
            type="text"
            bind:value={editForm.display_name}
            disabled={isSaving}
            placeholder="商戶顯示名稱"
          />
        </div>
        <div class="merchants-field">
          <label for="edit-name">識別名稱</label>
          <input
            id="edit-name"
            type="text"
            bind:value={editForm.name}
            disabled={isSaving}
            placeholder="唯一識別碼"
          />
        </div>
        <div class="merchants-field">
          <label for="edit-currency">幣別</label>
          <input
            id="edit-currency"
            type="text"
            bind:value={editForm.currency}
            disabled={isSaving}
            placeholder="例：TWD、USD"
          />
        </div>

        {#if editError}
          <p class="merchants-modal-error" role="alert">{editError}</p>
        {/if}

        <div class="merchants-modal-footer">
          <button type="button" class="merchants-btn-secondary" onclick={closeEdit}>取消</button>
          <button
            type="submit"
            class="merchants-btn-primary"
            disabled={!isFormValid || isSaving}
          >{isSaving ? '儲存中…' : '儲存'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
