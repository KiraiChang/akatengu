export interface Row {
  account_id: string;
  name:       string;
  value:      string;
  parent_id:  string | null;
  has_child:  boolean;
  depth:      number;
}

export function visibleRows(rows: Row[], expanded: Set<string>): Row[] {
  return rows.filter(r => {
    if (r.parent_id === null) return true;
    let pid: string | null = r.parent_id;
    while (pid !== null) {
      if (!expanded.has(pid)) return false;
      pid = rows.find(p => p.account_id === pid)?.parent_id ?? null;
    }
    return true;
  });
}

export function fmt(v: string): string {
  const n = parseFloat(v);
  return isNaN(n) ? v : n.toLocaleString();
}
