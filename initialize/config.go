package initialize

import (
	"github.com/joho/godotenv"
	"groqai2api/global"
	"os"
)

func InitConfig() {
	_ = godotenv.Load(".env")
	global.Host = os.Getenv("SERVER_HOST")
	if global.Host == "" {
		global.Host = "127.0.0.1"
	}
	global.Port = os.Getenv("SERVER_PORT")
	if global.Port == "" {
		global.Port = "8080"
	}
	global.ChinaPrompt = os.Getenv("CHINA_PROMPT")
	global.Authorization = os.Getenv("Authorization")
	global.OpenAuthSecret = os.Getenv("OpenAuthSecret")
	global.AuthSecret = os.Getenv("AuthSecret")
	if global.AuthSecret == "" {
		if global.Authorization == "" {
			global.AuthSecret = "root"
		} else {
			global.AuthSecret = global.Authorization
		}
	}
}
