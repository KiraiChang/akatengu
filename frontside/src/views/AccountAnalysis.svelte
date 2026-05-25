<script lang="ts">
  import { getAccountAll } from '../api/account';
  import { getLedgerAccountAll } from '../api/ledger';
  import { getChildrenBalance, getAccountEntries, getMonthlyBalance } from '../api/accountAnalysis';
  import AccountSelect from '../components/AccountSelect.svelte';
  import type { Account } from '../types/account';
  import type { LedgerAccount } from '../types/ledger';
  import type {
    AccountChildrenBalanceResult,
    AccountJournalEntryRow,
    AccountMonthlyBalance,
  } from '../types/accountAnalysis';

  type Tab = 'children' | 'entries' | 'monthly';

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;
  const currentYM  = today.slice(0, 7);
  const prevYearYM = `${String(parseInt(currentYM.slice(0, 4)) - 1)}-${currentYM.slice(5)}`;

  const PAGE_SIZE = 20;

  let allAccounts = $state<Account[]>([]);
  let allLedgers  = $state<LedgerAccount[]>([]);
  let selectedId  = $state('');
  let activeTab   = $state<Tab>('children');

  // Children tab — drill state
  let drillStack       = $state<{ id: string; name: string }[]>([]);
  const currentDrillId = $derived(
    drillStack.length > 0 ? drillStack[drillStack.length - 1].id : selectedId,
  );
  let childrenResult  = $state<AccountChildrenBalanceResult | null>(null);
  let childrenLoading = $state(false);
  let childrenError   = $state('');

  // Entries tab
  let eFrom       = $state(monthStart);
  let eTo         = $state(today);
  let ePage       = $state(1);
  let eData       = $state<AccountJournalEntryRow[]>([]);
  let eTotalPages = $state(0);
  let eLoading    = $state(false);
  let eError      = $state('');
  let eQueried    = $state(false);

  // Monthly tab
  let mFrom    = $state(prevYearYM);
  let mTo      = $state(currentYM);
  let mData    = $state<AccountMonthlyBalance[]>([]);
  let mLoading = $state(false);
  let mError   = $state('');
  let mQueried = $state(false);

  const selectedAccount = $derived(allAccounts.find(a => a.account_id === selectedId) ?? null);

  const mMaxVal = $derived(
    mData.length === 0
      ? 1
      : Math.max(
          ...mData.map(m =>
            Math.max(parseFloat(m.debit_total) || 0, parseFloat(m.credit_total) || 0),
          ),
          0.01,
        ),
  );

  $effect(() => {
    void Promise.all([getAccountAll(), getLedgerAccountAll()])
      .then(([accts, ldgrs]) => { allAccounts = accts; allLedgers = ldgrs; })
      .catch(() => {});
  });

  $effect(() => {
    if (!currentDrillId) { childrenResult = null; return; }
    void loadChildren(currentDrillId);
  });

  async function loadChildren(accountId: string): Promise<void> {
    childrenLoading = true;
    childrenError   = '';
    try {
      childrenResult = await getChildrenBalance(accountId);
    } catch (err) {
      childrenError  = err instanceof Error ? err.message : '載入失敗';
      childrenResult = null;
    } finally {
      childrenLoading = false;
    }
  }

  function selectAccount(id: string): void {
    if (id === selectedId) return;
    selectedId  = id;
    drillStack  = [];
    eData       = [];
    eTotalPages = 0;
    eQueried    = false;
    mData       = [];
    mQueried    = false;
  }

  function drillInto(id: string, name: string): void {
    drillStack = [...drillStack, { id, name }];
  }

  function drillTo(index: number): void {
    drillStack = index < 0 ? [] : drillStack.slice(0, index + 1);
  }

  async function queryEntries(e: Event): Promise<void> {
    e.preventDefault();
    if (!selectedId) return;
    ePage    = 1;
    eQueried = true;
    await loadEntries();
  }

  async function loadEntries(): Promise<void> {
    eLoading = true;
    eError   = '';
    try {
      const res = await getAccountEntries(selectedId, eFrom, eTo, ePage - 1, PAGE_SIZE);
      eData       = res.data ?? [];
      eTotalPages = res.meta?.total_pages ?? 0;
    } catch (err) {
      eError = err instanceof Error ? err.message : '查詢失敗';
      eData  = [];
    } finally {
      eLoading = false;
    }
  }

  async function changePage(delta: number): Promise<void> {
    ePage = Math.max(1, Math.min(eTotalPages, ePage + delta));
    await loadEntries();
  }

  async function queryMonthly(e: Event): Promise<void> {
    e.preventDefault();
    if (!selectedId) return;
    mLoading = true;
    mError   = '';
    mQueried = true;
    try {
      mData = await getMonthlyBalance(selectedId, mFrom, mTo);
    } catch (err) {
      mError = err instanceof Error ? err.message : '查詢失敗';
      mData  = [];
    } finally {
      mLoading = false;
    }
  }

  function barWidth(val: string): string {
    const n = parseFloat(val) || 0;
    return `${Math.round((n / mMaxVal) * 100)}%`;
  }

  function fmtAmount(s: string): string {
    const n = parseFloat(s);
    if (isNaN(n)) return '—';
    return n.toLocaleString('zh-TW', { minimumFractionDigits: 2 });
  }

  function netClass(debit: string, credit: string): string {
    const net = (parseFloat(debit) || 0) - (parseFloat(credit) || 0);
    if (net >  0.001) return 'aa-net--pos';
    if (net < -0.001) return 'aa-net--neg';
    return 'aa-net--zero';
  }

  function fmtNet(debit: string, credit: string): string {
    const net = (parseFloat(debit) || 0) - (parseFloat(credit) || 0);
    return net.toLocaleString('zh-TW', { minimumFractionDigits: 2, signDisplay: 'exceptZero' });
  }
