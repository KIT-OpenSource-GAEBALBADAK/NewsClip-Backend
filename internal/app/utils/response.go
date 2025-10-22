package utils

import "github.com/gin-gonic/gin"

// 표준 성공 응답
func SendSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(200, gin.H{ // 201 Created 등 상황에 맞게 수정 가능
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

// 표준 에러 응답
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"status":  "error",
		"message": message,
		"data":    nil,
	})
}
