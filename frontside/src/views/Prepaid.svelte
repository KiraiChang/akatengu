<script lang="ts">
  import { getAllPrepaids, getPrepaidAmortizations, createPrepaid, createPrepaidWithInstallment, amortizePrepaid, disposePrepaid } from '../api/prepaid';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import { getEntries } from '../api/transaction';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import TxnEntryPanel from '../components/TxnEntryPanel.svelte';
  import type { Prepaid, PrepaidAmortization, PrepaidStatus } from '../types/prepaid';
  import InstallmentTermsSection from '../components/InstallmentTermsSection.svelte';
  import type { InterestType } from '../types/installment';
  import type { Entry } from '../types/transaction';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import { getAllPrepaidCategories, createPrepaidCategory, updatePrepaidCategory, deletePrepaidCategory } from '../api/prepaidCategory';
  import type { PrepaidCategory } from '../types/prepaidCategory';

  type PrepaidPaymentType = 'CASH' | 'INSTALLMENT';

  const PAYMENT_LABELS: Record<PrepaidPaymentType, string> = {
    CASH:        '現金',
    INSTALLMENT: '分期',
  };
  const PAYMENT_TYPES: PrepaidPaymentType[] = ['CASH', 'INSTALLMENT'];

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
    name:                   string;
    account_id:             string;
    expense_account_id:     string;
    ledger_id:              string;
    total_amount:           string;
    periods:                string;
    start_date:             string;
    memo:                   string;
    note:                   string;
    payment_type:           PrepaidPaymentType;
    installment_count:      string;
    installment_start_date: string;
    interest_type:          InterestType;
    annual_rate:            string;
  }

  let showCreateModal = $state(false);
  let isCreating      = $state(false);
  let createError     = $state('');
  let createForm      = $state<CreateForm>(emptyCreateForm());

  const isCreateValid = $derived(
    createForm.name.trim()              !== '' &&
    createForm.account_id               !== '' &&
    createForm.expense_account_id       !== '' &&
    parseFloat(createForm.total_amount) > 0 &&
    parseInt(createForm.periods, 10)    > 0 &&
    createForm.start_date               !== '' &&
    (createForm.payment_type === 'CASH'
      ? createForm.ledger_id !== ''
      : createForm.ledger_id                                !== '' &&
        parseInt(createForm.installment_count, 10)          > 0 &&
        createForm.installment_start_date                   !== '' &&
        (createForm.interest_type !== 'FIXED_RATE' || parseFloat(createForm.annual_rate) > 0))
  );

  // ── 攤提 Modal ──
  let showAmortizeModal  = $state(false);
  let isAmortizing       = $state(false);
  let amortizeError      = $state('');
  let amortizePrepaidId  = $state(0);
  let amortizePrepaidUUID = $state('');
  let amortizeDate       = $state('');

  // ── 終止 Modal ──
  let showDisposeModal  = $state(false);
  let isDisposing       = $state(false);
  let disposeError      = $state('');
  let disposePrepaidId  = $state(0);
  let disposePrepaidUUID = $state('');
  let disposeDate       = $state('');
  let disposeMemo       = $state('');

  // ── 類別管理 ──
  interface CatForm {
    name:               string;
    account_id:         string;
    expense_account_id: string;
  }

  let categories         = $state<PrepaidCategory[]>([]);
  let isCatLoading       = $state(false);
  let catError           = $state('');
  let showCatCreateModal = $state(false);
  let isCreatingCat      = $state(false);
  let catCreateError     = $state('');
  let catCreateForm      = $state<CatForm>(emptyCatForm());
  let showCatEditModal   = $state(false);
  let isUpdatingCat      = $state(false);
  let catEditError       = $state('');
  let editingCat         = $state<PrepaidCategory | null>(null);
  let catEditForm        = $state<CatForm>(emptyCatForm());
  let showCatCloseModal  = $state(false);
  let isClosingCat       = $state(false);
  let catCloseError      = $state('');
  let closingCat         = $state<PrepaidCategory | null>(null);

  const isCatCreateValid = $derived(
    catCreateForm.name.trim()        !== '' &&
    catCreateForm.account_id         !== '' &&
    catCreateForm.expense_account_id !== ''
  );

  const isCatEditValid = $derived(
    catEditForm.name.trim()        !== '' &&
    catEditForm.account_id         !== '' &&
    catEditForm.expense_account_id !== ''
  );

  $effect(() => {
    void load();
    void loadCategories();
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
      const createLedger = activeLedgers.find(l => String(l.ledger_id) === createForm.ledger_id);
      if (createForm.payment_type === 'INSTALLMENT') {
        await createPrepaidWithInstallment({
          name:               createForm.name.trim(),
          account_id:         createForm.account_id,
          expense_account_id: createForm.expense_account_id,
          total_amount:       createForm.total_amount,
          periods:            parseInt(createForm.periods, 10),
          start_date:         createForm.start_date,
          memo:               createForm.memo.trim(),
          note:               createForm.note.trim(),
          installment: {
            installment_count: parseInt(createForm.installment_count, 10),
            start_date:        createForm.installment_start_date,
            interest_type:     createForm.interest_type,
            annual_rate:       createForm.interest_type === 'FIXED_RATE' ? createForm.annual_rate : '0',
            ledger_uuid:       createLedger?.ledger_uuid ?? '',
            memo:              createForm.memo.trim(),
            note:              createForm.note.trim(),
          },
        });
      } else {
        await createPrepaid({
          name:               createForm.name.trim(),
          account_id:         createForm.account_id,
          expense_account_id: createForm.expense_account_id,
          ledger_uuid:        createLedger?.ledger_uuid ?? '',
          total_amount:       createForm.total_amount,
          periods:            parseInt(createForm.periods, 10),
          start_date:         createForm.start_date,
          memo:               createForm.memo.trim(),
          note:               createForm.note.trim(),
        });
      }
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
    amortizePrepaidId   = pp.id;
    amortizePrepaidUUID = pp.prepaid_uuid;
    amortizeDate        = '';
    amortizeError       = '';
    showAmortizeModal   = true;
  }

  async function handleAmortize(e: Event): Promise<void> {
    e.preventDefault();
    if (!amortizeDate) return;
    isAmortizing  = true;
    amortizeError = '';
    try {
      await amortizePrepaid({ prepaid_uuid: amortizePrepaidUUID, period_date: amortizeDate });
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
    disposePrepaidId   = pp.id;
    disposePrepaidUUID = pp.prepaid_uuid;
    disposeDate        = '';
    disposeMemo        = '';
    disposeError       = '';
    showDisposeModal   = true;
  }

  async function handleDispose(e: Event): Promise<void> {
    e.preventDefault();
    if (!disposeDate) return;
    isDisposing  = true;
    disposeError = '';
    try {
      await disposePrepaid({ prepaid_uuid: disposePrepaidUUID, disposal_date: disposeDate, memo: disposeMemo.trim() });
      showDisposeModal = false;
      await load();
    } catch (err) {
      disposeError = err instanceof Error ? err.message : '終止失敗，請稍後再試。';
    } finally {
      isDisposing = false;
    }
  }

  function emptyCreateForm(): CreateForm {
    return { name: '', account_id: '', expense_account_id: '', ledger_id: '', total_amount: '', periods: '', start_date: '', memo: '', note: '', payment_type: 'CASH', installment_count: '', installment_start_date: '', interest_type: 'FREE', annual_rate: '' };
  }

  function fmtAmt(val: string): string {
    const n = parseFloat(val);
    return isNaN(n) ? val : n.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function progressPct(pp: Prepaid): number {
    return pp.periods === 0 ? 0 : Math.round((pp.amortized_periods / pp.periods) * 100);
  }

  function emptyCatForm(): CatForm {
    return { name: '', account_id: '', expense_account_id: '' };
  }

  async function loadCategories(): Promise<void> {
    isCatLoading = true;
    catError     = '';
    try {
      categories = await getAllPrepaidCategories();
    } catch (err) {
      catError = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isCatLoading = false;
    }
  }

  async function openCatCreateModal(): Promise<void> {
    catCreateForm  = emptyCatForm();
    catCreateError = '';
    await ensureAccounts();
    showCatCreateModal = true;
  }

  async function handleCatCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isCatCreateValid) return;
    isCreatingCat  = true;
    catCreateError = '';
    try {
      await createPrepaidCategory({
        name:               catCreateForm.name.trim(),
        account_id:         catCreateForm.account_id,
        expense_account_id: catCreateForm.expense_account_id,
      });
      showCatCreateModal = false;
      await loadCategories();
    } catch (err) {
      catCreateError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreatingCat = false;
    }
  }

  async function openCatEditModal(cat: PrepaidCategory): Promise<void> {
    editingCat  = cat;
    catEditForm = {
      name:               cat.name,
      account_id:         cat.account_id,
      expense_account_id: cat.expense_account_id,
    };
    catEditError = '';
    await ensureAccounts();
    showCatEditModal = true;
  }

  async function handleCatEdit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isCatEditValid || !editingCat) return;
    isUpdatingCat  = true;
    catEditError   = '';
    try {
      await updatePrepaidCategory(editingCat.category_uuid, {
        expected_version:   editingCat.version,
        name:               catEditForm.name.trim(),
        account_id:         catEditForm.account_id,
        expense_account_id: catEditForm.expense_account_id,
      });
      showCatEditModal = false;
      await loadCategories();
    } catch (err) {
      catEditError = err instanceof Error ? err.message : '修改失敗，請稍後再試。';
    } finally {
      isUpdatingCat = false;
    }
  }

  function openCatCloseModal(cat: PrepaidCategory): void {
    closingCat    = cat;
    catCloseError = '';
    showCatCloseModal = true;
  }

  async function handleCatClose(e: Event): Promise<void> {
    e.preventDefault();
    if (!closingCat) return;
    isClosingCat  = true;
    catCloseError = '';
    try {
      await deletePrepaidCategory(closingCat.category_uuid, closingCat.version);
      showCatCloseModal = false;
      await loadCategories();
    } catch (err) {
      catCloseError = err instanceof Error ? err.message : '關閉失敗，請稍後再試。';
    } finally {
      isClosingCat = false;
    }
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

<section class="section">
  <header class="section-header">
    <h2 class="section-title">預付費用類別</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isCatLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openCatCreateModal}>＋ 新增類別</button>
    </div>
  </header>

  {#if catError}
    <p class="query-error" role="alert">{catError}</p>
  {/if}

  <div class="table-wrap pp-cat-table-wrap">
    <table class="data-table" aria-label="預付費用類別">
      <thead>
        <tr>
          <th>名稱</th>
          <th>預付費用科目</th>
          <th>費用科目</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isCatLoading && categories.length === 0}
          <tr><td colspan="6" class="table-empty">載入中...</td></tr>
        {:else if categories.length === 0}
          <tr><td colspan="6" class="table-empty">無資料</td></tr>
        {:else}
          {#each categories as cat (cat.id)}
            <tr>
              <td>{cat.name}</td>
              <td class="mono" style="font-size:11px;">{cat.account_id}</td>
              <td class="mono" style="font-size:11px;">{cat.expense_account_id}</td>
              <td class="hidden md:table-cell">{cat.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{cat.updated_at ?? '—'}</td>
              <td>
                <span class="badge {cat.is_active ? 'pp-cat-status-active' : 'pp-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
                {#if cat.is_active}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openCatEditModal(cat)}>修改</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={() => openCatCloseModal(cat)}>關閉</button>
                {/if}
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="pp-cat-card-list">
    {#if isCatLoading && categories.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if categories.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each categories as cat (cat.id)}
        <div class="pp-cat-card">
          <div class="pp-cat-card-head">
            <span class="pp-cat-card-title">{cat.name}</span>
            <span class="badge {cat.is_active ? 'pp-cat-status-active' : 'pp-cat-status-closed'}">{cat.is_active ? '使用中' : '已關閉'}</span>
          </div>
          <div class="pp-cat-card-accts">
            <span>預付 {cat.account_id}</span>
            <span>費用 {cat.expense_account_id}</span>
          </div>
          {#if cat.is_active}
            <div class="pp-cat-card-footer">
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openCatEditModal(cat)}>修改</button>
              <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={() => openCatCloseModal(cat)}>關閉</button>
            </div>
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
          <label class="form-label" for="pp-payment-type">付款方式 *</label>
          <select id="pp-payment-type" class="form-select" bind:value={createForm.payment_type}>
            {#each PAYMENT_TYPES as pt}
              <option value={pt}>{PAYMENT_LABELS[pt]}</option>
            {/each}
          </select>
        </div>
        {#if createForm.payment_type === 'CASH'}
          <div class="form-group">
            <span class="form-label">付款帳戶 *</span>
            <LedgerSelect ledgers={activeLedgers} value={createForm.ledger_id} onselect={(id) => { createForm.ledger_id = id; }} />
          </div>
        {:else}
          <div class="form-group">
            <span class="form-label">信用卡帳戶 *</span>
            <LedgerSelect ledgers={activeLedgers} value={createForm.ledger_id} onselect={(id) => { createForm.ledger_id = id; }} />
          </div>
          <InstallmentTermsSection
            idPrefix="pp"
            bind:installmentCount={createForm.installment_count}
            bind:startDate={createForm.installment_start_date}
            bind:interestType={createForm.interest_type}
            bind:annualRate={createForm.annual_rate}
          />
        {/if}
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
          <label class="form-label" for="pp-amort-date">攤提月份 *</label>
          <input id="pp-amort-date" class="form-input" type="month" bind:value={amortizeDate} required />
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

<!-- ── 新增類別 Modal ──────────────────────── -->
{#if showCatCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCatCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-create-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-create-title">新增預付費用類別</h2>
        <button class="modal-close" onclick={() => { showCatCreateModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCatCreate}>
        {#if catCreateError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{catCreateError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-name">名稱 *</label>
          <input id="pp-cat-name" class="form-input" type="text" bind:value={catCreateForm.name} placeholder="例：辦公室租金" required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={catCreateForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { catCreateForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={catCreateForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { catCreateForm.expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCatCreateModal = false; }} disabled={isCreatingCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreatingCat || !isCatCreateValid}>{isCreatingCat ? '建立中…' : '建立類別'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 修改類別 Modal ──────────────────────── -->
{#if showCatEditModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCatEditModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-edit-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-edit-title">修改預付費用類別</h2>
        <button class="modal-close" onclick={() => { showCatEditModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCatEdit}>
        {#if catEditError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{catEditError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="pp-cat-edit-name">名稱 *</label>
          <input id="pp-cat-edit-name" class="form-input" type="text" bind:value={catEditForm.name} required />
        </div>
        <div class="form-group">
          <span class="form-label">預付費用科目 *</span>
          <AccountSelect {accounts} value={catEditForm.account_id} placeholder="選擇預付費用資產科目…" onselect={(id) => { catEditForm.account_id = id; }} />
        </div>
        <div class="form-group">
          <span class="form-label">費用科目 *</span>
          <AccountSelect {accounts} value={catEditForm.expense_account_id} placeholder="選擇攤提費用科目…" onselect={(id) => { catEditForm.expense_account_id = id; }} />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCatEditModal = false; }} disabled={isUpdatingCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isUpdatingCat || !isCatEditValid}>{isUpdatingCat ? '更新中…' : '儲存修改'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 關閉類別 Modal ──────────────────────── -->
{#if showCatCloseModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCatCloseModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pp-cat-close-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pp-cat-close-title">關閉預付費用類別</h2>
        <button class="modal-close" onclick={() => { showCatCloseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleCatClose}>
        {#if catCloseError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{catCloseError}</p>
        {/if}
        <p style="font-size:13px;color:#c0bdb4;margin-bottom:16px;">確定要關閉類別「{closingCat?.name}」？關閉後將無法用於新增預付費用。</p>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCatCloseModal = false; }} disabled={isClosingCat}>取消</button>
          <button type="submit" class="btn-primary" disabled={isClosingCat}>{isClosingCat ? '處理中…' : '確認關閉'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