</script>

<div class="content-header">
  <h1 class="content-title">科目分析</h1>
</div>

<section class="section">

  <!-- ── Account selector ────────────────── -->
  <div class="aa-selector">
    <span class="aa-selector-label">選擇科目</span>
    <div class="aa-selector-wrap">
      <AccountSelect
        accounts={allAccounts}
        value={selectedId}
        placeholder="選擇要分析的科目…"
        onselect={(id) => selectAccount(id)}
      />
    </div>
  </div>

  {#if !selectedId}
    <div class="table-empty" style="padding:40px;text-align:center;">請選擇科目以開始分析</div>
  {:else}

    <!-- ── Tab nav ──────────────────────── -->
    <nav class="tab-nav" aria-label="分析類型">
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === 'children'}
        onclick={() => (activeTab = 'children')}
      >子科目餘額</button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === 'entries'}
        onclick={() => (activeTab = 'entries')}
      >分錄明細</button>
      <button
        type="button"
        class="tab-btn"
        class:active={activeTab === 'monthly'}
        onclick={() => (activeTab = 'monthly')}
      >月別趨勢</button>
    </nav>

    <!-- ── Tab: Children Balance ─────────── -->
    {#if activeTab === 'children'}

      <!-- Breadcrumb -->
      <div class="aa-breadcrumb">
        {#if drillStack.length === 0}
          <span class="aa-breadcrumb-current">{selectedAccount?.name ?? selectedId}</span>
        {:else}
          <button class="aa-breadcrumb-btn" onclick={() => drillTo(-1)}>
            {selectedAccount?.name ?? selectedId}
          </button>
          {#each drillStack as item, i}
            <span class="aa-breadcrumb-sep" aria-hidden="true">›</span>
            {#if i < drillStack.length - 1}
              <button class="aa-breadcrumb-btn" onclick={() => drillTo(i)}>{item.name}</button>
            {:else}
              <span class="aa-breadcrumb-current">{item.name}</span>
            {/if}
          {/each}
        {/if}
      </div>

      {#if childrenError}
        <p class="query-error" role="alert">{childrenError}</p>
      {:else if childrenLoading}
        <div class="table-empty" style="padding:32px;text-align:center;">載入中…</div>
      {:else if childrenResult && childrenResult.children.length === 0}
        <div class="table-empty" style="padding:24px;text-align:center;">此科目無直屬子科目</div>
      {:else if childrenResult}
        <div class="table-wrap">
          <table class="data-table" aria-label="子科目餘額">
            <thead>
              <tr>
                <th>科目編號</th>
                <th>科目名稱</th>
                <th class="num">借方合計</th>
                <th class="num">貸方合計</th>
                <th class="num">淨額（借－貸）</th>
              </tr>
            </thead>
            <tbody>
              {#each childrenResult.children as child (child.account_id)}
                <tr>
                  <td class="mono" style="font-size:11px;color:#7a8096">{child.account_id}</td>
                  <td>
                    {#if child.has_child}
                      <button
                        class="aa-drill-btn"
                        onclick={() => drillInto(child.account_id, child.name)}
                      >{child.name}</button>
                    {:else}
                      {child.name}
                    {/if}
                    {#if child.is_summary}
                      <span style="font-size:10px;color:#5c6278;margin-left:4px;">[彙總]</span>
                    {/if}
                  </td>
                  <td class="mono num">{fmtAmount(child.debit_total)}</td>
                  <td class="mono num">{fmtAmount(child.credit_total)}</td>
                  <td class="mono num {netClass(child.debit_total, child.credit_total)}">
                    {fmtNet(child.debit_total, child.credit_total)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

    <!-- ── Tab: Journal Entries ──────────── -->
    {:else if activeTab === 'entries'}

      <form class="aa-query-bar" onsubmit={queryEntries}>
        <div class="form-group" style="margin:0">
          <label class="form-label" for="aa-e-from">開始日期</label>
          <input id="aa-e-from" class="form-input" type="date" bind:value={eFrom} required />
        </div>
        <div class="form-group" style="margin:0">
          <label class="form-label" for="aa-e-to">結束日期</label>
          <input id="aa-e-to" class="form-input" type="date" bind:value={eTo} required />
        </div>
        <button type="submit" class="btn-primary" disabled={eLoading} style="align-self:flex-end">
          {eLoading ? '查詢中…' : '查詢'}
        </button>
      </form>

      {#if eError}
        <p class="query-error" role="alert">{eError}</p>
      {:else if !eQueried}
        <div class="table-empty" style="padding:24px;text-align:center;">請設定日期範圍並點擊查詢</div>
      {:else if eLoading && eData.length === 0}
        <div class="table-empty" style="padding:24px;text-align:center;">載入中…</div>
      {:else if eData.length === 0}
        <div class="table-empty" style="padding:24px;text-align:center;">此期間無分錄資料</div>
      {:else}
        <div class="table-wrap">
          <table class="data-table" aria-label="分錄明細">
            <thead>
              <tr>
                <th>日期</th>
                <th>摘要</th>
                <th>帳戶</th>
                <th class="num">借方</th>
                <th class="num">貸方</th>
                <th>備註</th>
              </tr>
            </thead>
            <tbody>
              {#each eData as row (row.entry_id)}
                {@const ldgr = allLedgers.find(l => l.ledger_id === row.ledger_id)}
                <tr>
                  <td class="mono" style="font-size:11px">{row.txn_date}</td>
                  <td style="color:#dedad3">{row.description}</td>
                  <td style="font-size:11px;color:#5c6278">
                    {#if ldgr}
                      {ldgr.institution}<span style="color:#3d4258"> · </span>{ldgr.name}
                    {:else if row.ledger_id}
                      #{row.ledger_id}
                    {:else}
                      —
                    {/if}
                  </td>
                  <td class="mono num">
                    {parseFloat(row.debit) > 0 ? fmtAmount(row.debit) : ''}
                  </td>
                  <td class="mono num">
                    {parseFloat(row.credit) > 0 ? fmtAmount(row.credit) : ''}
                  </td>
                  <td style="font-size:11px;color:#5c6278">{row.note ?? '—'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        {#if eTotalPages > 1}
          <div class="pagination">
            <button class="page-btn" disabled={ePage <= 1} onclick={() => changePage(-1)}>‹</button>
            <span class="page-info">{ePage} / {eTotalPages}</span>
            <button class="page-btn" disabled={ePage >= eTotalPages} onclick={() => changePage(1)}>›</button>
          </div>
        {/if}
      {/if}

    <!-- ── Tab: Monthly Balance ───────────── -->
    {:else if activeTab === 'monthly'}

      <form class="aa-query-bar" onsubmit={queryMonthly}>
        <div class="form-group" style="margin:0">
          <label class="form-label" for="aa-m-from">起始月份</label>
          <input id="aa-m-from" class="form-input" type="month" bind:value={mFrom} required />
        </div>
        <div class="form-group" style="margin:0">
          <label class="form-label" for="aa-m-to">結束月份</label>
          <input id="aa-m-to" class="form-input" type="month" bind:value={mTo} required />
        </div>
        <button type="submit" class="btn-primary" disabled={mLoading} style="align-self:flex-end">
          {mLoading ? '查詢中…' : '查詢'}
        </button>
      </form>

      {#if mError}
        <p class="query-error" role="alert">{mError}</p>
      {:else if !mQueried}
        <div class="table-empty" style="padding:24px;text-align:center;">請設定月份範圍並點擊查詢</div>
      {:else if mLoading && mData.length === 0}
        <div class="table-empty" style="padding:24px;text-align:center;">載入中…</div>
      {:else if mData.length === 0}
        <div class="table-empty" style="padding:24px;text-align:center;">此期間無資料</div>
      {:else}
        <div class="table-wrap">
          <table class="data-table" aria-label="月別餘額趨勢">
            <thead>
              <tr>
                <th>月份</th>
                <th class="num">借方合計</th>
                <th class="aa-monthly-bar-cell"></th>
                <th class="num">貸方合計</th>
                <th class="aa-monthly-bar-cell"></th>
                <th class="num">淨額（借－貸）</th>
              </tr>
            </thead>
            <tbody>
              {#each mData as row (row.month)}
                <tr>
                  <td class="mono" style="font-size:12px">
                    {row.month}
                    {#if row.has_snapshot}
                      <span class="aa-snapshot-dot" title="含快照資料"></span>
                    {/if}
                  </td>
                  <td class="mono num" style="font-size:12px">{fmtAmount(row.debit_total)}</td>
                  <td class="aa-monthly-bar-cell">
                    <div class="aa-monthly-track">
                      <div class="aa-monthly-bar--debit" style="width:{barWidth(row.debit_total)}"></div>
                    </div>
                  </td>
                  <td class="mono num" style="font-size:12px">{fmtAmount(row.credit_total)}</td>
                  <td class="aa-monthly-bar-cell">
                    <div class="aa-monthly-track">
                      <div class="aa-monthly-bar--credit" style="width:{barWidth(row.credit_total)}"></div>
                    </div>
                  </td>
                  <td class="mono num {netClass(row.debit_total, row.credit_total)}" style="font-size:12px">
                    {fmtNet(row.debit_total, row.credit_total)}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

    {/if}
  {/if}

</section>
