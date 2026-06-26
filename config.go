package main

import (
	"errors"
	"os"
	"strconv"
)

type ChatConfig struct {
	Source string
	ChatID int64
	URL    string
}

type Config struct {
	Chats []ChatConfig
}

var cityDefs = []struct {
	Source string
	Prefix string
}{
	{"臺南", "TAINAN"},
	{"高雄", "KAOHSIUNG"},
	{"新竹", "HSINCHU"},
}

// LoadConfig reads chat configuration from environment variables.
// Each city uses the pattern {PREFIX}_CHAT and {PREFIX}_URL, e.g.:
//
//	TAINAN_CHAT=-1002309286627
//	TAINAN_URL=https://119dts.tncfd.gov.tw/DTS/caselist/html
//
// Cities with neither env var set are skipped. At least one city is required.
func LoadConfig() (*Config, error) {
	var chats []ChatConfig

	for _, d := range cityDefs {
		chatStr := os.Getenv(d.Prefix + "_CHAT")
		urlStr := os.Getenv(d.Prefix + "_URL")
		if chatStr == "" && urlStr == "" {
			continue
		}

		chatID, err := strconv.ParseInt(chatStr, 10, 64)
		if err != nil {
			return nil, errors.New("invalid " + d.Prefix + "_CHAT: " + err.Error())
		}

		chats = append(chats, ChatConfig{
			Source: d.Source,
			ChatID: chatID,
			URL:    urlStr,
		})
	}

	if len(chats) == 0 {
		return nil, errors.New("no city env vars configured (e.g. TAINAN_CHAT, TAINAN_URL)")
	}

	return &Config{Chats: chats}, nil
}
