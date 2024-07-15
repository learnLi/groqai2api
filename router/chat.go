package router

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	groq "github.com/learnLi/groq_client"
	"groqai2api/global"
	"groqai2api/middlewares"
	"groqai2api/pkg/bogdanfinn"
	"net/http"
	"strings"
	"time"
)

var (
	SupportModels = map[string]string{
		"gemma2-9b-it":       "gemma2-9b-it",
		"gemma-7b-it":        "gemma-7b-it",
		"llama3-70b-8192":    "llama3-70b-8192",
		"llama3-8b-8192":     "llama3-8b-8192",
		"mixtral-8x7b-32768": "mixtral-8x7b-32768",
		"gpt-3.5-turbo":      "llama3-70b-8192",
		"gpt-4":              "llama3-70b-8192",
	}
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
	var apiReq groq.APIRequest
	if err := c.ShouldBindJSON(&apiReq); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	}

	if apiReq.Model != "" {
		if strings.HasPrefix(apiReq.Model, "gpt-3.5") {
			apiReq.Model = "gpt-3.5-turbo"
		}

		if strings.HasPrefix(apiReq.Model, "gpt-4") {
			apiReq.Model = "gpt-4"
		}
	} else {
		// 防呆，给默认值
		apiReq.Model = "llama3-70b-8192"
	}

	// 处理模型映射
	if _, ok := SupportModels[apiReq.Model]; ok {
		apiReq.Model = SupportModels[apiReq.Model]
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

	account, apiKey, err := processAPIKey(client, authorization)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
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
	account, apiKey, err := processAPIKey(client, authorization)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	response, err := groq.GetModels(client, apiKey, account.Organization, "")
	defer response.Body.Close()
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

func processAPIKey(client groq.HTTPClient, authorization string) (*groq.Account, string, error) {
	var (
		isApiKey = false
		apiKey   string
	)
	account := global.AccountPool.Get()
	if authorization != "" {
		customToken := strings.Replace(authorization, "Bearer ", "", 1)
		if customToken != "" {
			// 如果支持apikey调用，且以gsk_开头的字符串，说明传递的是apikey
			if global.SupportApikey == "true" && strings.HasPrefix(customToken, global.ApiKeyPrefix) {
				isApiKey = true
				apiKey = customToken
				account = groq.NewAccountWithAPIKey(customToken, "", true)
			}
			// 说明传递的是session_token
			if strings.HasPrefix(customToken, "eyJhbGciOiJSUzI1NiI") {
				account = groq.NewAccount("", "")
				err := authSessionHandler(client, account, customToken, "")
				if err != nil {
					return account, "", err
				}
			}
			if len(customToken) == global.SessionTokenLen {
				account = groq.NewAccount(customToken, "")
				err := authRefreshHandler(client, account, customToken, "")
				if err != nil {
					return account, "", err
				}
			}
		}
	}

	if account == nil {
		return account, "", errors.New("found account")
	}

	if !isApiKey && !account.IsAPIKey {
		if _, ok := global.Cache.Get(account.Organization); !ok {
			err := authRefreshHandler(client, account, account.SessionToken, "")
			if err != nil {
				return account, "", err
			}
		}
		cacheKey, _ := global.Cache.Get(account.Organization)
		apiKey = cacheKey.(string)
		return account, apiKey, nil
	}
	return account, account.SessionToken, nil
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
