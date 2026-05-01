export interface JournalEntryLine {
  account_id: string;
  debit:      number;
  credit:     number;
  note:       string | null;
}

export interface JournalEntry {
  journal_entry_id: string;
  date:             string;
  description:      string;
  lines:            JournalEntryLine[];
  created_at:       string;
}

export interface CreateJournalEntryRequest {
  date:        string;
  description: string;
  lines:       JournalEntryLine[];
}
