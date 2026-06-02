<script lang="ts">
  import { getLedgerAccountTypeConfigs, updateLedgerAccountTypeConfig } from '../api/setting';
  import { getAccountAll } from '../api/account';
  import AccountSelect from './AccountSelect.svelte';
  import type { LedgerAccountTypeConfigResult, LedgerAccountType } from '../types/setting';
  import type { Account } from '../types/account';

  const LEDGER_TYPE_LABELS: Record<LedgerAccountType, string> = {
    BANK_ACCOUNT: '銀行帳戶',
    CREDIT_CARD:  '信用卡',
    LOAN:         '貸款',
  };
  const ALL_LEDGER_TYPES: LedgerAccountType[] = ['BANK_ACCOUNT', 'CREDIT_CARD', 'LOAN'];

  let configs   = $state<LedgerAccountTypeConfigResult[]>([]);
  let accounts  = $state<Account[]>([]);
  let isLoading = $state(false);
  let error     = $state('');

  let showModal      = $state(false);
  let isSaving       = $state(false);
  let saveError      = $state('');
  let editingType    = $state<LedgerAccountType>('BANK_ACCOUNT');
  let editAccountId  = $state('');

  const configMap    = $derived(new Map(configs.map(c => [c.type, c])));
  const accountMap   = $derived(new Map(accounts.map(a => [a.account_id, a])));

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      [configs, accounts] = await Promise.all([getLedgerAccountTypeConfigs(), getAccountAll()]);
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function openEditModal(type: LedgerAccountType): void {
    editingType   = type;
    editAccountId = configMap.get(type)?.account_id ?? '';
    saveError     = '';
    showModal     = true;
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!editAccountId) return;
    isSaving  = true;
    saveError = '';
    try {
      await updateLedgerAccountTypeConfig(editingType, editAccountId);
      showModal = false;
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : '更新失敗';
    } finally {
      isSaving = false;
    }
  }
</script>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">帳戶類型科目對應</h2>
    {#if isLoading}
      <span class="query-loading">
        <span class="spinner" aria-hidden="true"></span>
        載入中
      </span>
    {/if}
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="帳戶類型科目對應">
      <thead>
        <tr>
          <th>帳戶類型</th>
          <th>配置科目編號</th>
          <th>配置科目名稱</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && configs.length === 0}
          <tr><td colspan="4" class="table-empty">載入中...</td></tr>
        {:else}
          {#each ALL_LEDGER_TYPES as type (type)}
            {@const cfg  = configMap.get(type)}
            {@const acct = cfg?.account_id ? accountMap.get(cfg.account_id) : undefined}
            <tr>
              <td>{LEDGER_TYPE_LABELS[type]}</td>
              <td class="mono">{cfg?.account_id || '—'}</td>
              <td>{acct?.name ?? '—'}</td>
              <td>
                <button
                  class="btn-ghost"
                  style="padding:2px 10px;font-size:11px;"
                  onclick={() => openEditModal(type)}
                >編輯</button>
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
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="ledger-type-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="ledger-type-modal-title">修改帳戶類型科目對應</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        <div class="form-group">
          <span class="form-label">帳戶類型</span>
          <p style="font-size:13px;color:#dedad3;margin:0;padding:8px 0;">{LEDGER_TYPE_LABELS[editingType]}</p>
        </div>

        <div class="form-group">
          <label class="form-label" for="ledger-type-account">配置科目 *</label>
          <AccountSelect
            {accounts}
            value={editAccountId}
            placeholder="選擇會計科目…"
            onselect={(id) => { editAccountId = id; }}
          />
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !editAccountId}
          >
            {isSaving ? '儲存中…' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
