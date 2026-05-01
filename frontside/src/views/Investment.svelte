<script lang="ts">
  import { getInvestmentPaged, createInvestment, updateInvestment, buyInvestment, sellInvestment, getOpenLotsPaged, getPosition, getLotDisposalsPaged } from '../api/investment';
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import type { Investment, AssetType, CostMethod, InvestmentCreatedPayload, InvestmentLot, InvestmentPosition, InvestmentLotDisposal, LotStatus } from '../types/investment';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type { PaginatedMeta } from '../types/pagination';

  const PAGE_SIZE = 20;

  const ASSET_TYPE_LABELS: Record<AssetType, string> = {
    STOCK: '股票',
    FUND:  '基金',
    GOLD:  '黃金',
    FX:    '外匯',
  };

  const COST_METHOD_LABELS: Record<CostMethod, string> = {
    AVG:  '平均成本',
    FIFO: '先進先出',
  };

  const ASSET_TYPES:  AssetType[]  = ['STOCK', 'FUND', 'GOLD', 'FX'];
  const COST_METHODS: CostMethod[] = ['AVG', 'FIFO'];

  let page        = $state(1);
  let investments = $state<Investment[]>([]);
  let meta        = $state<PaginatedMeta | null>(null);
  let isLoading   = $state(false);
  let error       = $state('');

  // ── 新增 / 修改投資 modal ──
  let showModal   = $state(false);
  let mode        = $state<'create' | 'edit'>('create');
  let isSaving    = $state(false);
  let saveError   = $state('');
  let accounts    = $state<Account[]>([]);
  let editId      = $state(0);
  let editVersion = $state(0);
  let form        = $state<InvestmentCreatedPayload>({
    account_id:  '',
    asset_type:  'STOCK',
    currency:    'TWD',
    symbol:      '',
    name:        '',
    cost_method: 'AVG',
    is_active:   true,
  });

  const isFormValid = $derived(
    form.account_id.trim() !== '' &&
    form.symbol.trim()     !== '' &&
    form.name.trim()       !== '' &&
    form.currency.trim()   !== ''
  );

  // ── 持倉展開 ──
  interface HoldingState {
    lots:     InvestmentLot[];
    position: InvestmentPosition | null;
    loading:  boolean;
    error:    string;
  }

  let expandedId  = $state<number | null>(null);
  let holdingMap  = $state(new Map<number, HoldingState>());

  async function toggleHolding(inv: Investment): Promise<void> {
    if (expandedId === inv.investment_id) {
      expandedId = null;
      return;
    }
    expandedId = inv.investment_id;
    if (holdingMap.has(inv.investment_id)) return;

    holdingMap = new Map(holdingMap).set(inv.investment_id, {
      lots: [], position: null, loading: true, error: '',
    });

    try {
      if (inv.cost_method === 'FIFO') {
        const res = await getOpenLotsPaged(inv.investment_id, { page: 1, pageSize: 50 });
        holdingMap = new Map(holdingMap).set(inv.investment_id, {
          lots: res.data, position: null, loading: false, error: '',
        });
      } else {
        const pos = await getPosition(inv.investment_id);
        holdingMap = new Map(holdingMap).set(inv.investment_id, {
          lots: [], position: pos, loading: false, error: '',
        });
      }
    } catch (err) {
      holdingMap = new Map(holdingMap).set(inv.investment_id, {
        lots: [], position: null, loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  const LOT_STATUS_LABELS: Record<LotStatus, string> = {
    OPEN:    '持有',
    PARTIAL: '部分',
    CLOSED:  '已清',
  };

  // ── 批次處分展開 ──
  interface DisposalState {
    items:   InvestmentLotDisposal[];
    loading: boolean;
    error:   string;
  }

  let expandedLotId = $state<number | null>(null);
  let disposalMap   = $state(new Map<number, DisposalState>());

  async function toggleDisposal(lot: InvestmentLot): Promise<void> {
    if (expandedLotId === lot.lot_id) {
      expandedLotId = null;
      return;
    }
    expandedLotId = lot.lot_id;
    if (disposalMap.has(lot.lot_id)) return;

    disposalMap = new Map(disposalMap).set(lot.lot_id, {
      items: [], loading: true, error: '',
    });

    try {
      const res = await getLotDisposalsPaged(lot.lot_id, { page: 1, pageSize: 50 });
      disposalMap = new Map(disposalMap).set(lot.lot_id, {
        items: res.data, loading: false, error: '',
      });
    } catch (err) {
      disposalMap = new Map(disposalMap).set(lot.lot_id, {
        items: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  function fmtDec(val: string): string {
    const n = parseFloat(val);
    if (isNaN(n)) return val;
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 6 });
  }

  // ── 購買 / 出售 modal ──
  interface TxnForm {
    date:          string;
    quantity:      string;
    unit_price:    string;
    exchange_rate: string;
    fee:           string;
    tax:           string;
    ledger_id:     string;
  }

  let showTxnModal = $state(false);
  let txnMode      = $state<'buy' | 'sell'>('buy');
  let isTxnSaving  = $state(false);
  let txnError     = $state('');
  let ledgers      = $state<LedgerAccount[]>([]);
  let txnInvId     = $state(0);
  let txnForm      = $state<TxnForm>({
    date:          today(),
    quantity:      '',
    unit_price:    '',
    exchange_rate: '1',
    fee:           '0',
    tax:           '0',
    ledger_id:     '',
  });

  const isTxnValid = $derived(
    txnForm.ledger_id !== '' &&
    parseFloat(txnForm.quantity)      > 0 &&
    parseFloat(txnForm.unit_price)    > 0 &&
    parseFloat(txnForm.exchange_rate) > 0
  );

  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  $effect(() => {
    void load(page);
  });

  async function load(p: number): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      const result = await getInvestmentPaged({ page: p, pageSize: PAGE_SIZE });
      investments = result.data;
      meta        = result.meta;
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  async function openCreateModal(): Promise<void> {
    await ensureAccounts();
    mode        = 'create';
    editId      = 0;
    editVersion = 0;
    form = {
      account_id:  '',
      asset_type:  'STOCK',
      currency:    'TWD',
      symbol:      '',
      name:        '',
      cost_method: 'AVG',
      is_active:   true,
    };
    saveError = '';
    showModal = true;
  }

  async function openEditModal(inv: Investment): Promise<void> {
    await ensureAccounts();
    mode        = 'edit';
    editId      = inv.investment_id;
    editVersion = inv.version;
    form = {
      account_id:  inv.account_id,
      asset_type:  inv.asset_type,
      currency:    inv.currency,
      symbol:      inv.symbol,
      name:        inv.name,
      cost_method: inv.cost_method,
      is_active:   inv.is_active,
    };
    saveError = '';
    showModal = true;
  }

  async function openTxnModal(inv: Investment, m: 'buy' | 'sell'): Promise<void> {
    await ensureLedgers();
    txnMode  = m;
    txnInvId = inv.investment_id;
    txnForm  = {
      date:          today(),
      quantity:      '',
      unit_price:    '',
      exchange_rate: inv.currency === 'TWD' ? '1' : '',
      fee:           '0',
      tax:           '0',
      ledger_id:     '',
    };
    txnError     = '';
    showTxnModal = true;
  }

  async function ensureAccounts(): Promise<void> {
    if (accounts.length === 0) {
      try { accounts = await getAccountAll(); } catch { /* 非致命 */ }
    }
  }

  async function ensureLedgers(): Promise<void> {
    if (ledgers.length === 0) {
      try { ledgers = await getLedgerAccountAll(); } catch { /* 非致命 */ }
    }
  }

  function today(): string {
    return new Date().toISOString().slice(0, 10);
  }

  function closeModal(): void { showModal = false; }
  function closeTxnModal(): void { showTxnModal = false; }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isFormValid) return;
    isSaving  = true;
    saveError = '';
    try {
      if (mode === 'create') {
        await createInvestment({
          ...form,
          symbol:   form.symbol.trim().toUpperCase(),
          currency: form.currency.trim().toUpperCase(),
          name:     form.name.trim(),
        });
      } else {
        await updateInvestment({
          ...form,
          symbol:        form.symbol.trim().toUpperCase(),
          currency:      form.currency.trim().toUpperCase(),
          name:          form.name.trim(),
          investment_id: editId,
          version:       editVersion,
        });
      }
      showModal = false;
      if (page === 1) { await load(1); } else { page = 1; }
    } catch (err) {
      saveError = err instanceof Error ? err.message : (mode === 'create' ? '新增失敗' : '修改失敗') + '，請稍後再試。';
    } finally {
      isSaving = false;
    }
  }

  async function handleTxnSubmit(e: Event): Promise<void> {
    e.preventDefault();
    if (!isTxnValid) return;
    isTxnSaving = true;
    txnError    = '';
    try {
      const payload = {
        investment_id: txnInvId,
        date:          txnForm.date,
        quantity:      txnForm.quantity,
        unit_price:    txnForm.unit_price,
        exchange_rate: txnForm.exchange_rate,
        fee:           txnForm.fee || '0',
        tax:           txnForm.tax || '0',
        ledger_id:     parseInt(txnForm.ledger_id, 10),
      };
      if (txnMode === 'buy') {
        await buyInvestment(payload);
      } else {
        await sellInvestment(payload);
      }
      showTxnModal = false;
      // 清除持倉快取，確保下次展開時重新拉取
      const nextHolding = new Map(holdingMap);
      nextHolding.delete(txnInvId);
      holdingMap = nextHolding;
      disposalMap = new Map();
      if (expandedId === txnInvId) expandedId = null;
      expandedLotId = null;
      if (page === 1) { await load(1); } else { page = 1; }
    } catch (err) {
      txnError = err instanceof Error ? err.message : (txnMode === 'buy' ? '購買失敗' : '出售失敗') + '，請稍後再試。';
    } finally {
      isTxnSaving = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">投資管理</h1>
  {#if meta}
    <span class="content-date">共 {meta.total_count} 筆</span>
  {/if}
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">投資清單</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading">
          <span class="spinner" aria-hidden="true"></span>
          載入中
        </span>
      {/if}
      <button class="section-action" onclick={openCreateModal}>＋ 新增投資</button>
    </div>
  </header>

  <div class="table-wrap">
    <table class="data-table" aria-label="投資清單">
      <thead>
        <tr>
          <th>名稱</th>
          <th>資產類型</th>
          <th>代號</th>
          <th>幣別</th>
          <th>關聯科目</th>
          <th>計價方法</th>
          <th>狀態</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && investments.length === 0}
          <tr><td colspan="8" class="table-empty">載入中...</td></tr>
        {:else if investments.length === 0}
          <tr><td colspan="8" class="table-empty">無資料</td></tr>
        {:else}
          {#each investments as inv (inv.investment_id)}
            {@const holding = holdingMap.get(inv.investment_id)}
            {@const isExpanded = expandedId === inv.investment_id}
            <tr
              class:inv-row-expanded={isExpanded}
              style="cursor:pointer;"
              onclick={() => toggleHolding(inv)}
            >
              <td>
                <span class="inv-expand-icon">{isExpanded ? '▼' : '▶'}</span>
                {inv.name}
              </td>
              <td>
                <span class="badge inv-type-{inv.asset_type}">
                  {ASSET_TYPE_LABELS[inv.asset_type]}
                </span>
              </td>
              <td class="mono">{inv.symbol}</td>
              <td>{inv.currency}</td>
              <td class="mono">{inv.account_id}</td>
              <td>{COST_METHOD_LABELS[inv.cost_method]}</td>
              <td>
                {#if inv.is_active}
                  <span class="badge approved">持有中</span>
                {:else}
                  <span class="badge pending">已結清</span>
                {/if}
              </td>
              <td style="white-space:nowrap;" onclick={(e) => e.stopPropagation()}>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;" onclick={() => openTxnModal(inv, 'buy')}>購買</button>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;margin-left:4px;" onclick={() => openTxnModal(inv, 'sell')}>出售</button>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;margin-left:4px;" onclick={() => openEditModal(inv)}>編輯</button>
              </td>
            </tr>

            {#if isExpanded}
              <tr class="inv-holding-row">
                <td colspan="8" class="inv-holding-cell">
                  {#if holding?.loading}
                    <span class="inv-holding-msg">查詢中…</span>
                  {:else if holding?.error}
                    <span class="inv-holding-msg inv-holding-error">{holding.error}</span>
                  {:else if inv.cost_method === 'FIFO'}
                    {#if holding && holding.lots.length > 0}
                      <div class="inv-lot-table">
                        <div class="inv-lot-header">
                          <span>取得日期</span>
                          <span class="num">數量</span>
                          <span class="num">成本單價</span>
                          <span class="num">總成本</span>
                          <span class="num">剩餘數量</span>
                          <span>狀態</span>
                        </div>
                        {#each holding.lots as lot (lot.lot_id)}
                          {@const isLotExpanded = expandedLotId === lot.lot_id}
                          {@const disposal = disposalMap.get(lot.lot_id)}
                          <div
                            class="inv-lot-row"
                            class:inv-lot-row-expanded={isLotExpanded}
                            role="button"
                            tabindex="0"
                            onclick={() => toggleDisposal(lot)}
                            onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleDisposal(lot); }}
                          >
                            <span class="mono">
                              <span class="inv-expand-icon">{isLotExpanded ? '▼' : '▶'}</span>
                              {lot.acquired_date}
                            </span>
                            <span class="num mono">{fmtDec(lot.quantity)}</span>
                            <span class="num mono">{fmtDec(lot.unit_cost)}</span>
                            <span class="num mono">{fmtDec(lot.total_cost)}</span>
                            <span class="num mono">{fmtDec(lot.remaining_qty)}</span>
                            <span>
                              <span class="badge inv-lot-{lot.status.toLowerCase()}">{LOT_STATUS_LABELS[lot.status]}</span>
                            </span>
                          </div>
                          {#if isLotExpanded}
                            <div class="inv-disposal-panel">
                              {#if disposal?.loading}
                                <span class="inv-holding-msg">查詢中…</span>
                              {:else if disposal?.error}
                                <span class="inv-holding-msg inv-holding-error">{disposal.error}</span>
                              {:else if disposal && disposal.items.length > 0}
                                <div class="inv-disposal-table">
                                  <div class="inv-disposal-header">
                                    <span>處分日期</span>
                                    <span class="num">數量</span>
                                    <span class="num">成本基礎</span>
                                    <span class="num">出售金額</span>
                                    <span class="num">資本利得</span>
                                    <span class="num">持有天數</span>
                                  </div>
                                  {#each disposal.items as d, i (i)}
                                    <div class="inv-disposal-row">
                                      <span class="mono">{d.disposal_date}</span>
                                      <span class="num mono">{fmtDec(d.quantity)}</span>
                                      <span class="num mono">{fmtDec(d.cost_basis)}</span>
                                      <span class="num mono">{fmtDec(d.sale_proceeds)}</span>
                                      <span class="num mono">{fmtDec(d.capital_gain)}</span>
                                      <span class="num">{d.holding_period_days ?? '—'}</span>
                                    </div>
                                  {/each}
                                </div>
                              {:else}
                                <span class="inv-holding-msg">尚無處分紀錄</span>
                              {/if}
                            </div>
                          {/if}
                        {/each}
                      </div>
                    {:else}
                      <span class="inv-holding-msg">尚無庫存批次</span>
                    {/if}
                  {:else}
                    {#if holding?.position}
                      {@const pos = holding.position}
                      <div class="inv-position">
                        <div class="inv-position-item">
                          <span class="inv-position-label">總數量</span>
                          <span class="inv-position-value">{fmtDec(pos.total_quantity)}</span>
                        </div>
                        <div class="inv-position-item">
                          <span class="inv-position-label">總成本</span>
                          <span class="inv-position-value">{fmtDec(pos.total_cost)}</span>
                        </div>
                        <div class="inv-position-item">
                          <span class="inv-position-label">平均成本</span>
                          <span class="inv-position-value">{fmtDec(pos.avg_cost)}</span>
                        </div>
                      </div>
                    {:else}
                      <span class="inv-holding-msg">尚無持倉資料</span>
                    {/if}
                  {/if}
                </td>
              </tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  {#if meta && meta.total_pages > 1}
    <div class="pagination">
      <button
        class="page-btn"
        onclick={() => { page -= 1; }}
        disabled={page <= 1 || isLoading}
        aria-label="上一頁"
      >←</button>
      <span class="page-info">
        第 <span class="page-num">{page}</span> 頁 &nbsp;/&nbsp; 共 {meta.total_pages} 頁
      </span>
      <button
        class="page-btn"
        onclick={() => { page += 1; }}
        disabled={page >= meta.total_pages || isLoading}
        aria-label="下一頁"
      >→</button>
    </div>
  {/if}
</section>

<!-- ── 新增 / 修改投資 Modal ───────────────── -->
{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeModal(); }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="modal-title">{mode === 'create' ? '新增投資項目' : '修改投資項目'}</h2>
        <button class="modal-close" onclick={closeModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        <div class="form-group">
          <!-- svelte-ignore a11y_label_has_associated_control -->
          <label class="form-label">關聯科目 *</label>
          <AccountSelect
            accounts={accounts}
            value={form.account_id}
            placeholder="選擇投資關聯的會計科目…"
            onselect={(id) => { form.account_id = id; }}
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-symbol">代號 *</label>
            <input id="f-symbol" class="form-input" type="text" bind:value={form.symbol} placeholder="例：2330" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="f-name">名稱 *</label>
            <input id="f-name" class="form-input" type="text" bind:value={form.name} placeholder="例：台積電" required />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-asset-type">資產類型 *</label>
            <select id="f-asset-type" class="form-select" onchange={(e) => { form.asset_type = (e.target as HTMLSelectElement).value as AssetType; }}>
              {#each ASSET_TYPES as t}
                <option value={t} selected={form.asset_type === t}>{ASSET_TYPE_LABELS[t]}</option>
              {/each}
            </select>
          </div>
          <div class="form-group">
            <label class="form-label" for="f-currency">幣別 *</label>
            <input id="f-currency" class="form-input" type="text" bind:value={form.currency} placeholder="TWD" required />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="f-cost-method">計價方法 *</label>
            <select id="f-cost-method" class="form-select" onchange={(e) => { form.cost_method = (e.target as HTMLSelectElement).value as CostMethod; }}>
              {#each COST_METHODS as m}
                <option value={m} selected={form.cost_method === m}>{COST_METHOD_LABELS[m]}</option>
              {/each}
            </select>
          </div>
          <div class="form-group" style="display:flex;align-items:center;padding-top:26px;">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
              <input type="checkbox" bind:checked={form.is_active} />
              啟用（持有中）
            </label>
          </div>
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeModal} disabled={isSaving}>取消</button>
          <button type="submit" class="btn-primary" disabled={isSaving || !isFormValid}>
            {isSaving ? '儲存中…' : mode === 'create' ? '新增投資' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 購買 / 出售 Modal ────────────────────── -->
{#if showTxnModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) closeTxnModal(); }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="txn-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="txn-modal-title">{txnMode === 'buy' ? '購買投資' : '出售投資'}</h2>
        <button class="modal-close" onclick={closeTxnModal} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleTxnSubmit}>
        {#if txnError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{txnError}</p>
        {/if}

        <div class="form-group">
          <!-- svelte-ignore a11y_label_has_associated_control -->
          <label class="form-label">{txnMode === 'buy' ? '扣款帳戶' : '入帳帳戶'} *</label>
          <LedgerSelect
            ledgers={activeLedgers}
            value={txnForm.ledger_id}
            onselect={(id) => { txnForm.ledger_id = id; }}
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="t-date">交易日期 *</label>
            <input id="t-date" class="form-input" type="date" bind:value={txnForm.date} required />
          </div>
          <div class="form-group">
            <label class="form-label" for="t-qty">數量 *</label>
            <input
              id="t-qty"
              class="form-input"
              type="number"
              min="0.0001"
              step="any"
              bind:value={txnForm.quantity}
              placeholder="0"
              required
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="t-price">單價 *</label>
            <input
              id="t-price"
              class="form-input"
              type="number"
              min="0.0001"
              step="any"
              bind:value={txnForm.unit_price}
              placeholder="0.00"
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="t-rate">匯率 *</label>
            <input
              id="t-rate"
              class="form-input"
              type="number"
              min="0.0001"
              step="any"
              bind:value={txnForm.exchange_rate}
              placeholder="1"
              required
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="t-fee">手續費</label>
            <input
              id="t-fee"
              class="form-input"
              type="number"
              min="0"
              step="any"
              bind:value={txnForm.fee}
              placeholder="0"
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="t-tax">稅金</label>
            <input
              id="t-tax"
              class="form-input"
              type="number"
              min="0"
              step="any"
              bind:value={txnForm.tax}
              placeholder="0"
            />
          </div>
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={closeTxnModal} disabled={isTxnSaving}>取消</button>
          <button type="submit" class="btn-primary" disabled={isTxnSaving || !isTxnValid}>
            {isTxnSaving ? '儲存中…' : txnMode === 'buy' ? '確認購買' : '確認出售'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
