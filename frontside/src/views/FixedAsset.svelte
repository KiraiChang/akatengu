<script lang="ts">
  import { getAllFixedAssets, getFixedAssetDepreciations, purchaseFixedAsset, depreciateFixedAsset, disposeFixedAsset } from '../api/fixedAsset';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import { getEntries } from '../api/transaction';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import TxnEntryPanel from '../components/TxnEntryPanel.svelte';
  import type { FixedAsset, FixedAssetDepreciation, FixedAssetStatus, AssetPaymentType } from '../types/fixedAsset';
  import type { Entry } from '../types/transaction';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';

  const STATUS_LABELS: Record<FixedAssetStatus, string> = {
    ACTIVE:   '使用中',
    DISPOSED: '已處分',
  };

  const STATUS_CSS: Record<FixedAssetStatus, string> = {
    ACTIVE:   'fa-status-active',
    DISPOSED: 'fa-status-disposed',
  };

  const PAYMENT_LABELS: Record<AssetPaymentType, string> = {
    CASH:  '現金',
    LEASE: '租賃',
  };

  const PAYMENT_TYPES: AssetPaymentType[] = ['CASH', 'LEASE'];

  // ── 列表 ──
  let assets    = $state<FixedAsset[]>([]);
  let isLoading = $state(false);
  let error     = $state('');

  // ── 折舊明細展開 ──
  interface DeprState {
    items:   FixedAssetDepreciation[];
    loading: boolean;
    error:   string;
  }

  let expandedId = $state<number | null>(null);
  let deprMap    = $state(new Map<number, DeprState>());

  // ── 分錄展開 ──
  interface EntriesState {
    rows:    Entry[];
    loading: boolean;
    error:   string;
  }

  let expandedMainEntryId = $state<number | null>(null);
  let mainEntriesMap      = $state(new Map<number, EntriesState>());
  let expandedDeprEntryId = $state<number | null>(null);
  let deprEntriesMap      = $state(new Map<number, EntriesState>());

  // ── 共用資料 ──
  let accounts = $state<Account[]>([]);
  let ledgers  = $state<LedgerAccount[]>([]);
  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  // ── 新增固定資產 Modal ──
  interface PurchaseForm {
    name:                            string;
    asset_account_id:                string;
    accum_depreciation_account_id:   string;
    depreciation_expense_account_id: string;
    cost:                            string;
    residual_value:                  string;
    useful_life_months:              string;
    payment_type:                    AssetPaymentType;
    ledger_id:                       string;
    liability_account_id:            string;
    purchase_date:                   string;
    memo:                            string;
    note:                            string;
  }

  let showPurchaseModal = $state(false);
  let isPurchasing      = $state(false);
  let purchaseError     = $state('');
  let purchaseForm      = $state<PurchaseForm>(emptyPurchaseForm());

  const isPurchaseValid = $derived(
    purchaseForm.name.trim()                              !== '' &&
    purchaseForm.asset_account_id                         !== '' &&
    purchaseForm.accum_depreciation_account_id            !== '' &&
    purchaseForm.depreciation_expense_account_id          !== '' &&
    parseFloat(purchaseForm.cost)                         > 0 &&
    parseInt(purchaseForm.useful_life_months, 10)         > 0 &&
    purchaseForm.purchase_date                            !== '' &&
    (purchaseForm.payment_type === 'CASH'
      ? purchaseForm.ledger_id          !== ''
      : purchaseForm.liability_account_id !== '')
  );

  // ── 折舊 Modal ──
  let showDeprModal  = $state(false);
  let isDepreciating = $state(false);
  let deprError      = $state('');
  let deprAssetId    = $state(0);
  let deprDate       = $state('');

  // ── 處分 Modal ──
  interface DisposeForm {
    disposal_date:      string;
    proceeds:           string;
    proceeds_ledger_id: string;
    gain_account_id:    string;
    loss_account_id:    string;
    memo:               string;
  }

  let showDisposeModal  = $state(false);
  let isDisposing       = $state(false);
  let disposeError      = $state('');
  let disposeAssetId    = $state(0);
  let disposeForm       = $state<DisposeForm>(emptyDisposeForm());

  const isDisposeValid = $derived(
    disposeForm.disposal_date   !== '' &&
    disposeForm.gain_account_id !== '' &&
    disposeForm.loss_account_id !== ''
  );

  $effect(() => {
    void load();
  });

  async function load(): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      assets = await getAllFixedAssets();
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  async function toggleDepr(asset: FixedAsset): Promise<void> {
    if (expandedId === asset.id) { expandedId = null; return; }
    expandedId = asset.id;
    if (deprMap.has(asset.id)) return;
    deprMap = new Map(deprMap).set(asset.id, { items: [], loading: true, error: '' });
    try {
      const items = await getFixedAssetDepreciations(asset.id);
      deprMap = new Map(deprMap).set(asset.id, { items, loading: false, error: '' });
    } catch (err) {
      deprMap = new Map(deprMap).set(asset.id, {
        items: [], loading: false, error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function ensureAccounts(): Promise<void> {
    if (accounts.length === 0) { try { accounts = await getAccountAll(); } catch { /* 非致命 */ } }
  }

  async function ensureLedgers(): Promise<void> {
    if (ledgers.length === 0) { try { ledgers = await getLedgerAccountAll(); } catch { /* 非致命 */ } }
  }

  async function toggleMainEntry(asset: FixedAsset, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (!asset.txn_id) return;
    const txnId = asset.txn_id;
    if (expandedMainEntryId === asset.id) { expandedMainEntryId = null; return; }
    expandedMainEntryId = asset.id;
    if (mainEntriesMap.has(asset.id)) return;
    mainEntriesMap = new Map(mainEntriesMap).set(asset.id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      mainEntriesMap = new Map(mainEntriesMap).set(asset.id, { rows, loading: false, error: '' });
    } catch (err) {
      mainEntriesMap = new Map(mainEntriesMap).set(asset.id, {
        rows: [], loading: false, error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function toggleDeprEntry(depr: FixedAssetDepreciation, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (expandedDeprEntryId === depr.id) { expandedDeprEntryId = null; return; }
    expandedDeprEntryId = depr.id;
    if (deprEntriesMap.has(depr.id)) return;
    deprEntriesMap = new Map(deprEntriesMap).set(depr.id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(depr.txn_id);
      deprEntriesMap = new Map(deprEntriesMap).set(depr.id, { rows, loading: false, error: '' });
    } catch (err) {
      deprEntriesMap = new Map(deprEntriesMap).set(depr.id, {
        rows: [], loading: false, error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function openPurchaseModal(): Promise<void> {
    purchaseForm  = emptyPurchaseForm();
    purchaseError = '';
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    showPurchaseModal = true;
  }

  async function handlePurchase(e: Event): Promise<void> {
    e.preventDefault();
    if (!isPurchaseValid) return;
    isPurchasing  = true;
    purchaseError = '';
    try {
      await purchaseFixedAsset({
        name:                            purchaseForm.name.trim(),
        asset_account_id:                purchaseForm.asset_account_id,
        accum_depreciation_account_id:   purchaseForm.accum_depreciation_account_id,
        depreciation_expense_account_id: purchaseForm.depreciation_expense_account_id,
        cost:                            purchaseForm.cost,
        residual_value:                  purchaseForm.residual_value || '0',
        useful_life_months:              parseInt(purchaseForm.useful_life_months, 10),
        payment_type:                    purchaseForm.payment_type,
        ledger_id:                       purchaseForm.payment_type === 'CASH' ? parseInt(purchaseForm.ledger_id, 10) : null,
        liability_account_id:            purchaseForm.payment_type === 'LEASE' ? purchaseForm.liability_account_id : '',
        purchase_date:                   purchaseForm.purchase_date,
        memo:                            purchaseForm.memo.trim(),
        note:                            purchaseForm.note.trim(),
      });
      showPurchaseModal = false;
      await load();
    } catch (err) {
      purchaseError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isPurchasing = false;
    }
  }

  function openDeprModal(asset: FixedAsset, e: MouseEvent): void {
    e.stopPropagation();
    deprAssetId   = asset.id;
    deprDate      = '';
    deprError     = '';
    showDeprModal = true;
  }

  async function handleDepr(e: Event): Promise<void> {
    e.preventDefault();
    if (!deprDate) return;
    isDepreciating = true;
    deprError      = '';
    try {
      await depreciateFixedAsset({ asset_id: deprAssetId, period_date: deprDate });
      showDeprModal = false;
      const next = new Map(deprMap);
      next.delete(deprAssetId);
      deprMap = next;
      await load();
    } catch (err) {
      deprError = err instanceof Error ? err.message : '折舊失敗，請稍後再試。';
    } finally {
      isDepreciating = false;
    }
  }

  async function openDisposeModal(asset: FixedAsset, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    disposeAssetId = asset.id;
    disposeForm    = emptyDisposeForm();
    disposeError   = '';
    showDisposeModal = true;
  }

  async function handleDispose(e: Event): Promise<void> {
    e.preventDefault();
    if (!isDisposeValid) return;
    isDisposing  = true;
    disposeError = '';
    try {
      await disposeFixedAsset({
        asset_id:           disposeAssetId,
        disposal_date:      disposeForm.disposal_date,
        proceeds:           disposeForm.proceeds || '0',
        proceeds_ledger_id: disposeForm.proceeds_ledger_id !== '' ? parseInt(disposeForm.proceeds_ledger_id, 10) : null,
        gain_account_id:    disposeForm.gain_account_id,
        loss_account_id:    disposeForm.loss_account_id,
        memo:               disposeForm.memo.trim(),
      });
      showDisposeModal = false;
      await load();
    } catch (err) {
      disposeError = err instanceof Error ? err.message : '處分失敗，請稍後再試。';
    } finally {
      isDisposing = false;
    }
  }

  function emptyPurchaseForm(): PurchaseForm {
    return { name: '', asset_account_id: '', accum_depreciation_account_id: '', depreciation_expense_account_id: '', cost: '', residual_value: '0', useful_life_months: '', payment_type: 'CASH', ledger_id: '', liability_account_id: '', purchase_date: '', memo: '', note: '' };
  }

  function emptyDisposeForm(): DisposeForm {
    return { disposal_date: '', proceeds: '0', proceeds_ledger_id: '', gain_account_id: '', loss_account_id: '', memo: '' };
  }

  function fmtAmt(val: string): string {
    const n = parseFloat(val);
    return isNaN(n) ? val : n.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function progressPct(asset: FixedAsset): number {
    return asset.useful_life_months === 0 ? 0 : Math.round((asset.depreciated_periods / asset.useful_life_months) * 100);
  }
</script>

<div class="content-header">
  <h1 class="content-title">固定資產</h1>
  <span class="content-date">共 {assets.length} 筆</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

{#snippet deprContent(asset: FixedAsset)}
  {@const state = deprMap.get(asset.id)}
  {#if state?.loading}
    <span class="fa-sub-msg">查詢中…</span>
  {:else if state?.error}
    <span class="fa-sub-msg fa-sub-error">{state.error}</span>
  {:else if state && state.items.length > 0}
    <div class="fa-depr-table">
      <div class="fa-depr-header">
        <span>折舊日</span>
        <span class="num">金額</span>
        <span>更新者</span>
        <span>更新時間</span>
        <span></span>
      </div>
      {#each state.items as depr (depr.id)}
        {@const isEntryExpanded = expandedDeprEntryId === depr.id}
        {@const entryState = deprEntriesMap.get(depr.id)}
        <div class="fa-depr-item">
          <span class="mono">{depr.period_date}</span>
          <span class="num mono">{fmtAmt(depr.amount)}</span>
          <span>{depr.updated_by ?? '—'}</span>
          <span>{depr.updated_at ?? '—'}</span>
          <span>
            <button class="btn-ghost" style="padding:2px 8px;font-size:10px;" class:acct-mode-btn--active={isEntryExpanded} onclick={(e) => toggleDeprEntry(depr, e)}>分錄</button>
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
    <span class="fa-sub-msg">尚無折舊紀錄</span>
  {/if}
{/snippet}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">固定資產主檔</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading"><span class="spinner" aria-hidden="true"></span>載入中</span>
      {/if}
      <button class="section-action" onclick={openPurchaseModal}>＋ 新增資產</button>
    </div>
  </header>

  <div class="table-wrap fa-table-wrap">
    <table class="data-table" aria-label="固定資產主檔">
      <thead>
        <tr>
          <th>名稱</th>
          <th>付款</th>
          <th class="fa-amount">成本</th>
          <th class="fa-amount">已折舊</th>
          <th>折舊進度</th>
          <th>購入日</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && assets.length === 0}
          <tr><td colspan="9" class="table-empty">載入中...</td></tr>
        {:else if assets.length === 0}
          <tr><td colspan="9" class="table-empty">無資料</td></tr>
        {:else}
          {#each assets as asset (asset.id)}
            {@const isExpanded = expandedId === asset.id}
            {@const isMainExpanded = expandedMainEntryId === asset.id}
            {@const mainEntry = mainEntriesMap.get(asset.id)}
            <tr class:fa-row-expanded={isExpanded} style="cursor:pointer;" onclick={() => toggleDepr(asset)}>
              <td><span class="fa-expand-icon">{isExpanded ? '▼' : '▶'}</span>{asset.name}</td>
              <td>{PAYMENT_LABELS[asset.payment_type]}</td>
              <td class="fa-amount mono">{fmtAmt(asset.cost)}</td>
              <td class="fa-amount mono">{fmtAmt(asset.total_depreciated)}</td>
              <td>
                <div class="inst-progress">
                  <div class="inst-progress-bar"><div class="inst-progress-fill" style="width:{progressPct(asset)}%"></div></div>
                  <span class="inst-progress-text">{asset.depreciated_periods}/{asset.useful_life_months}</span>
                </div>
              </td>
              <td class="mono">{asset.purchase_date}</td>
              <td class="hidden md:table-cell">{asset.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{asset.updated_at ?? '—'}</td>
              <td onclick={(e) => e.stopPropagation()}>
                <span class="badge {STATUS_CSS[asset.status]}">{STATUS_LABELS[asset.status]}</span>
                {#if asset.status === 'ACTIVE'}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={(e) => openDeprModal(asset, e)}>折舊</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" onclick={(e) => openDisposeModal(asset, e)}>處分</button>
                {/if}
                {#if asset.txn_id !== null}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;margin-left:4px;" class:acct-mode-btn--active={isMainExpanded} onclick={(e) => toggleMainEntry(asset, e)}>分錄</button>
                {/if}
              </td>
            </tr>
            {#if isMainExpanded}
              <tr><td colspan="9" class="fa-sub-cell">
                <div class="txn-entries"><TxnEntryPanel rows={mainEntry?.rows ?? []} loading={mainEntry?.loading ?? false} error={mainEntry?.error ?? ''} {accounts} {ledgers} /></div>
              </td></tr>
            {/if}
            {#if isExpanded}
              <tr class="fa-sub-row"><td colspan="9" class="fa-sub-cell">{@render deprContent(asset)}</td></tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="fa-card-list">
    {#if isLoading && assets.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if assets.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each assets as asset (asset.id)}
        {@const isExpanded = expandedId === asset.id}
        <div class="fa-card {isExpanded ? 'fa-card-expanded' : ''}">
          <div class="fa-card-main" role="button" tabindex="0"
            onclick={() => toggleDepr(asset)}
            onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && toggleDepr(asset)}
            style="cursor:pointer;"
          >
            <div class="fa-card-head">
              <span class="fa-card-title"><span class="fa-expand-icon">{isExpanded ? '▼' : '▶'}</span>{asset.name}</span>
              <span class="badge {STATUS_CSS[asset.status]}">{STATUS_LABELS[asset.status]}</span>
            </div>
            <div class="fa-card-meta">
              <span class="fa-card-label">付款</span>
              <span>{PAYMENT_LABELS[asset.payment_type]}</span>
              <span class="inst-card-sep">·</span>
              <span class="fa-card-label">成本</span>
              <span class="mono">{fmtAmt(asset.cost)}</span>
              <span class="inst-card-sep">·</span>
              <span class="fa-card-label">已折舊</span>
              <span class="mono">{fmtAmt(asset.total_depreciated)}</span>
            </div>
            <div class="inst-progress">
              <div class="inst-progress-bar"><div class="inst-progress-fill" style="width:{progressPct(asset)}%"></div></div>
              <span class="inst-progress-text">{asset.depreciated_periods}/{asset.useful_life_months}</span>
            </div>
            <div class="fa-card-footer">
              <span class="mono" style="font-size:11px;color:#5c6278;">購入 {asset.purchase_date}</span>
              <div style="display:flex;gap:4px;flex-wrap:wrap;">
                {#if asset.status === 'ACTIVE'}
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={(e) => openDeprModal(asset, e)}>折舊</button>
                  <button type="button" class="btn-ghost" style="padding:1px 6px;font-size:10px;" onclick={(e) => openDisposeModal(asset, e)}>處分</button>
                {/if}
              </div>
            </div>
          </div>
          {#if isExpanded}
            <div class="fa-sub-cell" style="padding:0;">{@render deprContent(asset)}</div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<!-- ── 新增固定資產 Modal ──────────────────── -->
{#if showPurchaseModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showPurchaseModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-purchase-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-purchase-title">新增固定資產</h2>
        <button class="modal-close" onclick={() => { showPurchaseModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handlePurchase}>
        {#if purchaseError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{purchaseError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-name">資產名稱 *</label>
          <input id="fa-name" class="form-input" type="text" bind:value={purchaseForm.name} placeholder="例：辦公室電腦" required />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-asset-acct">資產科目 *</label>
          <AccountSelect {accounts} value={purchaseForm.asset_account_id} placeholder="選擇資產科目…" onselect={(id) => { purchaseForm.asset_account_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-accum-depr">累計折舊科目 *</label>
          <AccountSelect {accounts} value={purchaseForm.accum_depreciation_account_id} placeholder="選擇累計折舊科目…" onselect={(id) => { purchaseForm.accum_depreciation_account_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-depr-expense">折舊費用科目 *</label>
          <AccountSelect {accounts} value={purchaseForm.depreciation_expense_account_id} placeholder="選擇折舊費用科目…" onselect={(id) => { purchaseForm.depreciation_expense_account_id = id; }} />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="fa-cost">成本 *</label>
            <input id="fa-cost" class="form-input" type="number" min="0.01" step="any" bind:value={purchaseForm.cost} placeholder="0.00" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="fa-residual">殘值</label>
            <input id="fa-residual" class="form-input" type="number" min="0" step="any" bind:value={purchaseForm.residual_value} placeholder="0.00" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="fa-life">耐用年限（月）*</label>
            <input id="fa-life" class="form-input" type="number" min="1" step="1" bind:value={purchaseForm.useful_life_months} placeholder="60" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="fa-payment-type">付款方式 *</label>
            <select id="fa-payment-type" class="form-select" bind:value={purchaseForm.payment_type}>
              {#each PAYMENT_TYPES as pt}
                <option value={pt}>{PAYMENT_LABELS[pt]}</option>
              {/each}
            </select>
          </div>
        </div>
        {#if purchaseForm.payment_type === 'CASH'}
          <div class="form-group">
            <span class="form-label">付款帳戶 *</span>
            <LedgerSelect ledgers={activeLedgers} value={purchaseForm.ledger_id} onselect={(id) => { purchaseForm.ledger_id = id; }} />
          </div>
        {:else}
          <div class="form-group">
            <label class="form-label" for="fa-liability">應付科目 *</label>
            <AccountSelect {accounts} value={purchaseForm.liability_account_id} placeholder="選擇應付科目…" onselect={(id) => { purchaseForm.liability_account_id = id; }} />
          </div>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-purchase-date">購入日期 *</label>
          <input id="fa-purchase-date" class="form-input" type="date" bind:value={purchaseForm.purchase_date} required />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-memo">說明</label>
          <input id="fa-memo" class="form-input" type="text" bind:value={purchaseForm.memo} placeholder="選填" />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-note">備註</label>
          <input id="fa-note" class="form-input" type="text" bind:value={purchaseForm.note} placeholder="選填" />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showPurchaseModal = false; }} disabled={isPurchasing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isPurchasing || !isPurchaseValid}>{isPurchasing ? '建立中…' : '建立資產'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 折舊 Modal ──────────────────────────── -->
{#if showDeprModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showDeprModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-depr-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-depr-title">執行折舊</h2>
        <button class="modal-close" onclick={() => { showDeprModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleDepr}>
        {#if deprError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{deprError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-depr-date">折舊日期 *</label>
          <input id="fa-depr-date" class="form-input" type="date" bind:value={deprDate} required />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showDeprModal = false; }} disabled={isDepreciating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isDepreciating || !deprDate}>{isDepreciating ? '折舊中…' : '確認折舊'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 處分 Modal ──────────────────────────── -->
{#if showDisposeModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showDisposeModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="fa-dispose-title">
      <div class="modal-header">
        <h2 class="modal-title" id="fa-dispose-title">處分固定資產</h2>
        <button class="modal-close" onclick={() => { showDisposeModal = false; }} aria-label="關閉">×</button>
      </div>
      <form class="modal-body" onsubmit={handleDispose}>
        {#if disposeError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{disposeError}</p>
        {/if}
        <div class="form-group">
          <label class="form-label" for="fa-dispose-date">處分日期 *</label>
          <input id="fa-dispose-date" class="form-input" type="date" bind:value={disposeForm.disposal_date} required />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="fa-proceeds">處分價款</label>
            <input id="fa-proceeds" class="form-input" type="number" min="0" step="any" bind:value={disposeForm.proceeds} placeholder="0.00" />
          </div>
        </div>
        <div class="form-group">
          <span class="form-label">收款帳戶（選填）</span>
          <LedgerSelect ledgers={activeLedgers} value={disposeForm.proceeds_ledger_id} onselect={(id) => { disposeForm.proceeds_ledger_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-gain-acct">處分利益科目 *</label>
          <AccountSelect {accounts} value={disposeForm.gain_account_id} placeholder="選擇處分利益科目…" onselect={(id) => { disposeForm.gain_account_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-loss-acct">處分損失科目 *</label>
          <AccountSelect {accounts} value={disposeForm.loss_account_id} placeholder="選擇處分損失科目…" onselect={(id) => { disposeForm.loss_account_id = id; }} />
        </div>
        <div class="form-group">
          <label class="form-label" for="fa-dispose-memo">說明</label>
          <input id="fa-dispose-memo" class="form-input" type="text" bind:value={disposeForm.memo} placeholder="選填" />
        </div>
        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showDisposeModal = false; }} disabled={isDisposing}>取消</button>
          <button type="submit" class="btn-primary" disabled={isDisposing || !isDisposeValid}>{isDisposing ? '處理中…' : '確認處分'}</button>
        </div>
      </form>
    </div>
  </div>
{/if}
