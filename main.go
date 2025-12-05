package main

import (
	"log"
	"time"
)

var (
	APIKey = ""
	chats  = map[int64]string{
		/* 臺南	*/ -1002309286627: "https://119dts.tncfd.gov.tw/DTS/caselist/html",
		/* 高雄	*/ -1003110857793: "https://119dts.fdkc.gov.tw/tyfdapp/webControlKC?page=Tfqve7Vz8sjTOllavM2iqQ==&f=IC2SZJqIMDj1EwKMezrgvw==",
		/* 新竹 */ -1003421899373: "https://119.hcfd.gov.tw/DTS/caselist/html",
	}
)

func filter(e Event) bool {
	return true
	// r := (e.Status != "已到院" && e.Status != "返隊中" && e.Status != "已返隊")
	// log.Printf("%v: %#v", r, e)
	// return r
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

	for {
		log.Println("Fetching...")
		for chat, url := range chats {
			go func() {
				events, err := fetcher.Fetch(url)
				if err != nil {
					log.Println(err)
					return
				}

				for _, event := range events {
					if err := bot.SendEvent(chat, &event); err != nil {
						log.Println(err)
					}
				}
			}()
		}
		bot.GC()

		time.Sleep(10 * time.Second)
	}
}
