# Architecture Decision Records (ADR)

---

## ADR-001：現金流量表分類來源改為 journal_entries

**日期：** 2026-05-16
**狀態：** 已採用

### 背景

原始設計將 `cash_flow_category`（OPERATING / INVESTING / FINANCING）掛在 `accounts` 資料表，
導致同一科目下的所有分錄必須歸入同一個現金流量類別，無法針對個別交易彈性分類。

例如：股利支付在 IFRS 可歸入 OPERATING 或 FINANCING；
應收帳款依交易性質可能是 OPERATING 或 INVESTING。

### 決策

將 OPERATING / INVESTING / FINANCING 分類移至 `journal_entries.cash_flow_category`（nullable）。
每筆分錄可獨立設定，前端依所選科目的 `accounts.cash_flow_category` 作為預填預設值，使用者可自行修改。

`accounts.cash_flow_category` 保留以下用途：
- `CASH`：識別現金及約當現金科目，用於計算期初 / 期末現金餘額（`queryCashBefore` / `queryCashUpTo`）。
- `OPERATING / INVESTING / FINANCING`：提供前端預設值參考，不再參與計算。

### 取捨

| 面向 | 舊設計（account-based） | 新設計（entry-based） |
|------|------------------------|----------------------|
| 彈性 | 低，同科目只能一個分類 | 高，每筆分錄獨立設定 |
| 資料完整性 | account 設定後即生效 | 需前端 / pipeline 主動填入 |
| 系統自動分錄 | 靠科目自動帶入 | 需逐案評估並補上（見 TODO.md） |
| 查詢複雜度 | 較低 | 略高，需 GROUP BY (account_id, cash_flow_category) |

### 不採用方案

- **保留 account-based**：無法支援同一科目多分類需求，排除。
- **同時維護兩個來源**：增加維護成本與資料一致性風險，排除。
