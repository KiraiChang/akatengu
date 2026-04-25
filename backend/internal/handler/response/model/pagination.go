package model

type PaginationParams struct {
	Page     int64 `json:"page"`      // 第幾頁，從 1 開始
	PageSize int64 `json:"page_size"` // 每頁幾筆資料
}

func (p *PaginationParams) SetDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

type PaginatedMeta struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"per_page"`
	TotalCount int64 `json:"total_count"`
	TotalPages int64 `json:"total_pages"`
}

type PaginatedResponse[T any] struct {
	Data []T           `json:"data"`
	Meta PaginatedMeta `json:"meta"`
}

func (p *PaginatedResponse[T]) SetTotalCount(total int64) {
	p.Meta.TotalCount = total
	p.Meta.TotalPages = (total + p.Meta.PageSize - 1) / p.Meta.PageSize
}
