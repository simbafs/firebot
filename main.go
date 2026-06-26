package main

import (
	"log"
	"time"

	"tainanfire/diff"
	"tainanfire/event"
	"tainanfire/fetch"
	"tainanfire/telegram"
)

var APIKey = ""

func filter(e event.Event) bool {
	r := e.Category == "火災" || len(e.Brigade) >= 2
	return r
}

func init() {
	APIKey = Getenv("API_KEY", APIKey)

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var bot telegram.Bot
	if APIKey != "" {
		bot = telegram.NewTGBot(telegram.WithAPIKey(APIKey))
	} else {
		log.Println("API_KEY not set, using LocalBot")
		bot = telegram.NewLocalBot()
	}

	for _, c := range cfg.Chats {
		if err := bot.UnpinAll(c.ChatID); err != nil {
			log.Printf("unpin all chat %d: %v", c.ChatID, err)
		}
	}

	fetcher := &fetch.Fetcher{
		Filter: filter,
	}

	differs := map[string]*diff.Differ{}
	for _, c := range cfg.Chats {
		differs[c.Source] = diff.New()
	}
	first := true

	log.Println("start fetching events...")
	for {
		for _, chat := range cfg.Chats {
			go func(c ChatConfig, f bool) {
				events, err := fetcher.Fetch(c.URL, c.Source, c.Kind)
				if err != nil {
					log.Println(err)
					return
				}

				differ := differs[c.Source]
				if f {
					differ.Init(events)
					return
				}

				result := differ.Diff(events)
				if err := bot.Broadcast(c.ChatID, result, false); err != nil {
					log.Println(err)
				}
			}(chat, first)
		}
		first = false

		time.Sleep(10 * time.Second)
	}
}
