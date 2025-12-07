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

// RakutenClient 楽天ブックスAPIクライアント
type RakutenClient struct {
	config     config.RakutenConfig
	httpClient *http.Client
}

// NewRakutenClient RakutenClientを生成
func NewRakutenClient(cfg config.RakutenConfig) *RakutenClient {
	return &RakutenClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchByISBN ISBNで書籍を検索
func (c *RakutenClient) SearchByISBN(ctx context.Context, isbn string) (*entity.RakutenBook, error) {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("isbn", isbn)
	params.Set("applicationId", c.config.ApplicationID)
	params.Set("affiliateId", c.config.AffiliateID)

	reqURL := fmt.Sprintf("%s?%s", c.config.BaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	var response entity.RakutenBookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no book found for ISBN: %s", isbn)
	}

	return &response.Items[0].Item, nil
}

// SearchByTitle タイトルで書籍を検索
func (c *RakutenClient) SearchByTitle(ctx context.Context, title string) ([]*entity.RakutenBook, error) {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("title", title)
	params.Set("applicationId", c.config.ApplicationID)
	params.Set("affiliateId", c.config.AffiliateID)
	params.Set("hits", "10") // 最大10件

	reqURL := fmt.Sprintf("%s?%s", c.config.BaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	var response entity.RakutenBookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	books := make([]*entity.RakutenBook, len(response.Items))
	for i := range response.Items {
		books[i] = &response.Items[i].Item
	}

	return books, nil
}

// SearchByKeyword キーワードで書籍を検索（著者名やタイトルの部分一致）
func (c *RakutenClient) SearchByKeyword(ctx context.Context, keyword string) ([]*entity.RakutenBook, error) {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("keyword", keyword)
	params.Set("applicationId", c.config.ApplicationID)
	params.Set("affiliateId", c.config.AffiliateID)
	params.Set("hits", "10")
	params.Set("booksGenreId", "001") // 本ジャンル

	reqURL := fmt.Sprintf("%s?%s", c.config.BaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	var response entity.RakutenBookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	books := make([]*entity.RakutenBook, len(response.Items))
	for i := range response.Items {
		books[i] = &response.Items[i].Item
	}

	return books, nil
}

// FetchBookInfo ISBNまたはタイトルで書籍情報を取得（ISBN優先）
func (c *RakutenClient) FetchBookInfo(ctx context.Context, isbn string, title string) (*entity.RakutenBook, error) {
	// ISBNがある場合はISBNで検索
	if isbn != "" {
		book, err := c.SearchByISBN(ctx, isbn)
		if err == nil {
			return book, nil
		}
		// ISBNで見つからない場合はタイトルで検索
	}

	// タイトルで検索
	if title != "" {
		books, err := c.SearchByTitle(ctx, title)
		if err != nil {
			return nil, fmt.Errorf("failed to search by title: %w", err)
		}
		if len(books) > 0 {
			return books[0], nil
		}
	}

	return nil, fmt.Errorf("no book found for isbn: %s, title: %s", isbn, title)
}
