package utils

import (
	"akatengu/internal/dto"
)

func PaginateWithTotal[T any](items []T, pagination dto.PaginationParams) dto.PaginatedResponse[T] {
	resp := dto.PaginatedResponse[T]{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Items:    paginateByPage(items, pagination.Page, pagination.PageSize),
	}

	resp.SetTotalCount(int64(len(items)))
	return resp
}

func paginateByPage[T any](items []T, page, pageSize int) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize
	return paginate(items, skip, pageSize)
}

func paginate[T any](items []T, skip, take int) []T {
	if skip < 0 {
		skip = 0
	}
	if take <= 0 {
		return []T{}
	}

	start := skip
	if start > len(items) {
		start = len(items)
	}

	end := start + take
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}
