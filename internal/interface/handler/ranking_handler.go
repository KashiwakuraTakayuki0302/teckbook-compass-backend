package handler

import (
	"strconv"
	"teckbook-compass-backend/internal/usecase"
	"teckbook-compass-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// RankingHandler ランキングハンドラ
type RankingHandler struct {
	rankingUsecase *usecase.RankingUsecase
}

// NewRankingHandler ランキングハンドラのコンストラクタ
func NewRankingHandler(rankingUsecase *usecase.RankingUsecase) *RankingHandler {
	return &RankingHandler{
		rankingUsecase: rankingUsecase,
	}
}

// GetRankings 総合ランキング取得API
// @Summary 総合ランキング取得
// @Description 技術書の総合ランキングを取得
// @Tags rankings
// @Accept json
// @Produce json
// @Param range query string false "ランキング期間 (all, monthly, yearly)" default(all)
// @Param limit query int false "取得件数" default(5) minimum(1) maximum(100)
// @Param offset query int false "オフセット" default(0) minimum(0)
// @Param category query string false "カテゴリID"
// @Success 200 {object} dto.RankingResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /rankings [get]
func (h *RankingHandler) GetRankings(c *gin.Context) {
	// クエリパラメータの取得とデフォルト値設定
	rangeType := c.DefaultQuery("range", "all")
	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")
	categoryID := c.Query("category")

	// バリデーション: range
	if rangeType != "all" && rangeType != "monthly" && rangeType != "yearly" {
		response.Error(c, 400, "range パラメータは all, monthly, yearly のいずれかである必要があります")
		return
	}

	// バリデーション: limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		response.Error(c, 400, "limit パラメータは 1 から 100 の整数である必要があります")
		return
	}

	// バリデーション: offset
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.Error(c, 400, "offset パラメータは 0 以上の整数である必要があります")
		return
	}

	// ユースケースを実行
	result, err := h.rankingUsecase.GetRankings(c.Request.Context(), rangeType, limit, offset, categoryID)
	if err != nil {
		response.Error(c, 500, "ランキングの取得に失敗しました")
		return
	}

	response.Success(c, result)
}
