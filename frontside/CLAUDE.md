# AKATENGU

使用 Svelte 開發的會計軟體前端。

## 技術棧

- 框架：svelte 5.55.4
- 語言：TypeScript 6.0.2（strict mode）
- 建置工具：Vite 8.0.9
- 套件管理：npm
- Node.js 版本：24.12.2

## 重要文件

## 常用指令

- `pnpm dev`：啟動開發伺服器（port 5173）
- `pnpm build`：建置專案
- `pnpm test`：執行測試
- `pnpm test:coverage`：執行測試並產出覆蓋率報告
- `pnpm lint`：執行 ESLint 檢查

## 開發規範

### 命名規範

- 變數與函式：camelCase（`getUserName`）
- 元件與型別：PascalCase（`UserProfile`、`UserType`）
- 常數：UPPER_SNAKE_CASE（`API_BASE_URL`）
- 元件檔案名稱：PascalCase（`UserProfile.vue`）
- 非元件檔案名稱：camelCase（`useUser.ts`、`userService.ts`）
- 布林值：`is`/`has`/`should` 前綴（`isActive`）

### TypeScript

- 禁止使用 `any`，必要時使用 `unknown` 搭配型別守衛
- 使用 `interface` 定義物件型別，使用 `type` 定義聯合型別或工具型別
- 所有匯出的函式都必須標註回傳型別

### svelte 元件開發

- 統一使用 `<script setup lang="ts">` 語法
- Props 使用 `defineProps<T>()` 搭配 TypeScript 型別定義
- Emit 使用 `defineEmits<T>()` 搭配 TypeScript 型別定義
- 共用邏輯抽成 composable（`composables/useXxx.ts`）
- 超過 200 行的元件應拆分為子元件

### 錯誤處理

- API 呼叫使用 try-catch 包裹
- 使用統一的 `AppError` 類別處理錯誤
- 全域錯誤使用 `app.config.errorHandler` 捕捉

## 回應規範

- 一律使用繁體中文回應
- commit message 使用 Conventional Commits 格式，描述使用繁體中文
- 程式碼中的變數名稱與註解使用英文

## 常見任務

### 建立新頁面

1. 在 `src/views/` 底下建立對應的 `.svelte` 檔案
2. 在 `src/router/` 中加入路由設定
3. 如果有共用邏輯，在 `src/composables/` 建立對應的 composable
4. 建立對應的測試檔案

### 建立新 API 串接

1. 在 `src/services/` 底下建立對應的 service 檔案
2. 在 `src/types/` 定義請求與回應的 TypeScript 型別
3. 使用 `apiClient` 封裝，不要直接使用 fetch 或 axios
4. 建立對應的測試檔案