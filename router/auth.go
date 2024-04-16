package router

import (
	"errors"
	"github.com/gin-gonic/gin"
	"groqai2api/global"
	"groqai2api/middlewares"
	"strings"
)

func validateTokenWithComma(tokens []string) ([]string, error) {
	var newTokens []string
	if len(tokens) == 0 {
		return nil, errors.New("token is empty")
	}
	for _, token := range tokens {
		if token != "" && len(token) == 44 {
			newTokens = append(newTokens, token)
		}
	}
	if len(newTokens) == 0 {
		return nil, errors.New("token is invalid")
	}
	return newTokens, nil
}

func getList(c *gin.Context) {
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "ok",
		"data": global.AccountPool.GetList(),
	})
}

func addTokens(c *gin.Context) {
	token := c.PostForm("session_token")
	// 支持多账号添加，使用,号分隔
	var tokens []string
	tokens = strings.Split(token, ",")
	tokens, err := validateTokenWithComma(tokens)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	err = global.AccountPool.Add(tokens)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func InitAuth(Router *gin.RouterGroup) {
	Router.Use(middlewares.AuthSecret)
	Router.POST("add", addTokens)
	Router.POST("list", getList)
}
