<script lang="ts">
  import { getBalanceSheet, getIncomeStatement } from '../api/report';
  import type { BalanceSheet, IncomeStatement } from '../types/report';
  import DonutChart from '../components/DonutChart.svelte';
  import HBarChart  from '../components/HBarChart.svelte';

  const ASSET_COLORS = ['#c4a86e', '#a07040', '#d4bc8a', '#b8944e', '#e0cfa0', '#887040', '#f0e0b0'];
  const LEQT_COLORS  = ['#a07890', '#7090b0', '#80a090', '#c09070', '#9080a0', '#70a080', '#b0a070'];

  type Tab = 'balance_sheet' | 'income_statement';
  interface Row {
    account_id: string;
    name:       string;
    value:      string;
    parent_id:  string | null;
    has_child:  boolean;
    depth:      number;
  }

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let activeTab = $state<Tab>('balance_sheet');

  let bsDate    = $state(today);
  let bsResult  = $state<BalanceSheet | null>(null);
  let bsLoading = $state(false);
  let bsError   = $state('');

  let expandedGroups = $state(new Set<string>());

  function toggleGroup(key: string): void {
    const next = new Set(expandedGroups);
    if (next.has(key)) next.delete(key);
    else next.add(key);
    expandedGroups = next;
  }

  let incBegin   = $state(monthStart);
  let incEnd     = $state(today);
  let incResult  = $state<IncomeStatement | null>(null);
  let incLoading = $state(false);
  let incError   = $state('');

  const bsAssets:      Row[] = $derived(bsResult?.assets.map(r      => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const bsLiabilities: Row[] = $derived(bsResult?.liabilities.map(r => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const bsEquity:      Row[] = $derived(bsResult?.equity.map(r      => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const incIncome:     Row[] = $derived(incResult?.income.map(r     => ({ account_id: r.account_id, name: r.name, value: r.amount,  parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const incExpenses:   Row[] = $derived(incResult?.expenses.map(r   => ({ account_id: r.account_id, name: r.name, value: r.amount,  parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);

  interface ChartItem { label: string; value: number; id: string; has_child: boolean; }
  interface BreadcrumbItem { label: string; isCurrent: boolean; onclick: () => void; }

  function rootChartRows(rows: Row[]): Row[] {
    return rows.filter(r => r.parent_id === null);
  }

  function childChartRows(rows: Row[], parentId: string): Row[] {
    return rows.filter(r => r.parent_id === parentId);
  }

  function visibleRows(rows: Row[], expanded: Set<string>): Row[] {
    return rows.filter(r => {
      if (r.parent_id === null) return true;
      let pid: string | null = r.parent_id;
      while (pid !== null) {
        if (!expanded.has(pid)) return false;
        pid = rows.find(p => p.account_id === pid)?.parent_id ?? null;
      }
      return true;
    });
  }

  let expandedBsRows = $state(new Set<string>());
  let expandedIsRows = $state(new Set<string>());

  function toggleBsRow(id: string): void {
    const next = new Set(expandedBsRows);
    if (next.has(id)) next.delete(id); else next.add(id);
    expandedBsRows = next;
  }
  function toggleIsRow(id: string): void {
    const next = new Set(expandedIsRows);
    if (next.has(id)) next.delete(id); else next.add(id);
    expandedIsRows = next;
  }

  let assetDrillStack = $state<string[]>([]);

  function handleAssetDrill(id: string): void {
    if (bsAssets.find(r => r.account_id === id)?.has_child) assetDrillStack = [...assetDrillStack, id];
  }
  const assetChartRows: Row[] = $derived(
    assetDrillStack.length === 0
      ? rootChartRows(bsAssets)
      : childChartRows(bsAssets, assetDrillStack[assetDrillStack.length - 1])
  );
  const assetChartItems: ChartItem[] = $derived(
    assetChartRows
      .filter(r => parseFloat(r.value) > 0)
      .map(r => ({ label: r.name, value: parseFloat(r.value), id: r.account_id, has_child: r.has_child }))
  );
  const assetBreadcrumbs: BreadcrumbItem[] = $derived([
    { label: '資產', isCurrent: assetDrillStack.length === 0, onclick: () => { assetDrillStack = []; } },
    ...assetDrillStack.map((id, i) => ({
      label: bsAssets.find(r => r.account_id === id)?.name ?? id,
      isCurrent: i === assetDrillStack.length - 1,
      onclick: () => { assetDrillStack = assetDrillStack.slice(0, i + 1); },
    })),
  ]);
  const bsChartLeqt: ChartItem[] = $derived(
    rootChartRows([...bsLiabilities, ...bsEquity])
      .filter(r => parseFloat(r.value) > 0)
      .map(r => ({ label: r.name, value: parseFloat(r.value), id: r.account_id, has_child: r.has_child }))
  );
  let incDrillStack = $state<string[]>([]);
  let expDrillStack = $state<string[]>([]);

  function handleIncomeDrill(id: string): void {
    if (incIncome.find(r => r.account_id === id)?.has_child) incDrillStack = [...incDrillStack, id];
  }
  function handleExpenseDrill(id: string): void {
    if (incExpenses.find(r => r.account_id === id)?.has_child) expDrillStack = [...expDrillStack, id];
  }

  const incChartIncome: ChartItem[] = $derived(
    (incDrillStack.length === 0
      ? rootChartRows(incIncome)
      : childChartRows(incIncome, incDrillStack[incDrillStack.length - 1])
    ).filter(r => parseFloat(r.value) > 0)
     .map(r => ({ label: r.name, value: parseFloat(r.value), id: r.account_id, has_child: r.has_child }))
  );
  const incChartExpenses: ChartItem[] = $derived(
    (expDrillStack.length === 0
      ? rootChartRows(incExpenses)
      : childChartRows(incExpenses, expDrillStack[expDrillStack.length - 1])
    ).filter(r => parseFloat(r.value) > 0)
     .map(r => ({ label: r.name, value: parseFloat(r.value), id: r.account_id, has_child: r.has_child }))
  );
  const incomeBreadcrumbs: BreadcrumbItem[] = $derived([
    { label: '收入', isCurrent: incDrillStack.length === 0, onclick: () => { incDrillStack = []; } },
    ...incDrillStack.map((id, i) => ({
      label: incIncome.find(r => r.account_id === id)?.name ?? id,
      isCurrent: i === incDrillStack.length - 1,
      onclick: () => { incDrillStack = incDrillStack.slice(0, i + 1); },
    })),
  ]);
  const expenseBreadcrumbs: BreadcrumbItem[] = $derived([
    { label: '費用', isCurrent: expDrillStack.length === 0, onclick: () => { expDrillStack = []; } },
    ...expDrillStack.map((id, i) => ({
      label: incExpenses.find(r => r.account_id === id)?.name ?? id,
      isCurrent: i === expDrillStack.length - 1,
      onclick: () => { expDrillStack = expDrillStack.slice(0, i + 1); },
    })),
  ]);

  async function queryBalanceSheet(e: Event): Promise<void> {
    e.preventDefault();
    bsLoading = true; bsError = '';
    try { bsResult = await getBalanceSheet(bsDate); expandedGroups = new Set<string>(); expandedBsRows = new Set<string>(); assetDrillStack = []; }
    catch (err) { bsError = err instanceof Error ? err.message : '查詢失敗'; bsResult = null; }
    finally { bsLoading = false; }
  }

  async function queryIncomeStatement(e: Event): Promise<void> {
    e.preventDefault();
    incLoading = true; incError = '';
    try { incResult = await getIncomeStatement(incBegin, incEnd); expandedIsRows = new Set<string>(); incDrillStack = []; expDrillStack = []; }
    catch (err) { incError = err instanceof Error ? err.message : '查詢失敗'; incResult = null; }
    finally { incLoading = false; }
  }

  function fmt(v: string): string {
    const n = parseFloat(v);
    return isNaN(n) ? v : n.toLocaleString();
  }
</script>

{#snippet bsGroup(key: string, title: string, rows: Row[], total: string)}
  <div class="bs-group">
    <button
      class="bs-group-header"
      class:expanded={expandedGroups.has(key)}
      aria-expanded={expandedGroups.has(key)}
      onclick={() => toggleGroup(key)}
    >
      <span class="expand-icon" aria-hidden="true">{expandedGroups.has(key) ? '▼' : '▶'}</span>
      <span class="bs-group-name">{title}</span>
      <span class="bs-group-total mono">{fmt(total)}</span>
    </button>
    {#if expandedGroups.has(key)}
      <div class="bs-group-body">
        <table class="data-table">
          <tbody>
            {#each visibleRows(rows, expandedBsRows) as row (row.account_id)}
              <tr>
                <td class="mono">{row.account_id}</td>
                <td>
                  <span class="row-indent" style="width:{row.depth * 14}px"></span>
                  {#if row.has_child}
                    <button
                      class="row-expand-btn"
                      aria-expanded={expandedBsRows.has(row.account_id)}
                      onclick={() => toggleBsRow(row.account_id)}
                    >{expandedBsRows.has(row.account_id) ? '▼' : '▶'}</button>
                  {/if}
                  {row.name}
                </td>
                <td class="mono" style="text-align:right">{fmt(row.value)}</td>
              </tr>
            {:else}
              <tr><td colspan="3" class="table-empty">無資料</td></tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
{/snippet}

{#snippet rowTable(title: string, rows: Row[], total: string, totalLabel: string)}
  <div class="report-group">
    <h3 class="report-group-title">{title}</h3>
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>科目編號</th>
            <th>科目名稱</th>
            <th style="text-align:right">金額</th>
          </tr>
        </thead>
        <tbody>
          {#each visibleRows(rows, expandedIsRows) as row (row.account_id)}
            <tr>
              <td class="mono">{row.account_id}</td>
              <td>
                <span class="row-indent" style="width:{row.depth * 14}px"></span>
                {#if row.has_child}
                  <button
                    class="row-expand-btn"
                    aria-expanded={expandedIsRows.has(row.account_id)}
                    onclick={() => toggleIsRow(row.account_id)}
                  >{expandedIsRows.has(row.account_id) ? '▼' : '▶'}</button>
                {/if}
                {row.name}
              </td>
              <td class="mono" style="text-align:right">{fmt(row.value)}</td>
            </tr>
          {:else}
            <tr><td colspan="3" class="table-empty">無資料</td></tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="tfoot-total">
            <td colspan="2">{totalLabel}</td>
            <td class="mono" style="text-align:right">{fmt(total)}</td>
          </tr>
        </tfoot>
      </table>
    </div>
  </div>
{/snippet}

<div class="content-header">
  <h1 class="content-title">財務報表</h1>
</div>

<div class="tab-nav" role="tablist">
  <button
    role="tab"
    class="tab-btn"
    class:active={activeTab === 'balance_sheet'}
    aria-selected={activeTab === 'balance_sheet'}
    onclick={() => { activeTab = 'balance_sheet'; }}
  >資產負債表</button>
  <button
    role="tab"
    class="tab-btn"
    class:active={activeTab === 'income_statement'}
    aria-selected={activeTab === 'income_statement'}
    onclick={() => { activeTab = 'income_statement'; }}
  >損益表</button>
</div>

{#if activeTab === 'balance_sheet'}
  <section class="section">
    <form class="query-bar" onsubmit={queryBalanceSheet}>
      <div class="form-group" style="margin:0">
        <label class="form-label" for="bs-date">報表日期</label>
        <input id="bs-date" class="form-input" type="date" bind:value={bsDate} required />
      </div>
      <button type="submit" class="btn-primary" disabled={bsLoading} style="align-self:flex-end">
        {bsLoading ? '查詢中…' : '查詢'}
      </button>
    </form>
    {#if bsError}
      <p class="query-error" role="alert">{bsError}</p>
    {/if}
    {#if bsResult}
      <p class="report-period">報表日期：{bsResult.report_date}</p>
      <div class="chart-row">
        <div class="chart-drilldown-wrap">
          <nav class="drill-breadcrumb" aria-label="資產類別導覽">
            {#each assetBreadcrumbs as crumb, i (i)}
              {#if i > 0}<span class="drill-breadcrumb-sep" aria-hidden="true">›</span>{/if}
              {#if crumb.isCurrent}
                <span class="drill-breadcrumb-item--current">{crumb.label}</span>
              {:else}
                <button class="drill-breadcrumb-item--link" onclick={crumb.onclick}>{crumb.label}</button>
              {/if}
            {/each}
          </nav>
          <DonutChart
            title=""
            items={assetChartItems}
            colors={ASSET_COLORS}
            onclick={handleAssetDrill}
          />
        </div>
        <DonutChart title="負債＋權益" items={bsChartLeqt} colors={LEQT_COLORS} />
      </div>
      <div class="bs-columns">
        <div class="bs-col">
          {@render bsGroup('assets',      '資產', bsAssets,      bsResult.total_assets)}
        </div>
        <div class="bs-col">
          {@render bsGroup('liabilities', '負債', bsLiabilities, bsResult.total_liabilities)}
          {@render bsGroup('equity',      '權益', bsEquity,      bsResult.total_equity)}
        </div>
      </div>
      <div class="report-summary">
        <span>淨資產（資產 − 負債）</span>
        <span class="mono">{fmt(bsResult.net_worth)}</span>
      </div>
    {/if}
  </section>
{:else}
  <section class="section">
    <form class="query-bar" onsubmit={queryIncomeStatement}>
      <div class="form-group" style="margin:0">
        <label class="form-label" for="inc-begin">開始日期</label>
        <input id="inc-begin" class="form-input" type="date" bind:value={incBegin} required />
      </div>
      <div class="form-group" style="margin:0">
        <label class="form-label" for="inc-end">結束日期</label>
        <input id="inc-end" class="form-input" type="date" bind:value={incEnd} required />
      </div>
      <button type="submit" class="btn-primary" disabled={incLoading} style="align-self:flex-end">
        {incLoading ? '查詢中…' : '查詢'}
      </button>
    </form>
    {#if incError}
      <p class="query-error" role="alert">{incError}</p>
    {/if}
    {#if incResult}
      <p class="report-period">期間：{incResult.start_date} 至 {incResult.end_date}</p>
      <HBarChart
        incomeItems={incChartIncome}
        expenseItems={incChartExpenses}
        netIncome={parseFloat(incResult.net_income)}
        {incomeBreadcrumbs}
        {expenseBreadcrumbs}
        onIncomeClick={handleIncomeDrill}
        onExpenseClick={handleExpenseDrill}
      />
      {@render rowTable('收入', incIncome, incResult.total_income, '收入合計')}
      {@render rowTable('費用', incExpenses, incResult.total_expenses, '費用合計')}
      <div class="report-summary">
        <span>本期淨利（收入 − 費用）</span>
        <span class="mono">{fmt(incResult.net_income)}</span>
      </div>
    {/if}
  </section>
{/if}
