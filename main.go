package main

import (
	"log"
	"time"
)

var (
	APIKey = ""
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
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var bot Bot
	if APIKey != "" {
		bot = NewTGBot(WithAPIKey(APIKey))
	} else {
		log.Println("API_KEY not set, using LocalBot")
		bot = NewLocalBot()
	}

	// Clear all pinned messages from previous run.
	for _, c := range cfg.Chats {
		if err := bot.UnpinAll(c.ChatID); err != nil {
			log.Printf("unpin all chat %d: %v", c.ChatID, err)
		}
	}

	fetcher := &Fetcher{
		filter: filter,
	}

	differs := map[string]*Differ{}
	for _, c := range cfg.Chats {
		differs[c.Source] = NewDiffer()
	}
	first := true

	for {
		log.Println("Fetching...")
		for _, chat := range cfg.Chats {
			go func(c ChatConfig, f bool) {
				events, err := fetcher.Fetch(c.URL, c.Source)
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
