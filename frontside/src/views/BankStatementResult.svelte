<script lang="ts">
  import { getImportResult, autoMatch, reMatch, manualMatch } from '../api/bankStatement';
  import { getLedgerAccountAll } from '../api/ledger';
  import type { BankStatementImport, BankStatementTxn } from '../types/bankStatement';
  import type { LedgerAccount } from '../types/ledger';

  interface Props { params: Record<string, string>; }
  const { params }: Props = $props();

  const importId = $derived(parseInt(params.import_id, 10));

  let importData   = $state<BankStatementImport | null>(null);
  let transactions = $state<BankStatementTxn[]>([]);
  let ledgers      = $state<LedgerAccount[]>([]);
  let isLoading    = $state(false);
  let error        = $state('');
  let actionMsg    = $state('');

  let isAutoMatching = $state(false);
  let isReMatching   = $state(false);

  let matchingTxnId  = $state<number | null>(null);
  let manualEntryId  = $state('');
  let isManualSaving = $state(false);
  let manualError    = $state('');

  $effect(() => { void load(); });

  async function load(): Promise<void> {
    isLoading = true; error = '';
    try {
      const [res, ls] = await Promise.all([getImportResult(importId), getLedgerAccountAll()]);
      importData   = res.import;
      transactions = res.transactions;
      ledgers      = ls;
    } catch (e) {
      error = e instanceof Error ? e.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  async function handleAutoMatch(): Promise<void> {
    isAutoMatching = true; actionMsg = '';
    try {
      await autoMatch(importId);
      actionMsg = '自動比對完成，已重新整理資料';
      await load();
    } catch (e) {
      actionMsg = e instanceof Error ? e.message : '自動比對失敗';
    } finally {
      isAutoMatching = false;
    }
  }

  async function handleReMatch(): Promise<void> {
    isReMatching = true; actionMsg = '';
    try {
      await reMatch(importId);
      actionMsg = '重新比對完成，已重新整理資料';
      await load();
    } catch (e) {
      actionMsg = e instanceof Error ? e.message : '重新比對失敗';
    } finally {
      isReMatching = false;
    }
  }

  function openManualMatch(txnId: number): void {
    matchingTxnId = txnId; manualEntryId = ''; manualError = '';
  }

  function closeManualMatch(): void { matchingTxnId = null; manualError = ''; }

  async function handleManualMatch(): Promise<void> {
    if (!matchingTxnId) return;
    const entryId = parseInt(manualEntryId, 10);
    if (isNaN(entryId) || entryId <= 0) { manualError = '請輸入有效的傳票明細 ID'; return; }
    isManualSaving = true; manualError = '';
    try {
      await manualMatch(importId, matchingTxnId, entryId);
      closeManualMatch();
      await load();
    } catch (e) {
      manualError = e instanceof Error ? e.message : '手動比對失敗';
    } finally {
      isManualSaving = false;
    }
  }

  function goBack(): void { window.location.hash = '#/home/bank-statement'; }

  function fmtLedger(ledgerId: number): string {
    const l = ledgers.find(l => l.ledger_id === ledgerId);
    return l ? `${l.institution} ${l.name}` : String(ledgerId);
  }

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    if (isNaN(n) || n === 0) return '—';
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  const MATCH_LABELS: Record<string, string> = {
    unmatched: '未比對',
    matched:   '已比對',
    approved:  '已核准',
    ignored:   '已忽略',
  };

  const CONF_LABELS: Record<string, string> = {
    high:   '高',
    medium: '中',
    low:    '低',
  };
</script>

<div class="content-header">
  <div style="display:flex;align-items:center;gap:12px;min-width:0">
    <button class="btn-ghost" style="padding:2px 10px;font-size:12px;white-space:nowrap" onclick={goBack}>← 返回清單</button>
    <h1 class="content-title" style="margin:0;min-width:0">交易明細</h1>
  </div>
  <div style="display:flex;gap:8px;flex-shrink:0">
    <button
      class="btn-ghost"
      onclick={handleAutoMatch}
      disabled={isAutoMatching || isReMatching || isLoading}
    >{isAutoMatching ? '比對中…' : '自動比對'}</button>
    <button
      class="btn-ghost"
      onclick={handleReMatch}
      disabled={isAutoMatching || isReMatching || isLoading}
    >{isReMatching ? '比對中…' : '重新比對'}</button>
    <button
      class="btn-primary"
      onclick={() => { window.location.hash = `#/home/bank-statement/${importId}/review`; }}
      disabled={isLoading}
    >前往審查</button>
  </div>
</div>

{#if actionMsg}
  <p class="bs-action-msg" role="status">{actionMsg}</p>
{/if}
{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

{#if importData}
  <section class="section" style="margin-bottom:16px;padding:12px 16px">
    <div class="bs-import-meta">
      <div><span class="bs-meta-label">帳戶</span>{fmtLedger(importData.ledger_id)}</div>
      <div><span class="bs-meta-label">對帳日期</span><span class="mono">{importData.statement_date}</span></div>
      <div><span class="bs-meta-label">檔名</span>{importData.filename ?? '—'}</div>
      <div>
        <span class="bs-meta-label">狀態</span>
        <span class="bs-import-status bs-import-status--{importData.status}">{importData.status}</span>
      </div>
    </div>
  </section>
{/if}

<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="交易明細列表">
      <thead>
        <tr>
          <th>日期</th>
          <th>說明</th>
          <th class="num">借方</th>
          <th class="num">貸方</th>
          <th class="num">餘額</th>
          <th>參考編號</th>
          <th>比對狀態</th>
          <th>信心度</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="9" class="table-empty">載入中…</td></tr>
        {:else if transactions.length === 0}
          <tr><td colspan="9" class="table-empty">無交易資料</td></tr>
        {:else}
          {#each transactions as txn (txn.bank_txn_id)}
            <tr>
              <td class="mono" style="font-size:12px">{txn.txn_date}</td>
              <td
                style="font-size:12px;max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
                title={txn.description}
              >{txn.description}</td>
              <td class="num" style="font-size:12px">{fmtAmount(txn.debit)}</td>
              <td class="num" style="font-size:12px">{fmtAmount(txn.credit)}</td>
              <td class="num" style="font-size:12px;color:#7a8096">
                {txn.balance != null ? fmtAmount(txn.balance) : '—'}
              </td>
              <td class="mono" style="font-size:11px;color:#7a8096">{txn.reference_no ?? '—'}</td>
              <td>
                <span class="bs-match-status bs-match-status--{txn.match_status}">
                  {MATCH_LABELS[txn.match_status] ?? txn.match_status}
                </span>
              </td>
              <td style="font-size:11px;color:#7a8096">
                {txn.match_confidence ? (CONF_LABELS[txn.match_confidence] ?? txn.match_confidence) : '—'}
              </td>
              <td style="white-space:nowrap">
                {#if matchingTxnId === txn.bank_txn_id}
                  <div class="bs-manual-match-row">
                    <input
                      class="form-input"
                      style="width:110px;padding:2px 6px;font-size:11px;"
                      type="text"
                      inputmode="numeric"
                      placeholder="傳票明細 ID"
                      bind:value={manualEntryId}
                    />
                    <button
                      class="btn-ghost"
                      style="padding:2px 8px;font-size:11px;"
                      onclick={handleManualMatch}
                      disabled={isManualSaving}
                    >{isManualSaving ? '…' : '確認'}</button>
                    <button
                      class="btn-ghost"
                      style="padding:2px 8px;font-size:11px;"
                      onclick={closeManualMatch}
                      disabled={isManualSaving}
                    >取消</button>
                    {#if manualError}
                      <span style="font-size:11px;color:#c07070">{manualError}</span>
                    {/if}
                  </div>
                {:else if txn.match_status === 'unmatched' || txn.match_status === 'matched'}
                  <button
                    class="btn-ghost"
                    style="padding:2px 10px;font-size:11px;"
                    onclick={() => openManualMatch(txn.bank_txn_id)}
                  >手動比對</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>
