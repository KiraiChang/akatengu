<script module lang="ts">
  export interface NewAccountForm {
    parent_id:      string | null;
    accountSuffix:  string;
    account_id:     string;
    name:           string;
    type:           string;
    normal_balance: string;
    currency:       string;
    is_summary:     boolean;
    is_active:      boolean;
  }

  export function emptyNewAccountForm(): NewAccountForm {
    return {
      parent_id:      null,
      accountSuffix:  '',
      account_id:     '',
      name:           '',
      type:           'ASSET',
      normal_balance: 'DEBIT',
      currency:       'TWD',
      is_summary:     false,
      is_active:      true,
    };
  }
</script>

<script lang="ts">
  import AccountSelect from './AccountSelect.svelte';
  import type { Account } from '../types/account';

  const ACCOUNT_TYPES = ['ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE'] as const;
  const ACCOUNT_TYPE_LABELS: Record<string, string> = {
    ASSET:     '資產',
    LIABILITY: '負債',
    EQUITY:    '權益',
    INCOME:    '收入',
    EXPENSE:   '費用',
  };

  interface Props {
    accounts:  Account[];
    form:      NewAccountForm;
    required?: boolean;
  }

  let {
    accounts,
    form = $bindable(),
    required = true,
  }: Props = $props();

  const accountParent = $derived(
    form.parent_id
      ? (accounts.find(a => a.account_id === form.parent_id) ?? null)
      : null,
  );
  const accountPrefix = $derived(accountParent?.account_id ?? '');

  $effect(() => {
    form.account_id = accountPrefix + form.accountSuffix;
  });
</script>

<div class="new-account-section">
  <div class="form-group">
    <label class="form-label" for="naf-parent-id">父類別科目</label>
    <AccountSelect
      {accounts}
      value={form.parent_id ?? ''}
      placeholder="無（頂層科目）"
      onselect={(id) => {
        form.parent_id     = id || null;
        form.accountSuffix = '';
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
      <label class="form-label" for="naf-account-id">科目編號 *</label>
      <div class="acct-id-field">
        {#if accountPrefix}
          <span class="acct-id-prefix">{accountPrefix}</span>
        {/if}
        <input
          id="naf-account-id"
          class="form-input acct-id-input"
          type="text"
          bind:value={form.accountSuffix}
          placeholder={accountPrefix ? '後綴，例：-01' : '例：1101'}
          {required}
        />
      </div>
      {#if accountPrefix && form.account_id}
        <span class="acct-id-preview">完整編號：{form.account_id}</span>
      {/if}
    </div>
    <div class="form-group">
      <label class="form-label" for="naf-name">科目名稱 *</label>
      <input
        id="naf-name"
        class="form-input"
        type="text"
        bind:value={form.name}
        placeholder="例：玉山銀行存款"
        {required}
      />
    </div>
  </div>

  <div class="form-row">
    <div class="form-group">
      <label class="form-label" for="naf-type">科目類型 *</label>
      <select id="naf-type" class="form-select" bind:value={form.type}>
        {#each ACCOUNT_TYPES as t}
          <option value={t}>{ACCOUNT_TYPE_LABELS[t]}</option>
        {/each}
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="naf-normal-balance">正常餘額 *</label>
      <select id="naf-normal-balance" class="form-select" bind:value={form.normal_balance}>
        <option value="DEBIT">借（Debit）</option>
        <option value="CREDIT">貸（Credit）</option>
      </select>
    </div>
  </div>

  <div class="form-row">
    <div class="form-group">
      <label class="form-label" for="naf-currency">幣別</label>
      <input
        id="naf-currency"
        class="form-input"
        type="text"
        bind:value={form.currency}
        placeholder="TWD"
      />
    </div>
    <div class="form-group" style="display:flex;gap:20px;padding-top:26px;">
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
</div>