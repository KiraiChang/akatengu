<script lang="ts">
  interface BarItem { label: string; value: number; id: string; has_child: boolean; }
  interface BreadcrumbItem { label: string; isCurrent: boolean; onclick: () => void; }
  interface Props {
    incomeItems:        BarItem[];
    expenseItems:       BarItem[];
    netIncome:          number;
    incomeBreadcrumbs:  BreadcrumbItem[];
    expenseBreadcrumbs: BreadcrumbItem[];
    onIncomeClick?:     (id: string) => void;
    onExpenseClick?:    (id: string) => void;
  }

  const { incomeItems, expenseItems, netIncome,
          incomeBreadcrumbs, expenseBreadcrumbs,
          onIncomeClick, onExpenseClick }: Props = $props();

  const maxVal = $derived(Math.max(
    ...(incomeItems.length  ? incomeItems.map((i: BarItem) => i.value)  : [0]),
    ...(expenseItems.length ? expenseItems.map((i: BarItem) => i.value) : [0]),
    1,
  ));

  const incomeSubtotal  = $derived(incomeItems.reduce((s: number, i: BarItem) => s + i.value, 0));
  const expenseSubtotal = $derived(expenseItems.reduce((s: number, i: BarItem) => s + i.value, 0));

  function barPct(v: number): string {
    return `${Math.min(100, (v / maxVal) * 100).toFixed(1)}%`;
  }
</script>

<div class="hbar-chart">
  {#if incomeItems.length > 0 || incomeBreadcrumbs.length > 1}
    <div class="hbar-section">
      <div class="hbar-section-header">
        <nav class="drill-breadcrumb" aria-label="收入類別導覽">
          {#each incomeBreadcrumbs as crumb, i (i)}
            {#if i > 0}<span class="drill-breadcrumb-sep" aria-hidden="true">›</span>{/if}
            {#if crumb.isCurrent}
              <span class="drill-breadcrumb-item--current">{crumb.label}</span>
            {:else}
              <button class="drill-breadcrumb-item--link" onclick={crumb.onclick}>{crumb.label}</button>
            {/if}
          {/each}
        </nav>
      </div>
      {#each incomeItems as item (item.label)}
        {#if item.has_child}
          <button
            class="hbar-row hbar-row--clickable"
            onclick={() => onIncomeClick?.(item.id)}
          >
            <span class="hbar-label">{item.label}</span>
            <div class="hbar-track">
              <div class="hbar-bar hbar-bar--income" style={`width:${barPct(item.value)}`}></div>
            </div>
            <span class="hbar-value">{item.value.toLocaleString()}</span>
          </button>
        {:else}
          <div class="hbar-row">
            <span class="hbar-label">{item.label}</span>
            <div class="hbar-track">
              <div class="hbar-bar hbar-bar--income" style={`width:${barPct(item.value)}`}></div>
            </div>
            <span class="hbar-value">{item.value.toLocaleString()}</span>
          </div>
        {/if}
      {/each}
      <div class="hbar-row hbar-subtotal">
        <span class="hbar-label">收入合計</span>
        <div class="hbar-track"></div>
        <span class="hbar-value">{incomeSubtotal.toLocaleString()}</span>
      </div>
    </div>
  {/if}

  {#if expenseItems.length > 0 || expenseBreadcrumbs.length > 1}
    <div class="hbar-section">
      <div class="hbar-section-header">
        <nav class="drill-breadcrumb" aria-label="費用類別導覽">
          {#each expenseBreadcrumbs as crumb, i (i)}
            {#if i > 0}<span class="drill-breadcrumb-sep" aria-hidden="true">›</span>{/if}
            {#if crumb.isCurrent}
              <span class="drill-breadcrumb-item--current">{crumb.label}</span>
            {:else}
              <button class="drill-breadcrumb-item--link" onclick={crumb.onclick}>{crumb.label}</button>
            {/if}
          {/each}
        </nav>
      </div>
      {#each expenseItems as item (item.label)}
        {#if item.has_child}
          <button
            class="hbar-row hbar-row--clickable"
            onclick={() => onExpenseClick?.(item.id)}
          >
            <span class="hbar-label">{item.label}</span>
            <div class="hbar-track">
              <div class="hbar-bar hbar-bar--expense" style={`width:${barPct(item.value)}`}></div>
            </div>
            <span class="hbar-value">{item.value.toLocaleString()}</span>
          </button>
        {:else}
          <div class="hbar-row">
            <span class="hbar-label">{item.label}</span>
            <div class="hbar-track">
              <div class="hbar-bar hbar-bar--expense" style={`width:${barPct(item.value)}`}></div>
            </div>
            <span class="hbar-value">{item.value.toLocaleString()}</span>
          </div>
        {/if}
      {/each}
      <div class="hbar-row hbar-subtotal">
        <span class="hbar-label">費用合計</span>
        <div class="hbar-track"></div>
        <span class="hbar-value">{expenseSubtotal.toLocaleString()}</span>
      </div>
    </div>
  {/if}

  <div class="hbar-net" class:hbar-net--pos={netIncome >= 0} class:hbar-net--neg={netIncome < 0}>
    <span>本期淨利</span>
    <span class="mono">{netIncome.toLocaleString()}</span>
  </div>
</div>
