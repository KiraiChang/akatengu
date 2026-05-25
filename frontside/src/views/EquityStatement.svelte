<script lang="ts">
  import { getEquityStatement } from '../api/report';
  import type { EquityStatement } from '../types/report';
  import { fmt } from '../lib/reportUtils.svelte';

  const today      = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  let eqBegin   = $state(monthStart);
  let eqEnd     = $state(today);
  let eqResult  = $state<EquityStatement | null>(null);
  let eqLoading = $state(false);
  let eqError   = $state('');

  async function queryEquityStatement(e: Event): Promise<void> {
    e.preventDefault();
    eqLoading = true; eqError = '';
    try {
      eqResult = await getEquityStatement(eqBegin, eqEnd);
    } catch (err) {
      eqError = err instanceof Error ? err.message : '查詢失敗';
      eqResult = null;
    } finally {
      eqLoading = false;
    }
  }
</script>

<div class="content-header">
  <h1 class="content-title">權益變動表</h1>
</div>

<section class="section">
  <form class="query-bar" onsubmit={queryEquityStatement}>
    <div class="form-group" style="margin:0">
      <label class="form-label" for="eq-begin">開始日期</label>
      <input id="eq-begin" class="form-input" type="date" bind:value={eqBegin} required />
    </div>
    <div class="form-group" style="margin:0">
      <label class="form-label" for="eq-end">結束日期</label>
      <input id="eq-end" class="form-input" type="date" bind:value={eqEnd} required />
    </div>
    <button type="submit" class="btn-primary" disabled={eqLoading} style="align-self:flex-end">
      {eqLoading ? '查詢中…' : '查詢'}
    </button>
  </form>
  {#if eqError}
    <p class="query-error" role="alert">{eqError}</p>
  {/if}
  {#if eqResult}
    <p class="report-period">期間：{eqResult.start_date} 至 {eqResult.end_date}</p>
    <div class="table-wrap">
      <table class="data-table">
        <thead>
          <tr>
            <th>科目名稱</th>
            <th style="text-align:right">期初餘額</th>
            <th style="text-align:right">本期變動</th>
            <th style="text-align:right">期末餘額</th>
          </tr>
        </thead>
        <tbody>
          {#each eqResult.items as item (item.account_id)}
            <tr class:eq-row--summary={item.is_summary} class:eq-row--virtual={item.is_virtual}>
              <td>{item.name}</td>
              <td class="mono" style="text-align:right">{fmt(item.begin_balance)}</td>
              <td class="mono" style="text-align:right">{fmt(item.period_change)}</td>
              <td class="mono" style="text-align:right">{fmt(item.end_balance)}</td>
            </tr>
          {:else}
            <tr><td colspan="4" class="table-empty">無資料</td></tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="tfoot-total">
            <td>合計</td>
            <td class="mono" style="text-align:right">{fmt(eqResult.total_begin_balance)}</td>
            <td class="mono" style="text-align:right">{fmt(eqResult.total_period_change)}</td>
            <td class="mono" style="text-align:right">{fmt(eqResult.total_end_balance)}</td>
          </tr>
        </tfoot>
      </table>
    </div>
    <div class="report-summary">
      <span>本期淨利</span>
      <span class="mono">{fmt(eqResult.net_income)}</span>
    </div>
  {/if}
</section>
