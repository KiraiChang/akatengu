<script lang="ts">
  import { getInstallmentPaged, getInstallmentPaymentPaged, createInstallment, payInstallmentPeriod } from '../api/installment';
  import { getAccountAll, createAccount } from '../api/account';
  import { getLedgerAccountAll, createLedgerAccount, invalidateLedgerCache } from '../api/ledger';
  import { getEntries } from '../api/transaction';
  import AccountSelect from '../components/AccountSelect.svelte';
  import LedgerSelect from '../components/LedgerSelect.svelte';
  import { type NewAccountForm, emptyNewAccountForm } from '../components/NewAccountFormSection.svelte';
  import { type NewLedgerForm, emptyNewLedgerForm } from '../components/NewLedgerFormSection.svelte';
  import LedgerSelectSection from '../components/LedgerSelectSection.svelte';
  import TxnEntryPanel from '../components/TxnEntryPanel.svelte';
  import type { Installment, InstallmentPayment, InstallmentStatus, InstallmentPaymentStatus, InterestType, InstallmentCreatedPayload } from '../types/installment';
  import type { Entry } from '../types/transaction';
  import type { Account, CreateAccountRequest } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type { PaginatedMeta } from '../types/pagination';

  const PAGE_SIZE = 20;

  const STATUS_LABELS: Record<InstallmentStatus, string> = {
    ACTIVE:    '進行中',
    COMPLETED: '已完成',
    CANCELED:  '已取消',
  };

  const STATUS_CSS: Record<InstallmentStatus, string> = {
    ACTIVE:    'inst-status-active',
    COMPLETED: 'inst-status-completed',
    CANCELED:  'inst-status-canceled',
  };

  const PAY_STATUS_LABELS: Record<InstallmentPaymentStatus, string> = {
    PENDING: '待繳',
    PAID:    '已繳',
  };

  const PAY_STATUS_CSS: Record<InstallmentPaymentStatus, string> = {
    PENDING: 'inst-pay-pending',
    PAID:    'inst-pay-paid',
  };

  const INTEREST_TYPES: InterestType[] = ['FREE', 'FIXED_RATE'];
  const INTEREST_LABELS: Record<InterestType, string> = {
    FREE:       '免息',
    FIXED_RATE: '固定利率',
  };

  // ── 列表 ──
  let page         = $state(1);
  let installments = $state<Installment[]>([]);
  let meta         = $state<PaginatedMeta | null>(null);
  let isLoading    = $state(false);
  let error        = $state('');

  // ── 每期明細展開 ──
  interface PaymentState {
    items:   InstallmentPayment[];
    loading: boolean;
    error:   string;
  }

  let expandedId = $state<number | null>(null);
  let paymentMap = $state(new Map<number, PaymentState>());

  // ── 分錄展開（共用型別）──
  interface EntriesState {
    rows:    Entry[];
    loading: boolean;
    error:   string;
  }

  // 主檔分錄
  let expandedInstEntryId = $state<number | null>(null);
  let instEntriesMap      = $state(new Map<number, EntriesState>());

  // 每期分錄
  let expandedPayEntryPaymentId = $state<number | null>(null);
  let payEntriesMap             = $state(new Map<number, EntriesState>());

  // ── 共用資料 ──
  let accounts = $state<Account[]>([]);
  let ledgers  = $state<LedgerAccount[]>([]);

  const activeLedgers = $derived(ledgers.filter(l => l.is_active));

  // ── 新增分期 Modal ──
  interface CreateForm {
    memo:              string;
    note:              string;
    account_id:        string;
    ledger_id:         string;
    amount:            string;
    installment_count: string;
    start_date:        string;
    interest_type:     InterestType;
    annual_rate:       string;
  }

  let showCreateModal      = $state(false);
  let isCreating           = $state(false);
  let createError          = $state('');
  let createForm           = $state<CreateForm>(emptyCreateForm());
  let createNewLedger      = $state(false);
  let newLedgerForm        = $state<NewLedgerForm>(emptyNewLedgerForm());
  let newLedgerAccountId   = $state('');
  let createNewAccount     = $state(false);
  let newAccountForm       = $state<NewAccountForm>(emptyNewAccountForm());

  const isCreateValid = $derived(
    createForm.account_id !== '' &&
    parseFloat(createForm.amount)           > 0 &&
    parseInt(createForm.installment_count)  > 0 &&
    createForm.start_date !== '' &&
    (createForm.interest_type !== 'FIXED_RATE' || parseFloat(createForm.annual_rate) > 0) &&
    (createNewLedger
      ? newLedgerForm.institution.trim() !== '' &&
        newLedgerForm.name.trim()        !== '' &&
        (createNewAccount
          ? newAccountForm.account_id.trim() !== '' && newAccountForm.name.trim() !== ''
          : newLedgerAccountId !== '')
      : createForm.ledger_id !== '')
  );

  // ── 標記已繳 Modal ──
  interface PayForm {
    paid_date:      string;
    paid_ledger_id: string;
  }

  let showPayModal = $state(false);
  let isPaying     = $state(false);
  let payError     = $state('');
  let payInstId    = $state(0);
  let payInstUuid  = $state('');
  let payPeriod    = $state(0);
  let payForm      = $state<PayForm>({ paid_date: '', paid_ledger_id: '' });

  const isPayValid = $derived(
    payForm.paid_date      !== '' &&
    payForm.paid_ledger_id !== ''
  );

  $effect(() => {
    void load(page);
  });

  async function load(p: number): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      const result = await getInstallmentPaged({ page: p, pageSize: PAGE_SIZE });
      installments = result.data;
      meta         = result.meta;
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  async function togglePayment(inst: Installment): Promise<void> {
    if (expandedId === inst.installment_id) {
      expandedId = null;
      return;
    }
    expandedId = inst.installment_id;
    if (paymentMap.has(inst.installment_id)) return;

    paymentMap = new Map(paymentMap).set(inst.installment_id, {
      items: [], loading: true, error: '',
    });

    try {
      const res = await getInstallmentPaymentPaged(inst.installment_id, { page: 1, pageSize: 100 });
      paymentMap = new Map(paymentMap).set(inst.installment_id, {
        items: res.data, loading: false, error: '',
      });
    } catch (err) {
      paymentMap = new Map(paymentMap).set(inst.installment_id, {
        items: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
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

  async function toggleInstEntries(inst: Installment, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (!inst.transaction_id) return;
    const txnId = inst.transaction_id;

    if (expandedInstEntryId === inst.installment_id) {
      expandedInstEntryId = null;
      return;
    }
    expandedInstEntryId = inst.installment_id;
    if (instEntriesMap.has(inst.installment_id)) return;

    instEntriesMap = new Map(instEntriesMap).set(inst.installment_id, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      instEntriesMap = new Map(instEntriesMap).set(inst.installment_id, { rows, loading: false, error: '' });
    } catch (err) {
      instEntriesMap = new Map(instEntriesMap).set(inst.installment_id, {
        rows: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function togglePayEntries(paymentId: number, txnId: number, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (expandedPayEntryPaymentId === paymentId) {
      expandedPayEntryPaymentId = null;
      return;
    }
    expandedPayEntryPaymentId = paymentId;
    if (payEntriesMap.has(paymentId)) return;

    payEntriesMap = new Map(payEntriesMap).set(paymentId, { rows: [], loading: true, error: '' });
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    try {
      const rows = await getEntries(txnId);
      payEntriesMap = new Map(payEntriesMap).set(paymentId, { rows, loading: false, error: '' });
    } catch (err) {
      payEntriesMap = new Map(payEntriesMap).set(paymentId, {
        rows: [], loading: false,
        error: err instanceof Error ? err.message : '查詢失敗',
      });
    }
  }

  async function openCreateModal(): Promise<void> {
    createForm         = emptyCreateForm();
    createNewLedger    = false;
    newLedgerForm      = emptyNewLedgerForm();
    newLedgerAccountId = '';
    createNewAccount   = false;
    newAccountForm     = emptyNewAccountForm();
    createError        = '';
    await Promise.all([ensureAccounts(), ensureLedgers()]);
    showCreateModal = true;
  }

  async function handleCreate(e: Event): Promise<void> {
    e.preventDefault();
    if (!isCreateValid) return;
    isCreating  = true;
    createError = '';
    try {
      let finalLedgerId: number;

      if (createNewLedger) {
        // 1. 建立科目（如需要）
        let ledgerAccountId: string;
        if (createNewAccount) {
          if (!newAccountForm.account_id.trim() || !newAccountForm.name.trim()) {
            createError = '請填寫完整的會計科目資料';
            return;
          }
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
          ledgerAccountId = newAccountForm.account_id.trim();
        } else {
          ledgerAccountId = newLedgerAccountId;
        }

        // 2. 建立帳戶
        const creditLimit = newLedgerForm.type === 'CREDIT_CARD' && newLedgerForm.creditLimitInput
          ? newLedgerForm.creditLimitInput : null;
        await createLedgerAccount({
          account_id:   ledgerAccountId,
          institution:  newLedgerForm.institution.trim(),
          name:         newLedgerForm.name.trim(),
          type:         newLedgerForm.type,
          account_no:   newLedgerForm.account_no?.trim() || null,
          currency:     newLedgerForm.currency,
          credit_limit: creditLimit,
          billing_day:  newLedgerForm.type === 'CREDIT_CARD' ? newLedgerForm.billing_day || null : null,
          due_day:      newLedgerForm.type === 'CREDIT_CARD' ? newLedgerForm.due_day || null : null,
          is_active:    newLedgerForm.is_active,
          note:         null,
        });

        // 3. 取得新帳戶的 ledger_id
        invalidateLedgerCache();
        const freshLedgers = await getLedgerAccountAll(true);
        ledgers = freshLedgers;
        const created = freshLedgers.find(l => l.account_id === ledgerAccountId);
        if (!created) throw new Error('無法取得新建帳戶，請重試');
        finalLedgerId = created.ledger_id;

      } else {
        finalLedgerId = parseInt(createForm.ledger_id, 10);
      }

      // 4. 建立分期
      const payload: InstallmentCreatedPayload = {
        amount:            createForm.amount,
        installment_count: parseInt(createForm.installment_count, 10),
        start_date:        createForm.start_date,
        interest_type:     createForm.interest_type,
        annual_rate:       createForm.interest_type === 'FIXED_RATE' ? createForm.annual_rate : '0',
        account_id:        createForm.account_id,
        ledger_id:         finalLedgerId,
        memo:              createForm.memo.trim(),
        note:              createForm.note.trim(),
      };
      await createInstallment(payload);
      showCreateModal = false;
      if (page === 1) { await load(1); } else { page = 1; }
    } catch (err) {
      createError = err instanceof Error ? err.message : '新增失敗，請稍後再試。';
    } finally {
      isCreating = false;
    }
  }

  async function openPayModal(payment: InstallmentPayment, e: MouseEvent): Promise<void> {
    e.stopPropagation();
    await ensureLedgers();
    payInstId = payment.installment_id;
    payInstUuid = payment.installment_uuid
    payPeriod = Number(payment.period);
    payForm   = { paid_date: payment.due_date, paid_ledger_id: '' };
    payError  = '';
    showPayModal = true;
  }

  async function handlePay(e: Event): Promise<void> {
    e.preventDefault();
    if (!isPayValid) return;
    isPaying = true;
    payError = '';
    const ledger = activeLedgers.find((item) => item.ledger_id === parseInt(payForm.paid_ledger_id, 10) )
    if (ledger == undefined || ledger == null) {
      return;
    }
    try {
      await payInstallmentPeriod({
        installment_uuid: payInstUuid,
        period:         payPeriod,
        paid_date:      payForm.paid_date,
        paid_ledger_uuid: ledger.ledger_uuid,
      });
      showPayModal = false;
      const next = new Map(paymentMap);
      next.delete(payInstId);
      paymentMap = next;
      if (expandedId === payInstId) expandedId = null;
      if (page === 1) { await load(1); } else { page = 1; }
    } catch (err) {
      payError = err instanceof Error ? err.message : '儲存失敗，請稍後再試。';
    } finally {
      isPaying = false;
    }
  }

  function emptyCreateForm(): CreateForm {
    return {
      memo:              '',
      note:              '',
      account_id:        '',
      ledger_id:         '',
      amount:            '',
      installment_count: '',
      start_date:        '',
      interest_type:     'FREE',
      annual_rate:       '',
    };
  }

  function fmtAmt(val: string): string {
    const n = parseFloat(val);
    if (isNaN(n)) return val;
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function fmtTotal(amount: string, interest: string): string {
    const a = parseFloat(amount)   || 0;
    const i = parseFloat(interest) || 0;
    return (a + i).toLocaleString('zh-TW', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  }

  function progressPct(inst: Installment): number {
    if (inst.total_periods === 0) return 0;
    return Math.round((inst.paid_periods / inst.total_periods) * 100);
  }

  function interestLabel(inst: Installment): string {
    if (inst.interest_type === 'FIXED_RATE') {
      const r = parseFloat(inst.interest_rate);
      return isNaN(r) ? INTEREST_LABELS.FIXED_RATE : `${r}%/年`;
    }
    return INTEREST_LABELS.FREE;
  }
</script>

<div class="content-header">
  <h1 class="content-title">分期管理</h1>
  {#if meta}
    <span class="content-date">共 {meta.total_count} 筆</span>
  {/if}
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

{#snippet instPaymentContent(inst: Installment)}
  {@const payment = paymentMap.get(inst.installment_id)}
  {#if payment?.loading}
    <span class="inst-payment-msg">查詢中…</span>
  {:else if payment?.error}
    <span class="inst-payment-msg inst-payment-error">{payment.error}</span>
  {:else if payment && payment.items.length > 0}
    <div class="inst-payment-table">
      <div class="inst-payment-header">
        <span>期</span>
        <span>到期日</span>
        <span class="num">本金</span>
        <span class="num">利息</span>
        <span class="num">合計</span>
        <span>繳費日</span>
        <span>狀態</span>
        <span>更新者</span>
        <span>更新時間</span>
        <span></span>
      </div>
      {#each payment.items as p (p.payment_id)}
        {@const isEntryExpanded = expandedPayEntryPaymentId === p.payment_id}
        {@const entryState = payEntriesMap.get(p.payment_id)}
        <div class="inst-payment-item">
          <span class="mono">{p.period}</span>
          <span class="mono">{p.due_date}</span>
          <span class="num mono">{fmtAmt(p.amount)}</span>
          <span class="num mono">{fmtAmt(p.interest)}</span>
          <span class="num mono">{fmtTotal(p.amount, p.interest)}</span>
          <span class="mono">{p.paid_date ?? '—'}</span>
          <span>
            <span class="badge {PAY_STATUS_CSS[p.status]}">{PAY_STATUS_LABELS[p.status]}</span>
          </span>
          <span>{p.updated_by ?? '—'}</span>
          <span>{p.updated_at ?? '—'}</span>
          <span style="display:flex;align-items:center;gap:4px;flex-wrap:wrap;">
            {#if p.status === 'PENDING'}
              <button
                class="btn-ghost"
                style="padding:2px 8px;font-size:10px;"
                onclick={(e) => openPayModal(p, e)}
              >標記已繳</button>
            {/if}
            {#if p.transaction_id !== null}
              <button
                class="btn-ghost"
                style="padding:2px 8px;font-size:10px;"
                class:acct-mode-btn--active={isEntryExpanded}
                onclick={(e) => togglePayEntries(p.payment_id, p.transaction_id!, e)}
              >分錄</button>
            {/if}
          </span>
        </div>
        {#if isEntryExpanded}
          <div class="txn-entries">
            <TxnEntryPanel
              rows={entryState?.rows ?? []}
              loading={entryState?.loading ?? false}
              error={entryState?.error ?? ''}
              accounts={accounts}
              ledgers={ledgers}
            />
          </div>
        {/if}
      {/each}
    </div>
  {:else}
    <span class="inst-payment-msg">尚無繳款紀錄</span>
  {/if}
{/snippet}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">分期付款主檔</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading">
          <span class="spinner" aria-hidden="true"></span>
          載入中
        </span>
      {/if}
      <button class="section-action" onclick={openCreateModal}>＋ 新增分期</button>
    </div>
  </header>

  <div class="table-wrap inst-table-wrap">
    <table class="data-table" aria-label="分期付款主檔">
      <thead>
        <tr>
          <th>說明</th>
          <th>帳戶</th>
          <th class="inst-amount">總金額</th>
          <th>繳款進度</th>
          <th class="inst-amount">每期金額</th>
          <th>開始日</th>
          <th>利息</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && installments.length === 0}
          <tr><td colspan="10" class="table-empty">載入中...</td></tr>
        {:else if installments.length === 0}
          <tr><td colspan="10" class="table-empty">無資料</td></tr>
        {:else}
          {#each installments as inst (inst.installment_id)}
            {@const isExpanded = expandedId === inst.installment_id}
            {@const isInstEntryExpanded = expandedInstEntryId === inst.installment_id}
            {@const instEntryState = instEntriesMap.get(inst.installment_id)}
            <tr
              class:inst-row-expanded={isExpanded}
              style="cursor:pointer;"
              onclick={() => togglePayment(inst)}
            >
              <td>
                <span class="inst-expand-icon">{isExpanded ? '▼' : '▶'}</span>
                {inst.description || '—'}
              </td>
              <td class="mono">{inst.ledger_id}</td>
              <td class="inst-amount mono">{fmtAmt(inst.total_amount)}</td>
              <td>
                <div class="inst-progress">
                  <div class="inst-progress-bar">
                    <div class="inst-progress-fill" style="width:{progressPct(inst)}%"></div>
                  </div>
                  <span class="inst-progress-text">{inst.paid_periods}/{inst.total_periods}</span>
                </div>
              </td>
              <td class="inst-amount mono">{fmtAmt(inst.amount_per_period)}</td>
              <td class="mono">{inst.start_date}</td>
              <td>{interestLabel(inst)}</td>
              <td class="hidden md:table-cell">{inst.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{inst.updated_at ?? '—'}</td>
              <td onclick={(e) => e.stopPropagation()}>
                <span class="badge {STATUS_CSS[inst.status]}">{STATUS_LABELS[inst.status]}</span>
                {#if inst.transaction_id !== null}
                  <button
                    type="button"
                    class="btn-ghost"
                    style="padding:1px 6px;font-size:10px;margin-left:6px;"
                    class:acct-mode-btn--active={isInstEntryExpanded}
                    onclick={(e) => toggleInstEntries(inst, e)}
                  >分錄</button>
                {/if}
              </td>
            </tr>

            {#if isInstEntryExpanded}
              <tr>
                <td colspan="10" class="inst-payment-cell">
                  <div class="txn-entries">
                    <TxnEntryPanel
                      rows={instEntryState?.rows ?? []}
                      loading={instEntryState?.loading ?? false}
                      error={instEntryState?.error ?? ''}
                      accounts={accounts}
                      ledgers={ledgers}
                    />
                  </div>
                </td>
              </tr>
            {/if}

            {#if isExpanded}
              <tr class="inst-payment-row">
                <td colspan="10" class="inst-payment-cell">
                  {@render instPaymentContent(inst)}
                </td>
              </tr>
            {/if}
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="inst-card-list">
    {#if isLoading && installments.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if installments.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each installments as inst (inst.installment_id)}
        {@const isExpanded = expandedId === inst.installment_id}
        {@const isInstEntryExpanded = expandedInstEntryId === inst.installment_id}
        {@const instEntryState = instEntriesMap.get(inst.installment_id)}
        <div class="inst-card {isExpanded ? 'inst-card-expanded' : ''}">
          <div class="inst-card-main" role="button" tabindex="0" onclick={() => togglePayment(inst)} onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && togglePayment(inst)} style="cursor:pointer;">
            <div class="inst-card-head">
              <span class="inst-card-title">
                <span class="inst-expand-icon">{isExpanded ? '▼' : '▶'}</span>
                {inst.description || '—'}
              </span>
              <span class="badge {STATUS_CSS[inst.status]}">{STATUS_LABELS[inst.status]}</span>
            </div>
            <div class="inst-card-amounts">
              <span class="inst-card-label">總額</span>
              <span class="mono">{fmtAmt(inst.total_amount)}</span>
              <span class="inst-card-sep">·</span>
              <span class="inst-card-label">每期</span>
              <span class="mono">{fmtAmt(inst.amount_per_period)}</span>
            </div>
            <div class="inst-progress">
              <div class="inst-progress-bar">
                <div class="inst-progress-fill" style="width:{progressPct(inst)}%"></div>
              </div>
              <span class="inst-progress-text">{inst.paid_periods}/{inst.total_periods}</span>
            </div>
            <div class="inst-card-footer">
              <span class="inst-card-meta">
                <span class="inst-card-label">開始</span>
                <span class="mono">{inst.start_date}</span>
                <span class="inst-card-sep">·</span>
                {interestLabel(inst)}
              </span>
              {#if inst.transaction_id !== null}
                <button
                  type="button"
                  class="btn-ghost"
                  style="padding:1px 6px;font-size:10px;"
                  class:acct-mode-btn--active={isInstEntryExpanded}
                  onclick={(e) => toggleInstEntries(inst, e)}
                >分錄</button>
              {/if}
            </div>
          </div>
          {#if isInstEntryExpanded}
            <div style="padding:8px 16px 12px; border-top:1px solid #1a1c28;">
              <TxnEntryPanel
                rows={instEntryState?.rows ?? []}
                loading={instEntryState?.loading ?? false}
                error={instEntryState?.error ?? ''}
                accounts={accounts}
                ledgers={ledgers}
              />
            </div>
          {/if}
          {#if isExpanded}
            <div class="inst-payment-cell" style="padding:0;">
              {@render instPaymentContent(inst)}
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

<!-- ── 新增分期 Modal ─────────────────────── -->
{#if showCreateModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showCreateModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="create-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="create-modal-title">新增分期付款</h2>
        <button class="modal-close" onclick={() => { showCreateModal = false; }} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleCreate}>
        {#if createError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{createError}</p>
        {/if}

        <!-- 基本資訊 -->
        <div class="form-group">
          <label class="form-label" for="c-memo">說明</label>
          <input id="c-memo" class="form-input" type="text" bind:value={createForm.memo} placeholder="例：MacBook Pro 24 期分期" />
        </div>

        <div class="form-group">
          <label class="form-label" for="c-account">歸屬科目 *</label>
          <AccountSelect
            accounts={accounts}
            value={createForm.account_id}
            placeholder="選擇購買品項歸屬科目…"
            onselect={(id) => { createForm.account_id = id; }}
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="c-amount">總金額 *</label>
            <input
              id="c-amount"
              class="form-input"
              type="number"
              min="0.01"
              step="any"
              bind:value={createForm.amount}
              placeholder="0.00"
              required
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="c-count">期數 *</label>
            <input
              id="c-count"
              class="form-input"
              type="number"
              min="1"
              step="1"
              bind:value={createForm.installment_count}
              placeholder="12"
              required
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="c-start">第一期到期日 *</label>
            <input id="c-start" class="form-input" type="date" bind:value={createForm.start_date} required />
          </div>
          <div class="form-group">
            <label class="form-label" for="c-interest-type">利息類型 *</label>
            <select id="c-interest-type" class="form-select" bind:value={createForm.interest_type}>
              {#each INTEREST_TYPES as t}
                <option value={t}>{INTEREST_LABELS[t]}</option>
              {/each}
            </select>
          </div>
        </div>

        {#if createForm.interest_type === 'FIXED_RATE'}
          <div class="form-group">
            <label class="form-label" for="c-rate">年利率（%）*</label>
            <input
              id="c-rate"
              class="form-input"
              type="number"
              min="0.01"
              step="any"
              bind:value={createForm.annual_rate}
              placeholder="例：12.5"
            />
          </div>
        {/if}

        <LedgerSelectSection
          ledgers={activeLedgers}
          accounts={accounts}
          bind:ledgerId={createForm.ledger_id}
          bind:createNew={createNewLedger}
          bind:newLedgerForm={newLedgerForm}
          bind:createNewAccount={createNewAccount}
          bind:newAccountForm={newAccountForm}
          bind:accountId={newLedgerAccountId}
        />

        <div class="form-group">
          <label class="form-label" for="c-note">備註</label>
          <input id="c-note" class="form-input" type="text" bind:value={createForm.note} placeholder="選填" />
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showCreateModal = false; }} disabled={isCreating}>取消</button>
          <button type="submit" class="btn-primary" disabled={isCreating || !isCreateValid}>
            {isCreating ? '建立中…' : '建立分期'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<!-- ── 標記已繳 Modal ──────────────────────── -->
{#if showPayModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showPayModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="pay-modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="pay-modal-title">標記第 {payPeriod} 期已繳</h2>
        <button class="modal-close" onclick={() => { showPayModal = false; }} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handlePay}>
        {#if payError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{payError}</p>
        {/if}

        <div class="form-group">
          <label class="form-label" for="p-date">繳費日 *</label>
          <input id="p-date" class="form-input" type="date" bind:value={payForm.paid_date} required />
        </div>

        <div class="form-group">
          <span class="form-label">扣款帳戶 *</span>
          <LedgerSelect
            ledgers={activeLedgers}
            value={payForm.paid_ledger_id}
            onselect={(id) => { payForm.paid_ledger_id = id; }}
          />
        </div>

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showPayModal = false; }} disabled={isPaying}>取消</button>
          <button type="submit" class="btn-primary" disabled={isPaying || !isPayValid}>
            {isPaying ? '儲存中…' : '確認已繳'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}