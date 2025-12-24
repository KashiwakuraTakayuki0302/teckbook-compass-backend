package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"teckbook-compass-backend/internal/infrastructure/config"
)

// SlackClient Slack通知クライアント
type SlackClient struct {
	config     config.SlackConfig
	httpClient *http.Client
}

// NewSlackClient SlackClientを生成
func NewSlackClient(cfg config.SlackConfig) *SlackClient {
	return &SlackClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackMessage Slackメッセージ構造体
type SlackMessage struct {
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment Slack添付ファイル構造体
type SlackAttachment struct {
	Color  string `json:"color,omitempty"`
	Title  string `json:"title,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
}

// IsEnabled Slack通知が有効かどうか
func (c *SlackClient) IsEnabled() bool {
	return c.config.Enabled && c.config.WebhookURL != ""
}

// SendStartMessage バッチ開始メッセージを送信
func (c *SlackClient) SendStartMessage(fetchMode string) error {
	if !c.IsEnabled() {
		return nil
	}

	emoji := "📚"
	if fetchMode == "過去記事取得" {
		emoji = "📖"
	}

	message := fmt.Sprintf("%s *TeckBook Compass バッチ処理開始*\n取得モード: *%s*\n開始時刻: %s",
		emoji, fetchMode, time.Now().Format("2006-01-02 15:04:05"))

	return c.sendWebhook(SlackMessage{Text: message})
}

// SendResultMessage バッチ結果メッセージを送信
func (c *SlackClient) SendResultMessage(fetchMode string, processedArticles, newArticles, processedBooks, errors int, nextPage int, duration time.Duration, fetchStats *FetchStats) error {
	if !c.IsEnabled() {
		return nil
	}

	// 結果の絵文字と色
	emoji := "✅"
	color := "good"
	if errors > 0 {
		emoji = "⚠️"
		color = "warning"
	}

	text := fmt.Sprintf("%s *TeckBook Compass バッチ処理完了*", emoji)

	resultText := fmt.Sprintf(
		"• 取得モード: %s\n• 処理した記事数: %d\n• 新規記事数: %d\n• 処理した書籍数: %d\n• エラー数: %d\n• 処理時間: %v",
		fetchMode, processedArticles, newArticles, processedBooks, errors, duration.Round(time.Second),
	)

	if nextPage > 0 {
		resultText += fmt.Sprintf("\n• 次回開始ページ: %d", nextPage)
	}

	attachments := []SlackAttachment{
		{
			Color:  color,
			Title:  "処理結果",
			Text:   resultText,
			Footer: fmt.Sprintf("終了時刻: %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	// Qiita取得統計を追加
	if fetchStats != nil && len(fetchStats.QueryStats) > 0 {
		statsText := ""
		for _, qs := range fetchStats.QueryStats {
			statsText += fmt.Sprintf("• %s: 取得%d件, 新規%d件, 重複%d件\n", qs.Query, qs.Fetched, qs.New, qs.Duplicates)
		}
		statsText += fmt.Sprintf("─────────────────\n*合計: %d件*", fetchStats.Total)

		attachments = append(attachments, SlackAttachment{
			Color: "#36a64f",
			Title: "📊 Qiita記事取得詳細",
			Text:  statsText,
		})
	}

	return c.sendWebhook(SlackMessage{
		Text:        text,
		Attachments: attachments,
	})
}

// SendError エラーメッセージを送信
func (c *SlackClient) SendError(title, errorMessage string) error {
	if !c.IsEnabled() {
		return nil
	}

	text := "🚨 *TeckBook Compass エラー*"

	attachments := []SlackAttachment{
		{
			Color:  "danger",
			Title:  title,
			Text:   errorMessage,
			Footer: fmt.Sprintf("発生時刻: %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	return c.sendWebhook(SlackMessage{
		Text:        text,
		Attachments: attachments,
	})
}

// SendLog ログメッセージを送信（何もしない - 親メッセージのみ）
func (c *SlackClient) SendLog(message string) error {
	return nil
}

// SendLogf フォーマット付きログメッセージを送信（何もしない - 親メッセージのみ）
func (c *SlackClient) SendLogf(format string, args ...interface{}) error {
	return nil
}

// SendAmazonBatchStartMessage Amazon URL取得バッチ開始メッセージを送信
func (c *SlackClient) SendAmazonBatchStartMessage(targetCount int) error {
	if !c.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf("🛒 *Amazon URL取得バッチ開始*\n処理対象: *%d件*\n開始時刻: %s",
		targetCount, time.Now().Format("2006-01-02 15:04:05"))

	return c.sendWebhook(SlackMessage{Text: message})
}

// SendAmazonBatchResultMessage Amazon URL取得バッチ結果メッセージを送信
func (c *SlackClient) SendAmazonBatchResultMessage(processedBooks, updatedBooks, notFoundBooks, errors int, duration time.Duration) error {
	if !c.IsEnabled() {
		return nil
	}

	// 結果の絵文字と色
	emoji := "✅"
	color := "good"
	if errors > 0 {
		emoji = "⚠️"
		color = "warning"
	}

	text := fmt.Sprintf("%s *Amazon URL取得バッチ完了*", emoji)

	resultText := fmt.Sprintf(
		"• 処理した書籍数: %d\n• 更新した書籍数: %d\n• 未発見書籍数: %d\n• エラー数: %d\n• 処理時間: %v",
		processedBooks, updatedBooks, notFoundBooks, errors, duration.Round(time.Second),
	)

	attachments := []SlackAttachment{
		{
			Color:  color,
			Title:  "処理結果",
			Text:   resultText,
			Footer: fmt.Sprintf("終了時刻: %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	return c.sendWebhook(SlackMessage{
		Text:        text,
		Attachments: attachments,
	})
}

// SendCategorizeBatchStartMessage カテゴライズバッチ開始メッセージを送信
func (c *SlackClient) SendCategorizeBatchStartMessage(targetCount int) error {
	if !c.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf("🏷️ *カテゴライズバッチ開始*\n処理対象: *%d件*\n開始時刻: %s",
		targetCount, time.Now().Format("2006-01-02 15:04:05"))

	return c.sendWebhook(SlackMessage{Text: message})
}

// SendCategorizeBatchResultMessage カテゴライズバッチ結果メッセージを送信
func (c *SlackClient) SendCategorizeBatchResultMessage(processedBooks, categorizedBooks, errors int, duration time.Duration, promptTokens, completionTokens, totalTokens int) error {
	if !c.IsEnabled() {
		return nil
	}

	// 結果の絵文字と色
	emoji := "✅"
	color := "good"
	if errors > 0 {
		emoji = "⚠️"
		color = "warning"
	}

	text := fmt.Sprintf("%s *カテゴライズバッチ完了*", emoji)

	resultText := fmt.Sprintf(
		"• 処理した書籍数: %d\n• カテゴライズ成功: %d\n• エラー数: %d\n• 処理時間: %v",
		processedBooks, categorizedBooks, errors, duration.Round(time.Second),
	)

	attachments := []SlackAttachment{
		{
			Color:  color,
			Title:  "処理結果",
			Text:   resultText,
			Footer: fmt.Sprintf("終了時刻: %s", time.Now().Format("2006-01-02 15:04:05")),
		},
	}

	// トークン使用量がある場合は追加
	if totalTokens > 0 {
		tokenText := fmt.Sprintf(
			"• プロンプトトークン: %d\n• 完了トークン: %d\n• 合計トークン: %d",
			promptTokens, completionTokens, totalTokens,
		)
		attachments = append(attachments, SlackAttachment{
			Color: "#4A90D9",
			Title: "🤖 ChatGPT トークン使用量",
			Text:  tokenText,
		})
	}

	return c.sendWebhook(SlackMessage{
		Text:        text,
		Attachments: attachments,
	})
}

// sendWebhook Webhookでメッセージを送信
func (c *SlackClient) sendWebhook(msg SlackMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := c.httpClient.Post(c.config.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}
