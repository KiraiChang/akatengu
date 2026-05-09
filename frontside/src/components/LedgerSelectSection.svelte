<script lang="ts">
  import type { LedgerAccount } from '../types/ledger';
  import type { Account } from '../types/account';
  import LedgerSelect from './LedgerSelect.svelte';
  import NewLedgerFormSection, { type NewLedgerForm, emptyNewLedgerForm } from './NewLedgerFormSection.svelte';
  import { type NewAccountForm, emptyNewAccountForm } from './NewAccountFormSection.svelte';

  interface Props {
    ledgers:          LedgerAccount[];
    accounts:         Account[];
    ledgerId:         string;
    createNew:        boolean;
    newLedgerForm:    NewLedgerForm;
    createNewAccount: boolean;
    newAccountForm:   NewAccountForm;
    accountId:        string;
    label?:           string;
  }

  let {
    ledgers,
    accounts,
    ledgerId         = $bindable(),
    createNew        = $bindable(),
    newLedgerForm    = $bindable(),
    createNewAccount = $bindable(),
    newAccountForm   = $bindable(),
    accountId        = $bindable(),
    label            = '帳戶',
  }: Props = $props();
</script>

<div class="form-group" style="margin-bottom:4px;">
  <span class="form-label">{label} *</span>
  <div class="acct-mode-toggle">
    <button
      type="button"
      class="acct-mode-btn"
      class:acct-mode-btn--active={!createNew}
      onclick={() => { createNew = false; }}
    >選擇現有帳戶</button>
    <button
      type="button"
      class="acct-mode-btn"
      class:acct-mode-btn--active={createNew}
      onclick={() => {
        createNew        = true;
        newLedgerForm    = emptyNewLedgerForm();
        accountId        = '';
        createNewAccount = false;
        newAccountForm   = emptyNewAccountForm();
      }}
    >＋ 新增帳戶</button>
  </div>
</div>

{#if !createNew}
  <div class="form-group">
    <LedgerSelect {ledgers} value={ledgerId} onselect={(id) => { ledgerId = id; }} />
  </div>
{:else}
  <NewLedgerFormSection
    {accounts}
    bind:form={newLedgerForm}
    required={true}
    bind:createNewAccount
    bind:newAccountForm
    bind:accountId
  />
{/if}