package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Bot interface {
	Broadcast(chat int64, result DiffResult) error
}

type TGBot struct {
	bot    *gotgbot.Bot
	mu     sync.Mutex
	msgIDs map[string]int64 // UID → pinned Telegram message_id
}

type BotOpt struct {
	APIKey string
}

func WithAPIKey(apikey string) func(*BotOpt) {
	return func(opt *BotOpt) {
		opt.APIKey = apikey
	}
}

func NewTGBot(opts ...func(*BotOpt)) *TGBot {
	defaultOpt := &BotOpt{}

	for _, opt := range opts {
		opt(defaultOpt)
	}

	bot, err := gotgbot.NewBot(defaultOpt.APIKey, nil)
	if err != nil {
		panic(err)
	}
	return &TGBot{
		bot:    bot,
		msgIDs: make(map[string]int64),
	}
}

func (b *TGBot) storeMsgID(uid string, msgID int64) {
	b.mu.Lock()
	b.msgIDs[uid] = msgID
	b.mu.Unlock()
}

func (b *TGBot) getMsgID(uid string) (int64, bool) {
	b.mu.Lock()
	id, ok := b.msgIDs[uid]
	b.mu.Unlock()
	return id, ok
}

func (b *TGBot) delMsgID(uid string) {
	b.mu.Lock()
	delete(b.msgIDs, uid)
	b.mu.Unlock()
}

func (b *TGBot) Broadcast(chat int64, result DiffResult) error {
	for i := range result.New {
		event := &result.New[i]
		markdown := event.RichMarkdown()

		msg, err := sendRichMessage(b.bot.Token, chat, markdown)
		if err != nil {
			return err
		}

		if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
			DisableNotification: true,
		}); err != nil {
			log.Println("[pin-err]", event.UID, err)
		}

		b.storeMsgID(event.UID, msg.MessageId)
		log.Println("[new]", event.UID)
	}

	for i := range result.Updated {
		ed := &result.Updated[i]
		markdown := ed.RichDiffMarkdown()

		msgID, ok := b.getMsgID(ed.New.UID)
		if ok {
			if _, err := editRichMessage(b.bot.Token, chat, msgID, markdown); err != nil {
				log.Println("[edit-err]", ed.New.UID, err)
				// Fallback: send as new message + pin
				msg, err2 := sendRichMessage(b.bot.Token, chat, markdown)
				if err2 != nil {
					return err2
				}
				if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
					DisableNotification: true,
				}); err != nil {
					log.Println("[pin-err]", ed.New.UID, err)
				}
				b.storeMsgID(ed.New.UID, msg.MessageId)
			}
		} else {
			// Previously untracked — treat as new
			msg, err := sendRichMessage(b.bot.Token, chat, markdown)
			if err != nil {
				return err
			}
			if _, err := b.bot.PinChatMessage(chat, msg.MessageId, &gotgbot.PinChatMessageOpts{
				DisableNotification: true,
			}); err != nil {
				log.Println("[pin-err]", ed.New.UID, err)
			}
			b.storeMsgID(ed.New.UID, msg.MessageId)
		}
		log.Println("[update]", ed.New.UID)
	}

	for i := range result.Deleted {
		event := &result.Deleted[i]
		msgID, ok := b.getMsgID(event.UID)
		if !ok {
			continue
		}

		if _, err := b.bot.UnpinChatMessage(chat, &gotgbot.UnpinChatMessageOpts{
			MessageId: &msgID,
		}); err != nil {
			log.Println("[unpin-err]", event.UID, err)
		}
		b.delMsgID(event.UID)
		log.Println("[delete]", event.UID)
	}

	return nil
}

type LocalBot struct{}

func NewLocalBot() *LocalBot {
	return &LocalBot{}
}

func (b *LocalBot) Broadcast(chat int64, result DiffResult) error {
	for i := range result.New {
		fmt.Printf("[Chat %d] [new] %s\n%s\n\n", chat, result.New[i].UID, result.New[i].RichMarkdown())
	}
	for i := range result.Updated {
		fmt.Printf("[Chat %d] [update] %s\n%s\n\n", chat, result.Updated[i].New.UID, result.Updated[i].RichDiffMarkdown())
	}
	for i := range result.Deleted {
		fmt.Printf("[Chat %d] [delete] unpin %s\n", chat, result.Deleted[i].UID)
	}
	return nil
}
