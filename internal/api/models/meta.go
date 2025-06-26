package models

type Links struct {
	First string  `json:"first"`
	Last  string  `json:"last"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
}

type PaginationMeta struct {
	PageSize    int  `json:"page-size"`
	CurrentPage int  `json:"current-page"`
	NextPage    *int `json:"next-page"`
	PrevPage    *int `json:"prev-page"`
	TotalPages  int  `json:"total-pages"`
	TotalCount  int  `json:"total-count"`
}

type Meta struct {
	Pagination PaginationMeta `json:"pagination"`
}
