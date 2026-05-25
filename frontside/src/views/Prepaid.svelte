<script lang="ts">
  import { getAllPrepaids, getPrepaidAmortizations, createPrepaid, amortizePrepaid, disposePrepaid } from '../api/prepaid';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import { getEntries } from '../api/transaction';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import TxnEntryPanel from '../components/TxnEntryPanel.svelte';
  import type { Prepaid, PrepaidAmortization, PrepaidStatus } from '../types/prepaid';
  import type { Entry } from '../types/transaction';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';

  const STATUS_LABELS: Record<PrepaidStatus, string> = {
    ACTIVE:    '進行中',
    COMPLETED: '已完成',
    DISPOSED:  '已終止',
  };

  const STATUS_CSS: Record<PrepaidStatus, string> = {
    ACTIVE:    'pp-status-active',
    COMPLETED: 'pp-status-completed',
    DISPOSED:  'pp-status-disposed',
  };

  // ── 列表 ──
  let prepaids  = $state<Prepaid[]>([]);
  let isLoading = $state(false);
  let error     = $state('');

  // ── 攤提明細展開 ──
  interface AmortState {
    items:   PrepaidAmortization[];
    loading: boolean;
    error:   string;
  }

  let expandedId = $state<number | null>(null);
  let amortMap   = $state(new Map<number, AmortState>());

  // ── 分錄展開 ──
  interface EntriesState {
    rows:    Entry[];
    loading: boolean;
    error:   string;
  }

  let expandedMainEntryId  = $state<number | null>(null);
  let mainEntriesMap       = $state(new Map<number, EntriesState>());
  let expandedAmortEntryId = $state<number | null>(null);
  let amortEntriesMap      = $state(new Map<number, EntriesState>());

  // ── 共用資料 ──
  let accounts = $state<Account[]>([]);
  let ledgers  = $state<LedgerAccount[]>([]);
  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  // ── 新增預付費用 Modal ──
  interface CreateForm {
    name:               string;
    account_id:         string;
    expense_account_id: string;
    ledger_id:          string;
    total_amount:       string;
    periods:            string;
    start_date:         string;
    memo:               string;
    note:               string;
  }

  let showCreateModal = $state(false);
  let isCreating      = $state(false);
  let createError     = $state('');
  let createForm      = $state<CreateForm>(emptyCreateForm());

  const isCreateValid = $derived(
    createForm.name.trim()              !== '' &&
    createForm.account_id               !== '' &&
    createForm.expense_account_id       !== '' &&
    createForm.ledger_id                !== '' &&
    parseFloat(createForm.total_amount) > 0 &&
    parseInt(createForm.periods, 10)    > 0 &&
    createForm.start_date               !== ''
  );

  // ── 攤提 Modal ──
  let showAmortizeModal = $state(false);
  let isAmortizing      = $state(false);
  let amortizeError     = $state('');
  let amortizePrepaidId = $state(0);
  let amortizeDate      = $state('');

  // ── 終止 Modal ──
  let showDisposeModal = $state(false);
  let isDisposing      = $state(false);
  let disposeError     = $state('');
  let disposePrepaidId = $state(0);
  let disposeDate      = $state('');
  let disposeMemo      = $state('');

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      prepaids = await getAllPrepaids();
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  async function toggleAmort(pp: Prepaid): Promise<void> {
    if (expandedId === pp.id) { expandedId = null; return; }
    expandedId = pp.id;
    if (amortMap.has(pp.id)) return;
    amortMap = new Map(amortMap).set(pp.id, { items: [], loading: true, error: '' });
    try {
      const items = await getPrepaidAmortizations(pp.id);
      amortMap = new Map(amortMap).set(pp.id, { items, loading: false, error: '' });
    } catch (err) {
      amortMap = new Map(amortMap).set(pp.id, {
        items: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function ensureAccounts(): Promise<void> {
    if (accounts.length === 0) { try { accounts = await getAccountAll(); } catch { /* 非致命 */ } }
  }

  async function ensureLedgers(): Promise<void> {
    if (ledgers.length === 0) { try { ledgers = await getLedgerAccountAll(); } catch { /* 非致命 */ } }
  }

  async function toggleMainEntry(pp: Prepaid, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (!pp.txn_id) return;
    const txnId = pp.txn_id;
    if (expandedMainEntryId === pp.id) { expandedMainEntryId = null; return; }
    expandedMainEntryId = pp.id;
    if (mainEntriesMap.has(pp.id)) return;
    mainEntriesMap = new Map(mainEntriesMap).set(pp.id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      mainEntriesMap = new Map(mainEntriesMap).set(pp.id, { rows, loading: false, error: '' });
    } catch (err) {
      mainEntriesMap = new Map(mainEntriesMap).set(pp.id, {
        rows: [], loading: false, error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function toggleAmortEntry(amort: PrepaidAmortization, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (expandedAmortEntryId === amort.id) { expandedAmortEntryId = null; return; }
    expandedAmortEntryId = amort.id;
    if (amortEntriesMap.has(amort.id)) return;
    amortEntriesMap = new Map(amortEntriesMap).set(amort.id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(amort.txn_id);
      amortEntriesMap = new Map(amortEntriesMap).set(amort.id, { rows, loading: false, error: '' });
    } catch (err) {
      amortEntriesMap = new Map(amortEntriesMap).set(amort.id, {
        rows: [], loading: false, error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function openCreateModal(): Promise<void> {
    createForm  = emptyCreateForm();
    createError = '';
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    showCreateModal = true;
  }

  async function handleCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isCreateValid) return;
    isCreating  = true;
    createError = '';
    try {
      await createPrepaid({
        name:               createForm.name.trim(),
        account_id:         createForm.account_id,
        expense_account_id: createForm.expense_account_id,
        ledger_id:          parseInt(createForm.ledger_id, 10),
        total_amount:       createForm.total_amount,
        periods:            parseInt(createForm.periods, 10),
        start_date:         createForm.start_date,
        memo:               createForm.memo.trim(),
        note:               createForm.note.trim(),
      });
      showCreateModal = false;
      await load();
    } catch (err) {
      createError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreating = false;
    }
  }

  function openAmortizeModal(pp: Prepaid, e: MouseEvent): void {
    e.stopPropagation();
    amortizePrepaidId = pp.id;
    amortizeDate      = '';
    amortizeError     = '';
    showAmortizeModal  = true;
  }

  async function handleAmortize(e: Event): Promise<void> {
    e.preventDefault();
    if (!amortizeDate) return;
    isAmortizing  = true;
    amortizeError = '';
    try {
      await amortizePrepaid({ prepaid_id: amortizePrepaidId, period_date: amortizeDate });
      showAmortizeModal = false;
      const next = new Map(amortMap);
      next.delete(amortizePrepaidId);
      amortMap = next;
      await load();
    } catch (err) {
      amortizeError = err instanceof Error ? err.message : '攤提失敗，請稍後再試。';
    } finally {
      isAmortizing = false;
    }
  }

  function openDisposeModal(pp: Prepaid, e: MouseEvent): void {
    e.stopPropagation();
    disposePrepaidId = pp.id;
    disposeDate      = '';
    disposeMemo      = '';
    disposeError     = '';
    showDisposeModal  = true;
  }

  async function handleDispose(e: Event): Promise<void> {
    e.preventDefault();
    if (!disposeDate) return;
    isDisposing  = true;
    disposeError = '';
    try {
      await disposePrepaid({ prepaid_id: disposePrepaidId, disposal_date: disposeDate, memo: disposeMemo.trim() });
      showDisposeModal = false;
      await load();
    } catch (err) {
      disposeError = err instanceof Error ? err.message : '終止失敗，請稍後再試。';
    } finally {
      isDisposing = false;
    }
  }

  function emptyCreateForm(): CreateForm {
    return { name: '', account_id: '', expense_account_id: '', ledger_id: '', total_amount: '', periods: '', start_date: '', memo: '', note: '' };
  }

  function fmtAmt(val: string): string {
    const n = parseFloat(val);
    return isNaN(n) ? val : n.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function progressPct(pp: Prepaid): number {
    return pp.periods === 0 ? 0 : Math.round((pp.amortized_periods / pp.periods) * 100);
  }
</script>

<div class="content-header">
  <h1 class="content-title">預付費用</h1>
  <span class="content-date">共 {prepaids.length} 筆</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

{#snippet amortContent(pp: Prepaid)}
  {@const state = amortMap.get(pp.id)}
  {#if state?.loading}
    <span class="pp-sub-msg">查詢中…</span>
  {:else if state?.error}
    <span class="pp-sub-msg pp-sub-error">{state.error}</span>
  {:else if state && state.items.length > 0}
    <div class="pp-amort-table">
      <div class="pp-amort-header">
        <span>攤提日</span>
        <span class="num">金額</span>
        <span>更新者</span>
        <span>更新時間</span>
        <span></span>
      </div>
      {#each state.items as amort (amort.id)}
        {@const isEntryExpanded = expandedAmortEntryId === amort.id}
        {@const entryState = amortEntriesMap.get(amort.id)}
        <div class="pp-amort-item">
          <span class="mono">{amort.period_date}</span>
          <span class="num mono">{fmtAmt(amort.amount)}</span>
          <span>{amort.updated_by ?? '—'}</span>
          <span>{amort.updated_at ?? '—'}</span>
          <span>
            <button
              class="btn-ghost"
              style="padding:2px 8px;font-size:10px;"
              class:acct-mode-btn--active={isEntryExpanded}
              onclick={(e) => toggleAmortEntry(amort, e)}
            >分錄</button>
          </span>
        </div>
        {#if isEntryExpanded}
          <div class="txn-entries" style="padding:8px 16px 12px;">
            <TxnEntryPanel rows={entryState?.rows ?? []} loading={entryState?.loading ?? false} error={entryState?.error ?? ''} {accounts} {ledgers} />
          </div>
        {/if}
      {/each}
    </div>
  {:else}
    <span class="pp-sub-msg">尚無攤提紀錄</span>
  {/if}
{/snippet}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">預付費用主檔</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openCreateModal}>＋ 新增預付費用</button>
    </div>
  </header>

  <div class="table-wrap pp-table-wrap">
    <table class="data-table" aria-label="預付費用主檔">
      <thead>
        <tr>
          <th>名稱</th>
          <th class="pp-amount">總金額</th>
          <th class="pp-amount">已攤提</th>
          <th>進度</th>
          <th>開始日</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && prepaids.length === 0}
          <tr><td colspan="8" class="table-empty">載入中...</td></tr>
        {:else if prepaids.length === 0}
          <tr><td colspan="8" class="table-empty">無資料</td></tr>
        {:else}
          {#each prepaids as pp (pp.id)}
            {@const isExpanded = expandedId === pp.id}
            {@const isMainExpanded = expandedMainEntryId === pp.id}
            {@const mainEntry = mainEntriesMap.get(pp.id)}
            <tr class:pp-row-expanded={isExpanded} style="cursor:pointer;" onclick={() => toggleAmort(pp)}>
              <td><span class="pp-expand-icon">{isExpanded ? '▼' : '▶'}</span>{pp.name}</td>
              <td class="pp-amount mono">{fmtAmt(pp.total_amount)}</td>
              <td class="pp-amount mono">{fmtAmt(pp.amortized_amount)}</td>
              <td>
                <div class="inst-progress">
                  <div class="inst-progress-bar"><div class="inst-progress-fill" style="width:{progressPct(pp)}%"></div></div>
                  <span class="inst-progress-text">{pp.amortized_periods}/{pp.periods}</span>
                </div>
              </td>
              <td class="mono">{pp.start_date}</td>
              <td class="hidden md:table-cell">{pp.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{pp.updated_at ?? '—'}</td>
              <td onclick={(e) => e.stopPropagation()}>
                <span class="badge {STATUS_CSS[pp.status]}">{STATUS_LABELS[pp.status]}</span>
                {#if pp.status === 'ACTIVE'}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={(e) => openAmortizeModal(pp, e)}>攤提</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={(e) => openDisposeModal(pp, e)}>終止</button>
                {/if}
                {#if pp.txn_id !== null}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" class:acct-mode-btn--active={isMainExpanded} onclick={(e) => toggleMainEntry(pp, e)}>分錄</button>
                {/if}
              </td>
            </tr>
            {#if isMainExpanded}
              <tr><td colspan="8" class="pp-sub-cell">
                <div class="txn-entries"><TxnEntryPanel rows={mainEntry?.rows ?? []} loading={mainEntry?.loading ?? false} error={mainEntry?.error ?? ''} {accounts} {ledgers} /></div>
              </td></tr>
            {/if}
            {#if isExpanded}
              <tr class="pp-sub-row"><td colspan="8" class="pp-sub-cell">{@render amortContent(pp)}</td></tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="pp-card-list">
    {#if isLoading && prepaids.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if prepaids.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each prepaids as pp (pp.id)}
        {@const isExpanded = expandedId === pp.id}
        <div class="pp-card {isExpanded ? 'pp-card-expanded' : ''}">
          <div class="pp-card-main" role="button" tabindex="0"
            onclick={() => toggleAmort(pp)}
            onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && toggleAmort(pp)}
            style="cursor:pointer;"
          >
            <div class="pp-card-head">
              <span class="pp-card-title"><span class="pp-expand-icon">{isExpanded ? '▼' : '▶'}</span>{pp.name}</span>
              <span class="badge {STATUS_CSS[pp.status]}">{STATUS_LABELS[pp.status]}</span>
            </div>
            <div class="pp-card-amounts">
              <span class="inst-card-label">總額</span>
              <span class="mono">{fmtAmt(pp.total_amount)}</span>
              <span class="inst-card-sep">·</span>
              <span class="inst-card-label">已攤提</span>
              <span class="mono">{fmtAmt(pp.amortized_amount)}</span>
            </div>
            <div class="inst-progress">
              <div class="inst-progress-bar"><div class="inst-progress-fill" style="width:{progressPct(pp)}%"></div></div>
              <span class="inst-progress-text">{pp.amortized_periods}/{pp.periods}</span>
            </div>
            <div class="pp-card-footer">
              <span class="mono" style="font-size:11px;color:#5c6278;">開始 {pp.start_date}</span>
              <div style="display:flex;gap:4px;flex-wrap:wrap;">
                {#if pp.status === 'ACTIVE'}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={(e) => openAmortizeModal(pp, e)}>攤提</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={(e) => openDisposeModal(pp, e)}>終止</button>
                {/if}
              </div>
            </div>
          </div>
          {#if isExpanded}
            <div class="pp-sub-cell" style="padding:0;">{@render amortContent(pp)}</div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<!-- ── 新增預付費用 Modal ────────────────────── -->
{#if showCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-create-title">新增預付費用</h2>
        <button class="modal-close" onclick={() => { showCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCreate}>
        {#if createError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{createError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-name">名稱 *</label>
          <input id="pp-name" class="form-input" type="text" bind:value={createForm.name} placeholder="例：辦公室租金預付" required />
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-account">預付費用科目 *</label>
          <AccountSelect {accounts} value={createForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { createForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-expense">費用科目 *</label>
          <AccountSelect {accounts} value={createForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { createForm.expense_account_id = id; }} />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="pp-amount">總金額 *</label>
            <input id="pp-amount" class="form-input" type="number" min="0.01" step="any" bind:value={createForm.total_amount} placeholder="0.00" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="pp-periods">攤提期數 *</label>
            <input id="pp-periods" class="form-input" type="number" min="1" step="1" bind:value={createForm.periods} placeholder="12" required />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-start">起始日 *</label>
          <input id="pp-start" class="form-input" type="date" bind:value={createForm.start_date} required />
        </div>
        <div class="form-group">
          <span class="form-label">付款帳戶 *</span>
          <LedgerSelect ledgers={activeLedgers} value={createForm.ledger_id} onselect={(id) => { createForm.ledger_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-memo">說明</label>
          <input id="pp-memo" class="form-input" type="text" bind:value={createForm.memo} placeholder="選填" />
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-note">備註</label>
          <input id="pp-note" class="form-input" type="text" bind:value={createForm.note} placeholder="選填" />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCreateModal = false; }} disabled={isCreating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreating || !isCreateValid}>{isCreating ? '建立中…' : '建立預付費用'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 攤提 Modal ──────────────────────────── -->
{#if showAmortizeModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showAmortizeModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-amort-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-amort-title">執行攤提</h2>
        <button class="modal-close" onclick={() => { showAmortizeModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleAmortize}>
        {#if amortizeError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{amortizeError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-amort-date">攤提日期 *</label>
          <input id="pp-amort-date" class="form-input" type="date" bind:value={amortizeDate} required />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showAmortizeModal = false; }} disabled={isAmortizing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isAmortizing || !amortizeDate}>{isAmortizing ? '攤提中…' : '確認攤提'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 提前終止 Modal ──────────────────────── -->
{#if showDisposeModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showDisposeModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-dispose-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-dispose-title">提前終止預付費用</h2>
        <button class="modal-close" onclick={() => { showDisposeModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleDispose}>
        {#if disposeError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{disposeError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-dispose-date">終止日期 *</label>
          <input id="pp-dispose-date" class="form-input" type="date" bind:value={disposeDate} required />
        </div>
        <div class="form-group">
          <label class="form-label" for="pp-dispose-memo">說明</label>
          <input id="pp-dispose-memo" class="form-input" type="text" bind:value={disposeMemo} placeholder="選填" />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showDisposeModal = false; }} disabled={isDisposing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isDisposing || !disposeDate}>{isDisposing ? '處理中…' : '確認終止'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
