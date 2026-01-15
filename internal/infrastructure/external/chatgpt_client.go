package external

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"teckbook-compass-backend/internal/infrastructure/config"
)

// ChatGPT API関連のエラー
var (
	ErrChatGPTAPIError     = errors.New("chatgpt api error")
	ErrChatGPTInvalidJSON  = errors.New("invalid json response from chatgpt")
	ErrChatGPTRateLimited  = errors.New("chatgpt api rate limited")
	ErrChatGPTUnauthorized = errors.New("chatgpt api unauthorized")
)

// ChatGPTClient ChatGPT APIクライアント
type ChatGPTClient struct {
	config     config.ChatGPTConfig
	httpClient *http.Client
}

// NewChatGPTClient ChatGPTClientを生成
func NewChatGPTClient(cfg config.ChatGPTConfig) *ChatGPTClient {
	return &ChatGPTClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // タイムアウトを2分に延長
		},
	}
}

// IsEnabled ChatGPT APIが有効かどうかを判定
func (c *ChatGPTClient) IsEnabled() bool {
	return c.config.Enabled && c.config.APIKey != ""
}

// ChatGPTRequest ChatGPT APIリクエスト
type ChatGPTRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatGPTMessage `json:"messages"`
	Temperature float64          `json:"temperature"`
	MaxTokens   int              `json:"max_tokens"`
}

// ChatGPTMessage ChatGPT APIメッセージ
type ChatGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatGPTResponse ChatGPT APIレスポンス
type ChatGPTResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *ChatGPTErrorDetail `json:"error,omitempty"`
}

// ChatGPTErrorDetail ChatGPT APIエラー詳細
type ChatGPTErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// CategoryResult カテゴリ分類結果
type CategoryResult struct {
	CategoryCode string `json:"category_code"`
	// トークン使用量
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// CategorizeBook 書籍をカテゴリに分類する
func (c *ChatGPTClient) CategorizeBook(ctx context.Context, title, overview, isbn13 string, categories []CategoryInfo) (*CategoryResult, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("chatgpt api is disabled")
	}

	prompt := c.buildCategorizationPrompt(title, overview, isbn13, categories)

	req := ChatGPTRequest{
		Model: c.config.Model,
		Messages: []ChatGPTMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // 一貫性を重視
		MaxTokens:   100, // JSON応答は短いので十分
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// ステータスコードチェック
	if resp.StatusCode == 401 {
		return nil, ErrChatGPTUnauthorized
	}
	if resp.StatusCode == 429 {
		return nil, ErrChatGPTRateLimited
	}
	if resp.StatusCode != 200 {
		log.Printf("ChatGPT API error: status=%d, body=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%w: status=%d", ErrChatGPTAPIError, resp.StatusCode)
	}

	var chatResp ChatGPTResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("%w: %s", ErrChatGPTAPIError, chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no choices in response", ErrChatGPTAPIError)
	}

	content := chatResp.Choices[0].Message.Content
	log.Printf("ChatGPT response: %s (tokens: prompt=%d, completion=%d, total=%d)",
		content, chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens, chatResp.Usage.TotalTokens)

	// 有効なカテゴリコードのマップを作成
	validCodes := make(map[string]bool, len(categories))
	for _, cat := range categories {
		validCodes[cat.Code] = true
	}

	// JSONを抽出してパース
	result, err := c.parseCategorizationResult(content, validCodes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrChatGPTInvalidJSON, err)
	}

	// トークン使用量を設定
	result.PromptTokens = chatResp.Usage.PromptTokens
	result.CompletionTokens = chatResp.Usage.CompletionTokens
	result.TotalTokens = chatResp.Usage.TotalTokens

	return result, nil
}

// CategoryInfo カテゴリ情報
type CategoryInfo struct {
	Code string // カテゴリコード（例: "ai-ml"）
	Name string // カテゴリ名（例: "AI / 機械学習"）
}

// buildCategorizationPrompt 分類用プロンプトを構築
func (c *ChatGPTClient) buildCategorizationPrompt(title, overview, isbn13 string, categories []CategoryInfo) string {
	// カテゴリ一覧を構築（コードと名前のみ）
	categoryList := ""
	for _, cat := range categories {
		categoryList += fmt.Sprintf("- %s: %s\n", cat.Code, cat.Name)
	}

	prompt := fmt.Sprintf(`あなたは技術書を分類する専門家です。  
以下の「技術書タイトル」「技術書概要」「ISBN13」をもとに、  
指定されたカテゴリの中から **最も適切な1つだけ** を選択してください。

### 分類ルール

- 必ず **1カテゴリのみ** 選択すること
- 複数に該当しそうな場合でも、**書籍の主題として最も中心的なカテゴリを選ぶ**
- ISBN13は「書籍の種類・分野を補助的に判断する情報」として使用してよい
- 判断が難しい場合は「想定読者が最も多いカテゴリ」を基準にする

### カテゴリ一覧（コード: カテゴリ名）

%s

---

### 入力情報

- 技術書タイトル：  
    「%s」
- 技術書概要：  
    「%s」
- ISBN13：  
    「%s」

---

### 出力形式（JSONのみで返すこと）

category_codeには上記カテゴリ一覧のコード（例: "ai-ml", "web"など）を指定してください。

`+"```json\n{\"category_code\": \"string\"}\n```", categoryList, title, overview, isbn13)

	return prompt
}

// parseCategorizationResult ChatGPTの応答からカテゴリ結果をパース
func (c *ChatGPTClient) parseCategorizationResult(content string, validCodes map[string]bool) (*CategoryResult, error) {
	// JSONブロックを抽出（```json ... ``` または { ... }）
	jsonPattern := regexp.MustCompile("(?s)```json\\s*(.+?)\\s*```|\\{[^}]*\"category_code\"[^}]*\\}")
	matches := jsonPattern.FindStringSubmatch(content)

	var jsonStr string
	if len(matches) > 1 && matches[1] != "" {
		jsonStr = matches[1]
	} else if len(matches) > 0 {
		jsonStr = matches[0]
	} else {
		// 直接JSONとしてパースを試みる
		jsonStr = content
	}

	var result CategoryResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse json: %w, content: %s", err, content)
	}

	// カテゴリコードが空の場合
	if result.CategoryCode == "" {
		return nil, fmt.Errorf("category_code is empty, content: %s", content)
	}

	// カテゴリコードが有効なものか検証
	if !validCodes[result.CategoryCode] {
		validCodesList := make([]string, 0, len(validCodes))
		for code := range validCodes {
			validCodesList = append(validCodesList, code)
		}
		return nil, fmt.Errorf("invalid category_code: %q (valid codes: %v)", result.CategoryCode, validCodesList)
	}

	return &result, nil
}
