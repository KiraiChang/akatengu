<script lang="ts">
  import { getReviewItems, approveTxn, ignoreTxn, completeImport, getImportResult } from '../api/bankStatement';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { ReviewItem, ApproveRequest } from '../types/bankStatement';
  import type { BankStatementImport } from '../types/bankStatement';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';

  interface ApproveForm {
    account_id:         string;
    counter_account_id: string;
    ledger_id:          string;
    txn_date:           string;
    description:        string;
    note:               string;
  }

  interface Props { params: Record<string, string>; }
  const { params }: Props = $props();
  const importId = $derived(parseInt(params.import_id, 10));

  let importData   = $state<BankStatementImport | null>(null);
  let items        = $state<ReviewItem[]>([]);
  let allAccounts  = $state<Account[]>([]);
  let ledgers      = $state<LedgerAccount[]>([]);
  let isLoading    = $state(false);
  let error        = $state('');
  let actionMsg    = $state('');

  let approveTarget = $state<ReviewItem | null>(null);
  let approveForm   = $state<ApproveForm>({ account_id: '', counter_account_id: '', ledger_id: '', txn_date: '', description: '', note: '' });
  let isSaving      = $state(false);
  let saveError     = $state('');

  let ignoreTarget  = $state<ReviewItem | null>(null);
  let isIgnoring    = $state(false);
  let ignoreError   = $state('');

  let showComplete  = $state(false);
  let isCompleting  = $state(false);
  let completeError = $state('');

  $effect(() => { void load(); });

  async function load(): Promise<void> {
    isLoading = true; error = '';
    try {
      const [result, reviewItems, accounts, ls] = await Promise.all([
        getImportResult(importId),
        getReviewItems(importId),
        getAccountAll(),
        getLedgerAccountAll(),
      ]);
      importData  = result.import;
      items       = reviewItems;
      allAccounts = accounts;
      ledgers     = ls;
    } catch (e) {
      error = e instanceof Error ? e.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  function openApprove(item: ReviewItem): void {
    const s = item.suggested_entry;
    approveForm = {
      account_id:         s?.ledger_account_id  ?? '',
      counter_account_id: s?.counter_account_id ?? '',
      ledger_id:          s ? String(s.ledger_id) : '',
      txn_date:           item.txn_date,
      description:        item.description,
      note:               '',
    };
    saveError     = '';
    approveTarget = item;
  }

  function closeApprove(): void { approveTarget = null; }

  async function handleApprove(e: Event): Promise<void> {
    e.preventDefault();
    if (!approveTarget) return;
    isSaving = true; saveError = '';
    try {
      const req: ApproveRequest = {
        ledger_id:          parseInt(approveForm.ledger_id, 10),
        account_id:         approveForm.account_id,
        counter_account_id: approveForm.counter_account_id,
        txn_date:           approveForm.txn_date,
        description:        approveForm.description,
        note:               approveForm.note.trim() || null,
      };
      await approveTxn(importId, approveTarget.bank_txn_id, req);
      closeApprove();
      actionMsg = '已核准';
      await load();
    } catch (e) {
      saveError = e instanceof Error ? e.message : '核准失敗';
    } finally {
      isSaving = false;
    }
  }

  async function handleIgnore(): Promise<void> {
    if (!ignoreTarget) return;
    isIgnoring = true; ignoreError = '';
    try {
      await ignoreTxn(importId, ignoreTarget.bank_txn_id);
      ignoreTarget = null;
      actionMsg = '已忽略';
      await load();
    } catch (e) {
      ignoreError = e instanceof Error ? e.message : '忽略失敗';
    } finally {
      isIgnoring = false;
    }
  }

  async function handleComplete(): Promise<void> {
    if (!importData) return;
    isCompleting = true; completeError = '';
    try {
      await completeImport(importData.import_id, importData.import_uuid, importData.version);
      showComplete = false;
      actionMsg = '匯入已完成';
      await load();
    } catch (e) {
      completeError = e instanceof Error ? e.message : '完成匯入失敗';
    } finally {
      isCompleting = false;
    }
  }

  function goBack(): void { window.location.hash = `#/home/bank-statement/${importId}/result`; }

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    return isNaN(n) || n === 0 ? '—' : n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  const activeAccounts = $derived(allAccounts.filter(a => a.is_active));
  const activeLedgers  = $derived(ledgers.filter(l => l.is_active));

  const MATCH_LABELS: Record<string, string> = {
    unmatched: '未比對', matched: '已比對', approved: '已核准', ignored: '已忽略',
    UNMATCHED: '未比對', MATCHED: '已比對', APPROVED: '已核准', IGNORED: '已忽略',
  };
</script>

<div class="content-header">
  <div style="display:flex;align-items:center;gap:12px;min-width:0">
    <button class="btn-ghost" style="padding:2px 10px;font-size:12px;white-space:nowrap" onclick={goBack}>← 返回明細</button>
    <h1 class="content-title" style="margin:0;min-width:0">審查交易</h1>
  </div>
  <div style="display:flex;align-items:center;gap:8px;flex-shrink:0">
    {#if showComplete}
      {#if completeError}<span style="font-size:11px;color:#c07070">{completeError}</span>{/if}
      <span style="font-size:12px;color:#7a8096">確認完成匯入？此操作無法復原</span>
      <button class="btn-primary" onclick={handleComplete} disabled={isCompleting}>{isCompleting ? '處理中…' : '確認'}</button>
      <button class="btn-ghost" onclick={() => { showComplete = false; completeError = ''; }} disabled={isCompleting}>取消</button>
    {:else}
      <button class="btn-primary" onclick={() => showComplete = true} disabled={isLoading || !importData}>完成匯入</button>
    {/if}
  </div>
</div>

{#if actionMsg}<p class="bs-action-msg" role="status">{actionMsg}</p>{/if}
{#if error}<p class="query-error" role="alert">{error}</p>{/if}

<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="審查交易列表">
      <thead>
        <tr>
          <th>日期</th><th>說明</th><th class="num">借方</th><th class="num">貸方</th>
          <th>比對狀態</th><th>建議科目</th><th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="7" class="table-empty">載入中…</td></tr>
        {:else if items.length === 0}
          <tr><td colspan="7" class="table-empty">無待審查項目</td></tr>
        {:else}
          {#each items as item (item.bank_txn_id)}
            <tr>
              <td class="mono" style="font-size:12px">{item.txn_date}</td>
              <td style="font-size:12px;max-width:180px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title={item.description}>{item.description}</td>
              <td class="num" style="font-size:12px">{fmtAmount(item.debit)}</td>
              <td class="num" style="font-size:12px">{fmtAmount(item.credit)}</td>
              <td><span class="bs-match-status bs-match-status--{item.match_status.toLowerCase()}">{MATCH_LABELS[item.match_status] ?? item.match_status}</span></td>
              <td style="font-size:11px;color:#7a8096">{item.suggested_entry?.counter_account_id || '—'}</td>
              <td style="white-space:nowrap">
                {#if item.match_status === 'APPROVED' || item.match_status === 'approved'}
                  <span style="font-size:11px;color:#4a9070">已核准</span>
                {:else if item.match_status === 'IGNORED' || item.match_status === 'ignored'}
                  <span style="font-size:11px;color:#808080">已忽略</span>
                {:else}
                  <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openApprove(item)}>核准</button>
                  {#if ignoreTarget?.bank_txn_id === item.bank_txn_id}
                    <span style="font-size:11px;color:#c07070;margin:0 4px">確認忽略？</span>
                    <button class="btn-ghost" style="padding:2px 8px;font-size:11px;color:#c07070;border-color:rgba(192,112,112,0.4);" onclick={handleIgnore} disabled={isIgnoring}>{isIgnoring ? '…' : '確認'}</button>
                    <button class="btn-ghost" style="padding:2px 8px;font-size:11px;margin-left:4px;" onclick={() => { ignoreTarget = null; ignoreError = ''; }} disabled={isIgnoring}>取消</button>
                  {:else}
                    <button class="btn-ghost" style="padding:2px 10px;font-size:11px;margin-left:4px;" onclick={() => { ignoreTarget = item; ignoreError = ''; }}>忽略</button>
                  {/if}
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
  {#if ignoreError}<p class="query-error" style="margin-top:8px" role="alert">{ignoreError}</p>{/if}
</section>

{#if approveTarget}
  <div class="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="bs-approve-title">
    <div class="modal-panel modal-panel--wide">
      <header class="modal-header">
        <h2 class="modal-title" id="bs-approve-title">核准交易</h2>
        <button class="modal-close" onclick={closeApprove} aria-label="關閉">×</button>
      </header>
      <form class="modal-body" onsubmit={handleApprove}>
        {#if saveError}<p class="query-error" style="margin-bottom:16px" role="alert">{saveError}</p>{/if}

        <div class="bs-form-grid">
          <div class="form-group">
            <label class="form-label" for="bs-ap-date">交易日期 *</label>
            <input id="bs-ap-date" class="form-input" type="date" bind:value={approveForm.txn_date} required />
          </div>
          <div class="form-group" style="grid-column:span 2">
            <label class="form-label" for="bs-ap-desc">說明 *</label>
            <input id="bs-ap-desc" class="form-input" type="text" bind:value={approveForm.description} required />
          </div>
        </div>

        <div class="bs-form-grid">
          <div class="form-group">
            <label class="form-label" for="bs-ap-ledger">帳戶 *</label>
            <select id="bs-ap-ledger" class="form-input" bind:value={approveForm.ledger_id} required>
              <option value="">請選擇…</option>
              {#each activeLedgers as l (l.ledger_id)}
                <option value={String(l.ledger_id)}>{l.institution} {l.name}</option>
              {/each}
            </select>
          </div>
          <div class="form-group">
            <div class="form-label">銀行帳戶科目 *</div>
            <AccountSelect
              accounts={activeAccounts}
              value={approveForm.account_id}
              placeholder="選擇科目…"
              onselect={(id) => { approveForm.account_id = id; }}
            />
          </div>
          <div class="form-group">
            <div class="form-label">對應科目 *</div>
            <AccountSelect
              accounts={activeAccounts}
              value={approveForm.counter_account_id}
              placeholder="選擇對應科目…"
              onselect={(id) => { approveForm.counter_account_id = id; }}
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label" for="bs-ap-note">備註</label>
          <input id="bs-ap-note" class="form-input" type="text" placeholder="選填" bind:value={approveForm.note} />
        </div>

        <div class="je-actions">
          <button type="button" class="btn-ghost" onclick={closeApprove} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !approveForm.account_id || !approveForm.counter_account_id || !approveForm.ledger_id}
          >{isSaving ? '儲存中…' : '確認核准'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
