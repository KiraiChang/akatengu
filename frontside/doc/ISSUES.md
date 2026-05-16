# ISSUES

開發過程中遇到的問題、技術障礙、已知 Bug。

格式：
- **狀態**：🔴 未解決 / 🟡 暫緩 / 🟢 已解決
- **日期**：發現日期
- **描述**：問題內容與根本原因
- **解法**：已採取或計畫中的解法

---

## 🟢 2026-05-16｜npm run lint 執行失敗（eslint: command not found）

**現象**：執行 `npm run lint` 時回傳 `sh: eslint: command not found`，即使在正確的 frontside 目錄下也無法執行。

**根本原因**：Node.js / npm 執行環境異常，`node_modules/.bin/eslint` 未被正確解析（可能為 PATH 問題或 nvm/fnm 環境未載入）。

**解法**：由使用者修正執行環境後恢復正常。

**後續**：遇到相同情況時，向使用者說明狀況並請求支援，不嘗試自行繞過（見 CLAUDE.md 行為規範）。
