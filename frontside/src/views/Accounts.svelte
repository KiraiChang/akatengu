<script lang="ts">
  import { getAccountAll, createAccount, updateAccount } from '../api/account';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { Account, UpdateAccountRequest } from '../types/account';

  const ACCOUNT_TYPES = ['ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE'] as const;
  const ACCOUNT_TYPE_LABELS: Record<string, string> = {
    ASSET:     '資產',
    LIABILITY: '負債',
    EQUITY:    '權益',
    INCOME:    '收入',
    EXPENSE:   '費用',
  };

  const NORMAL_BALANCE_LABELS: Record<string, string> = {
    DEBIT:  '借',
    CREDIT: '貸',
  };

  let accounts    = $state<Account[]>([]);
  let isLoading   = $state(false);
  let error       = $state('');
  let expandedIds = $state(new Set<string>());

  const rootAccounts = $derived(accounts.filter(a => a.parent_id === null));

  const childrenMap = $derived((() => {
    const map = new Map<string, Account[]>();
    for (const a of accounts) {
      if (a.parent_id !== null) {
        const list = map.get(a.parent_id) ?? [];
        list.push(a);
        map.set(a.parent_id, list);
      }
    }
    return map;
  })());

  // modal state
  let showModal     = $state(false);
  let mode          = $state<'create' | 'edit'>('create');
  let isSaving      = $state(false);
  let saveError     = $state('');
  let accountSuffix = $state('');
  let form          = $state<UpdateAccountRequest>({
    account_id:     '',
    parent_id:      null,
    name:           '',
    type:           'ASSET',
    normal_balance: 'DEBIT',
    currency:       'TWD',
    is_summary:     false,
    is_active:      true,
    note:           null,
    version:        0,
  });

  const parentAccount = $derived(
    form.parent_id ? (accounts.find(a => a.account_id === form.parent_id) ?? null) : null
  );
  const accountPrefix = $derived(parentAccount?.account_id ?? '');

  $effect(() => {
    if (mode === 'create') {
      form.account_id = accountPrefix + accountSuffix;
    }
  });

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      accounts = await getAccountAll();
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function toggleExpand(accountId: string): void {
    const next = new Set(expandedIds);
    if (next.has(accountId)) {
      next.delete(accountId);
    } else {
      next.add(accountId);
    }
    expandedIds = next;
  }

  async function openModal(): Promise<void> {
    mode          = 'create';
    saveError     = '';
    accountSuffix = '';
    form = {
      account_id:     '',
      parent_id:      null,
      name:           '',
      type:           'ASSET',
      normal_balance: 'DEBIT',
      currency:       'TWD',
      is_summary:     false,
      is_active:      true,
      note:           null,
      version:        0,
    };
    showModal = true;
  }

  async function openEditModal(account: Account): Promise<void> {
    mode          = 'edit';
    saveError     = '';
    accountSuffix = '';
    form = {
      account_id:     account.account_id,
      parent_id:      account.parent_id,
      name:           account.name,
      type:           account.type,
      normal_balance: account.normal_balance,
      currency:       account.currency,
      is_summary:     account.is_summary,
      is_active:      account.is_active,
      note:           account.note,
      version:        account.version,
    };
    showModal = true;
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!form.account_id.trim() || !form.name.trim()) return;
    isSaving  = true;
    saveError = '';
    try {
      if (mode === 'create') {
        const { version: _v, ...createPayload } = form;
        await createAccount({
          ...createPayload,
          parent_id: form.parent_id || null,
          note:      form.note?.trim() || null,
        });
      } else {
        await updateAccount({
          ...form,
          parent_id: form.parent_id || null,
          note:      form.note?.trim() || null,
        });
      }
      showModal = false;
      await load();
    } catch (err) {
      saveError = err instanceof Error ? err.message : mode === 'create' ? '新增失敗' : '修改失敗';
    } finally {
      isSaving = false;
    }
  }

  function labelOf(map: Record<string, string>, key: string): string {
    return map[key] ?? key;
  }
</script>

