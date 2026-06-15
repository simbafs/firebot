package main

import (
	"log"
	"time"
)

var (
	APIKey = ""
	chats  = []struct {
		Source string
		ChatID int64
		URL    string
	}{
		{"臺南", -1002309286627, "https://119dts.tncfd.gov.tw/DTS/caselist/html"},
		{"高雄", -1003110857793, "https://119dts.fdkc.gov.tw/DTS/caselist/html"},
		{"新竹", -1003421899373, "https://119.hcfd.gov.tw/DTS/caselist/html"},
	}
)

func filter(e Event) bool {
	r := e.Category == "火災" || len(e.Brigade) >= 2
	return r
}

func init() {
	APIKey = Getenv("API_KEY", APIKey)

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func main() {
	bot := NewBot(WithAPIKey(APIKey))
	fetcher := &Fetcher{
		filter: filter,
	}
	first := true

	for {
		log.Println("Fetching...")
		for _, chat := range chats {
			go func(c struct {
				Source string
				ChatID int64
				URL    string
			}, f bool) {
				events, err := fetcher.Fetch(c.URL, c.Source)
				if err != nil {
					log.Println(err)
					return
				}

				for _, event := range events {
					if err := bot.SendEvent(c.ChatID, &event, f); err != nil {
						log.Println(err)
					}
				}
			}(chat, first)
		}
		first = false
		bot.GC()

		time.Sleep(10 * time.Second)
	}
}
