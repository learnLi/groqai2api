package router

import (
	"errors"
	"github.com/gin-gonic/gin"
	groq "github.com/learnLi/groq_client"
	"groqai2api/global"
	"groqai2api/pkg/bogdanfinn"
	"strings"
)

func getAPIKeyList(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		c.Abort()
		return
	}
	client := bogdanfinn.NewStdClient()
	proxyUrl := global.ProxyPool.GetProxyIP()
	if proxyUrl != "" {
		client.SetProxy(proxyUrl)
	}
	account, status, err := handleAuthorizationWithAccount(client, authorization)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	cacheKey, _ := global.Cache.Get(account.Organization)
	apikeyslist, err := groq.GetAPIKEYSLIST(client, cacheKey.(string), account.Organization, "")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(200, apikeyslist)
}

func handleAuthorizationWithAccount(client groq.HTTPClient, authorization string) (*groq.Account, int, error) {
	customToken := strings.Replace(authorization, "Bearer ", "", 1)
	if customToken != "" {
		if len(customToken) != global.SessionTokenLen {
			return nil, 401, errors.New("请输入正确的 Session Token")
		}
	}
	account := groq.NewAccount(customToken, "")
	err := authRefreshHandler(client, account, customToken, "")
	if err != nil {
		return nil, 400, err
	}

	if _, ok := global.Cache.Get(account.Organization); !ok {
		if err := authRefreshHandler(client, account, account.SessionToken, ""); err != nil {
			return nil, 400, err
		}
	}
	return account, 200, nil
}

func generateAPIKEY(c *gin.Context) {
	token := c.DefaultPostForm("api_key_name", "_test")
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		c.Abort()
		return
	}
	client := bogdanfinn.NewStdClient()
	proxyUrl := global.ProxyPool.GetProxyIP()
	if proxyUrl != "" {
		client.SetProxy(proxyUrl)
	}
	account, status, err := handleAuthorizationWithAccount(client, authorization)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	cacheKey, _ := global.Cache.Get(account.Organization)
	apikey, err := groq.GenerateAPIKEY(client, cacheKey.(string), account.Organization, token, "")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(200, apikey)
}

func deleteAPIKEY(c *gin.Context) {
	// 从URL中获取参数
	apiKey := c.Param("apiKeyID")
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		c.Abort()
		return
	}
	client := bogdanfinn.NewStdClient()
	proxyUrl := global.ProxyPool.GetProxyIP()
	if proxyUrl != "" {
		client.SetProxy(proxyUrl)
	}
	account, status, err := handleAuthorizationWithAccount(client, authorization)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	cacheKey, _ := global.Cache.Get(account.Organization)
	if err := groq.DeleteAPIKEY(client, cacheKey.(string), account.Organization, apiKey, ""); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(200, gin.H{"message": "API Key 删除成功"})
}

func InitPlatform(Router *gin.RouterGroup) {
	Router.GET("api_keys", getAPIKeyList)
	Router.POST("api_keys", generateAPIKEY)
	Router.DELETE("api_keys/:apiKeyID", deleteAPIKEY)
}
