package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Bot interface {
	Broadcast(chat int64, result DiffResult) error
}

type TGBot struct {
	bot    *gotgbot.Bot
	mu     sync.Mutex
	msgIDs map[string]int64 // UID → Telegram message_id
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

func escapeMarkdown(s string) string {
	return strings.ReplaceAll(s, "-", "\\-")
}

func (b *TGBot) sendMsg(chat int64, text string) (*gotgbot.Message, error) {
	return b.bot.SendMessage(chat, escapeMarkdown(text), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})
}

func (b *TGBot) editMsg(chat int64, msgID int64, text string) error {
	_, _, err := b.bot.EditMessageText(escapeMarkdown(text), &gotgbot.EditMessageTextOpts{
		ChatId:    chat,
		MessageId: msgID,
		ParseMode: "MarkdownV2",
	})
	return err
}

// storeMsgID saves the mapping from event UID to Telegram message_id.
func (b *TGBot) storeMsgID(uid string, msgID int64) {
	b.mu.Lock()
	b.msgIDs[uid] = msgID
	b.mu.Unlock()
}

// getMsgID retrieves the Telegram message_id for an event UID.
func (b *TGBot) getMsgID(uid string) (int64, bool) {
	b.mu.Lock()
	id, ok := b.msgIDs[uid]
	b.mu.Unlock()
	return id, ok
}

func (b *TGBot) Broadcast(chat int64, result DiffResult) error {
	for i := range result.New {
		event := &result.New[i]
		text := "新事件\n" + event.String()
		msg, err := b.sendMsg(chat, text)
		if err != nil {
			return err
		}
		b.storeMsgID(event.UID, msg.MessageId)
		log.Println("[new]", event.UID, text)
	}

	for i := range result.Updated {
		ed := &result.Updated[i]
		diff := ed.Old.Diff(&ed.New)
		text := "事件更新\n" + diff + "\n" + ed.New.String()

		// Try to edit the existing Telegram message for this event.
		// If editing fails (e.g., message too old or deleted), send a new message.
		if msgID, ok := b.getMsgID(ed.New.UID); ok {
			if err := b.editMsg(chat, msgID, text); err == nil {
				log.Println("[edit]", ed.New.UID)
				continue
			}
		}

		msg, err := b.sendMsg(chat, text)
		if err != nil {
			return err
		}
		b.storeMsgID(ed.New.UID, msg.MessageId)
		log.Println("[update]", ed.New.UID, text)
	}

	for i := range result.Deleted {
		event := &result.Deleted[i]
		// Edit the original message to mark the event as resolved.
		if msgID, ok := b.getMsgID(event.UID); ok {
			text := "~~事件結束~~\n" + event.String()
			if err := b.editMsg(chat, msgID, text); err != nil {
				log.Println("[delete-edit-err]", event.UID, err)
			} else {
				log.Println("[delete]", event.UID)
			}
		}
	}

	return nil
}

type LocalBot struct{}

func NewLocalBot() *LocalBot {
	return &LocalBot{}
}

func (b *LocalBot) Broadcast(chat int64, result DiffResult) error {
	for i := range result.New {
		text := "新事件\n" + result.New[i].String()
		fmt.Printf("[Chat %d] [new] %s %s\n", chat, result.New[i].UID, text)
	}

	for i := range result.Updated {
		ed := &result.Updated[i]
		diff := ed.Old.Diff(&ed.New)
		text := "事件更新\n" + diff + "\n" + ed.New.String()
		fmt.Printf("[Chat %d] [update] %s %s\n", chat, ed.New.UID, text)
	}

	for i := range result.Deleted {
		text := "事件結束\n" + result.Deleted[i].String()
		fmt.Printf("[Chat %d] [delete] %s %s\n", chat, result.Deleted[i].UID, text)
	}

	return nil
}
