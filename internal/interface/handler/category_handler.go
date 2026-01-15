package handler

import (
	"strconv"
	"teckbook-compass-backend/internal/usecase"
	"teckbook-compass-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// CategoryHandler カテゴリハンドラ
type CategoryHandler struct {
	categoryUsecase *usecase.CategoryUsecase
}

// NewCategoryHandler カテゴリハンドラのコンストラクタ
func NewCategoryHandler(categoryUsecase *usecase.CategoryUsecase) *CategoryHandler {
	return &CategoryHandler{
		categoryUsecase: categoryUsecase,
	}
}

// GetCategoriesWithBooks カテゴリ別書籍取得API
// @Summary カテゴリ別書籍取得
// @Description カテゴリとそのカテゴリに属する書籍のトップ3を取得
// @Tags categories
// @Accept json
// @Produce json
// @Param max_categories query int false "返却するカテゴリ数の上限（指定なしは全件）"
// @Param limit query int false "各カテゴリ内の書籍数の上限（指定なしは全件）"
// @Success 200 {object} dto.CategoryWithBooksResponse
// @Failure 500 {object} map[string]string
// @Router /categories/with-books [get]
func (h *CategoryHandler) GetCategoriesWithBooks(c *gin.Context) {
	// リクエストパラメータを取得
	maxCategories := 0 // デフォルト: 全件
	if maxCategoriesStr := c.Query("max_categories"); maxCategoriesStr != "" {
		if val, err := strconv.Atoi(maxCategoriesStr); err == nil && val > 0 {
			maxCategories = val
		}
	}

	limit := 0 // デフォルト: 全件
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	params := usecase.GetCategoriesWithBooksParams{
		MaxCategories: maxCategories,
		Limit:         limit,
	}

	result, err := h.categoryUsecase.GetCategoriesWithBooks(c.Request.Context(), params)
	if err != nil {
		response.Error(c, 500, "カテゴリの取得に失敗しました")
		return
	}

	response.Success(c, result)
}
