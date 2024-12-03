package middleware

import (
	"KazeFrame/internal/config"

	"github.com/gin-gonic/gin"
)

// TokenMatcherMiddleware 是一个中间件，用于检查URL中的token参数是否匹配预设的token值
func ParamAuth() gin.HandlerFunc {
	paramToken := config.GetConfig().Token.ParamKey
	return func(c *gin.Context) {
		token := c.Query("token")
		if token != paramToken {
			c.JSON(403, gin.H{"code": 403, "message": "无权访问"})
			c.Abort()
			return
		}
		c.Next()
	}
}
