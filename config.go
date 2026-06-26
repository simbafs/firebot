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
	Kind   string // "dts" or "asp"
}

type Config struct {
	Chats []ChatConfig
}

var cityDefs = []struct {
	Source string
	Prefix string
	URL    string
	Kind   string
}{
	{"臺南", "TAINAN", "https://119dts.tncfd.gov.tw/DTS/caselist/html", "dts"},
	{"高雄", "KAOHSIUNG", "https://119dts.fdkc.gov.tw/DTS/caselist/html", "dts"},
	{"新竹", "HSINCHU", "https://119.hcfd.gov.tw/DTS/caselist/html", "dts"},
	{"苗栗", "MIAOLI", "https://119mlfire.mlfd.gov.tw/DTS/caselist/html", "dts"},
	{"雲林", "YUNLIN", "https://119.ylfire.gov.tw/DTS/caselist/html", "dts"},
	{"臺中", "TAICHUNG", "https://www.fire.taichung.gov.tw/caselist/index.asp?Parser=99,8,226", "asp"},
	{"彰化", "CHANGHUA", "https://www.chfd.gov.tw/RealInfo/index.aspx?Parser=99,3,29", "asp"},
	{"桃園", "TAOYUAN", "https://www.tyfd.gov.tw/cht/index.php?act=caselist", "taoyuan"},
	{"新北", "NEWTAIPEI", "https://e.ntpc.gov.tw/v3/api/map/dynamic/layer/rescue", "ntpc_json"},
	{"嘉義縣", "CHIAYI", "https://cycfb.cyhg.gov.tw/DisasterPrevent.aspx?n=5F10482409025004&sms=ED4E0CDDC2EA92E6", "chiayi"},
}

// LoadConfig reads chat configuration from environment variables.
// If ALL_CHAT is set, all cities are routed to that single chat (for testing).
// Otherwise each city uses {PREFIX}_CHAT. Cities without {PREFIX}_CHAT are skipped.
// URLs are hardcoded in cityDefs.
//
//	ALL_CHAT=-1001234567890       # all cities → this chat
//	TAINAN_CHAT=-1002309286627    # individual city chat
func LoadConfig() (*Config, error) {
	var chats []ChatConfig

	allChatStr := os.Getenv("ALL_CHAT")
	if allChatStr != "" {
		allChatID, err := strconv.ParseInt(allChatStr, 10, 64)
		if err != nil {
			return nil, errors.New("invalid ALL_CHAT: " + err.Error())
		}

		for _, d := range cityDefs {
			chats = append(chats, ChatConfig{
				Source: d.Source,
				ChatID: allChatID,
				URL:    d.URL,
				Kind:   d.Kind,
			})
		}
		return &Config{Chats: chats}, nil
	}

	for _, d := range cityDefs {
		chatStr := os.Getenv(d.Prefix + "_CHAT")
		if chatStr == "" {
			continue
		}

		chatID, err := strconv.ParseInt(chatStr, 10, 64)
		if err != nil {
			return nil, errors.New("invalid " + d.Prefix + "_CHAT: " + err.Error())
		}

		chats = append(chats, ChatConfig{
			Source: d.Source,
			ChatID: chatID,
			URL:    d.URL,
			Kind:   d.Kind,
		})
	}

	if len(chats) == 0 {
		return nil, errors.New("no city env vars configured (e.g. TAINAN_CHAT, TAINAN_URL)")
	}

	return &Config{Chats: chats}, nil
}
