<script lang="ts">
  import { getSysAccounts, updateSysAccount } from '../api/sys';
  import { getAccountAll } from '../api/account';
  import AccountSelect from './AccountSelect.svelte';
  import type { SysAccount } from '../types/sys';
  import type { Account } from '../types/account';

  let sysAccounts = $state<SysAccount[]>([]);
  let accounts    = $state<Account[]>([]);
  let isLoading   = $state(false);
  let error       = $state('');

  let showModal = $state(false);
  let isSaving  = $state(false);
  let saveError = $state('');
  let editForm  = $state<SysAccount>({ sys_code: '', description: '', account_id: '' });

  const accountMap = $derived(new Map(accounts.map(a => [a.account_id, a])));

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      [sysAccounts, accounts] = await Promise.all([getSysAccounts(), getAccountAll()]);
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function openEditModal(entry: SysAccount): void {
    editForm  = { sys_code: entry.sys_code, description: entry.description, account_id: entry.account_id };
    saveError = '';
    showModal = true;
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!editForm.account_id) return;
    isSaving  = true;
    saveError = '';
    try {
      await updateSysAccount(editForm);
      showModal = false;
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : '更新失敗';
    } finally {
      isSaving = false;
    }
  }
</script>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">系統科目對應</h2>
    {#if isLoading}
      <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
    {/if}
  </header>

  {#if error}
    <p class="query-error" role="alert">{error}</p>
  {/if}

  <div class="table-wrap">
    <table class="data-table" aria-label="系統科目對應">
      <thead>
        <tr>
          <th>用途</th>
          <th class="mono">系統代碼</th>
          <th>對應科目編號</th>
          <th>對應科目名稱</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && sysAccounts.length === 0}
          <tr><td colspan="5" class="table-empty">載入中...</td></tr>
        {:else if sysAccounts.length === 0}
          <tr><td colspan="5" class="table-empty">無資料</td></tr>
        {:else}
          {#each sysAccounts as entry (entry.sys_code)}
            {@const acct = accountMap.get(entry.account_id)}
            <tr>
              <td>{entry.description}</td>
              <td class="mono">{entry.sys_code}</td>
              <td class="mono">{entry.account_id || '—'}</td>
              <td>{acct?.name ?? '—'}</td>
              <td>
                <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openEditModal(entry)}>編輯</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeModal(); }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="sys-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="sys-modal-title">修改系統科目對應</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}
        <div class="form-group">
          <span class="form-label">用途</span>
          <p style="font-size:13px;color:#dedad3;margin:0;padding:8px 0;">{editForm.description}</p>
        </div>
        <div class="form-group">
          <span class="form-label">系統代碼</span>
          <p class="mono" style="font-size:11px;color:#5c6278;margin:0;padding:8px 0;">{editForm.sys_code}</p>
        </div>
        <div class="form-group">
          <label class="form-label" for="sys-account-id">對應科目 *</label>
          <AccountSelect {accounts} value={editForm.account_id} placeholder="選擇會計科目…" onselect={(id) => { editForm.account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button type="submit" class="btn-primary" disabled={isSaving || !editForm.account_id}>{isSaving ? '儲存中…' : '儲存修改'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
