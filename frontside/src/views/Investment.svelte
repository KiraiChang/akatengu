<script lang="ts">
  import { getInvestmentPaged, createInvestment, updateInvestment, buyInvestment, sellInvestment, getOpenLotsPaged, getPosition, getLotDisposalsPaged, getMovementPaged } from '../api/investment';
  import { getAccountAll, createAccount } from '../api/account';
  import { getLedgerAccountAll, createLedgerAccount, invalidateLedgerCache } from '../api/ledger';
  import { getEntries } from '../api/transaction';
  import AccountSelect from '../components/AccountSelect.svelte';
  import NewAccountFormSection, { type NewAccountForm, emptyNewAccountForm } from '../components/NewAccountFormSection.svelte';
  import LedgerSelectSection from '../components/LedgerSelectSection.svelte';
  import TxnEntryPanel from '../components/TxnEntryPanel.svelte';
  import { type NewLedgerForm, emptyNewLedgerForm } from '../components/NewLedgerFormSection.svelte';
  import type { Investment, AssetType, CostMethod, IFRSCategory, InvestmentCreatedPayload, InvestmentLot, InvestmentPosition, InvestmentLotDisposal, LotStatus, InvestmentMovement, MovementType } from '../types/investment';
  import type { Entry } from '../types/transaction';
  import type { Account, CreateAccountRequest } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type { PaginatedMeta } from '../types/pagination';
  import { getLedgerAccountTypeConfigs, getAssetTypeConfigs } from '../api/setting';
  import type { LedgerAccountTypeConfigResult, AssetTypeAccountConfigResult } from '../types/setting';

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

  const IFRS_CATEGORY_LABELS: Record<IFRSCategory, string> = {
    FVTPL: '損益公允價值',
    FVOCI: '其他綜合損益',
    AC:    '攤銷成本',
  };

  const MOVEMENT_TYPE_LABELS: Record<MovementType, string> = {
    BUY:      '購入',
    SELL:     '賣出',
    DIVIDEND: '配息',
    SPLIT:    '分割',
    CONVERT:  '轉換',
  };

  const ASSET_TYPES:      AssetType[]    = ['STOCK', 'FUND', 'GOLD', 'FX'];
  const COST_METHODS:     CostMethod[]   = ['AVG', 'FIFO'];
  const IFRS_CATEGORIES:  IFRSCategory[] = ['FVTPL', 'FVOCI', 'AC'];

  let page        = $state(1);
  let investments = $state<Investment[]>([]);
  let meta        = $state<PaginatedMeta | null>(null);
  let isLoading   = $state(false);
  let error       = $state('');

  // ── 新增 / 修改投資 modal ──
  let showModal        = $state(false);
  let mode             = $state<'create' | 'edit'>('create');
  let isSaving         = $state(false);
  let saveError        = $state('');
  let accounts         = $state<Account[]>([]);
  let editId           = $state(0);
  let editVersion      = $state(0);
  let createNewAccount = $state(false);
  let newAccountForm   = $state<NewAccountForm>(emptyNewAccountForm());
  let form             = $state<InvestmentCreatedPayload>({
    account_id:    '',
    asset_type:    'STOCK',
    currency:      'TWD',
    symbol:        '',
    name:          '',
    cost_method:   'AVG',
    ifrs_category: 'FVTPL',
    is_active:     true,
  });

  const isFormValid = $derived(
    (createNewAccount
      ? newAccountForm.account_id.trim() !== '' && newAccountForm.name.trim() !== ''
      : form.account_id.trim() !== '') &&
    form.symbol.trim()   !== '' &&
    form.name.trim()     !== '' &&
    form.currency.trim() !== ''
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

  // ── 批次分錄展開 ──
  interface EntriesState {
    rows:    Entry[];
    loading: boolean;
    error:   string;
  }

  let expandedLotEntryId      = $state<number | null>(null);
  let lotEntriesMap           = $state(new Map<number, EntriesState>());
  let expandedDisposalEntryId = $state<number | null>(null);
  let disposalEntriesMap      = $state(new Map<number, EntriesState>());

  // ── 異動標籤 ──
  let holdingTabMap = $state(new Map<number, 'holding' | 'movement'>());

  // ── 異動列表 ──
  interface MovementState {
    items:   InvestmentMovement[];
    loading: boolean;
    error:   string;
    page:    number;
    meta:    PaginatedMeta | null;
  }

  let movementMap = $state(new Map<number, MovementState>());

  // ── 異動分錄展開 ──
  let expandedMovementEntryId = $state<number | null>(null);
  let movementEntriesMap      = $state(new Map<number, EntriesState>());

  async function toggleLotEntries(lot: InvestmentLot, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (!lot.txn_id) return;
    const txnId = lot.txn_id;

    if (expandedLotEntryId === lot.lot_id) {
      expandedLotEntryId = null;
      return;
    }
    expandedLotEntryId = lot.lot_id;
    if (lotEntriesMap.has(lot.lot_id)) return;

    lotEntriesMap = new Map(lotEntriesMap).set(lot.lot_id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      lotEntriesMap = new Map(lotEntriesMap).set(lot.lot_id, { rows, loading: false, error: '' });
    } catch (err) {
      lotEntriesMap = new Map(lotEntriesMap).set(lot.lot_id, {
        rows: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function toggleDisposalEntries(txnId: number): Promise<void> {
    if (expandedDisposalEntryId === txnId) {
      expandedDisposalEntryId = null;
      return;
    }
    expandedDisposalEntryId = txnId;
    if (disposalEntriesMap.has(txnId)) return;

    disposalEntriesMap = new Map(disposalEntriesMap).set(txnId, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      disposalEntriesMap = new Map(disposalEntriesMap).set(txnId, { rows, loading: false, error: '' });
    } catch (err) {
      disposalEntriesMap = new Map(disposalEntriesMap).set(txnId, {
        rows: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  function getHoldingTab(investmentId: number): 'holding' | 'movement' {
    return holdingTabMap.get(investmentId) ?? 'holding';
  }

  function setHoldingTab(investmentId: number, tab: 'holding' | 'movement'): void {
    holdingTabMap = new Map(holdingTabMap).set(investmentId, tab);
    if (tab === 'movement' && !movementMap.has(investmentId)) {
      void loadMovements(investmentId, 1);
    }
  }

  async function loadMovements(investmentId: number, p: number): Promise<void> {
    movementMap = new Map(movementMap).set(investmentId, {
      items: [], loading: true, error: '', page: p, meta: null,
    });
    try {
      const res = await getMovementPaged(investmentId, { page: p, pageSize: 50 });
      movementMap = new Map(movementMap).set(investmentId, {
        items: res.data, loading: false, error: '', page: p, meta: res.meta,
      });
    } catch (err) {
      movementMap = new Map(movementMap).set(investmentId, {
        items: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
        page: p, meta: null,
      });
    }
  }

  async function toggleMovementEntries(mov: InvestmentMovement, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (!mov.txn_id) return;
    const txnId = mov.txn_id;

    if (expandedMovementEntryId === mov.movement_id) {
      expandedMovementEntryId = null;
      return;
    }
    expandedMovementEntryId = mov.movement_id;
    if (movementEntriesMap.has(mov.movement_id)) return;

    movementEntriesMap = new Map(movementEntriesMap).set(mov.movement_id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      movementEntriesMap = new Map(movementEntriesMap).set(mov.movement_id, { rows, loading: false, error: '' });
    } catch (err) {
      movementEntriesMap = new Map(movementEntriesMap).set(mov.movement_id, {
        rows: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

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
  let txnCreateNewLedger    = $state(false);
  let txnNewLedgerForm      = $state<NewLedgerForm>(emptyNewLedgerForm());
  let txnNewLedgerAccountId = $state('');
  let txnCreateNewAccount   = $state(false);
  let txnNewAccountForm     = $state<NewAccountForm>(emptyNewAccountForm());
  let ledgerTypeConfigMap   = $state(new Map<string, LedgerAccountTypeConfigResult>());
  let assetTypeConfigMap    = $state(new Map<AssetType, AssetTypeAccountConfigResult>());

  const isTxnValid = $derived(
    (txnCreateNewLedger
      ? txnNewLedgerForm.institution.trim() !== '' &&
        txnNewLedgerForm.name.trim()        !== '' &&
        (txnCreateNewAccount
          ? txnNewAccountForm.account_id.trim() !== '' && txnNewAccountForm.name.trim() !== ''
          : txnNewLedgerAccountId                  !== '')
      : txnForm.ledger_id !== '') &&
    parseFloat(txnForm.quantity)      > 0 &&
    parseFloat(txnForm.unit_price)    > 0 &&
    parseFloat(txnForm.exchange_rate) > 0
  );

  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  const activeInvAssetTypeConfig = $derived(assetTypeConfigMap.get(form.asset_type) ?? null);

  const filteredInvAccounts = $derived((() => {
    if (!activeInvAssetTypeConfig?.account_id || activeInvAssetTypeConfig.descendants.length === 0) {
      return accounts;
    }
    const path: Account[] = [];
    let cur: string | null = activeInvAssetTypeConfig.account_id;
    while (cur) {
      const a = accounts.find(x => x.account_id === cur);
      if (!a) break;
      path.push(a);
      cur = a.parent_id ?? null;
    }
    return [...path.reverse(), ...activeInvAssetTypeConfig.descendants];
  })());

  const activeTxnLedgerTypeConfig = $derived(ledgerTypeConfigMap.get(txnNewLedgerForm.type) ?? null);
  const filteredTxnAccounts = $derived((() => {
    if (!activeTxnLedgerTypeConfig?.account_id || activeTxnLedgerTypeConfig.descendants.length === 0) {
      return accounts;
    }
    const path: Account[] = [];
    let cur: string | null = activeTxnLedgerTypeConfig.account_id;
    while (cur) {
      const a = accounts.find(x => x.account_id === cur);
      if (!a) break;
      path.push(a);
      cur = a.parent_id ?? null;
    }
    return [...path.reverse(), ...activeTxnLedgerTypeConfig.descendants];
  })());

  let _prevInvAssetType = $state('');
  $effect(() => {
    const t = form.asset_type as string;
    if (mode === 'create' && _prevInvAssetType !== '' && _prevInvAssetType !== t) {
      form.account_id = '';
    }
    _prevInvAssetType = t;
  });

  let _prevTxnLedgerType = $state('');
  $effect(() => {
    const t = txnNewLedgerForm.type as string;
    if (_prevTxnLedgerType !== '' && _prevTxnLedgerType !== t) {
      txnNewLedgerAccountId       = '';
      txnNewAccountForm.parent_id = null;
    }
    _prevTxnLedgerType = t;
  });

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
    await Promise.all([ensureAccounts(), ensureAssetTypeConfigs()]);
    mode             = 'create';
    editId           = 0;
    editVersion      = 0;
    createNewAccount = false;
    newAccountForm   = emptyNewAccountForm();
    form = {
      account_id:    '',
      asset_type:    'STOCK',
      currency:      'TWD',
      symbol:        '',
      name:          '',
      cost_method:   'AVG',
      ifrs_category: 'FVTPL',
      is_active:     true,
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
      account_id:    inv.account_id,
      asset_type:    inv.asset_type,
      currency:      inv.currency,
      symbol:        inv.symbol,
      name:          inv.name,
      cost_method:   inv.cost_method,
      ifrs_category: inv.ifrs_category,
      is_active:     inv.is_active,
    };
    saveError = '';
    showModal = true;
  }

  async function openTxnModal(inv: Investment, m: 'buy' | 'sell'): Promise<void> {
    await Promise.all([ensureLedgers(), ensureAccounts(), ensureConfigs()]);
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
    txnCreateNewLedger    = false;
    txnNewLedgerForm      = emptyNewLedgerForm();
    txnNewLedgerAccountId = '';
    txnCreateNewAccount   = false;
    txnNewAccountForm     = emptyNewAccountForm();
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

  async function ensureConfigs(): Promise<void> {
    if (ledgerTypeConfigMap.size === 0) {
      try {
        const configs = await getLedgerAccountTypeConfigs();
        ledgerTypeConfigMap = new Map(configs.map(c => [c.type, c]));
      } catch { /* 非致命 */ }
    }
  }

  async function ensureAssetTypeConfigs(): Promise<void> {
    if (assetTypeConfigMap.size === 0) {
      try {
        const configs = await getAssetTypeConfigs();
        assetTypeConfigMap = new Map(configs.map(c => [c.asset_type, c]));
      } catch { /* 非致命 */ }
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

    if (mode === 'create' && createNewAccount) {
      if (!newAccountForm.account_id.trim() || !newAccountForm.name.trim()) {
        saveError = '請填寫完整的會計科目資料';
        return;
      }
    }

    isSaving  = true;
    saveError = '';
    try {
      if (mode === 'create') {
        if (createNewAccount) {
          const accountPayload: CreateAccountRequest = {
            account_id:     newAccountForm.account_id.trim(),
            parent_id:      newAccountForm.parent_id || null,
            name:           newAccountForm.name.trim(),
            type:           newAccountForm.type,
            normal_balance: newAccountForm.normal_balance,
            currency:       newAccountForm.currency,
            is_summary:     newAccountForm.is_summary,
            is_active:      newAccountForm.is_active,
            note:           null,
          };
          await createAccount(accountPayload);
          accounts = [];
        }

        const targetAccountId = createNewAccount ? newAccountForm.account_id : form.account_id;
        await createInvestment({
          ...form,
          account_id: targetAccountId,
          symbol:     form.symbol.trim().toUpperCase(),
          currency:   form.currency.trim().toUpperCase(),
          name:       form.name.trim(),
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
      let finalLedgerId: number;

      if (txnCreateNewLedger) {
        let ledgerAccountId: string;
        if (txnCreateNewAccount) {
          if (!txnNewAccountForm.account_id.trim() || !txnNewAccountForm.name.trim()) {
            txnError = '請填寫完整的會計科目資料';
            return;
          }
          const accountPayload: CreateAccountRequest = {
            account_id:     txnNewAccountForm.account_id.trim(),
            parent_id:      txnNewAccountForm.parent_id || null,
            name:           txnNewAccountForm.name.trim(),
            type:           txnNewAccountForm.type,
            normal_balance: txnNewAccountForm.normal_balance,
            currency:       txnNewAccountForm.currency,
            is_summary:     txnNewAccountForm.is_summary,
            is_active:      txnNewAccountForm.is_active,
            note:           null,
          };
          await createAccount(accountPayload);
          accounts = [];
          ledgerAccountId = txnNewAccountForm.account_id.trim();
        } else {
          ledgerAccountId = txnNewLedgerAccountId;
        }

        const creditLimit = txnNewLedgerForm.type === 'CREDIT_CARD' && txnNewLedgerForm.creditLimitInput
          ? txnNewLedgerForm.creditLimitInput : null;
        await createLedgerAccount({
          account_id:   ledgerAccountId,
          institution:  txnNewLedgerForm.institution.trim(),
          name:         txnNewLedgerForm.name.trim(),
          type:         txnNewLedgerForm.type,
          account_no:   txnNewLedgerForm.account_no?.trim() || null,
          currency:     txnNewLedgerForm.currency,
          credit_limit: creditLimit,
          billing_day:  txnNewLedgerForm.type === 'CREDIT_CARD' ? txnNewLedgerForm.billing_day || null : null,
          due_day:      txnNewLedgerForm.type === 'CREDIT_CARD' ? txnNewLedgerForm.due_day || null : null,
          is_active:    txnNewLedgerForm.is_active,
          note:         null,
        });

        invalidateLedgerCache();
        const freshLedgers = await getLedgerAccountAll(true);
        ledgers = freshLedgers;
        const created = freshLedgers.find(l => l.account_id === ledgerAccountId);
        if (!created) throw new Error('無法取得新建帳戶，請重試');
        finalLedgerId = created.ledger_id;
      } else {
        finalLedgerId = parseInt(txnForm.ledger_id, 10);
      }

      const payload = {
        investment_id: txnInvId,
        date:          txnForm.date,
        quantity:      txnForm.quantity,
        unit_price:    txnForm.unit_price,
        exchange_rate: txnForm.exchange_rate,
        fee:           txnForm.fee || '0',
        tax:           txnForm.tax || '0',
        ledger_id:     finalLedgerId,
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
      const nextMovement = new Map(movementMap);
      nextMovement.delete(txnInvId);
      movementMap = nextMovement;
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

{#snippet invHoldingInner(inv: Investment)}
  {@const holding = holdingMap.get(inv.investment_id)}
  {@const activeTab = getHoldingTab(inv.investment_id)}
  <div class="inv-tab-bar">
    <button
      type="button"
      class="inv-tab-btn"
      class:inv-tab-btn--active={activeTab === 'holding'}
      onclick={() => setHoldingTab(inv.investment_id, 'holding')}
    >持倉</button>
    <button
      type="button"
      class="inv-tab-btn"
      class:inv-tab-btn--active={activeTab === 'movement'}
      onclick={() => setHoldingTab(inv.investment_id, 'movement')}
    >異動列表</button>
  </div>
  {#if activeTab === 'holding'}
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
          <span class="num">未實現單位公允(TWD)</span>
          <span>更新者</span>
          <span>更新時間</span>
          <span>狀態</span>
        </div>
        {#each holding.lots as lot (lot.lot_id)}
          {@const isLotExpanded = expandedLotId === lot.lot_id}
          {@const isEntryExpanded = expandedLotEntryId === lot.lot_id}
          {@const disposal = disposalMap.get(lot.lot_id)}
          {@const entryState = lotEntriesMap.get(lot.lot_id)}
          <div
            class="inv-lot-row"
            class:inv-lot-row-expanded={isLotExpanded}
            role="button"
            tabindex="0"
            onclick={() => toggleDisposal(lot)}
            onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleDisposal(lot); }}
          >
            <span class="mono" style="display:flex;align-items:center;gap:6px;">
              <span class="inv-expand-icon">{isLotExpanded ? '▼' : '▶'}</span>
              {lot.acquired_date}
              {#if lot.txn_id !== null}
                <button
                  type="button"
                  class="btn-ghost"
                  style="padding:1px 6px;font-size:10px;"
                  class:acct-mode-btn--active={isEntryExpanded}
                  onclick={(e) => toggleLotEntries(lot, e)}
                >分錄</button>
              {/if}
            </span>
            <span class="num mono">{fmtDec(lot.quantity)}</span>
            <span class="num mono">{fmtDec(lot.unit_cost)}</span>
            <span class="num mono">{fmtDec(lot.total_cost)}</span>
            <span class="num mono">{fmtDec(lot.remaining_qty)}</span>
            <span class="num mono">{fmtDec(lot.unrealized_unit_twd)}</span>
            <span>{lot.updated_by ?? '—'}</span>
            <span>{lot.updated_at ?? '—'}</span>
            <span>
              <span class="badge inv-lot-{lot.status.toLowerCase()}">{LOT_STATUS_LABELS[lot.status]}</span>
            </span>
          </div>
          {#if isEntryExpanded}
            <div class="inv-disposal-panel">
              <TxnEntryPanel
                rows={entryState?.rows ?? []}
                loading={entryState?.loading ?? false}
                error={entryState?.error ?? ''}
                accounts={accounts}
                ledgers={ledgers}
              />
            </div>
          {/if}
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
                  {#each disposal.items as d (d.id)}
                    {@const isDisposalEntryExpanded = d.txn_id !== null && expandedDisposalEntryId === d.txn_id}
                    {@const disposalEntryState = d.txn_id !== null ? disposalEntriesMap.get(d.txn_id) : undefined}
                    <div class="inv-disposal-row">
                      <span class="mono" style="display:flex;align-items:center;gap:6px;">
                        {d.disposal_date}
                        {#if d.txn_id !== null}
                          <button
                            type="button"
                            class="btn-ghost"
                            style="padding:1px 6px;font-size:10px;"
                            class:acct-mode-btn--active={isDisposalEntryExpanded}
                            onclick={() => toggleDisposalEntries(d.txn_id!)}
                          >分錄</button>
                        {/if}
                      </span>
                      <span class="num mono">{fmtDec(d.quantity)}</span>
                      <span class="num mono">{fmtDec(d.cost_basis)}</span>
                      <span class="num mono">{fmtDec(d.sale_proceeds)}</span>
                      <span class="num mono">{fmtDec(d.capital_gain)}</span>
                      <span class="num">{d.holding_period_days ?? '—'}</span>
                    </div>
                    {#if isDisposalEntryExpanded}
                      <div class="inv-disposal-panel">
                        <TxnEntryPanel
                          rows={disposalEntryState?.rows ?? []}
                          loading={disposalEntryState?.loading ?? false}
                          error={disposalEntryState?.error ?? ''}
                          accounts={accounts}
                          ledgers={ledgers}
                        />
                      </div>
                    {/if}
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
        <div class="inv-position-item">
          <span class="inv-position-label">公允單價(TWD)</span>
          <span class="inv-position-value">{fmtDec(pos.market_price_twd)}</span>
        </div>
      </div>
    {:else}
      <span class="inv-holding-msg">尚無持倉資料</span>
    {/if}
  {/if}
  {:else}
    {@const movState = movementMap.get(inv.investment_id)}
    {#if movState?.loading}
      <span class="inv-holding-msg">查詢中…</span>
    {:else if movState?.error}
      <span class="inv-holding-msg inv-holding-error">{movState.error}</span>
    {:else if movState && movState.items.length > 0}
      <div class="inv-movement-table">
        <div class="inv-movement-header">
          <span>異動日期</span>
          <span>類型</span>
          <span class="num">數量</span>
          <span class="num">單價(TWD)</span>
          <span class="num">手續費</span>
          <span class="num">稅金</span>
          <span class="num">資本利得</span>
          <span>更新者</span>
          <span>更新時間</span>
          <span></span>
        </div>
        {#each movState.items as mov (mov.movement_id)}
          {@const isMovEntryExpanded = expandedMovementEntryId === mov.movement_id}
          {@const movEntryState = movementEntriesMap.get(mov.movement_id)}
          <div class="inv-movement-row">
            <span class="mono">{mov.movement_date}</span>
            <span>
              <span class="badge inv-mov-{mov.movement_type}">
                {MOVEMENT_TYPE_LABELS[mov.movement_type]}
              </span>
            </span>
            <span class="num mono">{fmtDec(mov.quantity)}</span>
            <span class="num mono">{fmtDec(mov.unit_price_twd)}</span>
            <span class="num mono">{fmtDec(mov.fee)}</span>
            <span class="num mono">{fmtDec(mov.tax)}</span>
            <span class="num mono">{mov.realized_gain !== null ? fmtDec(mov.realized_gain) : '—'}</span>
            <span>{mov.updated_by ?? '—'}</span>
            <span>{mov.updated_at ?? '—'}</span>
            <span>
              {#if mov.txn_id !== null}
                <button
                  type="button"
                  class="btn-ghost"
                  style="padding:1px 6px;font-size:10px;"
                  class:acct-mode-btn--active={isMovEntryExpanded}
                  onclick={(e) => toggleMovementEntries(mov, e)}
                >分錄</button>
              {/if}
            </span>
          </div>
          {#if isMovEntryExpanded}
            <div class="inv-disposal-panel">
              <TxnEntryPanel
                rows={movEntryState?.rows ?? []}
                loading={movEntryState?.loading ?? false}
                error={movEntryState?.error ?? ''}
                accounts={accounts}
                ledgers={ledgers}
              />
            </div>
          {/if}
        {/each}
      </div>
    {:else}
      <span class="inv-holding-msg">尚無異動紀錄</span>
    {/if}
  {/if}
{/snippet}

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

  <div class="table-wrap inv-table-wrap">
    <table class="data-table inv-table" aria-label="投資清單">
      <thead>
        <tr>
          <th>名稱</th>
          <th>資產類型</th>
          <th>代號</th>
          <th>幣別</th>
          <th>關聯科目</th>
          <th>計價方法</th>
          <th>IFRS 分類</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && investments.length === 0}
          <tr><td colspan="11" class="table-empty">載入中...</td></tr>
        {:else if investments.length === 0}
          <tr><td colspan="11" class="table-empty">無資料</td></tr>
        {:else}
          {#each investments as inv (inv.investment_id)}
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
                <span class="badge inv-ifrs-{inv.ifrs_category.toLowerCase()}">
                  {IFRS_CATEGORY_LABELS[inv.ifrs_category]}
                </span>
              </td>
              <td class="hidden md:table-cell">{inv.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{inv.updated_at ?? '—'}</td>
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
                <td colspan="11" class="inv-holding-cell">
                  {@render invHoldingInner(inv)}
                </td>
              </tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="inv-card-list">
    {#if isLoading && investments.length === 0}
      <p class="query-loading" style="padding:16px 0;">載入中...</p>
    {:else if investments.length === 0}
      <p style="padding:16px 0;font-size:12px;color:#3d4258;">無資料</p>
    {:else}
      {#each investments as inv (inv.investment_id)}
        {@const isExpanded = expandedId === inv.investment_id}
        <div class="inv-card {isExpanded ? 'inv-card-expanded' : ''}">
          <div class="inv-card-main" role="button" tabindex="0" onclick={() => toggleHolding(inv)} onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && toggleHolding(inv)} style="cursor:pointer;">
            <div class="inv-card-head">
              <span class="inv-card-title">
                <span class="inv-expand-icon">{isExpanded ? '▼' : '▶'}</span>
                {inv.name}
              </span>
              <span class="badge inv-type-{inv.asset_type}">{ASSET_TYPE_LABELS[inv.asset_type]}</span>
            </div>
            <div class="inv-card-meta-row">
              <span class="inv-card-symbol mono">{inv.symbol}</span>
              <span class="inv-card-sep">·</span>
              <span class="inv-card-currency">{inv.currency}</span>
            </div>
            <div class="inv-card-acct mono">{inv.account_id}</div>
            <div class="inv-card-attrs">
              <span class="inv-card-attr">{COST_METHOD_LABELS[inv.cost_method]}</span>
              <span class="badge inv-ifrs-{inv.ifrs_category.toLowerCase()}">{IFRS_CATEGORY_LABELS[inv.ifrs_category]}</span>
            </div>
            <div class="inv-card-footer">
              {#if inv.is_active}
                <span class="badge approved">持有中</span>
              {:else}
                <span class="badge pending">已結清</span>
              {/if}
              <div>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;" onclick={(e) => { e.stopPropagation(); openTxnModal(inv, 'buy'); }}>購買</button>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;margin-left:4px;" onclick={(e) => { e.stopPropagation(); openTxnModal(inv, 'sell'); }}>出售</button>
                <button class="btn-ghost" style="padding:2px 8px;font-size:11px;margin-left:4px;" onclick={(e) => { e.stopPropagation(); openEditModal(inv); }}>編輯</button>
              </div>
            </div>
          </div>
          {#if isExpanded}
            <div class="inv-card-holding">
              {@render invHoldingInner(inv)}
            </div>
          {/if}
        </div>
      {/each}
    {/if}
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

        {#if mode === 'create'}
          <div class="form-group" style="margin-bottom:4px;">
            <span class="form-label">關聯科目 *</span>
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
            <div class="form-group">
              <AccountSelect
                accounts={filteredInvAccounts}
                value={form.account_id}
                placeholder="選擇投資關聯的會計科目…"
                onselect={(id) => { form.account_id = id; }}
              />
            </div>
          {:else}
            <NewAccountFormSection
              accounts={accounts}
              bind:form={newAccountForm}
              required={createNewAccount}
            />
          {/if}
        {:else}
          <div class="form-group">
            <label class="form-label" for="account_id">關聯科目 *</label>
            <AccountSelect
              accounts={accounts}
              value={form.account_id}
              placeholder="選擇投資關聯的會計科目…"
              onselect={(id) => { form.account_id = id; }}
            />
          </div>
        {/if}

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
          <div class="form-group">
            <label class="form-label" for="f-ifrs-category">IFRS 分類 *</label>
            <select id="f-ifrs-category" class="form-select" onchange={(e) => { form.ifrs_category = (e.target as HTMLSelectElement).value as IFRSCategory; }}>
              {#each IFRS_CATEGORIES as c}
                <option value={c} selected={form.ifrs_category === c}>{IFRS_CATEGORY_LABELS[c]}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="form-group" style="display:flex;align-items:center;">
          <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
            <input type="checkbox" bind:checked={form.is_active} />
            啟用（持有中）
          </label>
        </div>

        {#if mode === 'edit'}
          {@const inv = investments.find(i => i.investment_id === editId)}
          {#if inv}
            <div class="form-row" style="margin-top:8px;">
              <div class="form-group">
                <label class="form-label" for="f-inv-updated-by">更新者</label>
                <input id="f-inv-updated-by" class="form-input" type="text" value={inv.updated_by ?? '—'} readonly style="background:#f5f0e8;cursor:default;" />
              </div>
              <div class="form-group">
                <label class="form-label" for="f-inv-updated-at">更新時間</label>
                <input id="f-inv-updated-at" class="form-input" type="text" value={inv.updated_at ?? '—'} readonly style="background:#f5f0e8;cursor:default;" />
              </div>
            </div>
          {/if}
        {/if}

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

        <LedgerSelectSection
          ledgers={activeLedgers}
          accounts={filteredTxnAccounts}
          accountsForParent={filteredTxnAccounts}
          bind:ledgerId={txnForm.ledger_id}
          bind:createNew={txnCreateNewLedger}
          bind:newLedgerForm={txnNewLedgerForm}
          bind:createNewAccount={txnCreateNewAccount}
          bind:newAccountForm={txnNewAccountForm}
          bind:accountId={txnNewLedgerAccountId}
          label={txnMode === 'buy' ? '扣款帳戶' : '入帳帳戶'}
        />

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
