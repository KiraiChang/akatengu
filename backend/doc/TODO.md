# TODO

## 現金流量表（Cash Flow Statement）

### 系統自動產生分錄補上 cash_flow_category

目前系統內部自動組裝的 `TransactionCreatedPayload` 未設定 `cash_flow_category`，
導致以下流程產生的分錄不會出現在現金流量表的調整項中。
後續應逐一評估各事件類型的適當分類，並補上對應的 category。

- [ ] **分期付款（Installment）**
  - `applyInstallmentCreated`：分期購入資產時，資產那一側的分錄通常屬於 INVESTING；應付帳款那一側不分類。
  - `applyInstallmentPeriodPaid`：每期還款時，應付帳款 debit 那一側通常屬於 FINANCING（本金還款）。

- [ ] **投資買賣（Investment）**
  - `applyInvestmentBought`：買入投資時，投資科目那一側屬於 INVESTING。
  - `applyInvestmentSold`：賣出投資時，投資科目那一側屬於 INVESTING。
  - `applyDevidendReceived`：股利收入依 IFRS 可歸入 OPERATING 或 INVESTING，需依商業邏輯決定預設值。

- [ ] **年度結帳（Period Annual Close / Reopen）**
  - `applyPeriodAnnualClosed` / `applyPeriodAnnualReopened`：結帳分錄為科目間內部軋轉，不屬於現金流量表，應保持 `cash_flow_category = NULL`。

---

## 直接法現金流量表（Direct Method）

### 混合分類交易的精準拆分

目前 `queryDirectOperatingCash` 對含 INCOME/EXPENSE 分錄的交易，
會將整筆 CASH 移動歸入 Operating，不按比例拆分。

- [ ] **評估實際需求**：個人財務場景中是否真的會出現混合分類交易（如：同一筆 DR CASH 同時對應 INCOME + INVESTING）。
- [ ] **若需精準拆分**：可在 `journal_entries` 加入「Operating 現金比例」欄位，或要求使用者在 create entry 時拆成兩筆交易。

### 直接法明細行（Line Items）

目前直接法只提供 CashReceived / CashPaid 兩個加總數字。
前端若需要「收到薪資 3000 / 支付房租 1500」的明細清單，需另行設計。

- [ ] 新增 `queryDirectOperatingCashItems`，依科目彙總 CASH 移動明細（帳戶名稱 + 金額），格式類似 `CashFlowSection.Items`。

---

### 其他應評估的科目類型

- [ ] **折舊費用（Depreciation）**
  - 折舊為非現金費用，不涉及實際現金流出。在間接法現金流量表中，折舊通常列為「加回」調整項（OPERATING 正調整）。
  - 若日後新增折舊分錄流程，對應的「累計折舊」科目那一側應標記 `OPERATING`。

- [ ] **攤銷（Amortization）**
  - 邏輯同折舊，無形資產攤銷亦屬 OPERATING 非現金調整項。

- [ ] **應收 / 應付帳款變動**
  - 應收帳款增加 → OPERATING 負調整；減少（收回） → OPERATING 正調整。
  - 需確認前端預設值規則是否與科目 `accounts.cash_flow_category` 對應正確。
