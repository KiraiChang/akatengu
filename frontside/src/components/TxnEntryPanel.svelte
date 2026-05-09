<script lang="ts">
  import type { Entry } from '../types/transaction';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';

  interface Props {
    rows:     Entry[];
    loading:  boolean;
    error:    string;
    accounts: Account[];
    ledgers:  LedgerAccount[];
  }

  let { rows, loading, error, accounts, ledgers }: Props = $props();

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    if (isNaN(n)) return s;
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }
</script>

{#if loading}
  <div class="txn-entries-loading">載入分錄中…</div>
{:else if error}
  <div class="txn-entries-loading" style="color:var(--color-error,#c0392b);">{error}</div>
{:else if rows.length === 0}
  <div class="txn-entries-loading">無分錄資料</div>
{:else}
  <div class="entry-header">
    <span>會計科目</span>
    <span>帳戶</span>
    <span class="num">借方</span>
    <span class="num">貸方</span>
    <span>備註</span>
  </div>
  {#each rows as entry (entry.entry_id)}
    {@const acct = accounts.find(a => a.account_id === entry.account_id)}
    {@const ldgr = ledgers.find(l => l.ledger_id === entry.ledger_id)}
    <div class="entry-row">
      <span class="entry-cell">
        <span class="entry-stack">
          <span class="entry-stack-id">{entry.account_id}</span>
          <span class="entry-stack-name">{acct?.name ?? ''}</span>
        </span>
      </span>
      <span class="entry-cell entry-ledger">
        {#if ldgr}
          <span class="entry-stack">
            <span class="entry-stack-name">{ldgr.institution}</span>
            <span class="entry-stack-id">{ldgr.name}</span>
          </span>
        {:else}
          —
        {/if}
      </span>
      <span class="entry-cell num">{parseFloat(entry.debit)  > 0 ? fmtAmount(entry.debit)  : ''}</span>
      <span class="entry-cell num">{parseFloat(entry.credit) > 0 ? fmtAmount(entry.credit) : ''}</span>
      <span class="entry-cell entry-note">{entry.note ?? ''}</span>
    </div>
  {/each}
{/if}