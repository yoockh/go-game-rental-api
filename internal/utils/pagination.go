package utils

import (
	"github.com/labstack/echo/v4"
	myRequest "github.com/yoockh/go-api-utils/pkg-echo/request"
)

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func ParsePagination(c echo.Context) PaginationParams {
	page := myRequest.QueryInt(c, "page", 1)
	limit := myRequest.QueryInt(c, "limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func CreateMeta(params PaginationParams, total int64) PaginationMeta {
	totalPages := (total + int64(params.Limit) - 1) / int64(params.Limit)
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
