package dto

type PaginationParams struct {
	Page     int `json:"page"`      // 第幾頁，從 1 開始
	PageSize int `json:"page_size"` // 每頁幾筆資料
}

func (p *PaginationParams) SetDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

type PaginatedResponse[T any] struct {
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	Items      []T   `json:"items"`
}

func (p *PaginatedResponse[T]) SetTotalCount(total int64) {
	p.TotalCount = total
	p.TotalPages = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
}
