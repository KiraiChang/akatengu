<script lang="ts">
</script>

<div class="content-header">
  <h1 class="content-title">儀表板</h1>
  <span class="content-date">2026 / 04 / 23</span>
</div>

<div class="stat-grid">
  <div class="stat-card">
    <div class="stat-label">本月收入</div>
    <div class="stat-value income">1,240,800</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend up">▲ 12.4% 較上月</div>
  </div>
  <div class="stat-card">
    <div class="stat-label">本月支出</div>
    <div class="stat-value expense">873,250</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend down">▼ 3.1% 較上月</div>
  </div>
  <div class="stat-card accent">
    <div class="stat-label">帳戶餘額</div>
    <div class="stat-value">12,548,320</div>
    <div class="stat-unit">TWD</div>
    <div class="stat-trend neutral">上月淨增 +367,550</div>
  </div>
  <div class="stat-card">
    <div class="stat-label">待審傳票</div>
    <div class="stat-value">7</div>
    <div class="stat-unit">筆</div>
    <div class="stat-trend warn">需要處理</div>
  </div>
</div>

<section class="section">
  <header class="section-header">
    <h2 class="section-title">近期交易</h2>
    <button class="section-action">查看全部 →</button>
  </header>
  <div class="table-wrap">
    <table class="data-table" aria-label="近期交易明細">
      <thead>
        <tr>
          <th>日期</th>
          <th>傳票號</th>
          <th>摘要</th>
          <th>科目</th>
          <th class="num">借方</th>
          <th class="num">貸方</th>
          <th>狀態</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>2026-04-22</td>
          <td class="mono">JV-202604-0041</td>
          <td>辦公室租金</td>
          <td>租金費用</td>
          <td class="num">85,000</td>
          <td class="num">—</td>
          <td><span class="badge approved">已核准</span></td>
        </tr>
        <tr>
          <td>2026-04-22</td>
          <td class="mono">JV-202604-0040</td>
          <td>銷貨收入</td>
          <td>應收帳款</td>
          <td class="num">320,000</td>
          <td class="num">—</td>
          <td><span class="badge approved">已核准</span></td>
        </tr>
        <tr>
          <td>2026-04-21</td>
          <td class="mono">JV-202604-0039</td>
          <td>員工薪資</td>
          <td>薪資費用</td>
          <td class="num">—</td>
          <td class="num">480,000</td>
          <td><span class="badge pending">待審核</span></td>
        </tr>
        <tr>
          <td>2026-04-20</td>
          <td class="mono">JV-202604-0038</td>
          <td>設備採購</td>
          <td>固定資產</td>
          <td class="num">156,000</td>
          <td class="num">—</td>
          <td><span class="badge pending">待審核</span></td>
        </tr>
        <tr>
          <td>2026-04-19</td>
          <td class="mono">JV-202604-0037</td>
          <td>水電費</td>
          <td>水電費用</td>
          <td class="num">12,840</td>
          <td class="num">—</td>
          <td><span class="badge approved">已核准</span></td>
        </tr>
      </tbody>
    </table>
  </div>
</section>
