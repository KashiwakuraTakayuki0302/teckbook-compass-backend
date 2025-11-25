package handler

import (
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
// @Success 200 {object} dto.CategoryWithBooksResponse
// @Failure 500 {object} map[string]string
// @Router /categories/with-books [get]
func (h *CategoryHandler) GetCategoriesWithBooks(c *gin.Context) {
	result, err := h.categoryUsecase.GetCategoriesWithBooks(c.Request.Context())
	if err != nil {
		response.Error(c, 500, "カテゴリの取得に失敗しました")
		return
	}

	response.Success(c, result)
}
