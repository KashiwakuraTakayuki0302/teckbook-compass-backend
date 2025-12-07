package extractor

import (
	"regexp"
	"strings"
)

// BookExtractor 記事から書籍情報を抽出するエクストラクタ
type BookExtractor struct {
	// ISBN-10 パターン
	isbn10Pattern *regexp.Regexp
	// ISBN-13 パターン
	isbn13Pattern *regexp.Regexp
	// Amazon ASIN パターン
	asinPattern *regexp.Regexp
	// Amazon URL からASINを抽出するパターン
	amazonURLPattern *regexp.Regexp
	// 楽天 URL からISBNを抽出するパターン
	rakutenURLPattern *regexp.Regexp
	// 書籍タイトルパターン（「」や『』で囲まれたもの）
	titlePattern *regexp.Regexp
}

// ExtractedBook 抽出された書籍情報
type ExtractedBook struct {
	ISBN       string // ISBN（10桁または13桁）
	ASIN       string // Amazon ASIN
	Title      string // 抽出されたタイトル
	SourceType string // 抽出元の種類（"isbn", "asin", "amazon_url", "rakuten_url", "title"）
}

// NewBookExtractor BookExtractorを生成
func NewBookExtractor() *BookExtractor {
	return &BookExtractor{
		// ISBN-10: 10桁の数字（最後はXも可）
		isbn10Pattern: regexp.MustCompile(`\b(\d{9}[\dX])\b`),
		// ISBN-13: 978または979で始まる13桁
		isbn13Pattern: regexp.MustCompile(`\b(97[89]\d{10})\b`),
		// ASIN: Bで始まる10文字の英数字
		asinPattern: regexp.MustCompile(`\b(B[0-9A-Z]{9})\b`),
		// Amazon URL からASIN/ISBNを抽出
		amazonURLPattern: regexp.MustCompile(`amazon\.co\.jp/(?:dp|gp/product|exec/obidos/ASIN)/([A-Z0-9]{10})`),
		// 楽天 URL からISBNを抽出
		rakutenURLPattern: regexp.MustCompile(`books\.rakuten\.co\.jp/rb/(\d+)/`),
		// 書籍タイトルパターン（「」『』で囲まれたテキスト）
		titlePattern: regexp.MustCompile(`[「『]([^」』]{3,50})[」』]`),
	}
}

// ExtractFromText テキストから書籍情報を抽出
func (e *BookExtractor) ExtractFromText(text string) []ExtractedBook {
	var results []ExtractedBook
	seen := make(map[string]bool)

	// 1. Amazon URL からASIN/ISBNを抽出
	amazonMatches := e.amazonURLPattern.FindAllStringSubmatch(text, -1)
	for _, match := range amazonMatches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			book := ExtractedBook{SourceType: "amazon_url"}
			if strings.HasPrefix(match[1], "B") {
				book.ASIN = match[1]
			} else {
				book.ISBN = match[1]
			}
			results = append(results, book)
		}
	}

	// 2. 楽天 URL からIDを抽出（楽天の商品IDはISBNではないので別途処理が必要）
	rakutenMatches := e.rakutenURLPattern.FindAllStringSubmatch(text, -1)
	for _, match := range rakutenMatches {
		if len(match) > 1 && !seen["rakuten_"+match[1]] {
			seen["rakuten_"+match[1]] = true
			// 楽天のURLからはISBNが直接取れないことが多いので、スキップするか別途処理
		}
	}

	// 3. ISBN-13を抽出
	isbn13Matches := e.isbn13Pattern.FindAllString(text, -1)
	for _, isbn := range isbn13Matches {
		cleanISBN := cleanISBN(isbn)
		if !seen[cleanISBN] && isValidISBN13(cleanISBN) {
			seen[cleanISBN] = true
			results = append(results, ExtractedBook{
				ISBN:       cleanISBN,
				SourceType: "isbn13",
			})
		}
	}

	// 4. ISBN-10を抽出
	isbn10Matches := e.isbn10Pattern.FindAllString(text, -1)
	for _, isbn := range isbn10Matches {
		cleanISBN := cleanISBN(isbn)
		if !seen[cleanISBN] && isValidISBN10(cleanISBN) {
			// ISBN-10をISBN-13に変換
			isbn13 := convertISBN10to13(cleanISBN)
			if !seen[isbn13] {
				seen[isbn13] = true
				results = append(results, ExtractedBook{
					ISBN:       isbn13,
					SourceType: "isbn10",
				})
			}
		}
	}

	// 5. ASINを抽出
	asinMatches := e.asinPattern.FindAllString(text, -1)
	for _, asin := range asinMatches {
		if !seen[asin] {
			seen[asin] = true
			results = append(results, ExtractedBook{
				ASIN:       asin,
				SourceType: "asin",
			})
		}
	}

	// 6. 書籍タイトルを抽出（ISBN/ASINが見つからない場合の補助）
	titleMatches := e.titlePattern.FindAllStringSubmatch(text, -1)
	for _, match := range titleMatches {
		if len(match) > 1 {
			title := strings.TrimSpace(match[1])
			// 技術書らしいタイトルかどうかをフィルタリング
			if isTechBookTitle(title) && !seen["title_"+title] {
				seen["title_"+title] = true
				results = append(results, ExtractedBook{
					Title:      title,
					SourceType: "title",
				})
			}
		}
	}

	return results
}

