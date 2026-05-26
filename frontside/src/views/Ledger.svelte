<script lang="ts">
  import { getLedgerAccountAll, getLedgerBalances, invalidateLedgerCache, createLedgerAccount, updateLedgerAccount } from '../api/ledger';
  import { getAccountAll, createAccount } from '../api/account';
  import { getLedgerAccountTypeConfigs } from '../api/setting';
  import AccountSelect from '../components/AccountSelect.svelte';
  import { type NewAccountForm, emptyNewAccountForm } from '../components/NewAccountFormSection.svelte';
  import NewLedgerFormSection, { type NewLedgerForm, emptyNewLedgerForm } from '../components/NewLedgerFormSection.svelte';
  import type { Account, CreateAccountRequest } from '../types/account';
  import type { LedgerAccount, LedgerBalance, LedgerAccountCreatePayload, LedgerAccountUpdatePayload, LedgerAccountType } from '../types/ledger';
  import type { LedgerAccountTypeConfigResult } from '../types/setting';

  const TYPE_LABELS: Record<LedgerAccountType, string> = {
    BANK_ACCOUNT: '銀行帳戶',
    CREDIT_CARD:  '信用卡',
    LOAN:         '貸款',
  };

  const TYPE_CSS: Record<LedgerAccountType, string> = {
    BANK_ACCOUNT: 'ledger-type-bank',
    CREDIT_CARD:  'ledger-type-credit',
    LOAN:         'ledger-type-installment',
  };

  let ledgers            = $state<LedgerAccount[]>([]);
  let allAccounts        = $state<Account[]>([]);
  let balanceMap         = $state(new Map<number, string>());
  let isLoading          = $state(false);
  let error              = $state('');
  let ledgerTypeConfigMap = $state(new Map<string, LedgerAccountTypeConfigResult>());

  type LedgerForm = LedgerAccountCreatePayload & {
    creditLimitInput: string;
    ledger_id:        number;
    version:          number;
  };

  let showModal          = $state(false);
  let mode               = $state<'create' | 'edit'>('create');
  let isSaving           = $state(false);
  let saveError          = $state('');
  let form               = $state<LedgerForm>(emptyForm());
  let newLedgerForm      = $state<NewLedgerForm>(emptyNewLedgerForm());
  let newLedgerAccountId = $state('');
  let createNewAccount   = $state(false);
  let newAccountForm     = $state<NewAccountForm>(emptyNewAccountForm());

  const activeLedgerTypeConfig = $derived(ledgerTypeConfigMap.get(newLedgerForm.type) ?? null);
  // Build ancestor-inclusive list so AccountSelect tree navigation works:
  // without ancestors, AccountSelect starts at root (parent_id=null) and finds nothing.
  const filteredAccounts = $derived((() => {
    if (!activeLedgerTypeConfig?.account_id || activeLedgerTypeConfig.descendants.length === 0) {
      return allAccounts;
    }
    const path: Account[] = [];
    let cur: string | null = activeLedgerTypeConfig.account_id;
    while (cur) {
      const a = allAccounts.find(x => x.account_id === cur);
      if (!a) break;
      path.push(a);
      cur = a.parent_id ?? null;
    }
    return [...path.reverse(), ...activeLedgerTypeConfig.descendants];
  })());
  const filteredParentAccounts = $derived(filteredAccounts);

  let _prevLedgerType = $state<LedgerAccountType | ''>('');
  $effect(() => {
    const t = newLedgerForm.type;
    if (_prevLedgerType !== '' && _prevLedgerType !== t) {
      newLedgerAccountId       = '';
      newAccountForm.parent_id = null;
    }
    _prevLedgerType = t;
  });

  $effect(() => {
    void load();
  });

  async function load(force = false): Promise<void> {
    isLoading = true;
    error     = '';
    try {
      const [data, balances] = await Promise.all([
        getLedgerAccountAll(force),
        getLedgerBalances().catch(() => [] as LedgerBalance[]),
      ]);
      ledgers    = data;
      balanceMap = new Map(balances.map((b: LedgerBalance) => [b.ledger_id, b.balance]));
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗，請稍後再試。';
    } finally {
      isLoading = false;
    }
  }

  function fmtBalance(val: string | undefined): string {
    if (val === undefined) return '—';
    const n = parseFloat(val);
    if (isNaN(n)) return '—';
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  async function openModal(): Promise<void> {
    mode               = 'create';
    saveError          = '';
    form               = emptyForm();
    newLedgerForm      = emptyNewLedgerForm();
    newLedgerAccountId = '';
    createNewAccount   = false;
    newAccountForm     = emptyNewAccountForm();
    const [accounts, configs] = await Promise.all([
      allAccounts.length === 0 ? getAccountAll() : Promise.resolve(allAccounts),
      ledgerTypeConfigMap.size === 0
        ? getLedgerAccountTypeConfigs().catch(() => [] as LedgerAccountTypeConfigResult[])
        : Promise.resolve([...ledgerTypeConfigMap.values()]),
    ]);
    allAccounts         = accounts;
    ledgerTypeConfigMap = new Map(configs.map(c => [c.type, c]));
    showModal = true;
  }

  async function openEditModal(ledger: LedgerAccount): Promise<void> {
    mode      = 'edit';
    saveError = '';
    form = {
      account_id:       ledger.account_id,
      institution:      ledger.institution,
      name:             ledger.name,
      type:             ledger.type,
      account_no:       ledger.account_no,
      currency:         ledger.currency,
      credit_limit:     ledger.credit_limit,
      billing_day:      ledger.billing_day,
      due_day:          ledger.due_day,
      is_active:        ledger.is_active,
      note:             ledger.note,
      creditLimitInput: ledger.credit_limit ?? '',
      ledger_id:        ledger.ledger_id,
      version:          ledger.version,
    };
    if (allAccounts.length === 0) {
      allAccounts = await getAccountAll();
    }
    showModal = true;
  }

  async function handleSubmit(e: Event): Promise<void> {
    e.preventDefault();
    isSaving  = true;
    saveError = '';
    try {
      if (mode === 'create') {
        let ledgerAccountId: string;
        if (createNewAccount) {
          if (!newAccountForm.account_id.trim() || !newAccountForm.name.trim()) {
            saveError = '請填寫完整的會計科目資料';
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
          allAccounts = [];
          ledgerAccountId = newAccountForm.account_id.trim();
        } else {
          ledgerAccountId = newLedgerAccountId;
        }
        if (!ledgerAccountId) {
          saveError = '請選擇或新增會計科目';
          return;
        }
        const creditLimit = newLedgerForm.type === 'CREDIT_CARD' && newLedgerForm.creditLimitInput
          ? newLedgerForm.creditLimitInput : null;
        const payload: LedgerAccountCreatePayload = {
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
          note:         form.note?.trim() || null,
        };
        await createLedgerAccount(payload);
      } else {
        const creditLimit = form.type === 'CREDIT_CARD' && form.creditLimitInput ? form.creditLimitInput : null;
        const payload: LedgerAccountUpdatePayload = {
          ledger_id:    form.ledger_id,
          account_id:   form.account_id,
          institution:  form.institution.trim(),
          name:         form.name.trim(),
          account_no:   form.account_no?.trim() || null,
          currency:     form.currency,
          credit_limit: creditLimit,
          billing_day:  form.type === 'CREDIT_CARD' ? form.billing_day : null,
          due_day:      form.type === 'CREDIT_CARD' ? form.due_day : null,
          is_active:    form.is_active,
          note:         form.note?.trim() || null,
          version:      form.version,
          type:         form.type,
        };
        await updateLedgerAccount(payload);
      }
      invalidateLedgerCache();
      showModal = false;
      await load(true);
    } catch (err) {
      saveError = err instanceof Error ? err.message : mode === 'create' ? '新增失敗' : '修改失敗';
    } finally {
      isSaving = false;
    }
  }

  function emptyForm(): LedgerForm {
    return {
      account_id:       '',
      institution:      '',
      name:             '',
      type:             'BANK_ACCOUNT',
      account_no:       null,
      currency:         'TWD',
      credit_limit:     null,
      billing_day:      null,
      due_day:          null,
      is_active:        true,
      note:             null,
      creditLimitInput: '',
      ledger_id:        0,
      version:          0,
    };
  }

  const canSubmit = $derived(
    mode === 'create'
      ? newLedgerForm.name.trim()        !== '' &&
        newLedgerForm.institution.trim() !== '' &&
        (createNewAccount
          ? newAccountForm.account_id.trim() !== '' && newAccountForm.name.trim() !== ''
          : newLedgerAccountId !== '')
      : form.name.trim()        !== '' &&
        form.institution.trim() !== '' &&
        form.account_id         !== '',
  );
</script>

<div class="content-header">
  <h1 class="content-title">帳戶管理</h1>
  <span class="content-date">共 {ledgers.length} 筆</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<section class="section">
  <header class="section-header">
    <h2 class="section-title">帳戶列表</h2>
    <div style="display:flex;align-items:center;gap:16px;">
      {#if isLoading}
        <span class="query-loading">
          <span class="spinner" aria-hidden="true"></span>
          載入中
        </span>
      {/if}
      <button class="section-action" onclick={openModal}>＋ 新增帳戶</button>
    </div>
  </header>

  <div class="table-wrap ldgr-table-wrap">
    <table class="data-table" aria-label="帳戶列表">
      <thead>
        <tr>
          <th>帳戶名稱</th>
          <th>金融機構</th>
          <th>類型</th>
          <th>關聯科目</th>
          <th>帳號／卡號</th>
          <th>幣別</th>
          <th class="td-amount">餘額</th>
          <th>狀態</th>
          <th>備註</th>
          <th class="hidden md:table-cell">更新者</th>
          <th class="hidden md:table-cell">更新時間</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && ledgers.length === 0}
          <tr><td colspan="12" class="table-empty">載入中...</td></tr>
        {:else if ledgers.length === 0}
          <tr><td colspan="12" class="table-empty">無資料</td></tr>
        {:else}
          {#each ledgers as ledger (ledger.ledger_id)}
            <tr>
              <td>{ledger.name}</td>
              <td>{ledger.institution}</td>
              <td>
                <span class="badge {TYPE_CSS[ledger.type]}">
                  {TYPE_LABELS[ledger.type]}
                </span>
              </td>
              <td class="mono">{ledger.account_id}</td>
              <td class="mono">{ledger.account_no ?? '—'}</td>
              <td>{ledger.currency}</td>
              <td class="td-amount {balanceMap.has(ledger.ledger_id) ? (parseFloat(balanceMap.get(ledger.ledger_id)!) < 0 ? 'td-balance-neg' : 'td-balance-pos') : ''}">
                {fmtBalance(balanceMap.get(ledger.ledger_id))}
              </td>
              <td>
                {#if ledger.is_active}
                  <span class="badge approved">啟用</span>
                {:else}
                  <span class="badge pending">停用</span>
                {/if}
              </td>
              <td class="note-cell">{ledger.note ?? '—'}</td>
              <td class="hidden md:table-cell">{ledger.updated_by ?? '—'}</td>
              <td class="hidden md:table-cell">{ledger.updated_at ?? '—'}</td>
              <td>
                <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openEditModal(ledger)}>編輯</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>

  <div class="ldgr-card-list">
    {#if isLoading && ledgers.length === 0}
      <div class="table-empty">載入中...</div>
    {:else if ledgers.length === 0}
      <div class="table-empty">無資料</div>
    {:else}
      {#each ledgers as ledger (ledger.ledger_id)}
        {@const balAmt = balanceMap.get(ledger.ledger_id)}
        {@const balClass = balAmt !== undefined ? (parseFloat(balAmt) < 0 ? 'ldgr-card-balance-neg' : 'ldgr-card-balance-pos') : ''}
        <div class="ldgr-card">
          <div class="ldgr-card-head">
            <span class="ldgr-card-name">{ledger.name}</span>
            <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openEditModal(ledger)}>編輯</button>
          </div>
          {#if ledger.institution}
            <div class="ldgr-card-institution">{ledger.institution}</div>
          {/if}
          <div class="ldgr-card-row">
            <span class="badge {TYPE_CSS[ledger.type]}">{TYPE_LABELS[ledger.type]}</span>
            <span class="ldgr-card-meta">{ledger.currency}</span>
            {#if ledger.account_no}
              <span class="ldgr-card-meta mono">{ledger.account_no}</span>
            {/if}
          </div>
          <div class="ldgr-card-acct mono">{ledger.account_id}</div>
          <div class="ldgr-card-footer">
            <span class="ldgr-card-balance {balClass}">{fmtBalance(balAmt)}</span>
            {#if ledger.is_active}
              <span class="badge approved">啟用</span>
            {:else}
              <span class="badge pending">停用</span>
            {/if}
          </div>
          {#if ledger.note && ledger.note !== '—'}
            <div class="ldgr-card-note">{ledger.note}</div>
          {/if}
        </div>
      {/each}
    {/if}
  </div>
</section>

<!-- ── 新增帳戶 Modal ─────────────────────── -->
{#if showModal}
  <div class="modal-overlay" role="presentation" onclick={(e) => { if (e.target === e.currentTarget) showModal = false; }}>
    <div class="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div class="modal-header">
        <h2 class="modal-title" id="modal-title">{mode === 'create' ? '新增帳戶' : '修改帳戶'}</h2>
        <button class="modal-close" onclick={() => { showModal = false; }} aria-label="關閉">×</button>
      </div>

      <form class="modal-body" onsubmit={handleSubmit}>
        {#if saveError}
          <p class="query-error" role="alert" style="margin-bottom:16px;">{saveError}</p>
        {/if}

        {#if mode === 'create'}
          <NewLedgerFormSection
            accounts={filteredAccounts}
            accountsForParent={filteredParentAccounts}
            bind:form={newLedgerForm}
            required={true}
            bind:createNewAccount={createNewAccount}
            bind:newAccountForm={newAccountForm}
            bind:accountId={newLedgerAccountId}
          />
        {:else}
          <div class="form-row">
            <div class="form-group">
              <label class="form-label" for="f-institution">金融機構 *</label>
              <input
                id="f-institution"
                class="form-input"
                type="text"
                bind:value={form.institution}
                placeholder="例：玉山銀行"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="f-name">帳戶名稱 *</label>
              <input
                id="f-name"
                class="form-input"
                type="text"
                bind:value={form.name}
                placeholder="例：玉山數位帳戶"
                required
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label" for="f-type-display">帳戶類型</label>
              <input id="f-type-display" class="form-input" type="text" value={TYPE_LABELS[form.type]} readonly style="background:#f5f0e8;cursor:default;" />
            </div>
            <div class="form-group">
              <label class="form-label" for="f-account-no">帳號 ／ 卡號後四碼</label>
              <input
                id="f-account-no"
                class="form-input"
                type="text"
                value={form.account_no ?? ''}
                oninput={(e) => { form.account_no = (e.target as HTMLInputElement).value || null; }}
                placeholder="選填"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label" for="account_id">關聯會計科目 *</label>
              <AccountSelect
                accounts={allAccounts}
                value={form.account_id}
                placeholder="搜尋並選擇科目…"
                onselect={(id) => { form.account_id = id; }}
              />
            </div>
            <div class="form-group">
              <label class="form-label" for="f-currency">幣別</label>
              <input
                id="f-currency"
                class="form-input"
                type="text"
                bind:value={form.currency}
                placeholder="TWD"
              />
            </div>
          </div>

          {#if form.type === 'CREDIT_CARD'}
            <div class="form-row">
              <div class="form-group">
                <label class="form-label" for="f-credit-limit">信用額度</label>
                <input
                  id="f-credit-limit"
                  class="form-input"
                  type="number"
                  min="0"
                  step="1"
                  bind:value={form.creditLimitInput}
                  placeholder="例：100000"
                />
              </div>
              <div class="form-group"></div>
            </div>
            <div class="form-row">
              <div class="form-group">
                <label class="form-label" for="f-billing-day">帳單截止日（日）</label>
                <input
                  id="f-billing-day"
                  class="form-input"
                  type="number"
                  min="1"
                  max="31"
                  value={form.billing_day ?? ''}
                  oninput={(e) => { form.billing_day = (e.target as HTMLInputElement).value || null; }}
                  placeholder="例：25"
                />
              </div>
              <div class="form-group">
                <label class="form-label" for="f-due-day">繳費截止日（日）</label>
                <input
                  id="f-due-day"
                  class="form-input"
                  type="number"
                  min="1"
                  max="31"
                  value={form.due_day ?? ''}
                  oninput={(e) => { form.due_day = (e.target as HTMLInputElement).value || null; }}
                  placeholder="例：15"
                />
              </div>
            </div>
          {/if}

          <div class="form-group" style="display:flex;align-items:center;gap:8px;padding-top:4px;">
            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
              <input type="checkbox" bind:checked={form.is_active} />
              啟用此帳戶
            </label>
          </div>
        {/if}

        <div class="form-group" style="margin-top:16px;">
          <label class="form-label" for="f-note">備註</label>
          <input
            id="f-note"
            class="form-input"
            type="text"
            value={form.note ?? ''}
            oninput={(e) => { form.note = (e.target as HTMLInputElement).value || null; }}
            placeholder="選填"
          />
        </div>

        {#if mode === 'edit'}
          {@const ldgr = ledgers.find(l => l.ledger_id === form.ledger_id)}
          {#if ldgr}
            <div class="form-row" style="margin-top:8px;">
              <div class="form-group">
                <label class="form-label" for="f-ldgr-updated-by">更新者</label>
                <input id="f-ldgr-updated-by" class="form-input" type="text" value={ldgr.updated_by ?? '—'} readonly style="background:#f5f0e8;cursor:default;" />
              </div>
              <div class="form-group">
                <label class="form-label" for="f-ldgr-updated-at">更新時間</label>
                <input id="f-ldgr-updated-at" class="form-input" type="text" value={ldgr.updated_at ?? '—'} readonly style="background:#f5f0e8;cursor:default;" />
              </div>
            </div>
          {/if}
        {/if}

        <div class="modal-footer" style="padding:0;margin-top:8px;">
          <button type="button" class="btn-ghost" onclick={() => { showModal = false; }} disabled={isSaving}>取消</button>
          <button
            type="submit"
            class="btn-primary"
            disabled={isSaving || !canSubmit}
          >
            {isSaving ? '儲存中…' : mode === 'create' ? '新增帳戶' : '儲存修改'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}