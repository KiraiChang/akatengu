<script lang="ts">
  import { getIncomeStatement } from '../api/report';
  import type { IncomeStatement } from '../types/report';
  import { visibleRows, fmt } from '../lib/reportUtils.svelte';
  import type { Row } from '../lib/reportUtils.svelte';
  import HBarChart from '../components/HBarChart.svelte';
  import { goToAccountAnalysis } from '../lib/navigate';

  interface ChartItem { label: string; value: number; id: string; has_child: boolean; }
  interface BreadcrumbItem { label: string; isCurrent: boolean; onclick: () => void; }

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let incBegin   = $state(monthStart);
  let incEnd     = $state(today);
  let incResult  = $state<IncomeStatement | null>(null);
  let incLoading = $state(false);
  let incError   = $state('');

  let expandedIsRows = $state(new Set<string>());

  function toggleIsRow(id: string): void {
    const next = new Set(expandedIsRows);
    if (next.has(id)) next.delete(id); else next.add(id);
    expandedIsRows = next;
  }

  const incIncome:   Row[] = $derived(incResult?.income.map(r   => ({ account_id: r.account_id, name: r.name, value: r.amount, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const incExpenses: Row[] = $derived(incResult?.expenses.map(r => ({ account_id: r.account_id, name: r.name, value: r.amount, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);

  function rootChartRows(rows: Row[]): Row[] { return rows.filter(r => r.parent_id === null); }
  function childChartRows(rows: Row[], parentId: string): Row[] { return rows.filter(r => r.parent_id === parentId); }

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

  async function queryIncomeStatement(e: Event): Promise<void> {
    e.preventDefault();
    incLoading = true; incError = '';
    try {
      incResult = await getIncomeStatement(incBegin, incEnd);
      expandedIsRows = new Set<string>();
      incDrillStack = [];
      expDrillStack = [];
    } catch (err) {
      incError = err instanceof Error ? err.message : '查詢失敗';
      incResult = null;
    } finally {
      incLoading = false;
    }
  }
</script>

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
                <button class="account-link" onclick={() => goToAccountAnalysis(row.account_id)}>{row.name}</button>
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
  <h1 class="content-title">損益表</h1>
</div>

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
