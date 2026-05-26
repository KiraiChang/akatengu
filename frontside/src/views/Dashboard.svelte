<script lang="ts">
  import { getDashboardSummary, getMonthlyTrend, getLedgerBalances } from '../api/dashboard';
  import { getTransactionPaged } from '../api/transaction';
  import type { DashboardSummary, MonthlyTrendItem, LedgerBalance } from '../types/dashboard';
  import type { Transaction } from '../types/transaction';

  const today = new Date().toISOString().slice(0, 10);

  let summary   = $state<DashboardSummary | null>(null);
  let trend     = $state<MonthlyTrendItem[]>([]);
  let ledgers   = $state<LedgerBalance[]>([]);
  let txns      = $state<Transaction[]>([]);
  let isLoading = $state(true);
  let error     = $state('');

  const netProfit = $derived(
    summary ? parseFloat(summary.month_income) - parseFloat(summary.month_expense) : 0,
  );
  const trendMax = $derived(
    trend.length === 0
      ? 1
      : Math.max(...trend.flatMap(r => [parseFloat(r.income) || 0, parseFloat(r.expense) || 0]), 0.01),
  );

  $effect(() => {
    void Promise.all([
      getDashboardSummary(),
      getMonthlyTrend(12),
      getLedgerBalances(),
      getTransactionPaged({ page: 0, pageSize: 10 }),
    ])
      .then(([s, t, l, p]) => { summary = s; trend = t; ledgers = l; txns = p.data ?? []; })
      .catch(err => { error = err instanceof Error ? err.message : '載入失敗'; })
      .finally(() => { isLoading = false; });
  });

  const LEDGER_TYPE_LABELS: Record<string, string> = {
    BANK_ACCOUNT: '銀行', CREDIT_CARD: '信用卡', LOAN: '貸款',
  };
  const LEDGER_TAG_CLASS: Record<string, string> = {
    BANK_ACCOUNT: 'dash-ledger-tag--bank', CREDIT_CARD: 'dash-ledger-tag--card', LOAN: 'dash-ledger-tag--loan',
  };
  const STATUS_LABELS: Record<string, string> = {
    ACTIVE: '正常', CORRECTED: '已更正', VOIDED: '已沖銷', VOID_REF: '沖銷',
  };

  function fmtAmt(s: string): string {
    const n = parseFloat(s);
    return isNaN(n) ? '—' : Math.round(n).toLocaleString('zh-TW');
  }
  function fmtNet(n: number): string {
    return (n >= 0 ? '+' : '') + Math.round(n).toLocaleString('zh-TW');
  }
  function barW(s: string): string {
    return `${Math.round((Math.max(parseFloat(s) || 0, 0) / trendMax) * 100)}%`;
  }
  function txnBadge(status: string): string {
    return status === 'ACTIVE' ? 'badge approved' : status === 'CORRECTED' ? 'badge pending' : 'badge-plain';
  }
</script>

<div class="content-header">
  <h1 class="content-title">儀表板</h1>
  <span class="content-date">{summary?.as_of_date ?? today}</span>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<!-- ── Stat Cards ────────────────────────── -->
