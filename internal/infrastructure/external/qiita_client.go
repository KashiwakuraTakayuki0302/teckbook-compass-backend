package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"teckbook-compass-backend/internal/domain/entity"
	"teckbook-compass-backend/internal/infrastructure/config"
)

// QiitaClient Qiita APIクライアント
type QiitaClient struct {
	config     config.QiitaConfig
	httpClient *http.Client
}

// NewQiitaClient QiitaClientを生成
func NewQiitaClient(cfg config.QiitaConfig) *QiitaClient {
	return &QiitaClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchQueries 技術書関連の検索クエリ
var SearchQueries = []string{
	"技術書",
	"書籍",
	"本の紹介",
	"おすすめ本",
	"読んだ本",
	"入門書",
	"書評",
	"レビュー",
	"読書",
}

// SearchArticles 検索クエリで記事を取得
func (c *QiitaClient) SearchArticles(ctx context.Context, query string, page int, perPage int) ([]*entity.QiitaAPIArticle, error) {
	// URLを構築
	baseURL := fmt.Sprintf("%s/items", c.config.BaseURL)
	params := url.Values{}
	params.Set("query", query)
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("per_page", fmt.Sprintf("%d", perPage))

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// リクエストを作成
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーを設定
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	// リクエストを実行
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// ステータスコードを確認
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	// JSONをパース
	var articles []*entity.QiitaAPIArticle
	if err := json.Unmarshal(body, &articles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return articles, nil
}

// FetchAllArticlesForQuery 特定のクエリですべてのページの記事を取得（最大ページ数制限付き）
func (c *QiitaClient) FetchAllArticlesForQuery(ctx context.Context, query string, maxPages int) ([]*entity.QiitaAPIArticle, error) {
	var allArticles []*entity.QiitaAPIArticle
	perPage := 100 // Qiita APIの最大値

	for page := 1; page <= maxPages; page++ {
		articles, err := c.SearchArticles(ctx, query, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
		}

		allArticles = append(allArticles, articles...)

		// 取得した記事が少なければ次のページはない
		if len(articles) < perPage {
			break
		}

		// レートリミット対策：1秒待機
		time.Sleep(1 * time.Second)
	}

	return allArticles, nil
}

// FetchArticlesByQueries 複数のクエリで記事を取得
func (c *QiitaClient) FetchArticlesByQueries(ctx context.Context, queries []string, maxPagesPerQuery int) ([]*entity.QiitaAPIArticle, error) {
	var allArticles []*entity.QiitaAPIArticle
	seen := make(map[string]bool)

	fmt.Printf("\n=== Qiita記事取得詳細 ===\n")
	fmt.Printf("検索クエリ数: %d, 各クエリ最大ページ数: %d\n\n", len(queries), maxPagesPerQuery)

	for _, query := range queries {
		articles, err := c.FetchAllArticlesForQuery(ctx, query, maxPagesPerQuery)
		if err != nil {
			// エラーをログに記録して続行
			fmt.Printf("  [%s] エラー: %v\n", query, err)
			continue
		}

		// 重複を排除してカウント
		newCount := 0
		duplicateCount := 0
		for _, article := range articles {
			if !seen[article.ID] {
				seen[article.ID] = true
				allArticles = append(allArticles, article)
				newCount++
			} else {
				duplicateCount++
			}
		}

		fmt.Printf("  [%s] 取得: %d件, 新規: %d件, 重複: %d件, 累計: %d件\n",
			query, len(articles), newCount, duplicateCount, len(allArticles))

		// レートリミット対策：クエリ間で待機
		time.Sleep(2 * time.Second)
	}

	fmt.Printf("\n=== 合計: %d件（重複排除後） ===\n\n", len(allArticles))

	return allArticles, nil
}

// GetArticle 記事IDで記事詳細を取得
func (c *QiitaClient) GetArticle(ctx context.Context, articleID string) (*entity.QiitaAPIArticle, error) {
	reqURL := fmt.Sprintf("%s/items/%s", c.config.BaseURL, articleID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var article entity.QiitaAPIArticle
	if err := json.Unmarshal(body, &article); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &article, nil
}

// GetArticleStocksCount 記事のストック数を取得
func (c *QiitaClient) GetArticleStocksCount(ctx context.Context, articleID string) (int, error) {
	reqURL := fmt.Sprintf("%s/items/%s/stockers", c.config.BaseURL, articleID)

	params := url.Values{}
	params.Set("per_page", "1")
	reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Total-Count ヘッダーからストック数を取得
	totalCount := resp.Header.Get("Total-Count")
	if totalCount == "" {
		return 0, nil
	}

	var count int
	_, err = fmt.Sscanf(totalCount, "%d", &count)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Total-Count: %w", err)
	}

	return count, nil
}

// QueryStats クエリごとの取得統計
type QueryStats struct {
	Query      string // クエリ文字列
	Fetched    int    // 取得件数
	New        int    // 新規件数（重複排除後）
	Duplicates int    // 重複件数
}

// FetchStats 取得統計
type FetchStats struct {
	QueryStats []QueryStats // クエリごとの統計
	Total      int          // 合計件数
}

// FetchNewArticles 最新記事を取得（指定日時以降の記事）
func (c *QiitaClient) FetchNewArticles(ctx context.Context, queries []string, since *time.Time, maxPagesPerQuery int) ([]*entity.QiitaAPIArticle, *FetchStats, error) {
	var allArticles []*entity.QiitaAPIArticle
	seen := make(map[string]bool)
	stats := &FetchStats{QueryStats: make([]QueryStats, 0, len(queries))}

	fmt.Printf("\n=== 最新記事取得モード ===\n")
	if since != nil {
		fmt.Printf("取得対象: %s 以降の記事\n", since.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("検索クエリ数: %d, 各クエリ最大ページ数: %d\n\n", len(queries), maxPagesPerQuery)

	for _, query := range queries {
		// 検索クエリに日時フィルタを追加
		searchQuery := query
		if since != nil {
			// Qiita APIの検索では created:>YYYY-MM-DD 形式でフィルタ可能
			searchQuery = fmt.Sprintf("%s created:>%s", query, since.Format("2006-01-02"))
		}

		articles, err := c.FetchAllArticlesForQuery(ctx, searchQuery, maxPagesPerQuery)
		if err != nil {
			fmt.Printf("  [%s] エラー: %v\n", query, err)
			stats.QueryStats = append(stats.QueryStats, QueryStats{Query: query, Fetched: 0, New: 0, Duplicates: 0})
			continue
		}

		newCount := 0
		dupCount := 0
		for _, article := range articles {
			if !seen[article.ID] {
				seen[article.ID] = true
				allArticles = append(allArticles, article)
				newCount++
			} else {
				dupCount++
			}
		}

		stats.QueryStats = append(stats.QueryStats, QueryStats{
			Query:      query,
			Fetched:    len(articles),
			New:        newCount,
			Duplicates: dupCount,
		})

		fmt.Printf("  [%s] 取得: %d件, 新規: %d件, 累計: %d件\n",
			query, len(articles), newCount, len(allArticles))

		time.Sleep(2 * time.Second)
	}

	stats.Total = len(allArticles)
	fmt.Printf("\n=== 最新記事合計: %d件 ===\n\n", len(allArticles))
	return allArticles, stats, nil
}

// FetchHistoricalArticles 過去記事を取得（指定ページから）
func (c *QiitaClient) FetchHistoricalArticles(ctx context.Context, queries []string, startPage int, pagesPerQuery int) ([]*entity.QiitaAPIArticle, int, *FetchStats, error) {
	var allArticles []*entity.QiitaAPIArticle
	seen := make(map[string]bool)
	perPage := 100
	stats := &FetchStats{QueryStats: make([]QueryStats, 0, len(queries))}

	fmt.Printf("\n=== 過去記事取得モード ===\n")
	fmt.Printf("開始ページ: %d, 各クエリ取得ページ数: %d\n\n", startPage, pagesPerQuery)

	endPage := startPage + pagesPerQuery - 1
	hasMorePages := false

	for _, query := range queries {
		queryFetched := 0
		queryNew := 0
		queryDup := 0

		for page := startPage; page <= endPage; page++ {
			articles, err := c.SearchArticles(ctx, query, page, perPage)
			if err != nil {
				fmt.Printf("  [%s] ページ%d エラー: %v\n", query, page, err)
				break
			}

			queryFetched += len(articles)

			for _, article := range articles {
				if !seen[article.ID] {
					seen[article.ID] = true
					allArticles = append(allArticles, article)
					queryNew++
				} else {
					queryDup++
				}
			}

			// まだ次のページがあるかチェック
			if len(articles) == perPage {
				hasMorePages = true
			}

			if len(articles) < perPage {
				break
			}

			time.Sleep(1 * time.Second)
		}

		stats.QueryStats = append(stats.QueryStats, QueryStats{
			Query:      query,
			Fetched:    queryFetched,
			New:        queryNew,
			Duplicates: queryDup,
		})

		fmt.Printf("  [%s] ページ%d-%d: 取得%d件, 新規%d件, 累計: %d件\n",
			query, startPage, endPage, queryFetched, queryNew, len(allArticles))

		time.Sleep(2 * time.Second)
	}

	nextPage := startPage
	if hasMorePages {
		nextPage = endPage + 1
	}

	stats.Total = len(allArticles)
	fmt.Printf("\n=== 過去記事合計: %d件, 次回開始ページ: %d ===\n\n", len(allArticles), nextPage)
	return allArticles, nextPage, stats, nil
}
