<script lang="ts">
  import { getExchangeRates } from '../api/audit';
  import { createExchangeRate } from '../api/exchangeRate';
  import type { ExchangeRate } from '../types/audit';

  const today = new Date().toISOString().slice(0, 10);

  const SOURCE_LABELS: Record<string, string> = { MANUAL: '手動', IMPORT: '匯入' };

  let rates           = $state<ExchangeRate[]>([]);
  let isLoading       = $state(true);
  let error           = $state('');
  let filterCurrency  = $state('');

  let isSubmitting    = $state(false);
  let submitMsg       = $state('');
  let formCurrency    = $state('');
  let formDate        = $state(today);
  let formRate        = $state('');

  const canSubmit = $derived(
    formCurrency.trim().length > 0 && formDate !== '' && formRate.trim().length > 0 && !isNaN(parseFloat(formRate)),
  );

  async function search(): Promise<void> {
    isLoading = true;
    error = '';
    try {
      rates = await getExchangeRates(filterCurrency.trim() || undefined);
    } catch (err) {
      error = err instanceof Error ? err.message : '查詢失敗';
    } finally {
      isLoading = false;
    }
  }

  async function submit(): Promise<void> {
    if (!canSubmit || isSubmitting) return;
    isSubmitting = true;
    submitMsg    = '';
    try {
      await createExchangeRate({
        currency: formCurrency.trim().toUpperCase(),
        date:     formDate,
        rate_twd: formRate.trim(),
      });
      submitMsg    = '已儲存';
      formCurrency = '';
      formRate     = '';
      await search();
    } catch (err) {
      submitMsg = err instanceof Error ? err.message : '儲存失敗';
    } finally {
      isSubmitting = false;
    }
  }

  function sourceBadgeClass(source: string): string {
    return source === 'MANUAL' ? 'exrate-source--manual' : 'exrate-source--import';
  }

  $effect(() => { void search(); });
</script>

<div class="content-header">
  <h1 class="content-title">匯率管理</h1>
</div>

{#if error}
  <p class="query-error" role="alert">{error}</p>
{/if}

<!-- ── Filter ── -->
<div class="query-bar">
  <div class="exrate-field">
    <label class="form-label" for="filter-currency">幣別篩選</label>
    <input
      id="filter-currency"
      class="form-input"
      bind:value={filterCurrency}
      placeholder="USD"
      oninput={(e) => { filterCurrency = e.currentTarget.value.toUpperCase(); }}
      onkeydown={(e) => e.key === 'Enter' && search()}
    />
  </div>
  <button class="btn-ghost" onclick={search}>查詢</button>
  {#if filterCurrency}
    <button class="btn-ghost" onclick={() => { filterCurrency = ''; void search(); }}>清除</button>
  {/if}
</div>

<!-- ── Rates table ── -->
<section class="section">
  <div class="table-wrap">
    <table class="data-table" aria-label="匯率列表">
      <thead>
        <tr>
          <th>幣別</th>
          <th>日期</th>
          <th class="num">匯率 (TWD)</th>
          <th>來源</th>
        </tr>
      </thead>
      <tbody>
        {#if isLoading}
          <tr><td colspan="4" class="table-empty">載入中…</td></tr>
        {:else if rates.length === 0}
          <tr><td colspan="4" class="table-empty">無匯率資料</td></tr>
        {:else}
          {#each rates as rate (rate.rate_id)}
            <tr>
              <td class="mono">{rate.currency}</td>
              <td class="mono" style="font-size:11px">{rate.rate_date}</td>
              <td class="mono num">{rate.rate_twd}</td>
              <td><span class={sourceBadgeClass(rate.source)}>{SOURCE_LABELS[rate.source] ?? rate.source}</span></td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  </div>
</section>

<!-- ── Input form ── -->
<section class="section">
  <header class="section-header">
    <h2 class="section-title">新增 / 更新匯率</h2>
  </header>
  <div class="query-bar">
    <div class="exrate-field">
      <label class="form-label" for="form-currency">幣別</label>
      <input
        id="form-currency"
        class="form-input"
        bind:value={formCurrency}
        placeholder="USD"
        oninput={(e) => { formCurrency = e.currentTarget.value.toUpperCase(); }}
      />
    </div>
    <div class="exrate-field">
      <label class="form-label" for="form-date">日期</label>
      <input id="form-date" class="form-input" type="date" bind:value={formDate} />
    </div>
    <div class="exrate-field">
      <label class="form-label" for="form-rate">匯率 (TWD)</label>
      <input id="form-rate" class="form-input" bind:value={formRate} placeholder="32.50" />
    </div>
    <button
      class="btn-success"
      onclick={submit}
      disabled={!canSubmit || isSubmitting}
    >{isSubmitting ? '儲存中…' : '儲存'}</button>
  </div>
  {#if submitMsg}
    <p class:exrate-msg--ok={!submitMsg.includes('失敗') && !submitMsg.includes('error')}
       class:exrate-msg--err={submitMsg.includes('失敗') || submitMsg.includes('error')}>{submitMsg}</p>
  {/if}
</section>
