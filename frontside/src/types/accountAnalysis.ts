export interface AccountChildBalance {
  account_id:   string;
  parent_id:    string | null;
  name:         string;
  is_summary:   boolean;
  has_child:    boolean;
  debit_total:  string;
  credit_total: string;
}

export interface AccountChildrenBalanceResult {
  account_id: string;
  name:       string;
  children:   AccountChildBalance[];
}

export interface AccountJournalEntryRow {
  entry_id:   number;
  txn_id:     number;
  account_id: string;
  ledger_id:  number | null;
  debit:      string;
  credit:     string;
  note:       string | null;
  txn_date:   string;
  description: string;
}

export interface AccountMonthlyBalance {
  month:        string;
  debit_total:  string;
  credit_total: string;
  has_snapshot: boolean;
}

export interface PagedMeta {
  page:        number;
  page_size:   number;
  total_pages: number;
  total:       number;
}

export interface PagedResponse<T> {
  data: T[];
  meta: PagedMeta;
}
