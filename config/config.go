package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Configuration 项目配置
type Configuration struct {
	// gtp apikey
	ApiKey string `json:"api_key"`
	// 自动通过好友
	AutoPass bool `json:"auto_pass"`
	ObsidianGroup string `json:"obsidian_group"`
	ObsidianPath   string `json:"obsidian_daily_note"`
	// 启用私聊 AI 回复
	EnableUserChatAI bool `json:"enable_user_chat_ai"`
	SeparatorKeyword string `json:"SeparatorKeyword"` // <--- 新增这一行
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 从文件中读取
		config = &Configuration{}
		f, err := os.Open("config.json")
		if err != nil {
			log.Fatalf("open config err: %v", err)
			return
		}
		defer f.Close()
		encoder := json.NewDecoder(f)
		err = encoder.Decode(config)
		if err != nil {
			log.Fatalf("decode config err: %v", err)
			return
		}

		// 如果环境变量有配置，读取环境变量
		ApiKey := os.Getenv("ApiKey")
		AutoPass := os.Getenv("AutoPass")
		ObsidianGroup := os.Getenv("ObsidianGroup")  // 修复：原来读的是 ApiKey
		ObsidianPath := os.Getenv("ObsidianPath")
		EnableUserChatAI := os.Getenv("EnableUserChatAI")

		if ApiKey != "" {
			config.ApiKey = ApiKey
		}
		if AutoPass == "true" {
			config.AutoPass = true
		}
		if ObsidianGroup != "" {
			config.ObsidianGroup = ObsidianGroup
		}
		if ObsidianPath != "" {
			config.ObsidianPath = ObsidianPath
		}
		if EnableUserChatAI == "true" {
			config.EnableUserChatAI = true
		}
	})
	return config
}