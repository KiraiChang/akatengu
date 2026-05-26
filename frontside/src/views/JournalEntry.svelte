<script lang="ts">
  import { getTransactionPaged, getEntries, createTransaction } from '../api/transaction';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import { getTemplates, getTemplate, createTemplate } from '../api/template';
  import { goToAccountAnalysis } from '../lib/navigate';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import type { Transaction, Entry } from '../types/transaction';
  import { CASH_FLOW_CATEGORIES, CASH_FLOW_CATEGORY_LABELS } from '../types/account';
  import type { Account, CashFlowCategory } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type { TransactionTemplate } from '../types/template';

  const PAGE_SIZE = 20;

  let allAccounts = $state<Account[]>([]);
  let allLedgers  = $state<LedgerAccount[]>([]);
  let page        = $state(1);
  let total       = $state(0);
  let txns        = $state<Transaction[]>([]);
  let isLoading   = $state(false);
  let loadError   = $state('');

  let expandedId     = $state<number | null>(null);
  let entriesMap     = $state(new Map<number, Entry[]>());
  let entriesLoading = $state<number | null>(null);

  let showModal = $state(false);
  let isSaving  = $state(false);
  let saveError = $state('');

  let templates        = $state<TransactionTemplate[]>([]);
  let tplPickValue     = $state('');
  let tplApplying      = $state(false);

  let showSaveTplModal = $state(false);
  let saveTplTxnId     = $state<number | null>(null);
  let saveTplName      = $state('');
  let saveTplDesc      = $state('');
  let saveTplTag       = $state('');
  let saveTplSaving    = $state(false);
  let saveTplError     = $state('');
  let saveTplDone      = $state(false);

  type StringFormLineField = 'account_id' | 'ledger_id' | 'debit' | 'credit';

  interface FormLine {
    id:                 number;
    account_id:         string;
    ledger_id:          string;
    debit:              string;
    credit:             string;
    cash_flow_category: CashFlowCategory | null;
  }

  let formDate        = $state(today());
  let formDescription = $state('');
  let formCurrency    = $state('TWD');
  let formReceiptNo   = $state('');
  let formNote        = $state('');
  let lineSeq         = 2;
  let formLines       = $state<FormLine[]>([emptyLine(), emptyLine()]);

  $effect(() => {
    void loadPage(page);
  });

  const activeAccounts = $derived(allAccounts.filter(a => a.is_active));
  const activeLedgers  = $derived(allLedgers.filter(l => l.is_active));

  $effect(() => {
    void getAccountAll().then(a => { allAccounts = a; }).catch(() => {});
    void getLedgerAccountAll().then(l => { allLedgers = l; }).catch(() => {});
  });

  async function loadPage(p: number): Promise<void> {
    isLoading = true;
    loadError = '';
    try {
      const res = await getTransactionPaged({ page: p - 1, pageSize: PAGE_SIZE });
      txns  = res.data ?? [];
      total = res.meta.total_pages ?? 0;
    } catch (err) {
      loadError = err instanceof Error ? err.message : '載入失敗';
    } finally {
      isLoading = false;
    }
  }

  async function toggleExpand(txnId: number): Promise<void> {
    if (expandedId === txnId) {
      expandedId = null;
      return;
    }
    expandedId = txnId;
    if (entriesMap.has(txnId)) return;

    entriesLoading = txnId;
    try {
      const rows = await getEntries(txnId);
      entriesMap = new Map(entriesMap).set(txnId, rows);
    } catch {
      entriesMap = new Map(entriesMap).set(txnId, []);
    } finally {
      entriesLoading = null;
    }
  }

  function today(): string {
    return new Date().toISOString().slice(0, 10);
  }

  function emptyLine(): FormLine {
    return { id: lineSeq++, account_id: '', ledger_id: '', debit: '', credit: '', cash_flow_category: null };
  }

  function addLine(): void {
    formLines = [...formLines, emptyLine()];
  }

  function removeLine(id: number): void {
    if (formLines.length <= 2) return;
    formLines = formLines.filter(l => l.id !== id);
  }

  function updateLine(id: number, field: StringFormLineField, value: string): void {
    formLines = formLines.map(l => l.id === id ? { ...l, [field]: value } : l);
  }

  function updateLineCashFlow(id: number, value: CashFlowCategory | null): void {
    formLines = formLines.map(l => l.id === id ? { ...l, cash_flow_category: value } : l);
  }

  function selectLedger(lineId: number, ledgerIdStr: string): void {
    const ledger = activeLedgers.find(l => String(l.ledger_id) === ledgerIdStr);
    const newAccountId = ledgerIdStr === '' ? '' : (ledger?.account_id ?? '');
    const acct = newAccountId ? (allAccounts.find(a => a.account_id === newAccountId) ?? null) : null;
    formLines = formLines.map(l =>
      l.id === lineId
        ? {
            ...l,
            ledger_id:          ledgerIdStr,
            account_id:         ledgerIdStr === '' ? '' : (ledger?.account_id ?? l.account_id),
            cash_flow_category: acct?.cash_flow_category ?? null,
          }
        : l,
    );
  }

  const totalDebit  = $derived(formLines.reduce((s, l) => s + (parseFloat(l.debit)  || 0), 0));
  const totalCredit = $derived(formLines.reduce((s, l) => s + (parseFloat(l.credit) || 0), 0));
  const isBalanced  = $derived(Math.abs(totalDebit - totalCredit) < 0.001 && totalDebit > 0);

  async function loadTemplates(): Promise<void> {
    try { templates = await getTemplates(); } catch { templates = []; }
  }

  async function applyTemplate(idStr: string): Promise<void> {
    tplPickValue = idStr;
    if (!idStr) { formLines = [emptyLine(), emptyLine()]; return; }
    tplApplying = true;
    try {
      const detail = await getTemplate(parseInt(idStr, 10));
      formLines = detail.entries.length === 0
        ? [emptyLine(), emptyLine()]
        : detail.entries.map(e => ({
            id:                 lineSeq++,
            account_id:         e.account_id,
            ledger_id:          e.ledger_id ? String(e.ledger_id) : '',
            debit:              parseFloat(e.debit)  > 0 ? e.debit  : '',
            credit:             parseFloat(e.credit) > 0 ? e.credit : '',
            cash_flow_category: e.cash_flow_category ?? null,
          }));
    } catch { /* leave unchanged */ } finally { tplApplying = false; }
  }

  function openSaveTplModal(txnId: number): void {
    saveTplTxnId     = txnId;
    saveTplName      = '';
    saveTplDesc      = '';
    saveTplTag       = '';
    saveTplError     = '';
    saveTplDone      = false;
    showSaveTplModal = true;
  }

  function closeSaveTplModal(): void {
    showSaveTplModal = false;
    saveTplTxnId = null;
  }

  async function handleSaveTemplate(e: Event): Promise<void> {
    e.preventDefault();
    if (!saveTplTxnId || !saveTplName.trim()) return;
    const entries = entriesMap.get(saveTplTxnId) ?? [];
    if (entries.length === 0) return;
    saveTplSaving = true;
    saveTplError  = '';
    try {
      await createTemplate({
        name:        saveTplName.trim(),
        description: saveTplDesc.trim() || null,
        tag:         saveTplTag.trim()  || null,
        entries: entries.map((entry, i) => ({
          sort_order:         i + 1,
          account_id:         entry.account_id,
          ledger_id:          entry.ledger_id ?? null,
          debit:              parseFloat(entry.debit)  || 0,
          credit:             parseFloat(entry.credit) || 0,
          note:               entry.note               || null,
          cash_flow_category: entry.cash_flow_category ?? null,
        })),
      });
      saveTplDone = true;
      void loadTemplates();
    } catch (err) {
      saveTplError = err instanceof Error ? err.message : '儲存失敗';
    } finally {
      saveTplSaving = false;
    }
  }

  function openModal(): void {
    formDate        = today();
    formDescription = '';
    formCurrency    = 'TWD';
    formReceiptNo   = '';
    formNote        = '';
    formLines       = [emptyLine(), emptyLine()];
    saveError       = '';
    tplPickValue    = '';
    showModal       = true;
    void loadTemplates();
  }

  function closeModal(): void {
    showModal = false;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isBalanced || !formDescription.trim()) return;

    const validLines = formLines.filter(
      l => l.account_id && (parseFloat(l.debit) > 0 || parseFloat(l.credit) > 0),
    );
    if (validLines.length < 2) {
      saveError = '至少需要兩筆有效分錄';
      return;
    }

    isSaving  = true;
    saveError = '';
    try {
      await createTransaction({
        transaction_date: formDate,
        description:      formDescription.trim(),
        total_amount:     totalDebit,
        currency:         formCurrency,
        receipt_no:       formReceiptNo.trim() || null,
        note:             formNote.trim() || null,
        ref_txn_id:       null,
        entries: validLines.map(l => ({
          account_id:         l.account_id,
          ledger_id:          l.ledger_id ? parseInt(l.ledger_id, 10) : null,
          debit:              parseFloat(l.debit)  || 0,
          credit:             parseFloat(l.credit) || 0,
          cash_flow_category: l.cash_flow_category,
        })),
      });
      closeModal();
      await loadPage(page);
    } catch (err) {
      saveError = err instanceof Error ? err.message : '新增失敗';
    } finally {
      isSaving = false;
    }
  }

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    if (isNaN(n)) return s;
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  function fmtNum(n: number): string {
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  const totalPages = $derived(Math.max(1, Math.ceil(total / PAGE_SIZE)));

  const STATUS_LABEL: Record<string, string> = {
    ACTIVE:     '正常',
    CORRECTED:  '已更正',
    VOIDED:     '已作廢',
    VOID_REF:   '作廢參考',
  };
</script>

<div class="content-header">
  <h1 class="content-title">傳票管理</h1>
  <button class="btn-primary" onclick={openModal}>＋ 新增分錄</button>
</div>

{#if loadError}
  <p class="query-error" role="alert">{loadError}</p>
{/if}

<section class="section">
  <div class="txn-table">
    <div class="txn-thead">
      <span>日期</span>
      <span>摘要</span>
      <span class="num">金額</span>
      <span class="txn-cell--currency">幣別</span>
      <span>狀態</span>
      <span></span>
    </div>

    {#if isLoading}
      <div class="txn-loading">載入中…</div>
    {:else if txns.length === 0}
      <div class="txn-loading">尚無傳票記錄</div>
    {:else}
      {#each txns as txn (txn.txn_id)}
        <div class="txn-row {expandedId === txn.txn_id ? 'txn-row--expanded' : ''}">
          <div
            class="txn-main"
            role="button"
            tabindex="0"
            onclick={() => toggleExpand(txn.txn_id)}
            onkeydown={(e) => e.key === 'Enter' && toggleExpand(txn.txn_id)}
          >
            <span class="txn-cell">{txn.txn_date}</span>
            <span class="txn-cell txn-desc">{txn.description}</span>
            <span class="txn-cell num">{fmtAmount(txn.total_amount)}</span>
            <span class="txn-cell txn-cell--currency">{txn.currency}</span>
            <span class="txn-cell">
              <span class="txn-status txn-status--{txn.status.toLowerCase().replace('_', '-')}">
                {STATUS_LABEL[txn.status] ?? txn.status}
              </span>
            </span>
            <span class="txn-cell txn-chevron">{expandedId === txn.txn_id ? '▲' : '▼'}</span>
          </div>

          {#if expandedId === txn.txn_id}
            <div class="txn-entries">
              {#if entriesLoading === txn.txn_id}
                <div class="txn-entries-loading">載入分錄中…</div>
              {:else}
                {@const rows = entriesMap.get(txn.txn_id) ?? []}
                {#if rows.length === 0}
                  <div class="txn-entries-loading">無分錄資料</div>
                {:else}
                  <div class="entry-header">
                    <span>會計科目</span>
                    <span>帳戶</span>
                    <span>現金流量</span>
                    <span class="num">借方</span>
                    <span class="num">貸方</span>
                    <span class="entry-cell--note">備註</span>
                    <span class="entry-cell--audit">更新者</span>
                    <span class="entry-cell--audit">更新時間</span>
                  </div>
                  {#each rows as entry (entry.entry_id)}
                    {@const acct = allAccounts.find(a => a.account_id === entry.account_id)}
                    {@const ldgr = allLedgers.find(l => l.ledger_id === entry.ledger_id)}
                    <div class="entry-row">
                      <span class="entry-cell">
                        <button
                          class="entry-stack account-link"
                          onclick={() => goToAccountAnalysis(entry.account_id)}
                        >
                          <span class="entry-stack-id">{entry.account_id}</span>
                          <span class="entry-stack-name">{acct?.name ?? ''}</span>
                        </button>
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
                      <span class="entry-cell">{entry.cash_flow_category ? CASH_FLOW_CATEGORY_LABELS[entry.cash_flow_category] : '—'}</span>
                      <span class="entry-cell num">{parseFloat(entry.debit) > 0 ? fmtAmount(entry.debit) : ''}</span>
                      <span class="entry-cell num">{parseFloat(entry.credit) > 0 ? fmtAmount(entry.credit) : ''}</span>
                      <span class="entry-cell entry-note entry-cell--note">{entry.note ?? ''}</span>
                      <span class="entry-cell entry-cell--audit">{entry.updated_by ?? '—'}</span>
                      <span class="entry-cell entry-cell--audit">{entry.updated_at ?? '—'}</span>
                    </div>
                  {/each}
                  <div class="entry-tpl-footer">
                    <button
                      type="button"
                      class="entry-tpl-btn"
                      onclick={() => openSaveTplModal(txn.txn_id)}
                    >儲存為範本</button>
                  </div>
                  {@const txnRecord = txns.find(t => t.txn_id === expandedId)}
                  {#if txnRecord?.updated_by || txnRecord?.updated_at}
                    <div style="padding:6px 24px;font-size:10px;color:#3d4258;letter-spacing:0.06em;border-top:1px solid rgba(255,255,255,0.03);">
                      傳票最後更新：{txnRecord.updated_by ?? '—'} · {txnRecord.updated_at ?? '—'}
                    </div>
                  {/if}
                {/if}
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  {#if totalPages > 1}
    <div class="pagination">
      <button
        class="page-btn"
        disabled={page <= 1}
        onclick={() => { page -= 1; }}
      >‹</button>
      <span class="page-info">{page} / {totalPages}</span>
      <button
        class="page-btn"
        disabled={page >= totalPages}
        onclick={() => { page += 1; }}
      >›</button>
    </div>
  {/if}
</section>

{#if showSaveTplModal}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal-panel">
      <header class="modal-header">
        <h2 class="modal-title">儲存為範本</h2>
        <button class="modal-close" onclick={closeSaveTplModal} aria-label="關閉">×</button>
      </header>

      {#if saveTplDone}
        <div class="modal-body">
          <p style="font-size:13px;color:#6ab88a;letter-spacing:0.06em;margin:0 0 20px;">範本已儲存成功。</p>
          <div class="je-actions">
            <button type="button" class="btn-primary" onclick={closeSaveTplModal}>關閉</button>
          </div>
        </div>
      {:else}
        <form class="modal-body" onsubmit={handleSaveTemplate}>
          {#if saveTplError}
            <p class="query-error" style="margin-bottom:16px;" role="alert">{saveTplError}</p>
          {/if}
          <div class="form-group">
            <label class="form-label" for="tpl-name">範本名稱 *</label>
            <input
              id="tpl-name"
              class="form-input"
              type="text"
              bind:value={saveTplName}
              placeholder="例：月租費用"
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="tpl-desc">說明</label>
            <input
              id="tpl-desc"
              class="form-input"
              type="text"
              bind:value={saveTplDesc}
              placeholder="選填"
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="tpl-tag">標籤</label>
            <input
              id="tpl-tag"
              class="form-input"
              type="text"
              bind:value={saveTplTag}
              placeholder="選填，例：租金"
            />
          </div>
          <div class="je-actions" style="margin-top:8px;">
            <button type="button" class="btn-ghost" onclick={closeSaveTplModal} disabled={saveTplSaving}>取消</button>
            <button
              type="submit"
              class="btn-primary"
              disabled={saveTplSaving || !saveTplName.trim()}
            >
              {saveTplSaving ? '儲存中…' : '確認儲存'}
            </button>
          </div>
        </form>
      {/if}
    </div>
  </div>
{/if}

{#if showModal}
  <div class="modal-overlay" role="dialog" aria-modal="true">
    <div class="modal-panel modal-panel--wide">
      <header class="modal-header">
        <h2 class="modal-title">新增分錄</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </header>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" style="margin-bottom:16px;" role="alert">{saveError}</p>
        {/if}

        {#if templates.length > 0}
          <div class="je-tpl-bar">
            <span class="je-tpl-bar-label">從範本載入</span>
            <select
              class="je-tpl-select"
              value={tplPickValue}
              onchange={(e) => applyTemplate((e.target as HTMLSelectElement).value)}
              disabled={tplApplying}
            >
              <option value="">— 不使用範本 —</option>
              {#each templates as tpl (tpl.id)}
                <option value={String(tpl.id)}>{tpl.name}{tpl.tag ? ` [${tpl.tag}]` : ''}</option>
              {/each}
            </select>
            {#if tplApplying}
              <span class="je-tpl-loading">載入中…</span>
            {/if}
          </div>
        {/if}

        <div class="je-form-header je-form-header--4col">
          <div class="form-group">
            <label class="form-label" for="je-date">日期 *</label>
            <input id="je-date" class="form-input" type="date" bind:value={formDate} required />
          </div>
          <div class="form-group" style="grid-column: span 2;">
            <label class="form-label" for="je-desc">摘要 *</label>
            <input
              id="je-desc"
              class="form-input"
              type="text"
              bind:value={formDescription}
              placeholder="本筆分錄說明"
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="je-currency">幣別</label>
            <input id="je-currency" class="form-input" type="text" bind:value={formCurrency} placeholder="TWD" />
          </div>
          <div class="form-group">
            <label class="form-label" for="je-receipt">收據號碼</label>
            <input id="je-receipt" class="form-input" type="text" bind:value={formReceiptNo} placeholder="選填" />
          </div>
          <div class="form-group" style="grid-column: span 3;">
            <label class="form-label" for="je-note">備註</label>
            <input id="je-note" class="form-input" type="text" bind:value={formNote} placeholder="選填" />
          </div>
        </div>

        <div class="je-lines">
          <div class="je-lines-header je-lines-header--6col">
            <span>金融帳戶</span>
            <span>會計科目 *</span>
            <span>現金流量</span>
            <span class="num">借方金額</span>
            <span class="num">貸方金額</span>
            <span></span>
          </div>

          {#each formLines as line (line.id)}
            <div class="je-line je-line--6col">
              <div class="je-line-cell">
                <LedgerSelect
                  ledgers={activeLedgers}
                  value={line.ledger_id}
                  onselect={(id) => selectLedger(line.id, id)}
                />
              </div>
              <div class="je-line-cell">
                <AccountSelect
                  accounts={activeAccounts}
                  value={line.account_id}
                  placeholder="選擇科目…"
                  onselect={(id) => {
                    updateLine(line.id, 'account_id', id);
                    const acct = activeAccounts.find(a => a.account_id === id) ?? null;
                    updateLineCashFlow(line.id, acct?.cash_flow_category ?? null);
                  }}
                />
              </div>
              <div class="je-line-cell">
                <select
                  class="je-cf-select"
                  value={line.cash_flow_category ?? ''}
                  onchange={(e) => {
                    const v = (e.target as HTMLSelectElement).value;
                    updateLineCashFlow(line.id, v ? v as CashFlowCategory : null);
                  }}
                >
                  <option value="">—</option>
                  {#each CASH_FLOW_CATEGORIES as c}
                    <option value={c}>{CASH_FLOW_CATEGORY_LABELS[c]}</option>
                  {/each}
                </select>
              </div>
              <div class="je-line-cell">
                <input
                  class="je-amount-input"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="0.00"
                  value={line.debit}
                  oninput={(e) => {
                    updateLine(line.id, 'debit', (e.target as HTMLInputElement).value);
                    if ((e.target as HTMLInputElement).value) updateLine(line.id, 'credit', '');
                  }}
                />
              </div>
              <div class="je-line-cell">
                <input
                  class="je-amount-input"
                  type="number"
                  min="0"
                  step="0.01"
                  placeholder="0.00"
                  value={line.credit}
                  oninput={(e) => {
                    updateLine(line.id, 'credit', (e.target as HTMLInputElement).value);
                    if ((e.target as HTMLInputElement).value) updateLine(line.id, 'debit', '');
                  }}
                />
              </div>
              <div class="je-line-cell" style="padding:6px 8px;">
                <button
                  type="button"
                  class="je-remove-btn"
                  onclick={() => removeLine(line.id)}
                  disabled={formLines.length <= 2}
                  aria-label="移除此行"
                >×</button>
              </div>
            </div>
          {/each}

          <div class="je-totals">
            <div class="je-total-item">
              <span class="je-total-label">借方合計</span>
              <span class="je-total-value {isBalanced ? 'balanced' : totalDebit > 0 ? 'unbalanced' : ''}">
                {fmtNum(totalDebit)}
              </span>
            </div>
            <div class="je-total-item">
              <span class="je-total-label">貸方合計</span>
              <span class="je-total-value {isBalanced ? 'balanced' : totalCredit > 0 ? 'unbalanced' : ''}">
                {fmtNum(totalCredit)}
              </span>
            </div>
          </div>
        </div>

        <button type="button" class="je-add-line-btn" onclick={addLine}>＋ 新增一行</button>

        <div class="je-actions">
          {#if !isBalanced && (totalDebit > 0 || totalCredit > 0)}
            <span class="je-balance-hint">借貸不平衡，差額：{fmtNum(Math.abs(totalDebit - totalCredit))}</span>
          {/if}
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !isBalanced || !formDescription.trim()}
          >
            {isSaving ? '儲存中…' : '確認新增'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}