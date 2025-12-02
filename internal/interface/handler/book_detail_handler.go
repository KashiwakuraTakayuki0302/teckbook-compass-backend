package handler

import (
	"teckbook-compass-backend/internal/usecase"
	"teckbook-compass-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// BookDetailHandler 書籍詳細ハンドラ
type BookDetailHandler struct {
	bookDetailUsecase *usecase.BookDetailUsecase
}

// NewBookDetailHandler 書籍詳細ハンドラのコンストラクタ
func NewBookDetailHandler(bookDetailUsecase *usecase.BookDetailUsecase) *BookDetailHandler {
	return &BookDetailHandler{
		bookDetailUsecase: bookDetailUsecase,
	}
}

// GetBookDetail 書籍詳細取得API
// @Summary 書籍詳細取得
// @Description 指定された書籍IDの詳細情報を取得する
// @Tags books
// @Accept json
// @Produce json
// @Param bookId path string true "書籍ID"
// @Success 200 {object} dto.BookDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{bookId} [get]
func (h *BookDetailHandler) GetBookDetail(c *gin.Context) {
	// パスパラメータからbookIDを取得
	bookID := c.Param("bookId")

	// バリデーション: bookIDが空でないか確認
	if bookID == "" {
		response.Error(c, 400, "書籍IDは必須です")
		return
	}

	// ユースケースを実行
	result, err := h.bookDetailUsecase.GetBookDetail(c.Request.Context(), bookID)
	if err != nil {
		response.Error(c, 500, "書籍詳細の取得に失敗しました")
		return
	}

	// 書籍が見つからない場合
	if result == nil {
		response.Error(c, 404, "指定された書籍が見つかりません")
		return
	}

	response.Success(c, result)
}