// ExtractFromHTML HTMLテキストから書籍情報を抽出
func (e *BookExtractor) ExtractFromHTML(html string) []ExtractedBook {
	// HTMLタグを除去してテキストを抽出
	text := stripHTMLTags(html)
	return e.ExtractFromText(text)
}

// cleanISBN ISBNからハイフンやスペースを除去
func cleanISBN(isbn string) string {
	cleaned := strings.ReplaceAll(isbn, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	return cleaned
}

// isValidISBN10 ISBN-10のチェックデジットを検証
func isValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		sum += int(isbn[i]-'0') * (10 - i)
	}

	lastChar := isbn[9]
	if lastChar == 'X' || lastChar == 'x' {
		sum += 10
	} else if lastChar >= '0' && lastChar <= '9' {
		sum += int(lastChar - '0')
	} else {
		return false
	}

	return sum%11 == 0
}

// isValidISBN13 ISBN-13のチェックデジットを検証
func isValidISBN13(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}

	sum := 0
	for i := 0; i < 12; i++ {
		if isbn[i] < '0' || isbn[i] > '9' {
			return false
		}
		digit := int(isbn[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}

	checkDigit := (10 - (sum % 10)) % 10
	return int(isbn[12]-'0') == checkDigit
}

// convertISBN10to13 ISBN-10をISBN-13に変換
func convertISBN10to13(isbn10 string) string {
	if len(isbn10) != 10 {
		return isbn10
	}

	// 978 + ISBN-10の最初の9桁
	prefix := "978" + isbn10[:9]

	// チェックデジットを計算
	sum := 0
	for i := 0; i < 12; i++ {
		digit := int(prefix[i] - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10

	return prefix + string(rune('0'+checkDigit))
}

// stripHTMLTags HTMLタグを除去
func stripHTMLTags(html string) string {
	// HTMLタグを除去する簡易的な正規表現
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, " ")
}

// isTechBookTitle 技術書らしいタイトルかどうかを判定
func isTechBookTitle(title string) bool {
	// 技術書らしいキーワード
	techKeywords := []string{
		"入門", "実践", "基礎", "詳解", "徹底", "攻略",
		"プログラミング", "開発", "設計", "アーキテクチャ",
		"Java", "Python", "Go", "Ruby", "JavaScript", "TypeScript",
		"React", "Vue", "Angular", "Node",
		"AWS", "GCP", "Azure", "Docker", "Kubernetes",
		"機械学習", "深層学習", "AI", "データ",
		"セキュリティ", "ネットワーク", "Linux",
		"アルゴリズム", "データ構造",
		"Clean", "Agile", "TDD", "DDD",
	}

	titleLower := strings.ToLower(title)
	for _, keyword := range techKeywords {
		if strings.Contains(titleLower, strings.ToLower(keyword)) {
			return true
		}
	}

	// 一般的な書籍パターン
	if strings.HasSuffix(title, "入門") ||
		strings.HasSuffix(title, "実践") ||
		strings.Contains(title, "の教科書") ||
		strings.Contains(title, "ハンドブック") {
		return true
	}

	return false
}
