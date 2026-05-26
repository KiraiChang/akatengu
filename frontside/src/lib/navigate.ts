export function goToAccountAnalysis(accountId: string): void {
  window.location.hash = `#/home/account-analysis?account=${encodeURIComponent(accountId)}`;
}
