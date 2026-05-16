<script lang="ts">
  import { getBalanceSheet } from '../api/report';
  import type { BalanceSheet } from '../types/report';
  import { visibleRows, fmt } from '../lib/reportUtils.svelte';
  import type { Row } from '../lib/reportUtils.svelte';
  import DonutChart from '../components/DonutChart.svelte';

  const ASSET_COLORS = ['#c4a86e', '#a07040', '#d4bc8a', '#b8944e', '#e0cfa0', '#887040', '#f0e0b0'];
  const LEQT_COLORS  = ['#a07890', '#7090b0', '#80a090', '#c09070', '#9080a0', '#70a080', '#b0a070'];

  interface ChartItem { label: string; value: number; id: string; has_child: boolean; }
  interface BreadcrumbItem { label: string; isCurrent: boolean; onclick: () => void; }

  const today = new Date().toISOString().slice(0, 10);

  let bsDate    = $state(today);
  let bsResult  = $state<BalanceSheet | null>(null);
  let bsLoading = $state(false);
  let bsError   = $state('');

  let expandedGroups = $state(new Set<string>());
  let expandedBsRows = $state(new Set<string>());

  function toggleGroup(key: string): void {
    const next = new Set(expandedGroups);
    if (next.has(key)) next.delete(key); else next.add(key);
    expandedGroups = next;
  }

  function toggleBsRow(id: string): void {
    const next = new Set(expandedBsRows);
    if (next.has(id)) next.delete(id); else next.add(id);
    expandedBsRows = next;
  }

  const bsAssets:      Row[] = $derived(bsResult?.assets.map(r      => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const bsLiabilities: Row[] = $derived(bsResult?.liabilities.map(r => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);
  const bsEquity:      Row[] = $derived(bsResult?.equity.map(r      => ({ account_id: r.account_id, name: r.name, value: r.balance, parent_id: r.parent_id, has_child: r.has_child, depth: r.depth })) ?? []);

  function rootChartRows(rows: Row[]): Row[] {
    return rows.filter(r => r.parent_id === null);
  }

  function childChartRows(rows: Row[], parentId: string): Row[] {
    return rows.filter(r => r.parent_id === parentId);
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

  async function queryBalanceSheet(e: Event): Promise<void> {
    e.preventDefault();
    bsLoading = true; bsError = '';
    try {
      bsResult = await getBalanceSheet(bsDate);
      expandedGroups = new Set<string>();
      expandedBsRows = new Set<string>();
      assetDrillStack = [];
    } catch (err) {
      bsError = err instanceof Error ? err.message : '查詢失敗';
      bsResult = null;
    } finally {
      bsLoading = false;
    }
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

<div class="content-header">
  <h1 class="content-title">資產負債表</h1>
</div>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">資產負債表</h2>
  </header>
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
