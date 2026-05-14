package service

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

type PageInput struct {
	Page     int
	PageSize int
}

type PageResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func NormalizePageInput(in PageInput) PageInput {
	if in.Page <= 0 {
		in.Page = defaultPage
	}
	if in.PageSize <= 0 {
		in.PageSize = defaultPageSize
	}
	if in.PageSize > maxPageSize {
		in.PageSize = maxPageSize
	}
	return in
}

func NewPageResult[T any](items []T, total int, page PageInput) PageResult[T] {
	page = NormalizePageInput(page)
	if items == nil {
		items = []T{}
	}
	return PageResult[T]{
		Items:    items,
		Total:    total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}
}

func (p PageInput) Limit() int {
	return NormalizePageInput(p).PageSize
}

func (p PageInput) Offset() int {
	p = NormalizePageInput(p)
	return (p.Page - 1) * p.PageSize
}
