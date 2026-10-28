package models

// Page holds a slice of results plus pagination metadata.
type Page[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// Pagination describes the slice of a result set that was returned.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PageParams is the normalized pagination input for a query.
type PageParams struct {
	Page    int
	PerPage int
	Offset  int
}

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

// NewPageParams normalizes page/per_page values and computes the SQL offset.
func NewPageParams(page, perPage int) PageParams {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return PageParams{Page: page, PerPage: perPage, Offset: (page - 1) * perPage}
}

// NewPage builds a Page result from items, total count and the request params.
func NewPage[T any](items []T, total int, p PageParams) Page[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := 0
	if p.PerPage > 0 {
		totalPages = (total + p.PerPage - 1) / p.PerPage
	}
	return Page[T]{
		Items: items,
		Pagination: Pagination{
			Page:       p.Page,
			PerPage:    p.PerPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// SortOrder represents an ascending or descending sort direction.
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)
