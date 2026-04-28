package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

// DefaultPageSize 默认每页条数
const DefaultPageSize = 20

// MaxPageSize 最大每页条数
const MaxPageSize = 100

// getPageSizeQuery 兼容 pageSize 和 page_size 两种参数名
func getPageSizeQuery(c *gin.Context, defaultValue string) string {
	if v := c.Query("pageSize"); v != "" {
		return v
	}
	if v := c.Query("page_size"); v != "" {
		return v
	}
	return defaultValue
}

// ParsePagination 从 gin context 中解析安全的分页参数
// page: 默认为 1，最小为 1
// pageSize: 默认为 DefaultPageSize，最小为 1，最大为 MaxPageSize
func ParsePagination(c *gin.Context) PaginationParams {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(getPageSizeQuery(c, strconv.Itoa(DefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// ParsePaginationWithSize 自定义默认 pageSize 的分页解析
func ParsePaginationWithSize(c *gin.Context, defaultSize int) PaginationParams {
	if defaultSize < 1 {
		defaultSize = DefaultPageSize
	}
	if defaultSize > MaxPageSize {
		defaultSize = MaxPageSize
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(getPageSizeQuery(c, strconv.Itoa(defaultSize)))
	if err != nil || pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

func ParseIntQuery(c *gin.Context, key string, defaultValue, minValue, maxValue int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(defaultValue)))
	if err != nil {
		value = defaultValue
	}
	if value < minValue {
		value = minValue
	}
	if maxValue > 0 && value > maxValue {
		value = maxValue
	}
	return value
}
