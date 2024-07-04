package router

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	groq "github.com/learnLi/groq_client"
	"groqai2api/global"
	"groqai2api/middlewares"
	"groqai2api/pkg/bogdanfinn"
	"net/http"
	"strings"
	"time"
)

func authSessionHandler(client groq.HTTPClient, account *groq.Account, api_key string, proxy string) error {
	organization, err := groq.GerOrganizationId(client, api_key, proxy)
	if err != nil {
		return err
	}
	account.Organization = organization
	global.Cache.Set(organization, api_key, 3*time.Minute)
	return nil
}

func authRefreshHandler(client groq.HTTPClient, account *groq.Account, api_key string, proxy string) error {
	token, err := groq.GetSessionToken(client, api_key, "")
	if err != nil {
		return err
	}
	organization, err := groq.GerOrganizationId(client, token.Data.SessionJwt, proxy)
	if err != nil {
		return err
	}
	account.Organization = organization
	global.Cache.Set(organization, token.Data.SessionJwt, 3*time.Minute)
	return nil
}

func chat(c *gin.Context) {
	var (
		isApiKey = false
		apiKey   string
		apiReq   groq.APIRequest
	)
	if err := c.ShouldBindJSON(&apiReq); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	}
	// 默认插入中文prompt
	if global.ChinaPrompt == "true" {
		prompt := groq.APIMessage{
			Content: "使用中文回答，输出时不要带英文",
			Role:    "system",
		}
		apiReq.Messages = append([]groq.APIMessage{prompt}, apiReq.Messages...)
	}

	client := bogdanfinn.NewStdClient()
	proxyUrl := global.ProxyPool.GetProxyIP()
	if proxyUrl != "" {
		client.SetProxy(proxyUrl)
	}

	authorization := c.Request.Header.Get("Authorization")
	account := global.AccountPool.Get()
	if authorization != "" {
		customToken := strings.Replace(authorization, "Bearer ", "", 1)
		if customToken != "" {
			// 如果支持apikey调用，且以gsk_开头的字符串，说明传递的是apikey
			if global.SupportApikey == "true" && strings.HasPrefix(customToken, global.ApiKeyPrefix) {
				isApiKey = true
				apiKey = customToken
				account = groq.NewAccount(customToken, "")
			}
			// 说明传递的是session_token
			if strings.HasPrefix(customToken, "eyJhbGciOiJSUzI1NiI") {
				account = groq.NewAccount("", "")
				err := authSessionHandler(client, account, customToken, "")
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
			if len(customToken) == global.SessionTokenLen {
				account = groq.NewAccount(customToken, "")
				err := authRefreshHandler(client, account, customToken, "")
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
		}
	}

	if account == nil {
		c.JSON(400, gin.H{"error": "found account"})
		c.Abort()
		return
	}

	if !isApiKey {
		if _, ok := global.Cache.Get(account.Organization); !ok {
			err := authRefreshHandler(client, account, account.SessionToken, "")
			if err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		}
		cacheKey, _ := global.Cache.Get(account.Organization)
		apiKey = cacheKey.(string)
	}

	response, err := groq.ChatCompletions(client, apiReq, apiKey, account.Organization, "")
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	defer response.Body.Close()
	groq.NewReadWriter(c.Writer, response).StreamHandler()
}

func models(c *gin.Context) {
	client := bogdanfinn.NewStdClient()
	proxyUrl := global.ProxyPool.GetProxyIP()
	if proxyUrl != "" {
		client.SetProxy(proxyUrl)
	}
	authorization := c.Request.Header.Get("Authorization")
	account := global.AccountPool.Get()
	if authorization != "" {
		customToken := strings.Replace(authorization, "Bearer ", "", 1)
		if customToken != "" {
			// 说明传递的是session_token
			if strings.HasPrefix(customToken, "eyJhbGciOiJSUzI1NiI") {
				account = groq.NewAccount("", "")
				err := authSessionHandler(client, account, customToken, "")
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
			if len(customToken) == 44 {
				account = groq.NewAccount(customToken, "")
				err := authRefreshHandler(client, account, customToken, "")
				if err != nil {
					c.JSON(400, gin.H{"error": err.Error()})
					c.Abort()
					return
				}
			}
		}
	}

	if account == nil {
		c.JSON(400, gin.H{"error": "found account"})
		c.Abort()
		return
	}

	if _, ok := global.Cache.Get(account.Organization); !ok {
		err := authRefreshHandler(client, account, account.SessionToken, "")
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	}
	api_key, _ := global.Cache.Get(account.Organization)
	response, err := groq.GetModels(client, api_key.(string), account.Organization, "")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	var mo groq.Models

	if err := json.NewDecoder(response.Body).Decode(&mo); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, mo)
}

func InitChat(Router *gin.RouterGroup) {
	Router.Use(middlewares.Authorization)
	Router.GET("models", models)
	ChatRouter := Router.Group("chat")
	{
		ChatRouter.OPTIONS("/completions", middlewares.Options)
		ChatRouter.POST("/completions", chat)
	}
}
