package model

// PaginateWithTotal 資料庫做好分頁所以把total填入即可
func PaginateWithTotal[T any](items []T, pagination PaginationParams, totalCount int64) PaginatedResponse[T] {
	resp := PaginatedResponse[T]{
		Meta: PaginatedMeta{
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
		Data: items,
	}

	resp.SetTotalCount(totalCount)
	return resp
}

// PaginateWithoutTotal 資料庫取的所有資料，由程序分頁
func PaginateWithoutTotal[T any](items []T, pagination PaginationParams) PaginatedResponse[T] {
	resp := PaginatedResponse[T]{
		Meta: PaginatedMeta{
			Page:     pagination.Page,
			PageSize: pagination.PageSize,
		},
		Data: paginateByPage(items, pagination.Page, pagination.PageSize),
	}

	resp.SetTotalCount(int64(len(items)))
	return resp
}

func paginateByPage[T any](items []T, page, pageSize int64) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := (page - 1) * pageSize
	return paginate(items, skip, pageSize)
}

func paginate[T any](items []T, skip, take int64) []T {
	if skip < 0 {
		skip = 0
	}
	if take <= 0 {
		return []T{}
	}

	start := skip
	if start > int64(len(items)) {
		start = int64(len(items))
	}

	end := start + take
	if end > int64(len(items)) {
		end = int64(len(items))
	}

	return items[start:end]
}
