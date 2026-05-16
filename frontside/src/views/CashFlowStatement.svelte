<script lang="ts">
  import { getCashFlowStatement } from '../api/report';
  import type { CashFlowStatement } from '../types/report';
  import { fmt } from '../lib/reportUtils.svelte';

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let cfBegin   = $state(monthStart);
  let cfEnd     = $state(today);
  let cfResult  = $state<CashFlowStatement | null>(null);
  let cfLoading = $state(false);
  let cfError   = $state('');

  async function queryCashFlow(e: Event): Promise<void> {
    e.preventDefault();
    cfLoading = true; cfError = '';
    try {
      cfResult = await getCashFlowStatement(cfBegin, cfEnd);
    } catch (err) {
      cfError = err instanceof Error ? err.message : '查詢失敗';
      cfResult = null;
    } finally {
      cfLoading = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">現金流量表</h1>
</div>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">現金流量表</h2>
  </header>
  <form class="query-bar" onsubmit={queryCashFlow}>
    <div class="form-group" style="margin:0">
      <label class="form-label" for="cf-begin">開始日期</label>
      <input id="cf-begin" class="form-input" type="date" bind:value={cfBegin} required />
    </div>
    <div class="form-group" style="margin:0">
      <label class="form-label" for="cf-end">結束日期</label>
      <input id="cf-end" class="form-input" type="date" bind:value={cfEnd} required />
    </div>
    <button type="submit" class="btn-primary" disabled={cfLoading} style="align-self:flex-end">
      {cfLoading ? '查詢中…' : '查詢'}
    </button>
  </form>
  {#if cfError}
    <p class="query-error" role="alert">{cfError}</p>
  {/if}
  {#if cfResult}
    <p class="report-period">期間：{cfResult.start_date} 至 {cfResult.end_date}</p>

    <div class="cf-section">
      <h3 class="cf-section-title">營業活動現金流量</h3>
      <table class="data-table">
        <tbody>
          <tr class="cf-row--highlight">
            <td>本期淨利</td>
            <td class="mono" style="text-align:right">{fmt(cfResult.operating_activities.net_income)}</td>
          </tr>
          {#each cfResult.operating_activities.adjustments as item (item.account_id)}
            <tr>
              <td>{item.name}</td>
              <td class="mono" style="text-align:right">{fmt(item.amount)}</td>
            </tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="tfoot-total">
            <td>營業活動現金流量合計</td>
            <td class="mono" style="text-align:right">{fmt(cfResult.operating_activities.total)}</td>
          </tr>
        </tfoot>
      </table>
    </div>

    <div class="cf-section">
      <h3 class="cf-section-title">投資活動現金流量</h3>
      <table class="data-table">
        <tbody>
          {#each cfResult.investing_activities.items as item (item.account_id)}
            <tr>
              <td>{item.name}</td>
              <td class="mono" style="text-align:right">{fmt(item.amount)}</td>
            </tr>
          {:else}
            <tr><td colspan="2" class="table-empty">無資料</td></tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="tfoot-total">
            <td>投資活動現金流量合計</td>
            <td class="mono" style="text-align:right">{fmt(cfResult.investing_activities.total)}</td>
          </tr>
        </tfoot>
      </table>
    </div>

    <div class="cf-section">
      <h3 class="cf-section-title">融資活動現金流量</h3>
      <table class="data-table">
        <tbody>
          {#each cfResult.financing_activities.items as item (item.account_id)}
            <tr>
              <td>{item.name}</td>
              <td class="mono" style="text-align:right">{fmt(item.amount)}</td>
            </tr>
          {:else}
            <tr><td colspan="2" class="table-empty">無資料</td></tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="tfoot-total">
            <td>融資活動現金流量合計</td>
            <td class="mono" style="text-align:right">{fmt(cfResult.financing_activities.total)}</td>
          </tr>
        </tfoot>
      </table>
    </div>

    <div class="cf-summary">
      <div class="cf-summary-row">
        <span>期初現金</span>
        <span class="mono">{fmt(cfResult.beginning_cash)}</span>
      </div>
      <div class="cf-summary-row">
        <span>現金淨變動</span>
        <span class="mono">{fmt(cfResult.net_change)}</span>
      </div>
      <div class="cf-summary-row cf-summary-row--total">
        <span>期末現金</span>
        <span class="mono">{fmt(cfResult.ending_cash)}</span>
      </div>
    </div>
  {/if}
</section>
