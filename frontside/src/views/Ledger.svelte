<script lang="ts">
  import { getLedgerAccountAll, getLedgerBalances, invalidateLedgerCache, createLedgerAccount, updateLedgerAccount } from '../api/ledger';
  import { getAccountAll, createAccount } from '../api/account';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { Account, CreateAccountRequest } from '../types/account';
  import type { LedgerAccount, LedgerBalance, LedgerAccountCreatePayload, LedgerAccountUpdatePayload, LedgerAccountType } from '../types/ledger';

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

  const ACCOUNT_TYPES = ['ASSET', 'LIABILITY', 'EQUITY', 'INCOME', 'EXPENSE'] as const;
  const ACCOUNT_TYPE_LABELS: Record<string, string> = {
    ASSET:     '資產',
    LIABILITY: '負債',
    EQUITY:    '權益',
    INCOME:    '收入',
    EXPENSE:   '費用',
  };

  let ledgers     = $state<LedgerAccount[]>([]);
  let allAccounts = $state<Account[]>([]);
  let balanceMap  = $state(new Map<number, string>());
  let isLoading   = $state(false);
  let error       = $state('');

  type LedgerForm = LedgerAccountCreatePayload & {
    creditLimitInput: string;
    ledger_id:        number;
    version:          number;
  };

  interface NewAccountForm {
    parent_id:      string | null;
    accountSuffix:  string;
    account_id:     string;
    name:           string;
    type:           string;
    normal_balance: string;
    currency:       string;
    is_summary:     boolean;
    is_active:      boolean;
  }

  let showModal        = $state(false);
  let mode             = $state<'create' | 'edit'>('create');
  let isSaving         = $state(false);
  let saveError        = $state('');
  let form             = $state<LedgerForm>(emptyForm());
  let createNewAccount = $state(false);
  let newAccountForm   = $state<NewAccountForm>(emptyAccountForm());

  const accountParent = $derived(
    newAccountForm.parent_id
      ? (allAccounts.find(a => a.account_id === newAccountForm.parent_id) ?? null)
      : null,
  );
  const accountPrefix = $derived(accountParent?.account_id ?? '');

  $effect(() => {
    newAccountForm.account_id = accountPrefix + newAccountForm.accountSuffix;
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
    mode             = 'create';
    saveError        = '';
    createNewAccount = false;
    form             = emptyForm();
    newAccountForm   = emptyAccountForm();
    if (allAccounts.length === 0) {
      allAccounts = await getAccountAll();
    }
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
    if (!form.name.trim() || !form.institution.trim()) return;

    const ledgerAccountId = createNewAccount ? newAccountForm.account_id : form.account_id;
    if (!ledgerAccountId) return;

    isSaving  = true;
    saveError = '';
    try {
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
      }

      const creditLimit = form.type === 'CREDIT_CARD' && form.creditLimitInput ? form.creditLimitInput : null;
      if (mode === 'create') {
        const payload: LedgerAccountCreatePayload = {
          account_id:   ledgerAccountId,
          institution:  form.institution.trim(),
          name:         form.name.trim(),
          type:         form.type,
          account_no:   form.account_no?.trim() || null,
          currency:     form.currency,
          credit_limit: creditLimit,
          billing_day:  form.type === 'CREDIT_CARD' ? form.billing_day : null,
          due_day:      form.type === 'CREDIT_CARD' ? form.due_day : null,
          is_active:    form.is_active,
          note:         form.note?.trim() || null,
        };
        await createLedgerAccount(payload);
      } else {
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

  function emptyAccountForm(): NewAccountForm {
    return {
      parent_id:      null,
      accountSuffix:  '',
      account_id:     '',
      name:           '',
      type:           'ASSET',
      normal_balance: 'DEBIT',
      currency:       'TWD',
      is_summary:     false,
      is_active:      true,
    };
  }

  const canSubmit = $derived(
    form.name.trim() !== '' &&
    form.institution.trim() !== '' &&
    (createNewAccount
      ? newAccountForm.account_id.trim() !== '' && newAccountForm.name.trim() !== ''
      : form.account_id !== ''),
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

  <div class="table-wrap">
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
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading && ledgers.length === 0}
          <tr><td colspan="10" class="table-empty">載入中...</td></tr>
        {:else if ledgers.length === 0}
          <tr><td colspan="10" class="table-empty">無資料</td></tr>
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
              <td>
                <button class="btn-ghost" style="padding:2px 10px;font-size:11px;" onclick={() => openEditModal(ledger)}>編輯</button>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
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
          {#if mode === 'create'}
            <div class="form-group">
              <label class="form-label" for="f-type">帳戶類型 *</label>
              <select id="f-type" class="form-select" bind:value={form.type}>
                <option value="BANK_ACCOUNT">銀行帳戶</option>
                <option value="CREDIT_CARD">信用卡</option>
                <option value="LOAN">貸款</option>
              </select>
            </div>
          {:else}
            <div class="form-group">
              <label class="form-label" for="f-type-display">帳戶類型</label>
              <input id="f-type-display" class="form-input" type="text" value={TYPE_LABELS[form.type]} readonly style="background:#f5f0e8;cursor:default;" />
            </div>
          {/if}
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

        {#if mode === 'create'}
          <div class="form-group" style="margin-bottom:4px;">
            <span class="form-label">關聯會計科目 *</span>
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
                onclick={() => { createNewAccount = true; newAccountForm = emptyAccountForm(); }}
              >＋ 新增科目</button>
            </div>
          </div>

          {#if !createNewAccount}
            <div class="form-row">
              <div class="form-group">
                <!-- svelte-ignore a11y_label_has_associated_control -->
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
          {:else}
            <div class="new-account-section">
              <div class="form-group">
                <!-- svelte-ignore a11y_label_has_associated_control -->
                <label class="form-label">父類別科目</label>
                <AccountSelect
                  accounts={allAccounts}
                  value={newAccountForm.parent_id ?? ''}
                  placeholder="無（頂層科目）"
                  onselect={(id) => {
                    newAccountForm.parent_id   = id || null;
                    newAccountForm.accountSuffix = '';
                    const parent = allAccounts.find(a => a.account_id === id) ?? null;
                    if (parent) {
                      newAccountForm.type           = parent.type;
                      newAccountForm.normal_balance = parent.normal_balance;
                      newAccountForm.currency       = parent.currency;
                    }
                  }}
                />
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label class="form-label" for="f-new-acct-id">科目編號 *</label>
                  <div class="acct-id-field">
                    {#if accountPrefix}
                      <span class="acct-id-prefix">{accountPrefix}</span>
                    {/if}
                    <input
                      id="f-new-acct-id"
                      class="form-input acct-id-input"
                      type="text"
                      bind:value={newAccountForm.accountSuffix}
                      placeholder={accountPrefix ? '後綴，例：-01' : '例：1101'}
                      required={createNewAccount}
                    />
                  </div>
                  {#if accountPrefix && newAccountForm.account_id}
                    <span class="acct-id-preview">完整編號：{newAccountForm.account_id}</span>
                  {/if}
                </div>
                <div class="form-group">
                  <label class="form-label" for="f-new-acct-name">科目名稱 *</label>
                  <input
                    id="f-new-acct-name"
                    class="form-input"
                    type="text"
                    bind:value={newAccountForm.name}
                    placeholder="例：玉山銀行存款"
                    required={createNewAccount}
                  />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label class="form-label" for="f-new-acct-type">科目類型 *</label>
                  <select id="f-new-acct-type" class="form-select" bind:value={newAccountForm.type}>
                    {#each ACCOUNT_TYPES as t}
                      <option value={t}>{ACCOUNT_TYPE_LABELS[t]}</option>
                    {/each}
                  </select>
                </div>
                <div class="form-group">
                  <label class="form-label" for="f-new-acct-balance">正常餘額 *</label>
                  <select id="f-new-acct-balance" class="form-select" bind:value={newAccountForm.normal_balance}>
                    <option value="DEBIT">借（Debit）</option>
                    <option value="CREDIT">貸（Credit）</option>
                  </select>
                </div>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label class="form-label" for="f-new-acct-currency">幣別</label>
                  <input
                    id="f-new-acct-currency"
                    class="form-input"
                    type="text"
                    bind:value={newAccountForm.currency}
                    placeholder="TWD"
                  />
                </div>
                <div class="form-group" style="display:flex;gap:20px;padding-top:26px;">
                  <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
                    <input type="checkbox" bind:checked={newAccountForm.is_summary} />
                    摘要科目
                  </label>
                  <label style="display:flex;align-items:center;gap:8px;cursor:pointer;font-size:12px;color:#9a8a6a;letter-spacing:0.06em;">
                    <input type="checkbox" bind:checked={newAccountForm.is_active} />
                    啟用
                  </label>
                </div>
              </div>

              <div class="form-group" style="margin-bottom:0;">
                <label class="form-label" for="f-ledger-currency">帳戶幣別</label>
                <input
                  id="f-ledger-currency"
                  class="form-input"
                  type="text"
                  bind:value={form.currency}
                  placeholder="TWD"
                />
              </div>
            </div>
          {/if}
        {:else}
          <div class="form-row">
            <div class="form-group">
              <!-- svelte-ignore a11y_label_has_associated_control -->
              <label class="form-label">關聯會計科目 *</label>
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
        {/if}

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