<div class="stat-grid">
  <div class="stat-card">
    <div class="stat-label">本月收入</div>
    <div class="stat-value income">{summary ? fmtAmt(summary.month_income) : '—'}</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend" class:up={netProfit > 0} class:down={netProfit < 0} class:neutral={!summary}>
      {#if summary}本月淨利 {fmtNet(netProfit)}{/if}
    </div>
  </div>
  <div class="stat-card">
    <div class="stat-label">本月支出</div>
    <div class="stat-value expense">{summary ? fmtAmt(summary.month_expense) : '—'}</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend neutral">
      {#if summary && parseFloat(summary.month_income) > 0}
        佔收入 {Math.round(parseFloat(summary.month_expense) / parseFloat(summary.month_income) * 100)}%
      {/if}
    </div>
  </div>
  <div class="stat-card">
    <div class="stat-label">淨資產</div>
    <div class="stat-value">{summary ? fmtAmt(summary.total_equity) : '—'}</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend neutral">
      {#if summary}資產 {fmtAmt(summary.total_assets)} ／ 負債 {fmtAmt(summary.total_liabilities)}{/if}
    </div>
  </div>
  <div class="stat-card accent">
    <div class="stat-label">現金餘額</div>
    <div class="stat-value">{summary ? fmtAmt(summary.cash_balance) : '—'}</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend neutral">{#if summary}截至 {summary.as_of_date}{/if}</div>
  </div>
</div>

<!-- ── Two-column: Recent Txns + Ledger Balances ─── -->
<div class="dash-two-col">

  <section class="section">
    <header class="section-header">
      <h2 class="section-title">近期傳票</h2>
      <a href="#/home/journal-entry" class="section-action">查看全部 →</a>
    </header>
    <div class="table-wrap">
      <table class="data-table" aria-label="近期傳票">
        <thead>
          <tr>
            <th>日期</th>
            <th>摘要</th>
            <th class="num">金額</th>
            <th>幣別</th>
            <th>狀態</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr><td colspan="5" class="table-empty">載入中…</td></tr>
          {:else if txns.length === 0}
            <tr><td colspan="5" class="table-empty">無近期傳票</td></tr>
          {:else}
            {#each txns as txn (txn.txn_id)}
              <tr>
                <td class="mono" style="font-size:11px">{txn.txn_date}</td>
                <td>{txn.description}</td>
                <td class="mono num">{fmtAmt(txn.total_amount)}</td>
                <td style="font-size:11px;color:#5c6278">{txn.currency}</td>
                <td><span class="{txnBadge(txn.status)}">{STATUS_LABELS[txn.status] ?? txn.status}</span></td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

  <section class="section">
    <header class="section-header">
      <h2 class="section-title">帳戶餘額</h2>
    </header>
    <div class="table-wrap">
      <table class="data-table" aria-label="帳戶餘額">
        <tbody>
          {#if isLoading}
            <tr><td colspan="3" class="table-empty">載入中…</td></tr>
          {:else if ledgers.length === 0}
            <tr><td colspan="3" class="table-empty">無帳戶資料</td></tr>
          {:else}
            {#each ledgers as l (l.ledger_id)}
              <tr>
                <td><span class="dash-ledger-tag {LEDGER_TAG_CLASS[l.type] ?? ''}">{LEDGER_TYPE_LABELS[l.type] ?? l.type}</span></td>
                <td>
                  <div style="font-size:12px">{l.name}</div>
                  <div style="font-size:10px;color:#5c6278">{l.institution}</div>
                </td>
                <td class="mono num" class:dash-bal--neg={parseFloat(l.balance) < 0}>{fmtAmt(l.balance)}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </section>

</div>

<!-- ── Monthly Trend ─────────────────────── -->
<section class="section">
  <header class="section-header">
    <h2 class="section-title">月度損益趨勢</h2>
    <span class="content-date">近 12 個月</span>
  </header>
  {#if !isLoading && trend.length === 0}
    <div class="table-empty" style="padding:24px;text-align:center">尚無趨勢資料</div>
  {:else}
    <div class="table-wrap">
      <table class="data-table" aria-label="月度損益趨勢">
        <thead>
          <tr>
            <th>月份</th>
            <th class="num">收入</th>
            <th class="dash-bar-cell"></th>
            <th class="num">支出</th>
            <th class="dash-bar-cell"></th>
            <th class="num">淨額</th>
          </tr>
        </thead>
        <tbody>
          {#if isLoading}
            <tr><td colspan="6" class="table-empty">載入中…</td></tr>
          {:else}
            {#each trend as row (row.month)}
              {@const net = parseFloat(row.net)}
              <tr>
                <td class="mono" style="font-size:12px">{row.month}</td>
                <td class="mono num" style="font-size:12px">{fmtAmt(row.income)}</td>
                <td class="dash-bar-cell">
                  <div class="dash-bar-track"><div class="dash-bar--income" style="width:{barW(row.income)}"></div></div>
                </td>
                <td class="mono num" style="font-size:12px">{fmtAmt(row.expense)}</td>
                <td class="dash-bar-cell">
                  <div class="dash-bar-track"><div class="dash-bar--expense" style="width:{barW(row.expense)}"></div></div>
                </td>
                <td class="mono num" style="font-size:12px"
                  class:dash-net--pos={net > 0.001}
                  class:dash-net--neg={net < -0.001}
                >{fmtNet(net)}</td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  {/if}
</section>
