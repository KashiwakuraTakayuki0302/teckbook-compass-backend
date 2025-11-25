package response

import "github.com/gin-gonic/gin"

// Success 成功レスポンスを返す
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, data)
}

// Error エラーレスポンスを返す
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error": message,
	})
}
