package external

import (
	"context"
	"errors"
	"fmt"
	"log"

	"teckbook-compass-backend/internal/infrastructure/config"

	paapi5 "github.com/goark/pa-api"
	"github.com/goark/pa-api/entity"
	"github.com/goark/pa-api/query"
)

// AmazonClient Amazon Product Advertising APIクライアント
type AmazonClient struct {
	client  paapi5.Client
	enabled bool
}

// AmazonBook Amazon APIから取得した書籍情報
type AmazonBook struct {
	ASIN      string
	Title     string
	URL       string
	DetailURL string
}

// ErrAmazonAPIError Amazon APIエラー
var ErrAmazonAPIError = errors.New("amazon api error")

// ErrAmazonNotFound Amazon商品が見つからない
var ErrAmazonNotFound = errors.New("amazon product not found")

// NewAmazonClient AmazonClientを生成
func NewAmazonClient(cfg config.AmazonConfig) *AmazonClient {
	if !cfg.Enabled {
		log.Println("Amazon API: 無効")
		return &AmazonClient{enabled: false}
	}

	client := paapi5.New(
		paapi5.WithMarketplace(paapi5.LocaleJapan),
	).CreateClient(
		cfg.AccessKey,
		cfg.SecretKey,
		cfg.PartnerTag,
	)

	log.Println("Amazon API: 有効")
	return &AmazonClient{
		client:  client,
		enabled: true,
	}
}

// IsEnabled Amazon APIが有効かどうか
func (c *AmazonClient) IsEnabled() bool {
	return c.enabled
}

// SearchByISBN ISBNで書籍を検索してAmazon URLを取得
func (c *AmazonClient) SearchByISBN(ctx context.Context, isbn string) (*AmazonBook, error) {
	if !c.enabled {
		return nil, fmt.Errorf("amazon api is disabled")
	}

	// ISBNで検索（GetItems APIを使用）
	q := query.NewGetItems(
		c.client.Marketplace(),
		c.client.PartnerTag(),
		c.client.PartnerType(),
	).ASINs([]string{isbn}).EnableItemInfo().EnableOffers()

	// リクエスト実行
	body, err := c.client.RequestContext(ctx, q)
	if err != nil {
		log.Printf("Amazon API error: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAmazonAPIError, err)
	}

	// レスポンスをパース
	res, err := entity.DecodeResponse(body)
	if err != nil {
		log.Printf("Amazon API decode error: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAmazonAPIError, err)
	}

	// エラーチェック
	if len(res.Errors) > 0 {
		errMsg := res.Errors[0].Message
		log.Printf("Amazon API returned error: %s", errMsg)
		return nil, fmt.Errorf("%w: %s", ErrAmazonAPIError, errMsg)
	}

	// 結果がない場合
	if res.ItemsResult == nil || len(res.ItemsResult.Items) == 0 {
		return nil, ErrAmazonNotFound
	}

	item := res.ItemsResult.Items[0]
	return &AmazonBook{
		ASIN:      item.ASIN,
		Title:     item.ItemInfo.Title.DisplayValue,
		URL:       item.DetailPageURL,
		DetailURL: item.DetailPageURL,
	}, nil
}

// SearchByTitle タイトルで書籍を検索してAmazon URLを取得
func (c *AmazonClient) SearchByTitle(ctx context.Context, title string) (*AmazonBook, error) {
	if !c.enabled {
		return nil, fmt.Errorf("amazon api is disabled")
	}

	// タイトルで検索（SearchItems APIを使用）
	q := query.NewSearchItems(
		c.client.Marketplace(),
		c.client.PartnerTag(),
		c.client.PartnerType(),
	).Search(query.Keywords, title).
		Request(query.SearchIndex, "Books").
		EnableItemInfo().
		EnableOffers()

	// リクエスト実行
	body, err := c.client.RequestContext(ctx, q)
	if err != nil {
		log.Printf("Amazon API error: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAmazonAPIError, err)
	}

	// レスポンスをパース
	res, err := entity.DecodeResponse(body)
	if err != nil {
		log.Printf("Amazon API decode error: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrAmazonAPIError, err)
	}

	// エラーチェック
	if len(res.Errors) > 0 {
		errMsg := res.Errors[0].Message
		log.Printf("Amazon API returned error: %s", errMsg)
		return nil, fmt.Errorf("%w: %s", ErrAmazonAPIError, errMsg)
	}

	// 結果がない場合
	if res.SearchResult == nil || len(res.SearchResult.Items) == 0 {
		return nil, ErrAmazonNotFound
	}

	item := res.SearchResult.Items[0]
	return &AmazonBook{
		ASIN:      item.ASIN,
		Title:     item.ItemInfo.Title.DisplayValue,
		URL:       item.DetailPageURL,
		DetailURL: item.DetailPageURL,
	}, nil
}
