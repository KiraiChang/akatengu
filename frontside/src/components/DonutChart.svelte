<script lang="ts">
  import { arc, pie } from 'd3';
  import type { DefaultArcObject, PieArcDatum } from 'd3';

  interface Item { label: string; value: number; id: string; has_child: boolean; }
  interface Props {
    title:    string;
    items:    Item[];
    colors:   string[];
    onclick?: (id: string) => void;
  }

  const { title, items, colors, onclick }: Props = $props();

  const INNER_R = 52;
  const OUTER_R = 85;

  const arcFn = arc<DefaultArcObject>()
    .innerRadius(INNER_R)
    .outerRadius(OUTER_R)
    .cornerRadius(2)
    .padAngle(0.02);

  const pieFn = pie<number>().sort(null);

  const total = $derived(items.reduce((s: number, i: Item) => s + i.value, 0));
  const arcs  = $derived(pieFn(items.map((i: Item) => i.value)));

  let hovered = $state<number | null>(null);

  function toPath(d: PieArcDatum<number>): string {
    return arcFn({
      innerRadius: INNER_R,
      outerRadius: OUTER_R,
      startAngle:  d.startAngle,
      endAngle:    d.endAngle,
      padAngle:    d.padAngle,
    }) ?? '';
  }

  function colorAt(i: number): string {
    return colors[i % colors.length] ?? '#ccc';
  }

  function ariaLabel(i: number): string {
    const item = items[i];
    return item ? `${item.label}: ${item.value.toLocaleString()}` : '';
  }
</script>

<div class="chart-wrap">
  {#if title}
    <p class="chart-title">{title}</p>
  {/if}
  <div class="chart-svg-container">
    <svg viewBox="0 0 200 200" width="100%" role="img" aria-label={title}>
      {#if items.length === 0 || total === 0}
        <text x="100" y="105" text-anchor="middle" class="donut-empty">無資料</text>
      {:else}
        <g transform="translate(100, 100)">
          {#each arcs as d, i (i)}
            <path
              role="button"
              tabindex="0"
              aria-label={ariaLabel(i)}
              d={toPath(d)}
              fill={colorAt(i)}
              opacity={hovered === null || hovered === i ? 1 : 0.4}
              style="cursor:pointer;transition:opacity 0.15s;outline:none;"
              onfocus={() => { hovered = i; }}
              onblur={() => { hovered = null; }}
              onmouseenter={() => { hovered = i; }}
              onmouseleave={() => { hovered = null; }}
              onclick={() => { onclick?.(items[i]?.id ?? ''); }}
              onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') onclick?.(items[i]?.id ?? ''); }}
            />
          {/each}
          <text class="donut-center-label" x="0" y="-8" text-anchor="middle">合計</text>
          <text class="donut-center-value" x="0" y="14" text-anchor="middle">
            {total.toLocaleString()}
          </text>
        </g>
      {/if}
    </svg>
    {#if hovered !== null && items[hovered]}
      <div class="chart-tooltip">
        <span class="chart-tooltip-dot" style={`background:${colorAt(hovered)}`}></span>
        <span class="chart-tooltip-name">{items[hovered].label}</span>
        <span class="chart-tooltip-val">{items[hovered].value.toLocaleString()}</span>
        {#if total > 0}
          <span class="chart-tooltip-pct">({((items[hovered].value / total) * 100).toFixed(1)}%)</span>
        {/if}
      </div>
    {/if}
  </div>
</div>
