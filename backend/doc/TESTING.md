# Testing Strategy

本文件定義專案的測試規範與結構，包含測試層次劃分、BDD 撰寫規則與執行方式。

---

## 測試層次對照

| 層次 | 路徑 | 測試風格 |
|------|------|----------|
| Handler / Middleware | `internal/handler/**` | BDD（Ginkgo v2 + Gomega） |
| Services / Repos / Pkg | `internal/services/**`、`internal/repos/**`、`internal/pkg/**` | 標準 Go testing（table-driven） |

Handler/Middleware 情境分支多（token 有效性 × merchant 存在 × role 符合），GIVEN/WHEN/THEN 結構可清楚呈現情境組合。Services/Repos/Pkg 情境相對線性，table-driven 已足夠，不引入框架。

---

## Handler / Middleware：BDD 規則

### 檔案結構

每個需要測試的 package 建立三個檔案：

```
xxx_suite_test.go      ← Ginkgo bootstrap（每個 package 只需一個）
xxx_scenarios_test.go  ← 純資料：scenario struct + 測試情境清單
xxx_test.go            ← 純程式碼：stubs、helpers、spec runner
```

### Suite 檔（`xxx_suite_test.go`）

```go
package middleware

import (
    "testing"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestXxxSuite(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Xxx Suite")
}
```

### Scenarios 檔（`xxx_scenarios_test.go`）

純資料，不含任何執行邏輯。GIVEN/WHEN/THEN 以欄位形式記錄，與 spec runner 完全解耦。

```go
package middleware

import (
    "context"
    "net/http"
    // 僅 import 資料層所需套件
)

type xxxScenario struct {
    given    string
    when     string
    then     string
    ctx      context.Context  // 輸入：前提條件
    stubErr  error            // 輸入：依賴行為
    wantCode int              // 預期：HTTP 狀態碼
    wantBody string           // 預期：回應訊息（空字串表示不驗證）
}

var xxxScenarios = []xxxScenario{
    {
        given:    "某前提條件",
        when:     "收到任意 HTTP 請求",
        then:     "回傳 403 且訊息包含 forbidden",
        ctx:      someCtx(),   // helper 定義於 xxx_test.go，package 內可見
        wantCode: http.StatusForbidden,
        wantBody: "forbidden",
    },
}
```

> `someCtx()` 等 helper 定義在 `xxx_test.go`，同 package 內可跨檔案呼叫，無需額外 import。

### Spec 檔（`xxx_test.go`）

```go
package middleware

import (
    "context"
    "fmt"
    "net/http"
    "net/http/httptest"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    // ...
)

// ─── Stubs ───────────────────────────────────────────────────────────────────

type stubXxxRepo struct{ err error }

func (s *stubXxxRepo) SomeMethod(...) (...) { return ..., s.err }
// 其餘方法回傳零值

var _ query.XxxRepo = (*stubXxxRepo)(nil)  // 編譯期介面確認

// ─── Helpers ─────────────────────────────────────────────────────────────────

func someCtx() context.Context { ... }
func requestWithContext(ctx context.Context) *http.Request { ... }

// ─── Spec runner ─────────────────────────────────────────────────────────────

var _ = Describe("XxxMiddleware", func() {
    for _, s := range xxxScenarios {
        s := s
        label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)
        It(label, func() {
            stub := &stubXxxRepo{err: s.stubErr}
            rec  := httptest.NewRecorder()

            XxxMiddleware(stub)(okHandler("ok")).ServeHTTP(rec, requestWithContext(s.ctx))

            Expect(rec.Code).To(Equal(s.wantCode))
            if s.wantBody != "" {
                Expect(rec.Body.String()).To(ContainSubstring(s.wantBody))
            }
        })
    }
})
```

### Stub 原則

- Stub struct 透過介面注入，僅實作測試所需方法，其餘方法回傳零值
- `enumx.Enum[T]` 的零值為 `T{}`，**不可用 `""`**
- 一律加編譯期斷言：`var _ query.XxxRepo = (*stubXxxRepo)(nil)`

---

## Services / Repos / Pkg：標準 Go 測試

使用標準 `testing` 套件 + table-driven 測試。不引入 Ginkgo。

```go
func TestSomething(t *testing.T) {
    cases := []struct {
        name    string
        input   SomeType
        want    SomeResult
        wantErr bool
    }{
        { name: "正常情況", input: ..., want: ... },
        { name: "邊界值",   input: ..., want: ... },
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := Something(tc.input)
            if (err != nil) != tc.wantErr {
                t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
            }
            if got != tc.want {
                t.Errorf("want %v, got %v", tc.want, got)
            }
        })
    }
}
```

---

## 執行指令

```bash
# 全套測試
go test ./...

# 指定層
go test ./internal/handler/...
go test ./internal/services/...

# 指定 Ginkgo suite（含詳細輸出）
go test ./internal/handler/middleware/... -run TestMerchantSuite -v

# 全部 handler 層 BDD 輸出
go test ./internal/handler/... -v
```
