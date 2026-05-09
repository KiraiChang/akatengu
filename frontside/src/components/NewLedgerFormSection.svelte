<script module lang="ts">
  import type { LedgerAccountType } from '../types/ledger';

  export interface NewLedgerForm {
    institution:      string;
    name:             string;
    type:             LedgerAccountType;
    account_no:       string | null;
    currency:         string;
    creditLimitInput: string | null;
    billing_day:      string | null;
    due_day:          string | null;
    is_active:        boolean;
  }

  export function emptyNewLedgerForm(): NewLedgerForm {
    return {
      institution:      '',
      name:             '',
      type:             'BANK_ACCOUNT',
      account_no:       '',
      currency:         'TWD',
      creditLimitInput: '',
      billing_day:      '',
      due_day:          '',
      is_active:        true,
    };
  }
</script>

<script lang="ts">
  import AccountSelect from './AccountSelect.svelte';
  import NewAccountFormSection, { type NewAccountForm, emptyNewAccountForm } from './NewAccountFormSection.svelte';
  import type { Account } from '../types/account';

  const LEDGER_TYPE_LABELS: Record<LedgerAccountType, string> = {
    BANK_ACCOUNT: '銀行帳戶',
    CREDIT_CARD:  '信用卡',
    LOAN:         '貸款',
  };

  interface Props {
    accounts:         Account[];
    form:             NewLedgerForm;
    required?:        boolean;
    createNewAccount: boolean;
    newAccountForm:   NewAccountForm;
    accountId:        string;
  }

  let {
    accounts,
    form             = $bindable(),
    required         = true,
    createNewAccount = $bindable(),
    newAccountForm   = $bindable(),
    accountId        = $bindable(),
  }: Props = $props();
</script>

<div class="new-account-section">
  <div class="form-row">
    <div class="form-group">
      <label class="form-label" for="nl-institution">金融機構 *</label>
      <input
        id="nl-institution"
        class="form-input"
        type="text"
        bind:value={form.institution}
        placeholder="例：玉山銀行"
        required={required}
      />
    </div>
    <div class="form-group">
      <label class="form-label" for="nl-name">帳戶名稱 *</label>
      <input
        id="nl-name"
        class="form-input"
        type="text"
        bind:value={form.name}
        placeholder="例：玉山信用卡"
        required={required}
      />
    </div>
  </div>

  <div class="form-row">
    <div class="form-group">
      <label class="form-label" for="nl-type">帳戶類型 *</label>
      <select id="nl-type" class="form-select" bind:value={form.type}>
        {#each Object.entries(LEDGER_TYPE_LABELS) as [val, label]}
          <option value={val}>{label}</option>
        {/each}
      </select>
    </div>
    <div class="form-group">
      <label class="form-label" for="nl-account-no">帳號 ／ 卡號後四碼</label>
      <input
        id="nl-account-no"
        class="form-input"
        type="text"
        value={form.account_no ?? ''}
        oninput={(e) => { form.account_no = (e.target as HTMLInputElement).value || null; }}
        placeholder="選填"
      />
    </div>
  </div>

  {#if form.type === 'CREDIT_CARD'}
    <div class="form-row">
      <div class="form-group">
        <label class="form-label" for="nl-credit-limit">信用額度</label>
        <input
          id="nl-credit-limit"
          class="form-input"
          type="number"
          min="0"
          step="1"
          value={form.creditLimitInput ?? ''}
          oninput={(e) => { form.creditLimitInput = (e.target as HTMLInputElement).value || null; }}
          placeholder="例：100000"
        />
      </div>
      <div class="form-group"></div>
    </div>
    <div class="form-row">
      <div class="form-group">
        <label class="form-label" for="nl-billing-day">帳單截止日（日）</label>
        <input
          id="nl-billing-day"
          class="form-input"
          type="number"
          min="1"
          max="31"
          value={form.billing_day ?? ''}
          oninput={(e) => { form.billing_day = (e.target as HTMLInputElement).value || null; }}
          placeholder="例：25"
        />
      </div>
      <div class="form-group">
        <label class="form-label" for="nl-due-day">繳費截止日（日）</label>
        <input
          id="nl-due-day"
          class="form-input"
          type="number"
          min="1"
          max="31"
          value={form.due_day ?? ''}
          oninput={(e) => { form.due_day = (e.target as HTMLInputElement).value || null; }}
          placeholder="例：15"
        />
      </div>
    </div>
  {/if}

  <div class="form-group" style="margin-bottom:4px;margin-top:4px;">
    <span class="form-label">關聯會計科目 *</span>
    <div class="acct-mode-toggle">
      <button
        type="button"
        class="acct-mode-btn"
        class:acct-mode-btn--active={!createNewAccount}
        onclick={() => { createNewAccount = false; }}
      >選擇現有科目</button>
      <button
        type="button"
        class="acct-mode-btn"
        class:acct-mode-btn--active={createNewAccount}
        onclick={() => { createNewAccount = true; newAccountForm = emptyNewAccountForm(); }}
      >＋ 新增科目</button>
    </div>
  </div>

  {#if !createNewAccount}
    <div class="form-row">
      <div class="form-group">
        <AccountSelect
          {accounts}
          value={accountId}
          placeholder="搜尋並選擇科目…"
          onselect={(id) => { accountId = id; }}
        />
      </div>
      <div class="form-group">
        <label class="form-label" for="nl-currency">幣別</label>
        <input
          id="nl-currency"
          class="form-input"
          type="text"
          bind:value={form.currency}
          placeholder="TWD"
        />
      </div>
    </div>
  {:else}
    <NewAccountFormSection
      {accounts}
      bind:form={newAccountForm}
      required={createNewAccount}
    />
    <div class="form-group" style="margin-bottom:0;margin-top:4px;">
      <label class="form-label" for="nl-currency-new">帳戶幣別</label>
      <input
        id="nl-currency-new"
        class="form-input"
        type="text"
        bind:value={form.currency}
        placeholder="TWD"
      />
    </div>
  {/if}

  <div class="form-group" style="display:flex;align-items:center;gap:8px;padding-top:8px;margin-bottom:0;">
    <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
      <input type="checkbox" bind:checked={form.is_active} />
      啟用此帳戶
    </label>
  </div>
</div>