{#snippet accountRow(account: Account, depth: number)}
  <tr
    class:expandable={account.has_child}
    onclick={account.has_child ? () => toggleExpand(account.account_id) : undefined}
    aria-expanded={account.has_child ? expandedIds.has(account.account_id) : undefined}
  >
    <td class="mono">
      <span class="indent" style:padding-left="{depth * 20}px">
        {#if account.has_child}
          <span class="expand-icon" aria-hidden="true">
            {expandedIds.has(account.account_id) ? '▼' : '▶'}
          </span>
        {:else}
          <span class="expand-icon expand-icon--empty" aria-hidden="true"></span>
        {/if}
        {account.account_id}
      </span>
    </td>
    <td>{account.name}</td>
    <td>{labelOf(ACCOUNT_TYPE_LABELS, account.type)}</td>
    <td>{labelOf(NORMAL_BALANCE_LABELS, account.normal_balance)}</td>
    <td>{account.currency}</td>
    <td>
      {#if account.is_summary}
        <span class="badge approved">是</span>
      {:else}
        <span class="badge-plain">否</span>
      {/if}
    </td>
    <td>
      {#if account.is_active}
        <span class="badge approved">啟用</span>
      {:else}
        <span class="badge pending">停用</span>
      {/if}
    </td>
    <td class="note-cell">{account.note ?? '—'}</td>
    <td onclick={(e) => e.stopPropagation()}>
      <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openEditModal(account)}>編輯</button>
    </td>
  </tr>
  {#if expandedIds.has(account.account_id)}
    {#each childrenMap.get(account.account_id) ?? [] as child (child.account_id)}
      {@render accountRow(child, depth + 1)}
    {/each}
  {/if}
{/snippet}

<div class="content-header">
  <h1 class="content-title">科目查詢</h1>
  <span class="content-date">共 {accounts.length} 筆</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">科目列表</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading">
          <span class="spinner" aria-hidden="true"></span>
          載入中
        </span>
      {/if}
      <button class="section-action" onclick={openModal}>＋ 新增科目</button>
    </div>
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="科目列表">
      <thead>
        <tr>
          <th>科目編號</th>
          <th>科目名稱</th>
          <th>類型</th>
          <th>正常餘額</th>
          <th>幣別</th>
          <th>摘要科目</th>
          <th>狀態</th>
          <th>備註</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && accounts.length === 0}
          <tr><td colspan="9" class="table-empty">載入中...</td></tr>
        {:else if accounts.length === 0}
          <tr><td colspan="9" class="table-empty">無資料</td></tr>
        {:else}
          {#each rootAccounts as account (account.account_id)}
            {@render accountRow(account, 0)}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<!-- ── 新增科目 Modal ─────────────────────── -->
{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeModal(); }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="modal-title">{mode === 'create' ? '新增會計科目' : '修改會計科目'}</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        <div class="form-group">
          <!-- svelte-ignore a11y_label_has_associated_control -->
          <label class="form-label">父科目</label>
          <AccountSelect
            accounts={accounts}
            value={form.parent_id ?? ''}
            placeholder="無（頂層科目）"
            onselect={(id) => {
              form.parent_id = id || null;
              const parent = accounts.find(a => a.account_id === id) ?? null;
              if (parent) {
                form.type           = parent.type;
                form.normal_balance = parent.normal_balance;
                form.currency       = parent.currency;
              }
            }}
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-account-id">科目編號 *</label>
            {#if mode === 'edit'}
              <input id="f-account-id" class="form-input" type="text" value={form.account_id} readonly style="background:#f5f0e8;cursor:default;" />
            {:else}
              <div class="acct-id-field">
                {#if accountPrefix}
                  <span class="acct-id-prefix">{accountPrefix}</span>
                {/if}
                <input
                  id="f-account-id"
                  class="form-input acct-id-input"
                  type="text"
                  bind:value={accountSuffix}
                  placeholder={accountPrefix ? '後綴，例：-01' : '例：1101'}
                  required
                />
              </div>
              {#if accountPrefix}
                <span class="acct-id-preview">完整編號：{form.account_id || '—'}</span>
              {/if}
            {/if}
          </div>
          <div class="form-group">
            <label class="form-label" for="f-name">科目名稱 *</label>
            <input
              id="f-name"
              class="form-input"
              type="text"
              bind:value={form.name}
              placeholder="例：現金"
              required
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-type">科目類型 *</label>
            <select id="f-type" class="form-select" bind:value={form.type}>
              {#each ACCOUNT_TYPES as t}
                <option value={t}>{ACCOUNT_TYPE_LABELS[t]}</option>
              {/each}
            </select>
          </div>
          <div class="form-group">
            <label class="form-label" for="f-balance">正常餘額 *</label>
            <select id="f-balance" class="form-select" bind:value={form.normal_balance}>
              <option value="DEBIT">借（Debit）</option>
              <option value="CREDIT">貸（Credit）</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-currency">幣別</label>
            <input
              id="f-currency"
              class="form-input"
              type="text"
              bind:value={form.currency}
              placeholder="TWD"
            />
          </div>
          <div class="form-group" style="display:flex;gap:28px;padding-top:26px;">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
              <input type="checkbox" bind:checked={form.is_summary} />
              摘要科目
            </label>
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
              <input type="checkbox" bind:checked={form.is_active} />
              啟用
            </label>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label" for="f-note">備註</label>
          <input
            id="f-note"
            class="form-input"
            type="text"
            value={form.note ?? ''}
            oninput={(e) => { form.note = (e.target as HTMLInputElement).value || null; }}
            placeholder="選填"
          />
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !form.account_id.trim() || !form.name.trim()}
          >
            {isSaving ? '儲存中…' : mode === 'create' ? '新增科目' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}