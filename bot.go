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
	msgIDs map[string]int64 // UID → latest Telegram message_id (reply chains from the latest)
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

// sendReply sends a message that replies to the message with replyToMsgID.
func (b *TGBot) sendReply(chat int64, text string, replyToMsgID int64) (*gotgbot.Message, error) {
	return b.bot.SendMessage(chat, escapeMarkdown(text), &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: replyToMsgID,
		},
	})
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

		// Reply to the latest message for this event to build a reply chain.
		// If unavailable, send as a standalone message.
		if msgID, ok := b.getMsgID(ed.New.UID); ok {
			reply, err := b.sendReply(chat, text, msgID)
			if err != nil {
				log.Println("[update-reply-err]", ed.New.UID, err)
				// Fallback: send as standalone message
				msg, err2 := b.sendMsg(chat, text)
				if err2 != nil {
					return err2
				}
				b.storeMsgID(ed.New.UID, msg.MessageId)
			} else {
				// Update latest msgID so the next update replies to this one.
				b.storeMsgID(ed.New.UID, reply.MessageId)
			}
		} else {
			msg, err := b.sendMsg(chat, text)
			if err != nil {
				return err
			}
			b.storeMsgID(ed.New.UID, msg.MessageId)
		}
		log.Println("[update]", ed.New.UID, text)
	}

	// Deleted events are intentionally not broadcast — they simply disappear
	// from the website when resolved, which is the expected terminal state.

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

	return nil
}
