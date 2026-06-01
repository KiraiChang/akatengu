<script lang="ts">
  import type { InterestType } from '../types/installment';

  interface Props {
    installmentCount: string;
    startDate:        string;
    interestType:     InterestType;
    annualRate:       string;
    idPrefix?:        string;
  }

  let {
    installmentCount = $bindable(''),
    startDate        = $bindable(''),
    interestType     = $bindable<InterestType>('FREE'),
    annualRate       = $bindable(''),
    idPrefix         = 'inst',
  }: Props = $props();

  const INTEREST_TYPES: InterestType[] = ['FREE', 'FIXED_RATE'];
  const INTEREST_LABELS: Record<InterestType, string> = {
    FREE:       '免息',
    FIXED_RATE: '固定利率',
  };
</script>

<div class="form-row">
  <div class="form-group">
    <label class="form-label" for="{idPrefix}-count">期數 *</label>
    <input id="{idPrefix}-count" class="form-input" type="number" min="1" step="1" bind:value={installmentCount} placeholder="24" required />
  </div>
  <div class="form-group">
    <label class="form-label" for="{idPrefix}-start">第一期到期日 *</label>
    <input id="{idPrefix}-start" class="form-input" type="date" bind:value={startDate} required />
  </div>
</div>
<div class="form-group">
  <label class="form-label" for="{idPrefix}-interest-type">利息類型 *</label>
  <select id="{idPrefix}-interest-type" class="form-select" bind:value={interestType}>
    {#each INTEREST_TYPES as t}
      <option value={t}>{INTEREST_LABELS[t]}</option>
    {/each}
  </select>
</div>
{#if interestType === 'FIXED_RATE'}
  <div class="form-group">
    <label class="form-label" for="{idPrefix}-annual-rate">年利率（%）*</label>
    <input id="{idPrefix}-annual-rate" class="form-input" type="number" min="0.01" step="any" bind:value={annualRate} placeholder="例：12.5" />
  </div>
{/if}
