<script lang="ts">
  import type { LedgerAccount } from '../types/ledger';

  interface Props {
    ledgers:  LedgerAccount[];
    value:    string;
    onselect: (id: string) => void;
  }

  let { ledgers, value, onselect }: Props = $props();

  let open   = $state(false);
  let rootEl = $state<HTMLDivElement | null>(null);

  const selected = $derived(ledgers.find(l => String(l.ledger_id) === value) ?? null);

  $effect(() => {
    if (!open) return;
    function onOutside(e: MouseEvent): void {
      if (rootEl && !rootEl.contains(e.target as Node)) open = false;
    }
    document.addEventListener('mousedown', onOutside);
    return () => document.removeEventListener('mousedown', onOutside);
  });

  function select(ledger: LedgerAccount): void {
    onselect(String(ledger.ledger_id));
    open = false;
  }

  function clear(e: MouseEvent): void {
    e.stopPropagation();
    onselect('');
    open = false;
  }
</script>

<div class="ldgr-select" bind:this={rootEl}>
  <div
    class="ldgr-trigger"
    class:ldgr-trigger--open={open}
    onclick={() => { open = !open; }}
    role="combobox"
    aria-expanded={open}
    aria-haspopup="listbox"
    aria-controls="ldgr-panel"
    tabindex="0"
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); open = !open; } }}
  >
    {#if selected}
      <span class="ldgr-trigger-value">
        <span class="ldgr-trigger-institution">{selected.institution}</span>
        <span class="ldgr-trigger-name">{selected.name}</span>
      </span>
    {:else}
      <span class="ldgr-trigger-placeholder">— 不指定 —</span>
    {/if}
    <span class="ldgr-trigger-icons">
      {#if value}
        <button class="ldgr-clear" type="button" onclick={clear} tabindex="-1" aria-label="清除">×</button>
      {/if}
      <span class="ldgr-chevron" aria-hidden="true">{open ? '▲' : '▼'}</span>
    </span>
  </div>

  {#if open}
    <div class="ldgr-panel" id="ldgr-panel" role="listbox">
      <button
        class="ldgr-item ldgr-item--none"
        type="button"
        role="option"
        aria-selected={value === ''}
        onclick={() => { onselect(''); open = false; }}
      >— 不指定 —</button>
      {#each ledgers as ledger (ledger.ledger_id)}
        <button
          class="ldgr-item"
          class:ldgr-item--selected={String(ledger.ledger_id) === value}
          type="button"
          role="option"
          aria-selected={String(ledger.ledger_id) === value}
          onclick={() => select(ledger)}
        >
          <span class="ldgr-item-institution">{ledger.institution}</span>
          <span class="ldgr-item-name">{ledger.name}</span>
        </button>
      {/each}
    </div>
  {/if}
</div>