package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:static
var staticFiles embed.FS

// FileSystem 取出 static/ 子目錄作為 http.FileSystem
func FileSystem() http.FileSystem {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// SPAHandler 處理 SPA routing：
// - 靜態資源（有副檔名）直接回傳
// - 其他路徑一律回傳 index.html，讓前端 router 處理
func SPAHandler(fsys http.FileSystem) http.Handler {
	fileServer := http.FileServer(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 嘗試開啟該路徑
		f, err := fsys.Open(path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// 有副檔名但找不到 → 404（避免把 /api/missing.json 導回 SPA）
		if strings.Contains(path, ".") {
			http.NotFound(w, r)
			return
		}

		// 無副檔名的路徑 → 回傳 index.html，交給前端 router
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
