<script lang="ts">
  import { getCashFlowStatement, getDirectCashFlowStatement } from '../api/report';
  import type { CashFlowStatement, DirectCashFlowStatement } from '../types/report';
  import { fmt } from '../lib/reportUtils.svelte';

  type Method = 'indirect' | 'direct';

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let cfBegin    = $state(monthStart);
  let cfEnd      = $state(today);
  let method     = $state<Method>('indirect');
  let cfResult   = $state<CashFlowStatement | null>(null);
  let cfDirect   = $state<DirectCashFlowStatement | null>(null);
  let cfLoading  = $state(false);
  let cfError    = $state('');

  async function queryCashFlow(e: Event): Promise<void> {
    e.preventDefault();
    cfLoading = true; cfError = '';
    cfResult = null; cfDirect = null;
    try {
      if (method === 'indirect') {
        cfResult = await getCashFlowStatement(cfBegin, cfEnd);
      } else {
        cfDirect = await getDirectCashFlowStatement(cfBegin, cfEnd);
      }
    } catch (err) {
      cfError = err instanceof Error ? err.message : '查詢失敗';
    } finally {
      cfLoading = false;
    }
  }

  function switchMethod(m: Method): void {
    method = m;
    cfResult = null;
    cfDirect = null;
    cfError  = '';
  }

  const startDate = $derived(cfResult?.start_date ?? cfDirect?.start_date ?? '');
  const endDate   = $derived(cfResult?.end_date   ?? cfDirect?.end_date   ?? '');
  const hasResult = $derived(cfResult !== null || cfDirect !== null);

  const investing  = $derived(cfResult?.investing_activities ?? cfDirect?.investing_activities ?? null);
  const financing  = $derived(cfResult?.financing_activities ?? cfDirect?.financing_activities ?? null);
  const netChange  = $derived(cfResult?.net_change  ?? cfDirect?.net_change  ?? '');
  const beginCash  = $derived(cfResult?.beginning_cash ?? cfDirect?.beginning_cash ?? '');
  const endCash    = $derived(cfResult?.ending_cash   ?? cfDirect?.ending_cash     ?? '');
</script>

<div class="content-header">
  <h1 class="content-title">現金流量表</h1>
</div>

<section class="section">
  <nav class="tab-nav" aria-label="現金流量表方法">
    <button
      type="button"
      class="tab-btn"
      class:active={method === 'indirect'}
      onclick={() => switchMethod('indirect')}
    >間接法</button>
    <button
      type="button"
      class="tab-btn"
      class:active={method === 'direct'}
      onclick={() => switchMethod('direct')}
    >直接法</button>
  </nav>

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

  {#if hasResult}
    <p class="report-period">期間：{startDate} 至 {endDate}</p>

    <!-- Operating Activities -->
    <div class="cf-section">
      <h3 class="cf-section-title">營業活動現金流量</h3>
      <table class="data-table">
        {#if cfResult}
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
        {:else if cfDirect}
          <tbody>
            <tr>
              <td>收到的現金</td>
              <td class="mono" style="text-align:right">{fmt(cfDirect.operating_activities.cash_received)}</td>
            </tr>
            <tr>
              <td>支付的現金</td>
              <td class="mono" style="text-align:right">{fmt(cfDirect.operating_activities.cash_paid)}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="tfoot-total">
              <td>營業活動現金流量合計</td>
              <td class="mono" style="text-align:right">{fmt(cfDirect.operating_activities.total)}</td>
            </tr>
          </tfoot>
        {/if}
      </table>
    </div>

    <!-- Investing Activities -->
    {#if investing}
      <div class="cf-section">
        <h3 class="cf-section-title">投資活動現金流量</h3>
        <table class="data-table">
          <tbody>
            {#each investing.items as item (item.account_id)}
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
              <td class="mono" style="text-align:right">{fmt(investing.total)}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    {/if}

    <!-- Financing Activities -->
    {#if financing}
      <div class="cf-section">
        <h3 class="cf-section-title">融資活動現金流量</h3>
        <table class="data-table">
          <tbody>
            {#each financing.items as item (item.account_id)}
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
              <td class="mono" style="text-align:right">{fmt(financing.total)}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    {/if}

    <!-- Summary -->
    <div class="cf-summary">
      <div class="cf-summary-row">
        <span>期初現金</span>
        <span class="mono">{fmt(beginCash)}</span>
      </div>
      <div class="cf-summary-row">
        <span>現金淨變動</span>
        <span class="mono">{fmt(netChange)}</span>
      </div>
      <div class="cf-summary-row cf-summary-row--total">
        <span>期末現金</span>
        <span class="mono">{fmt(endCash)}</span>
      </div>
    </div>
  {/if}
</section>
